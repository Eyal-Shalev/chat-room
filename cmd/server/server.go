package main

import (
	"errors"
	"log"
	"log/slog"
	"net/http"
	"time"

	"chat-room/room"
	"chat-room/www"
	"github.com/go-chi/httplog/v2"
)

func main() {
	server := &www.Server{
		Room: room.New(nil),
		Logger: httplog.NewLogger("chat-room", httplog.Options{
			LogLevel:        slog.LevelDebug,
			Concise:         true,
			RequestHeaders:  true,
			ResponseHeaders: true,
		}),
	}

	log.Printf("Starting server at port 8888")
	defer log.Printf("Goodbye")

	go func() {
		time.Sleep(time.Millisecond)
		_, _ = http.Post("http://127.0.0.1:7331/_templ/reload/events", "", nil)
	}()

	err := http.ListenAndServe("127.0.0.1:8888", server)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("Error running server: %v", err)
	}
}
