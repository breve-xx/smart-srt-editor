package main

import (
	"bytes"
	"context"
	"io"
	"log"
	"net/http"

	"dev.codecrunchiness/smart-srt-editor/internal/ui"
	"github.com/martinlindhe/subtitles"
)

func main() {
	var subs subtitles.Subtitle

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

		subs, err = subtitles.Parse(buf.Bytes())
		if err != nil {
			http.Error(w, "Error parsing file", http.StatusInternalServerError)
			return
		}

		component := ui.Listing(handler, &subs)
		component.Render(context.Background(), w)
	})

	http.HandleFunc("/download", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Downloading file")

		w.Header().Set("Content-Disposition", "attachment; filename=edited.srt")
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Write([]byte(subs.AsSRT()))
	})

	log.Println("Server started on port 8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
