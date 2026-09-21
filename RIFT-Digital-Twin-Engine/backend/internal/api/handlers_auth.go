package api

import (
	"net/http"
	"time"

	"rift/internal/auth"
)

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token       string `json:"token"`
	UserID      string `json:"userId"`
	DisplayName string `json:"displayName"`
	Role        string `json:"role"`
	ExpiresAt   string `json:"expiresAt"`
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	user, err := s.AuthStore.Authenticate(req.Username, req.Password)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	token, err := s.Tokens.Issue(auth.Claims{UserID: user.ID, OrgID: user.OrgID, Role: string(user.Role)}, 12*time.Hour)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not issue token")
		return
	}
	s.Audit.Record(user.ID, "login", "user:"+user.ID, "", nil)
	writeJSON(w, http.StatusOK, loginResponse{
		Token: token, UserID: user.ID, DisplayName: user.DisplayName, Role: string(user.Role),
		ExpiresAt: time.Now().Add(12 * time.Hour).UTC().Format(time.RFC3339),
	})
}

type registerRequest struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	DisplayName string `json:"displayName"`
	Role        string `json:"role"`
}

// handleRegister lets an authenticated admin create additional users
// (implementing Organization/Workspace/User/Team management). The very
// first user of a fresh install is created by seedDemoData instead, since no
// admin token exists yet to call this endpoint with.
func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	role := req.Role
	if role == "" {
		role = "viewer"
	}
	u, err := s.AuthStore.CreateUser("org_default", req.Username, req.DisplayName, req.Password, roleOrDefault(role))
	if err != nil {
		writeError(w, http.StatusConflict, err.Error())
		return
	}
	s.Audit.Record(userIDFrom(r), "user_created", "user:"+u.ID, "", map[string]interface{}{"role": role})
	writeJSON(w, http.StatusCreated, u)
}
