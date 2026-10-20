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
