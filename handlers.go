package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

// ==========================================
// 1. HANDLER SUMBER DAYA (FITUR B1)
// ==========================================

// handleCreateSumberDaya: POST /sumber-daya (Status 201)
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

	// Validasi nama (di-trim, tidak boleh kosong, max 50 karakter Unicode)
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

// handleListSumberDaya: GET /sumber-daya (Status 200)
func (s *Server) handleListSumberDaya(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	res := s.sumberDaya
	if res == nil {
		res = []SumberDaya{}
	}
	writeJSON(w, http.StatusOK, res)
}

// handleGetSumberDayaByID: GET /sumber-daya/{id} (Status 200, 400 jika bukan angka, 404 jika tidak ada)
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

// ==========================================
// 2. HANDLER PEMINJAMAN (FITUR B2, B3, B6)
// ==========================================

// handleCreatePeminjaman: POST /peminjaman (Status 201)
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

	// Validasi peminjam
	peminjam := strings.TrimSpace(req.Peminjam)
	if peminjam == "" {
		writeError(w, http.StatusBadRequest, "peminjam wajib diisi")
		return
	}
	if utf8.RuneCountInString(peminjam) > 50 {
		writeError(w, http.StatusBadRequest, "nama peminjam maksimal 50 karakter")
		return
	}

	// Validasi tanggal YYYY-MM-DD
	tanggal := strings.TrimSpace(req.Tanggal)
	if len(tanggal) != 10 {
		writeError(w, http.StatusBadRequest, "format tanggal harus YYYY-MM-DD")
		return
	}
	if _, err := time.Parse("2006-01-02", tanggal); err != nil {
		writeError(w, http.StatusBadRequest, "tanggal tidak valid")
		return
	}

	// Validasi rentang slot (1-8)
	if req.SlotMulai < 1 || req.SlotMulai > 8 || req.SlotSelesai < 1 || req.SlotSelesai > 8 || req.SlotMulai > req.SlotSelesai {
		writeError(w, http.StatusBadRequest, "slot harus antara 1-8 dan slot_mulai <= slot_selesai")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Validasi sumber daya ada
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

	// Validasi jumlah unit
	if req.Jumlah < 1 || req.Jumlah > targetSD.Stok {
		writeError(w, http.StatusBadRequest, "jumlah minimal 1 dan tidak boleh melebihi stok")
		return
	}

	// Cek bentrok kapasitas slot (409)
	if s.cekBentrok(req.SumberDayaID, tanggal, req.SlotMulai, req.SlotSelesai, req.Jumlah, targetSD.Stok, 0) {
		writeError(w, http.StatusConflict, "kapasitas slot pada tanggal tersebut telah terlampaui")
		return
	}

	p := Peminjaman{
		ID:           s.nextPinjamanID,
		SumberDayaID: req.SumberDayaID,
		Peminjam:     peminjam,
		Tanggal:      tanggal,
		SlotMulai:    req.SlotMulai,
		SlotSelesai:  req.SlotSelesai,
		Jumlah:       req.Jumlah,
		Status:       "aktif",
		PaketID:      0, // Persona D: 0 untuk peminjaman biasa
	}
	s.nextPinjamanID++
	s.pinjaman = append(s.pinjaman, p)

	writeJSON(w, http.StatusCreated, p)
}

// handleListPeminjaman: GET /peminjaman (Status 200, dengan filter opsional)
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

// handleGetPeminjamanByID: GET /peminjaman/{id} (Status 200, 400 jika non-angka, 404 jika tidak ada)
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

