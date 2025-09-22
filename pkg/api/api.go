package api

import "net/http"

// Init — регистрируем все /api/... маршруты.
func Init(mux *http.ServeMux) {
	// открыт
	mux.HandleFunc("/api/signin",    signinHandler)
	mux.HandleFunc("/api/nextdate",  nextDateHandler)

	// защищено
	mux.HandleFunc("/api/task",      auth(taskHandler))     // POST/GET/PUT/DELETE
	mux.HandleFunc("/api/tasks",     auth(tasksHandler))    // список
	mux.HandleFunc("/api/task/done", auth(taskDoneHandler)) // done
}
