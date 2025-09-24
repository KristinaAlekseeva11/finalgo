package api

import (
	"encoding/json"
	"net/http"
)

type signinReq struct {
	Password string `json:"password"`
}
type signinResp struct {
	Token string `json:"token,omitempty"`
	Error string `json:"error,omitempty"`
}

func signinHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req signinReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "bad json", http.StatusBadRequest)
		return
	}
	if req.Password == "" || req.Password != pass() {
		writeJSON(w, signinResp{Error: "Неверный пароль"}, http.StatusUnauthorized)
		return
	}
	tok, err := makeTokenFromPassword(req.Password)
	if err != nil {
		writeError(w, "token error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, signinResp{Token: tok}, http.StatusOK)
}
