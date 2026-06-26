/* Project Encore: BFG - Localized Private Game Restoration Server
 * Copyright (C) 2026 Paficent <paficent@tutamail.com> & Contributors
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package main

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"paficent/bfg/utils"
)

type user struct {
	BBBID       int    `json:"bbb_id"`
	Username    string `json:"username"`
	PassHash    string `json:"pass_hash"`
	DateCreated int64  `json:"date_created"`
	IP          string `json:"ip"`
}

// just account info not actual save data, should be fine as a json
type userStore struct {
	mu     sync.Mutex
	path   string
	NextID int              `json:"next_id"`
	Users  map[string]*user `json:"users"`
	byID   map[int]*user
}

func loadUserStore(path string) *userStore {
	s := &userStore{path: path, NextID: 1, Users: map[string]*user{}, byID: map[int]*user{}}
	if raw, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(raw, s)
	}
	if s.Users == nil {
		s.Users = map[string]*user{}
	}
	if s.NextID < 1 {
		s.NextID = 1
	}
	for _, u := range s.Users {
		s.byID[u.BBBID] = u
	}
	return s
}

func (s *userStore) saveToDisk() {
	if raw, err := json.MarshalIndent(s, "", "  "); err == nil {
		_ = os.WriteFile(s.path, raw, 0o644)
	}
}

func (s *userStore) byUsername(name string) *user {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.Users[name]
}

func (s *userStore) byBBBID(id int) *user {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.byID[id]
}

func (s *userStore) create(name, passHash, ip string) *user {
	s.mu.Lock()
	defer s.mu.Unlock()
	u := &user{BBBID: s.NextID, Username: name, PassHash: passHash, DateCreated: time.Now().Unix(), IP: ip}
	s.NextID++
	s.Users[name] = u
	s.byID[u.BBBID] = u
	s.saveToDisk()
	return u
}

type authServer struct {
	cfg         config
	store       *userStore
	contentRoot string

	cmdMu    sync.Mutex
	commands []map[string]any
}

func newAuthServer(cfg config, usersPath, contentRoot string) *authServer {
	return &authServer{cfg: cfg, store: loadUserStore(usersPath), contentRoot: contentRoot}
}

func (s *authServer) handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /auth.php", s.handleAuth)
	mux.HandleFunc("POST /verify_user", s.handleVerifyUser)
	mux.HandleFunc("POST /command/gs_display_generic_message", s.handleDisplayMessage)
	mux.HandleFunc("GET /commands", s.handleCommands)
	mux.HandleFunc("GET /content/{ver}/files.json", s.handleFilesJSON)
	mux.HandleFunc("GET /content/{ver}/{file...}", s.handleServeFile)
	mux.HandleFunc("GET /logging_in.mp4", s.handleLoadingMP4)
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "service": "bfg-auth-go"})
	})
	return mux
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func errorResponse(w http.ResponseWriter, message string, status int) {
	writeJSON(w, status, map[string]any{"ok": false, "message": message})
}

func hostOnly(hostport string) string {
	if i := strings.IndexByte(hostport, ':'); i >= 0 {
		return hostport[:i]
	}
	return hostport
}

func (s *authServer) makeAccessToken(username, loginType, clientVersion string) string {
	payload, _ := json.Marshal(map[string]any{
		"username":       username,
		"login_type":     loginType,
		"client_version": clientVersion,
		"issued_at":      time.Now().Unix(),
	})
	return utils.EncryptToken(string(payload), s.cfg.IV, s.cfg.Key)
}

func (s *authServer) handleAuth(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	username := strings.TrimSpace(q.Get("u"))
	password := q.Get("p")
	loginType := q.Get("t")
	clientVersion := q.Get("client_version")
	ip := hostOnly(r.RemoteAddr)

	if username == "" || password == "" {
		errorResponse(w, "Username and password are required.", http.StatusBadRequest)
		return
	}
	if len(username) > 64 || len(password) > 128 {
		errorResponse(w, "Username or password too long.", http.StatusBadRequest)
		return
	}

	u := s.store.byUsername(username)
	if u != nil {
		if !utils.CheckPassword(password, u.PassHash) {
			writeJSON(w, http.StatusUnauthorized, map[string]any{
				"ok": false, "acc_exists": true, "message": "Incorrect password",
			})
			return
		}
	} else {
		u = s.store.create(username, utils.HashPassword(password), ip)
	}

	host := hostOnly(r.Host)
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":         true,
		"acc_exists": true,
		"sessId":     s.makeAccessToken(username, loginType, clientVersion),
		"bbbId":      u.BBBID,
		"username":   username,
		"serverIp":   host,
		"login_type": loginType,
		"contentUrl": "http://" + host + ":900/content/" + clientVersion + "/files.json",
		"friends":    []int{1},
		"auto_login": true,
	})
}

func (s *authServer) handleVerifyUser(w http.ResponseWriter, r *http.Request) {
	var body struct {
		BBBID  *json.Number `json:"bbb_id"`
		GameID json.Number  `json:"game_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.BBBID == nil {
		if body.BBBID == nil {
			errorResponse(w, "Missing required field: bbb_id.", http.StatusBadRequest)
			return
		}
		errorResponse(w, "Invalid or missing JSON body.", http.StatusBadRequest)
		return
	}

	id, err := body.BBBID.Int64()
	if err != nil {
		errorResponse(w, "bbb_id must be an integer.", http.StatusBadRequest)
		return
	}
	if s.store.byBBBID(int(id)) == nil {
		errorResponse(w, "User not found.", http.StatusNotFound)
		return
	}

	payload, _ := json.Marshal(map[string]any{
		"user_id":   id,
		"game_id":   body.GameID.String(),
		"timestamp": time.Now().Unix(),
	})
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":         true,
		"session_id": utils.EncryptToken(string(payload), s.cfg.IV, s.cfg.Key),
	})
}

func (s *authServer) queueDisplayMessage(message string, forceLogout bool, targetBBBID any) map[string]any {
	cmd := map[string]any{
		"command":       "gs_display_generic_message",
		"payload":       map[string]any{"force_logout": forceLogout, "msg": message},
		"target_bbb_id": targetBBBID,
	}
	s.cmdMu.Lock()
	s.commands = append(s.commands, cmd)
	s.cmdMu.Unlock()
	return cmd
}

func (s *authServer) handleDisplayMessage(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Message     string       `json:"message"`
		ForceLogout bool         `json:"force_logout"`
		BBBID       *json.Number `json:"bbb_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		errorResponse(w, "Invalid or missing JSON body.", http.StatusBadRequest)
		return
	}
	message := strings.TrimSpace(body.Message)
	if message == "" {
		errorResponse(w, "Missing required field: message.", http.StatusBadRequest)
		return
	}

	var target any
	if body.BBBID != nil {
		id, err := body.BBBID.Int64()
		if err != nil {
			errorResponse(w, "bbb_id must be an integer.", http.StatusBadRequest)
			return
		}
		target = id
	}

	cmd := s.queueDisplayMessage(message, body.ForceLogout, target)
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "queued": true, "command": cmd})
}

func (s *authServer) handleCommands(w http.ResponseWriter, r *http.Request) {
	s.cmdMu.Lock()
	pending := s.commands
	s.commands = nil
	s.cmdMu.Unlock()
	if pending == nil {
		pending = []map[string]any{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "commands": pending})
}
