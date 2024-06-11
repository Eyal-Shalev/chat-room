package www

import (
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	"chat-room/data"
	"chat-room/room"
	"chat-room/user"
	"github.com/a-h/templ"
	"github.com/go-chi/httplog/v2"
	"golang.org/x/text/language"
)

type Server struct {
	Room   *room.Room
	Logger *httplog.Logger
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	mux := http.NewServeMux()

	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	mux.HandleFunc("POST /enter", s.handleUserFormPost)
	mux.HandleFunc("POST /post", s.handleMessageFormPost)
	mux.HandleFunc("GET /chat-stream", s.handleChatStream)
	mux.Handle("GET /", templ.Handler(Index()))

	var handler http.Handler = mux

	handler = SetPreferredColorSchemaMiddleware(handler)
	handler = SetPageLanguageMiddleware(handler)
	handler = SetUserNameMiddleware(handler)

	handler = httplog.RequestLogger(s.Logger, []string{
		"static/script.js",
		"favicon.ico",
	})(handler)
	handler.ServeHTTP(w, r)
}

func (s *Server) handleUserFormPost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = userForm(&userFormParams{Error: err}).Render(r.Context(), w)
		return
	}

	name := r.PostFormValue("name")
	if len(name) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		_ = userForm(&userFormParams{
			NameError: fmt.Errorf("empty name"),
		}).Render(r.Context(), w)
		return
	}

	SetUserCookie(w, name)
	_ = UserComponent(name).Render(r.Context(), w)
}

func (s *Server) handleMessageFormPost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()

	var messageErr error
	var message string
	if err == nil {
		message = r.PostFormValue("message")
	}
	if len(message) == 0 {
		messageErr = fmt.Errorf("no message provided")
	}

	username, ok := GetUserName(r.Context())
	if !ok {
		err = fmt.Errorf("no username provided")
	}

	if err != nil || messageErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = messageForm(&messageFormParams{
			Error:        err,
			MessageError: messageErr,
			Message:      message,
		}).Render(r.Context(), w)
		return
	}

	err = s.Room.SendContext(r.Context(), data.UserMessage{
		UserName: username,
		Message:  message,
	})
	if err != nil {
		httplog.LogEntry(r.Context()).Error("Error sending message", "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	_ = messageForm(nil).Render(r.Context(), w)
}

func (s *Server) handleChatStream(w http.ResponseWriter, r *http.Request) {
	username, ok := GetUserName(r.Context())
	if !ok {
		http.Error(w, "no username provided", http.StatusUnauthorized)
		return
	}
	u := user.New(username)
	err := s.Room.Join(u)
	if err != nil {
		httplog.LogEntry(r.Context()).Error("Error joining chat", "error", err)
		http.Error(w, "failed to join chat", http.StatusInternalServerError)
		return
	}
	defer s.Room.Leave(u)

	strings.FieldsFunc("", func(r rune) bool {
		return r == '\n' || r == '\r'
	})

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-r.Context().Done():
			_, _ = w.Write([]byte("event: close\n"))
			_, _ = w.Write([]byte("\n\n"))
			w.(http.Flusher).Flush()
			return
		case msgs := <-u.Incoming:
			slices.Reverse(msgs)
			_, _ = w.Write([]byte("event: message\n"))
			_, _ = w.Write([]byte("data: "))
			_ = MessageRows(msgs).Render(r.Context(), w)
			_, _ = w.Write([]byte("\n\n"))
		case <-ticker.C:
			_, _ = w.Write([]byte(": keep alive\n\n"))
			w.(http.Flusher).Flush()
		}
	}
}

func SetUserNameMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userCookie, err := GetUserCookie(r)
		if err != nil && !errors.Is(err, http.ErrNoCookie) {
			httplog.LogEntry(r.Context()).Warn("Invalid user cookie", "error", err)
		} else if err == nil {
			r = r.WithContext(SetUserName(r.Context(), userCookie.Username))
		}
		next.ServeHTTP(w, r)
	})
}

const preferredColorSchemaHeader = "Sec-CH-Prefers-Color-Scheme"

func SetPreferredColorSchemaMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Accept-CH", preferredColorSchemaHeader)
		w.Header().Add("Vary-CH", preferredColorSchemaHeader)
		w.Header().Add("Critical-CH", preferredColorSchemaHeader)
		colorSchema := r.Header.Get(preferredColorSchemaHeader)
		if colorSchema != "" {
			r = r.WithContext(SetPreferredColorSchema(r.Context(), colorSchema))
		}

		next.ServeHTTP(w, r)
	})
}

func SetPageLanguageMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tags, _, err := language.ParseAcceptLanguage(r.Header.Get("Accept-Language"))
		if err == nil && len(tags) > 0 {
			r = r.WithContext(SetPageLanguage(r.Context(), tags[0].String()))
		}
		next.ServeHTTP(w, r)
	})
}
