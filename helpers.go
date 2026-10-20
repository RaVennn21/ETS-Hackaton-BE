package main

import (
	"encoding/json"
	"net/http"
)

// writeJSON mengirim response berformat JSON dengan header Content-Type application/json
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// writeError mengirim format error standar {"error": "pesan"}
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// cekBentrok memeriksa kapasitas stok per slot (1-8) pada tanggal tertentu (Fitur B5).
// excludeID dipakai untuk PUT agar tidak bentrok dengan data dirinya sendiri (isi 0 jika peminjaman baru).
func (s *Server) cekBentrok(sumberDayaID int, tanggal string, slotMulai, slotSelesai, jumlah, stok, excludeID int) bool {
	for slot := slotMulai; slot <= slotSelesai; slot++ {
		terpakai := 0
		for _, p := range s.pinjaman {
			if p.ID == excludeID {
				continue
			}
			if p.SumberDayaID == sumberDayaID && p.Tanggal == tanggal && p.Status == "aktif" {
				if slot >= p.SlotMulai && slot <= p.SlotSelesai {
					terpakai += p.Jumlah
				}
			}
		}
		if terpakai+jumlah > stok {
			return true
		}
	}
	return false
}
