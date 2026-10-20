# AI-LOG

Catat setiap penggunaan AI yang menghasilkan kode. Jika tidak memakai AI sama sekali, tulis satu baris: **tidak memakai AI**.

Log yang jujur tidak mengurangi nilai, sebanyak apa pun AI dipakai. Yang mengurangi nilai adalah log yang tidak sesuai dengan kode, atau isi log yang tidak bisa dijelaskan.

## (Nama fitur atau fungsi, misalnya: Deteksi bentrok slot, B5)
- Prompt: (ringkas, apa yang kamu minta)
- Dipakai: (bagian kode yang kamu ambil, misalnya nama fungsi dan file)
- Diubah: (apa yang kamu ubah atau perbaiki dari hasil AI, dan kenapa)

## Memulai pengerjaan
- Prompt: Apa yang harus ku lakukan untuk memulai pengerjaan
- Dipakai: main.go
- Diubah: menambahkan struct SumberDaya

## Melakukan Setup
- Prompt: melakukan setup untuk B1
- Dipakai:main.go
- Diubah: menambahkan barebone untuk pengerjaan B1,karena saya bingung apa yg harus dilakukan

## menyelesaikan B1
- Prompt: menyelesaikan B1 
- Dipakai: main.go
- Diubah: menyelesaikna semua requiremnt

## menyelesaikan B2
- Prompt: menyelesaikan B2
- Dipakai: main.go
- Diubah: menyelesaikna semua requiremnt

## menyelesaikan B3
- Prompt: Cara membuat endpoint melihat daftar dan detail peminjaman dengan filter query
- Dipakai: handleListPeminjaman dan handleGetPeminjamanByID di main.go
- Diubah: Menambahkan filter sumber_daya_id, tanggal, status serta validasi error 400 jika sumber_daya_id bukan angka

## Validasi dan Penanganan Error JSON (B4)
- Prompt: Cara memastikan format error selalu JSON dan aturan validasi input pada semua endpoint
- Dipakai: Helper writeError dan logika validasi di seluruh handler main.go
- Diubah: Memastikan Content-Type selalu application/json dan pesan error berformat {"error": "..."} dengan status code yang tepat (400, 404, 409)

## Deteksi Bentrok Kapasitas Slot (B5)
- Prompt: Buat fungsi Go untuk cek apakah rentang slot peminjaman melebihi stok kapasitas per slot
- Dipakai: Fungsi cekBentrok di main.go
- Diubah: Menghitung akumulasi jumlah unit aktif per slot (1-8) dan menambahkan parameter excludeID agar pada fungsi update (PUT) tidak bentrok dengan diri sendiri

## Ubah dan Hapus Peminjaman (B6)
- Prompt: "Buat handler PUT dan DELETE peminjaman di Go net/http dengan aturan urutan error"
- Dipakai: handleUpdatePeminjaman dan handleDeletePeminjaman di main.go
- Diubah: Memastikan pengecekan urutan error (400 ID non-angka, 404 jika tidak ditemukan, 400 body tidak valid, 409 bentrok) dan response 204 tanpa body untuk DELETE

## Refaktor Kode Menjadi Modular
- Prompt: "Refaktor kode menjadi modular 3-5 file agar lebih mudah dipahami dan dikelola"
- Dipakai: Pemisahan kode ke models.go, helpers.go, handlers.go, dan main.go
- Diubah: Memastikan seluruh file tetap berada di package main pada root repo sesuai ketentuan dosen, memisahkan fungsi handler, helper, dan struct tanpa merusak dependensi

## Keamanan Permintaan Bersamaan (B7)
- Prompt: "Cara memastikan pengujian concurrency aman menggunakan sync.Mutex di Go net/http"
- Dipakai: Penerapan s.mu.Lock() dan defer s.mu.Unlock() di seluruh operasi baca-tulis store
- Diubah: Memastikan pengecekan ketersediaan slot dan penambahan data peminjaman dilakukan dalam satu critical section kunci mutex yang sama agar tidak terjadi race condition

## Fitur Khas Persona D: Peminjaman Paket
- Prompt: "Buat endpoint POST /paket-peminjaman dan DELETE /paket-peminjaman/{id} secara atomik all-or-nothing"
- Dipakai: handleCreatePaketPeminjaman dan handleDeletePaketPeminjaman di handlers.go
- Diubah: Memvalidasi agar satu sumber daya tidak duplikat dalam satu paket, memastikan transaksi atomik (jika 1 item bentrok tolak 409 dan tidak ada ID yang terbuang), serta menghapus seluruh item paket sekaligus pada DELETE
