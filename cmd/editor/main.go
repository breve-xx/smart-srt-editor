package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"dev.codecrunchiness/smart-srt-editor/internal/editor/ctx"
	"dev.codecrunchiness/smart-srt-editor/internal/editor/ui"
	"github.com/google/uuid"
	"github.com/martinlindhe/subtitles"
)

var (
	sessions = make(map[uuid.UUID]*ctx.Session)
	mu       = sync.RWMutex{}
)

type EditedCaption struct {
	Seq   int       `json:"seq"`
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
	Text  []string  `json:"text"`
}

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

	http.HandleFunc("/edit/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Editing file")

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

		// Extract index from URL
		indexStr := r.URL.Path[len("/edit/"):]
		index, err := strconv.Atoi(indexStr)
		if err != nil {
			http.Error(w, "Invalid index", http.StatusBadRequest)
			return
		}

		// Parse request body
		var editedCaption EditedCaption
		if err := json.NewDecoder(r.Body).Decode(&editedCaption); err != nil {
			http.Error(w, "Failed to parse JSON", http.StatusBadRequest)
			return
		}

		// Log index and parsed data
		fmt.Printf("Editing index: %d\n", index)
		fmt.Printf("Parsed Caption: %+v\n", editedCaption)

		caption := session.Subs.Captions[index]
		if caption.Seq != editedCaption.Seq {
			http.Error(w, "Sequence number mismatch", http.StatusBadRequest)
			return
		}

		session.Subs.Captions[index].Text = editedCaption.Text

		// Respond with success
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Edit successful"))
	})

	log.Println("Server started on port 8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