// handleUpdatePeminjaman: PUT /peminjaman/{id} (Status 200)
func (s *Server) handleUpdatePeminjaman(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "id harus berupa angka")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	var loanIndex = -1
	for i := range s.pinjaman {
		if s.pinjaman[i].ID == id {
			loanIndex = i
			break
		}
	}
	if loanIndex == -1 {
		writeError(w, http.StatusNotFound, "peminjaman tidak ditemukan")
		return
	}

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

	peminjam := strings.TrimSpace(req.Peminjam)
	if peminjam == "" {
		writeError(w, http.StatusBadRequest, "peminjam wajib diisi")
		return
	}
	if utf8.RuneCountInString(peminjam) > 50 {
		writeError(w, http.StatusBadRequest, "nama peminjam maksimal 50 karakter")
		return
	}

	tanggal := strings.TrimSpace(req.Tanggal)
	if len(tanggal) != 10 {
		writeError(w, http.StatusBadRequest, "format tanggal harus YYYY-MM-DD")
		return
	}
	if _, err := time.Parse("2006-01-02", tanggal); err != nil {
		writeError(w, http.StatusBadRequest, "tanggal tidak valid")
		return
	}

	if req.SlotMulai < 1 || req.SlotMulai > 8 || req.SlotSelesai < 1 || req.SlotSelesai > 8 || req.SlotMulai > req.SlotSelesai {
		writeError(w, http.StatusBadRequest, "slot harus antara 1-8 dan slot_mulai <= slot_selesai")
		return
	}

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

	if req.Jumlah < 1 || req.Jumlah > targetSD.Stok {
		writeError(w, http.StatusBadRequest, "jumlah minimal 1 dan tidak boleh melebihi stok")
		return
	}

	// Cek bentrok (exclude ID peminjaman yang sedang diubah)
	if s.cekBentrok(req.SumberDayaID, tanggal, req.SlotMulai, req.SlotSelesai, req.Jumlah, targetSD.Stok, id) {
		writeError(w, http.StatusConflict, "kapasitas slot pada tanggal tersebut telah terlampaui")
		return
	}

	s.pinjaman[loanIndex].SumberDayaID = req.SumberDayaID
	s.pinjaman[loanIndex].Peminjam = peminjam
	s.pinjaman[loanIndex].Tanggal = tanggal
	s.pinjaman[loanIndex].SlotMulai = req.SlotMulai
	s.pinjaman[loanIndex].SlotSelesai = req.SlotSelesai
	s.pinjaman[loanIndex].Jumlah = req.Jumlah

	writeJSON(w, http.StatusOK, s.pinjaman[loanIndex])
}

// handleDeletePeminjaman: DELETE /peminjaman/{id} (Status 204 tanpa body)
func (s *Server) handleDeletePeminjaman(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "id harus berupa angka")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	var loanIndex = -1
	for i := range s.pinjaman {
		if s.pinjaman[i].ID == id {
			loanIndex = i
			break
		}
	}
	if loanIndex == -1 {
		writeError(w, http.StatusNotFound, "peminjaman tidak ditemukan")
		return
	}

	s.pinjaman = append(s.pinjaman[:loanIndex], s.pinjaman[loanIndex+1:]...)
	w.WriteHeader(http.StatusNoContent)
}

// ==========================================
// 3. FITUR KHAS PERSONA D (PANITIA ACARA)
// ==========================================

type PaketItemReq struct {
	SumberDayaID int `json:"sumber_daya_id"`
	Jumlah       int `json:"jumlah"`
}

type PaketPeminjamanReq struct {
	Peminjam    string         `json:"peminjam"`
	Tanggal     string         `json:"tanggal"`
	SlotMulai   int            `json:"slot_mulai"`
	SlotSelesai int            `json:"slot_selesai"`
	Item        []PaketItemReq `json:"item"`
}

type PaketResponse struct {
	PaketID    int          `json:"paket_id"`
	Peminjaman []Peminjaman `json:"peminjaman"`
}

