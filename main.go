package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

type SumberDaya struct {
	ID    int    `json:"id"`
	Nama  string `json:"nama"`
	Jenis string `json:"jenis"`
	Stok  int    `json:"stok"`
}

type Peminjaman struct {
	ID           int    `json:"id"`
	SumberDayaID int    `json:"sumber_daya_id"`
	Peminjam     string `json:"peminjam"`
	Tanggal      string `json:"tanggal"`
	SlotMulai    int    `json:"slot_mulai"`
	SlotSelesai  int    `json:"slot_selesai"`
	Jumlah       int    `json:"jumlah"`
	Status       string `json:"status"`   // "aktif"
	PaketID      int    `json:"paket_id"` // bernilai 0 untuk peminjaman biasa
}

type Server struct {
	mu             sync.Mutex
	sumberDaya     []SumberDaya
	nextID         int
	pinjaman       []Peminjaman
	nextPinjamanID int
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

// Handler POST /peminjaman (Status 201 Created) - Fitur B2
func (s *Server) handleCreatePeminjaman(w http.ResponseWriter, r *http.Request) {
	var req struct {
		SumberDayaID int    `json:"sumber_daya_id"`
		Peminjam     string `json:"peminjam"`
		Tanggal      string `json:"tanggal"`
		SlotMulai    int    `json:"slot_mulai"`
		SlotSelesai  int    `json:"slot_selesai"`
		Jumlah       int    `json:"jumlah"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "format body JSON tidak valid")
		return
	}

	// 1. Validasi peminjam
	peminjam := strings.TrimSpace(req.Peminjam)
	if peminjam == "" {
		writeError(w, http.StatusBadRequest, "peminjam wajib diisi")
		return
	}
	if utf8.RuneCountInString(peminjam) > 50 {
		writeError(w, http.StatusBadRequest, "nama peminjam maksimal 50 karakter")
		return
	}

	// 2. Validasi tanggal format YYYY-MM-DD
	tanggal := strings.TrimSpace(req.Tanggal)
	if len(tanggal) != 10 {
		writeError(w, http.StatusBadRequest, "format tanggal harus YYYY-MM-DD")
		return
	}
	if _, err := time.Parse("2006-01-02", tanggal); err != nil {
		writeError(w, http.StatusBadRequest, "tanggal tidak valid")
		return
	}

	// 3. Validasi slot (1-8 dan slot_mulai <= slot_selesai)
	if req.SlotMulai < 1 || req.SlotMulai > 8 || req.SlotSelesai < 1 || req.SlotSelesai > 8 || req.SlotMulai > req.SlotSelesai {
		writeError(w, http.StatusBadRequest, "slot harus antara 1-8 dan slot_mulai <= slot_selesai")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// 4. Validasi sumber daya ada (jika tidak ada -> 400)
	var targetSD *SumberDaya
	for i := range s.sumberDaya {
		if s.sumberDaya[i].ID == req.SumberDayaID {
			targetSD = &s.sumberDaya[i]
			break
		}
	}
	if targetSD == nil {
		writeError(w, http.StatusBadRequest, "sumber daya tidak ditemukan")
		return
	}

	// 5. Validasi jumlah (antara 1 dan stok)
	if req.Jumlah < 1 || req.Jumlah > targetSD.Stok {
		writeError(w, http.StatusBadRequest, "jumlah minimal 1 dan tidak boleh melebihi stok")
		return
	}

	// 6. Deteksi bentrok kapasitas slot (409 Conflict)
	for slot := req.SlotMulai; slot <= req.SlotSelesai; slot++ {
		terpakai := 0
		for _, p := range s.pinjaman {
			if p.SumberDayaID == req.SumberDayaID && p.Tanggal == tanggal && p.Status == "aktif" {
				if slot >= p.SlotMulai && slot <= p.SlotSelesai {
					terpakai += p.Jumlah
				}
			}
		}
		if terpakai+req.Jumlah > targetSD.Stok {
			writeError(w, http.StatusConflict, "kapasitas slot pada tanggal tersebut telah terlampaui")
			return
		}
	}

	// Buat peminjaman baru
	p := Peminjaman{
		ID:           s.nextPinjamanID,
		SumberDayaID: req.SumberDayaID,
		Peminjam:     peminjam,
		Tanggal:      tanggal,
		SlotMulai:    req.SlotMulai,
		SlotSelesai:  req.SlotSelesai,
		Jumlah:       req.Jumlah,
		Status:       "aktif",
		PaketID:      0, // Persona D: bernilai 0 untuk peminjaman biasa
	}
	s.nextPinjamanID++
	s.pinjaman = append(s.pinjaman, p)

	writeJSON(w, http.StatusCreated, p)
}

// Handler GET /peminjaman (Status 200) - Fitur B3: Daftar dengan filter opsional
func (s *Server) handleListPeminjaman(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	sdIDStr := q.Get("sumber_daya_id")
	var filterSdID int
	if sdIDStr != "" {
		var err error
		filterSdID, err = strconv.Atoi(sdIDStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "sumber_daya_id harus berupa angka")
			return
		}
	}

	filterTanggal := q.Get("tanggal")
	filterStatus := q.Get("status")

	s.mu.Lock()
	defer s.mu.Unlock()

	res := make([]Peminjaman, 0)
	for _, p := range s.pinjaman {
		if sdIDStr != "" && p.SumberDayaID != filterSdID {
			continue
		}
		if filterTanggal != "" && p.Tanggal != filterTanggal {
			continue
		}
		if filterStatus != "" && p.Status != filterStatus {
			continue
		}
		res = append(res, p)
	}

	writeJSON(w, http.StatusOK, res)
}

// Handler GET /peminjaman/{id} (Status 200, 400 jika bukan angka, 404 jika tidak ada) - Fitur B3
func (s *Server) handleGetPeminjamanByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "id harus berupa angka")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, p := range s.pinjaman {
		if p.ID == id {
			writeJSON(w, http.StatusOK, p)
			return
		}
	}

	writeError(w, http.StatusNotFound, "peminjaman tidak ditemukan")
}

// NewHandler membuat instance Server baru dan mendaftarkan route (wajib ada sesuai syarat ETS)
func NewHandler() http.Handler {
	s := &Server{
		sumberDaya:     make([]SumberDaya, 0),
		nextID:         1,
		pinjaman:       make([]Peminjaman, 0),
		nextPinjamanID: 1,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /sumber-daya", s.handleCreateSumberDaya)
	mux.HandleFunc("GET /sumber-daya", s.handleListSumberDaya)
	mux.HandleFunc("GET /sumber-daya/{id}", s.handleGetSumberDayaByID)
	mux.HandleFunc("POST /peminjaman", s.handleCreatePeminjaman)
	mux.HandleFunc("GET /peminjaman", s.handleListPeminjaman)
	mux.HandleFunc("GET /peminjaman/{id}", s.handleGetPeminjamanByID)

	return mux
}

func main() {
	http.ListenAndServe(":8080", NewHandler())
}
