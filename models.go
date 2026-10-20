package main

import "sync"

// SumberDaya merepresentasikan ruangan atau alat yang dapat dipinjam
type SumberDaya struct {
	ID    int    `json:"id"`
	Nama  string `json:"nama"`
	Jenis string `json:"jenis"`
	Stok  int    `json:"stok"`
}

// Peminjaman merepresentasikan satu transaksi peminjaman
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

// Server menyimpan state in-memory database dengan mutex untuk thread-safety
type Server struct {
	mu             sync.Mutex
	sumberDaya     []SumberDaya
	nextID         int
	pinjaman       []Peminjaman
	nextPinjamanID int
	nextPaketID    int
}
