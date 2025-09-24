package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/KristinaAlekseeva11/finalgo/pkg/db"
)

// единая точка на /api/task — рулим по методу
func taskHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		addTaskHandler(w, r)
	case http.MethodGet:
		getTaskHandler(w, r)
	case http.MethodPut:
		editTaskHandler(w, r)
	case http.MethodDelete:
		deleteTaskHandler(w, r)
	default:
		writeError(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

/************* GET /api/task?id=... *************/
func getTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.URL.Query().Get("id"))
	if id == "" {
		writeError(w, "id required", http.StatusBadRequest)
		return
	}
	t, err := db.GetTask(id)
	if err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, t, http.StatusOK)
}

/************* PUT /api/task (редактирование) *************/
func editTaskHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, "read body error", http.StatusBadRequest)
		return
	}
	var t db.Task
	if err := json.Unmarshal(body, &t); err != nil {
		writeError(w, "bad json", http.StatusBadRequest)
		return
	}
	t.ID = strings.TrimSpace(t.ID)
	t.Title = strings.TrimSpace(t.Title)
	t.Repeat = strings.TrimSpace(t.Repeat)

	if t.ID == "" {
		writeError(w, "id required", http.StatusBadRequest)
		return
	}
	if t.Title == "" {
		writeError(w, "empty title", http.StatusBadRequest)
		return
	}
	if err := normalizeAndCheckDate(&t); err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := db.UpdateTask(&t); err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{}`))
}

/************* POST /api/task (добавление) *************/
func addTaskHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, "read body error", http.StatusBadRequest)
		return
	}
	var t db.Task
	if err := json.Unmarshal(body, &t); err != nil {
		writeError(w, "bad json", http.StatusBadRequest)
		return
	}
	t.Title = strings.TrimSpace(t.Title)
	if t.Title == "" {
		writeError(w, "empty title", http.StatusBadRequest)
		return
	}
	if err := normalizeAndCheckDate(&t); err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}
	id, err := db.AddTask(&t)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]any{"id": id}, http.StatusOK)
}

/************* общая логика нормализации даты *************/
func normalizeAndCheckDate(t *db.Task) error {
	const dateFmt = "20060102"
	now := dateOnly(time.Now())
	t.Date = strings.TrimSpace(t.Date)

	if t.Date == "" {
		t.Date = now.Format(dateFmt)
	}

	d, err := time.Parse(dateFmt, t.Date)
	if err != nil {
		return errors.New("bad date format (need 20060102)")
	}
	d = dateOnly(d)

	rep := strings.TrimSpace(t.Repeat)
	var next string
	if rep != "" {
		next, err = NextDate(now, t.Date, rep)
		if err != nil {
			return err
		}
	}

	if !d.After(now) && !d.Equal(now) {
		if rep == "" {
			t.Date = now.Format(dateFmt)
		} else {
			t.Date = next
		}
	}
	return nil
}
