package server

import (
	"log"
	"net/http"
	"os"
	"strconv"
)

// ResolvePort возвращает порт из TODO_PORT или 7540 по умолчанию.
func ResolvePort() int {
	if v := os.Getenv("TODO_PORT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
		log.Printf("WARN: invalid TODO_PORT=%q, fallback to 7540", v)
	}
	return 7540
}

// NewMux настраивает маршруты для раздачи статики из ./web.
func NewMux() http.Handler {
	// Важно: отдаём файлы «как есть», без префиксов — этого ждут тесты.
	fs := http.FileServer(http.Dir("web"))
	mux := http.NewServeMux()
	mux.Handle("/", fs)
	return mux
}

// Run запускает HTTP-сервер.
func Run() error {
	addr := ":" + strconv.Itoa(ResolvePort())
	srv := &http.Server{
		Addr:    addr,
		Handler: NewMux(),
	}
	log.Printf("Server listening on http://localhost%s", addr)
	return srv.ListenAndServe()
}
