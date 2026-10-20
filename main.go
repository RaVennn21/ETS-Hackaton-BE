package main

import "net/http"

// NewHandler membuat instance Server baru dan mendaftarkan seluruh route API
func NewHandler() http.Handler {
	s := &Server{
		sumberDaya:     make([]SumberDaya, 0),
		nextID:         1,
		pinjaman:       make([]Peminjaman, 0),
		nextPinjamanID: 1,
		nextPaketID:    1,
	}

	mux := http.NewServeMux()

	// Route Sumber Daya (B1)
	mux.HandleFunc("POST /sumber-daya", s.handleCreateSumberDaya)
	mux.HandleFunc("GET /sumber-daya", s.handleListSumberDaya)
	mux.HandleFunc("GET /sumber-daya/{id}", s.handleGetSumberDayaByID)

	// Route Peminjaman (B2, B3, B6)
	mux.HandleFunc("POST /peminjaman", s.handleCreatePeminjaman)
	mux.HandleFunc("GET /peminjaman", s.handleListPeminjaman)
	mux.HandleFunc("GET /peminjaman/{id}", s.handleGetPeminjamanByID)
	mux.HandleFunc("PUT /peminjaman/{id}", s.handleUpdatePeminjaman)
	mux.HandleFunc("DELETE /peminjaman/{id}", s.handleDeletePeminjaman)

	// Route Fitur Khas Persona D (Paket Peminjaman)
	mux.HandleFunc("POST /paket-peminjaman", s.handleCreatePaketPeminjaman)
	mux.HandleFunc("DELETE /paket-peminjaman/{id}", s.handleDeletePaketPeminjaman)

	return mux
}

func main() {
	http.ListenAndServe(":8080", NewHandler())
}
