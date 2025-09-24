package api

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/KristinaAlekseeva11/finalgo/pkg/db"
)

type TasksResp struct {
	Tasks []*db.Task `json:"tasks"`
}

func tasksHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	q := r.URL.Query()
	search := strings.TrimSpace(q.Get("search"))

	// опционально можно передать ?limit=...
	limit := 50
	if s := q.Get("limit"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 && n <= 100 {
			limit = n
		}
	}

	tasks, err := db.Tasks(limit, search)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if tasks == nil {
		tasks = make([]*db.Task, 0)
	}

	writeJSON(w, TasksResp{Tasks: tasks}, http.StatusOK)
}
