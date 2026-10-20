# PROSES

## 1. Persona dan masalahnya
Pengguna persona D adalah seorang panitia sebuah acara yang membutuhkan peminjaman aula, proyektor, dan sound system sekaligus. Kebutuhan utamanya adalah peminjaman harus dipaketkan secara atomik (all-or-nothing): apabila salah satu alat atau ruangan tidak tersedia pada slot yang diinginkan, maka seluruh peminjaman harus dibatalkan agar panitia tidak terlanjur menyewa sebagian barang tanpa kelengkapan lainnya.

## 2. Rancangan struktur data
Struktur data yang digunakan:
```go
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
```
**Alasan pemilihan struktur:**
Data disimpan di memori menggunakan slice di dalam struct `Server`. Slice dipilih karena data otomatis terurut berdasarkan ID yang dibuat berurutan (`auto-increment`). Penambahan `sync.Mutex` pada struct `Server` memastikan seluruh operasi baca dan tulis aman dari race condition ketika banyak request datang bersamaan.

## 3. Satu keputusan desain
- **Keputusan yang diambil:** Melakukan pengecekan bentrok ketersediaan dengan menghitung akumulasi unit per slot (slot 1 sampai 8), bukan sekadar memeriksa irisan rentang slot (`slot_mulai` s.d. `slot_selesai`).
- **Pilihan lain yang dipertimbangkan:** Hanya mengecek tumpang tindih waktu rentang slot peminjaman secara matematis.
- **Alasan:** Sumber daya (khususnya alat) memiliki stok lebih dari 1. Jika hanya mengecek irisan rentang slot, peminjaman kedua pada jam yang sama akan langsung ditolak padahal stoknya masih mencukupi. Dengan menghitung akumulasi unit per slot, kapasitas stok terpakai secara akurat dan adil.

## 4. Bug atau kendala (minimal dua)
### Bug 1
- **Gejala:** Validasi panjang karakter string salah menghitung panjang input yang mengandung karakter non-ASCII sehingga tertolak dengan status 400.
- **Cara menemukan:** Mengetes batas panjang input nama menggunakan karakter beraksen/Unicode.
- **Penyebab dan solusi:** Fungsi `len()` di Go menghitung jumlah byte, bukan jumlah karakter (huruf non-ASCII berukuran 2-4 byte). Solusinya mengganti pengecekan menggunakan `utf8.RuneCountInString()`.

### Bug 2
- **Gejala:** Endpoint `GET /sumber-daya` dan `GET /peminjaman` mengembalikan `null` dalam JSON saat data masih kosong, bukan array kosong `[]`.
- **Cara menemukan:** Menguji endpoint GET sesaat setelah server baru dinyalakan melalui curl/Postman.
- **Penyebab dan solusi:** Slice Go yang bernilai `nil` secara bawaan di-encode menjadi `null` oleh `json.Marshal`. Solusinya adalah menginisialisasi slice kosong dengan `make([]..., 0)` atau mengembalikan slice kosong explisit saat nil agar respons JSON selalu berbentuk `[]`.

## 5. Change request
- **Apa yang diminta:** Endpoint `GET /ringkasan?tanggal=YYYY-MM-DD` untuk menyajikan dashboard ringkasan peminjaman harian, total unit slot terpakai, dan rincian per sumber daya (Change Request A).
- **Bagian kode yang diubah:** Menambahkan fungsi `handleRingkasan` dan struct response di `handlers.go`, serta mendaftarkan route di `main.go`.
- **Satu keputusan desain yang diambil:** Menggunakan hash map (`peminjamanPerSD`) untuk menghitung frekuensi peminjaman per sumber daya dalam satu kali iterasi slice peminjaman aktif ($O(N)$), lalu memetakan ke seluruh slice `sumberDaya` agar sumber daya yang tidak memiliki peminjaman tetap muncul dengan nilai 0 secara berurutan sesuai ID.
