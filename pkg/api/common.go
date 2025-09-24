package api

import (
	"encoding/json"
	"net/http"
)

type errResp struct {
	Error string `json:"error"`
}

func writeJSON(w http.ResponseWriter, data any, status int) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, msg string, status int) {
	writeJSON(w, errResp{Error: msg}, status)
}