// handleCreatePaketPeminjaman: POST /paket-peminjaman (Status 201 Created)
func (s *Server) handleCreatePaketPeminjaman(w http.ResponseWriter, r *http.Request) {
	var req PaketPeminjamanReq
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

	// 2. Validasi tanggal
	tanggal := strings.TrimSpace(req.Tanggal)
	if len(tanggal) != 10 {
		writeError(w, http.StatusBadRequest, "format tanggal harus YYYY-MM-DD")
		return
	}
	if _, err := time.Parse("2006-01-02", tanggal); err != nil {
		writeError(w, http.StatusBadRequest, "tanggal tidak valid")
		return
	}

	// 3. Validasi slot
	if req.SlotMulai < 1 || req.SlotMulai > 8 || req.SlotSelesai < 1 || req.SlotSelesai > 8 || req.SlotMulai > req.SlotSelesai {
		writeError(w, http.StatusBadRequest, "slot harus antara 1-8 dan slot_mulai <= slot_selesai")
		return
	}

	// 4. Validasi item (1 sampai 10 elemen)
	if len(req.Item) < 1 || len(req.Item) > 10 {
		writeError(w, http.StatusBadRequest, "item paket harus berisi 1 sampai 10 elemen")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// 5. Validasi duplikat sumber daya, keberadaan sumber daya, dan stok
	seenSD := make(map[int]bool)
	type verifiedItem struct {
		sd     *SumberDaya
		jumlah int
	}
	var itemsToBook []verifiedItem

	for _, it := range req.Item {
		if seenSD[it.SumberDayaID] {
			writeError(w, http.StatusBadRequest, "satu sumber daya tidak boleh muncul dua kali dalam satu paket")
			return
		}
		seenSD[it.SumberDayaID] = true

		var targetSD *SumberDaya
		for i := range s.sumberDaya {
			if s.sumberDaya[i].ID == it.SumberDayaID {
				targetSD = &s.sumberDaya[i]
				break
			}
		}
		if targetSD == nil {
			writeError(w, http.StatusBadRequest, "sumber daya tidak ditemukan")
			return
		}

		if it.Jumlah < 1 || it.Jumlah > targetSD.Stok {
			writeError(w, http.StatusBadRequest, "jumlah minimal 1 dan tidak boleh melebihi stok")
			return
		}

		itemsToBook = append(itemsToBook, verifiedItem{sd: targetSD, jumlah: it.Jumlah})
	}

	// 6. Cek bentrok atomik: Bila salah satu item bentrok, hasilnya 409 dan tidak ada peminjaman dibuat
	for _, it := range itemsToBook {
		if s.cekBentrok(it.sd.ID, tanggal, req.SlotMulai, req.SlotSelesai, it.jumlah, it.sd.Stok, 0) {
			writeError(w, http.StatusConflict, "kapasitas slot pada salah satu item telah terlampaui")
			return
		}
	}

	// Simpan paket peminjaman
	paketID := s.nextPaketID
	s.nextPaketID++

	createdList := make([]Peminjaman, 0, len(itemsToBook))
	for _, it := range itemsToBook {
		p := Peminjaman{
			ID:           s.nextPinjamanID,
			SumberDayaID: it.sd.ID,
			Peminjam:     peminjam,
			Tanggal:      tanggal,
			SlotMulai:    req.SlotMulai,
			SlotSelesai:  req.SlotSelesai,
			Jumlah:       it.jumlah,
			Status:       "aktif",
			PaketID:      paketID,
		}
		s.nextPinjamanID++
		s.pinjaman = append(s.pinjaman, p)
		createdList = append(createdList, p)
	}

	writeJSON(w, http.StatusCreated, PaketResponse{
		PaketID:    paketID,
		Peminjaman: createdList,
	})
}

// handleDeletePaketPeminjaman: DELETE /paket-peminjaman/{id} (Status 204 tanpa body)
func (s *Server) handleDeletePaketPeminjaman(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "id harus berupa angka")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	found := false
	newPinjaman := make([]Peminjaman, 0, len(s.pinjaman))
	for _, p := range s.pinjaman {
		if p.PaketID == id {
			found = true
		} else {
			newPinjaman = append(newPinjaman, p)
		}
	}

	if !found {
		writeError(w, http.StatusNotFound, "paket peminjaman tidak ditemukan")
		return
	}

	s.pinjaman = newPinjaman
	w.WriteHeader(http.StatusNoContent)
}

// ==========================================
// 4. CHANGE REQUEST A: DASHBOARD RINGKASAN
// ==========================================

type PerSumberDayaSummary struct {
	SumberDayaID     int    `json:"sumber_daya_id"`
	Nama             string `json:"nama"`
	JumlahPeminjaman int    `json:"jumlah_peminjaman"`
}

type RingkasanResponse struct {
	Tanggal          string                 `json:"tanggal"`
	JumlahPeminjaman int                    `json:"jumlah_peminjaman"`
	TotalUnitSlot    int                    `json:"total_unit_slot"`
	PerSumberDaya    []PerSumberDayaSummary `json:"per_sumber_daya"`
}

// handleRingkasan: GET /ringkasan?tanggal=YYYY-MM-DD
func (s *Server) handleRingkasan(w http.ResponseWriter, r *http.Request) {
	tanggal := strings.TrimSpace(r.URL.Query().Get("tanggal"))
	if len(tanggal) != 10 {
		writeError(w, http.StatusBadRequest, "format tanggal harus YYYY-MM-DD")
		return
	}
	if _, err := time.Parse("2006-01-02", tanggal); err != nil {
		writeError(w, http.StatusBadRequest, "tanggal tidak valid")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	jumlahPeminjaman := 0
	totalUnitSlot := 0

	// Hitung peminjaman per sumber daya pada tanggal tersebut
	peminjamanPerSD := make(map[int]int)
	for _, p := range s.pinjaman {
		if p.Tanggal == tanggal && p.Status == "aktif" {
			jumlahPeminjaman++
			slots := p.SlotSelesai - p.SlotMulai + 1
			totalUnitSlot += p.Jumlah * slots
			peminjamanPerSD[p.SumberDayaID]++
		}
	}

	// per_sumber_daya memuat SEMUA sumber daya terurut menurut id
	perSD := make([]PerSumberDayaSummary, 0, len(s.sumberDaya))
	for _, sd := range s.sumberDaya {
		perSD = append(perSD, PerSumberDayaSummary{
			SumberDayaID:     sd.ID,
			Nama:             sd.Nama,
			JumlahPeminjaman: peminjamanPerSD[sd.ID],
		})
	}

	writeJSON(w, http.StatusOK, RingkasanResponse{
		Tanggal:          tanggal,
		JumlahPeminjaman: jumlahPeminjaman,
		TotalUnitSlot:    totalUnitSlot,
		PerSumberDaya:    perSD,
	})
}

