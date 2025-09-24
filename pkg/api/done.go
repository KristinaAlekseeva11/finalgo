package api

import (
	"net/http"
	"strings"
	"time"

	"github.com/KristinaAlekseeva11/finalgo/pkg/db"
)

// POST /api/task/done?id=...
func taskDoneHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	id := strings.TrimSpace(r.URL.Query().Get("id"))
	if id == "" {
		writeError(w, "id required", http.StatusBadRequest)
		return
	}

	// 1) достаём задачу
	t, err := db.GetTask(id)
	if err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 2) если одноразовая — удаляем
	if strings.TrimSpace(t.Repeat) == "" {
		if err := db.DeleteTask(id); err != nil {
			writeError(w, err.Error(), http.StatusBadRequest)
			return
		}
		writeJSON(w, struct{}{}, http.StatusOK)
		return
	}

	// 3) повторяемая: считаем следующую дату ОТ ДАТЫ ЗАДАЧИ, а не от "сегодня"
	// это ключевой момент для TestDone
	taskDate, err := time.Parse(dateFmt, t.Date) // dateFmt = "20060102" (из api/nextdate.go)
	if err != nil {
		writeError(w, "bad task date", http.StatusBadRequest)
		return
	}
	taskDate = dateOnly(taskDate)

	next, err := NextDate(taskDate, t.Date, t.Repeat)
	if err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := db.UpdateDate(next, id); err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, struct{}{}, http.StatusOK)
}

// DELETE /api/task?id=...
func deleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.URL.Query().Get("id"))
	if id == "" {
		writeError(w, "id required", http.StatusBadRequest)
		return
	}
	if err := db.DeleteTask(id); err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, struct{}{}, http.StatusOK)
}
