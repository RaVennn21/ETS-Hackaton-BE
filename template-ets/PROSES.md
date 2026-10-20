# PROSES

## 1. Persona dan masalahnya
Pengguna persona D adalah seorang panitia sebuah acara yang membutuhan peminjaman aula, proyektor dan sound system sekaligus. Jadi diharuskan peminjaman harus dipaketkan. apabila satu alat tidak tersedia maka peminjaman akan dibatalkan.

## 2. Rancangan struktur data
(Tulis struct dan slice/map apa yang kamu simpan, dan kenapa dipilih begitu. Rancang ini sebelum menulis kode, lalu catat kalau berubah di tengah jalan.)

```go
type SumberDaya struct{
	ID int `json:"id"`
	Nama string `json:"nama"`
	Jenis string `json:"jenis"`
	Stok int `json:"stok"`
}


```

## 3. Satu keputusan desain
(Satu keputusan yang kamu ambil, pilihan lain yang kamu pertimbangkan, dan alasanmu memilih yang ini. Contoh: kenapa cek bentrok per slot, bukan per rentang.)

## 4. Bug atau kendala (minimal dua)
### Bug 1
- Gejala:
- Cara menemukan:
- Penyebab dan solusi:

### Bug 2
- Gejala:
- Cara menemukan:
- Penyebab dan solusi:

## 5. Change request
(Apa yang diminta, bagian kode mana yang kamu ubah, dan satu keputusan desain yang kamu ambil.)
