package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"unicode/utf8"
)

type SumberDaya struct {
	ID    int    `json:"id"`
	Nama  string `json:"nama"`
	Jenis string `json:"jenis"`
	Stok  int    `json:"stok"`
}

type Server struct {
	mu         sync.Mutex
	sumberDaya []SumberDaya
	nextID     int
}

// writeJSON mengirim response berformat JSON dengan header Content-Type
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// writeError mengirim format error standar {"error": "pesan"}
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// Handler POST /sumber-daya (Status 201)
func (s *Server) handleCreateSumberDaya(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Nama  string `json:"nama"`
		Jenis string `json:"jenis"`
		Stok  int    `json:"stok"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "format body JSON tidak valid")
		return
	}

	// Validasi nama (di-trim, tidak boleh kosong, max 50 karakter)
	nama := strings.TrimSpace(req.Nama)
	if nama == "" {
		writeError(w, http.StatusBadRequest, "nama wajib diisi")
		return
	}
	if utf8.RuneCountInString(nama) > 50 {
		writeError(w, http.StatusBadRequest, "nama maksimal 50 karakter")
		return
	}

	// Validasi jenis (hanya "ruang" atau "alat")
	jenis := strings.TrimSpace(req.Jenis)
	if jenis != "ruang" && jenis != "alat" {
		writeError(w, http.StatusBadRequest, "jenis harus 'ruang' atau 'alat'")
		return
	}

	// Validasi stok (minimal 1)
	if req.Stok < 1 {
		writeError(w, http.StatusBadRequest, "stok minimal harus 1")
		return
	}

	s.mu.Lock()
	sd := SumberDaya{
		ID:    s.nextID,
		Nama:  nama,
		Jenis: jenis,
		Stok:  req.Stok,
	}
	s.nextID++
	s.sumberDaya = append(s.sumberDaya, sd)
	s.mu.Unlock()

	writeJSON(w, http.StatusCreated, sd)
}

// Handler GET /sumber-daya (Status 200)
func (s *Server) handleListSumberDaya(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	res := s.sumberDaya
	if res == nil {
		res = []SumberDaya{}
	}
	writeJSON(w, http.StatusOK, res)
}

// Handler GET /sumber-daya/{id} (Status 200, 400 jika bukan angka, 404 jika tidak ada)
func (s *Server) handleGetSumberDayaByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "id harus berupa angka")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, sd := range s.sumberDaya {
		if sd.ID == id {
			writeJSON(w, http.StatusOK, sd)
			return
		}
	}

	writeError(w, http.StatusNotFound, "sumber daya tidak ditemukan")
}

// NewHandler membuat instance Server baru dan mendaftarkan route (wajib ada sesuai syarat ETS)
func NewHandler() http.Handler {
	s := &Server{
		sumberDaya: make([]SumberDaya, 0),
		nextID:     1,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /sumber-daya", s.handleCreateSumberDaya)
	mux.HandleFunc("GET /sumber-daya", s.handleListSumberDaya)
	mux.HandleFunc("GET /sumber-daya/{id}", s.handleGetSumberDayaByID)

	return mux
}

func main() {
	http.ListenAndServe(":8080", NewHandler())
}
