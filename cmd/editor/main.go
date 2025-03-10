package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"

	"dev.codecrunchiness/smart-srt-editor/internal/editor/ctx"
	"dev.codecrunchiness/smart-srt-editor/internal/editor/ui"
	"github.com/google/uuid"
	"github.com/martinlindhe/subtitles"
)

var (
	sessions = make(map[uuid.UUID]*ctx.Session)
	mu       = sync.RWMutex{}
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		component := ui.UploadPage()
		component.Render(context.Background(), w)
	})

	http.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Uploading file")

		file, handler, err := r.FormFile("srtFile")
		if err != nil {
			http.Error(w, "Error retrieving file", http.StatusInternalServerError)
			return
		}
		defer file.Close()

		buf := new(bytes.Buffer)
		_, err = io.Copy(buf, file)
		if err != nil {
			http.Error(w, "Error reading file", http.StatusInternalServerError)
			return
		}

		subs, err := subtitles.Parse(buf.Bytes())
		if err != nil {
			http.Error(w, "Error parsing file", http.StatusInternalServerError)
			return
		}

		sessionID, err := uuid.NewRandom()
		if err != nil {
			http.Error(w, "Error creating session", http.StatusInternalServerError)
			return
		}
		mu.Lock()
		sessions[sessionID] = ctx.NewSession(sessionID, handler, &subs)
		mu.Unlock()

		w.Header().Set("Authorization", fmt.Sprintf("Bearer %s", sessionID.String()))

		component := ui.Listing(sessions[sessionID])
		component.Render(context.Background(), w)
	})

	http.HandleFunc("/download", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Downloading file")

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || len(authHeader) <= 7 || authHeader[:7] != "Bearer " {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		sessionID, err := uuid.Parse(authHeader[7:])
		if err != nil {
			http.Error(w, "Invalid session ID", http.StatusUnauthorized)
			return
		}

		mu.Lock()
		session, exists := sessions[sessionID]
		mu.Unlock()

		if !exists {
			http.Error(w, "Session not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Disposition", "attachment; filename=edited.srt")
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Write([]byte(session.Subs.AsSRT()))
	})

	log.Println("Server started on port 8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
