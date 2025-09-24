package server

import (
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/KristinaAlekseeva11/finalgo/pkg/api"
	"github.com/KristinaAlekseeva11/finalgo/pkg/db"
)

// порт либо из TODO_PORT, либо 7540
func ResolvePort() int {
	if v := os.Getenv("TODO_PORT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
		log.Printf("WARN: неправильный порт %q, беру 7540", v)
	}
	return 7540
}

// собираем маршруты
func NewMux() http.Handler {
	mux := http.NewServeMux()

	// 1) API (важно: регистрируем до статики)
	api.Init(mux)

	// 2) Статика из папки web
	fs := http.FileServer(http.Dir("web"))
	mux.Handle("/", fs)

	return mux
}

func Run() error {
	// инициализируем базу (путь можно задать TODO_DBFILE)
	dbPath := os.Getenv("TODO_DBFILE")
	if err := db.Init(dbPath); err != nil {
		return err
	}
	defer db.Close()

	addr := ":" + strconv.Itoa(ResolvePort())
	srv := &http.Server{Addr: addr, Handler: NewMux()}

	log.Printf("Server запускается по адресу http://localhost%s", addr)
	return srv.ListenAndServe()
}
