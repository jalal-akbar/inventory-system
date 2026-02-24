# 📋 TESTING MANUAL — Inventory System (E2E)

> **Versi:** 1.0  
> **Tanggal:** 25 Februari 2026  
> **Tujuan:** Panduan testing manual end-to-end melalui browser untuk memastikan setiap fitur utama berjalan sesuai ekspektasi.  
> **Base URL:** `http://localhost:7070`

---

## 📌 Prasyarat

| #   | Item            | Detail                                                         |
| --- | --------------- | -------------------------------------------------------------- |
| 1   | Server berjalan | Jalankan `make run` — server aktif di port `7070`              |
| 2   | Database fresh  | Opsional: `make reset-db` lalu `make seed` untuk data bersih   |
| 3   | Akun Admin      | Username: `admin` / Password: `admin123`                       |
| 4   | Akun Staff      | Dibuat manual oleh Admin di halaman **Settings → Kelola User** |
| 5   | Browser         | Chrome / Firefox versi terbaru                                 |

---

## 🗂️ Daftar Fitur yang Diuji

1. [Autentikasi (Login & Logout)](#1-autentikasi-login--logout)
2. [Manajemen User (Admin)](#2-manajemen-user-admin)
3. [Input Produk Baru](#3-input-produk-baru)
4. [Approval Workflow (Admin)](#4-approval-workflow-admin)
5. [POS / Transaksi Penjualan](#5-pos--transaksi-penjualan)
6. [Validasi Stok & Status Produk](#6-validasi-stok--status-produk)
7. [Void Transaksi](#7-void-transaksi)

---

## 1. Autentikasi (Login & Logout)

### TC-01 — Login Admin (Happy Path)

|               |                                  |
| ------------- | -------------------------------- |
| **ID**        | TC-01                            |
| **Prioritas** | 🔴 High                          |
| **Prasyarat** | Akun admin sudah ada di database |

| #   | Langkah                            | Hasil yang Diharapkan                                                            |
| --- | ---------------------------------- | -------------------------------------------------------------------------------- |
| 1   | Buka `http://localhost:7070/login` | Halaman login tampil dengan field username & password                            |
| 2   | Isi **Username:** `admin`          | Field terisi                                                                     |
| 3   | Isi **Password:** `admin123`       | Field terisi (karakter tersembunyi)                                              |
| 4   | Klik tombol **Login**              | Redirect ke `/dashboard`                                                         |
| 5   | Periksa sidebar / topbar           | Nama user `admin` ditampilkan, menu admin (Approvals, Settings, Backup) terlihat |

**Status:** [ ]

---

### TC-02 — Login Staff (Happy Path)

|               |                                                  |
| ------------- | ------------------------------------------------ |
| **ID**        | TC-02                                            |
| **Prioritas** | 🔴 High                                          |
| **Prasyarat** | Akun staff sudah dibuat oleh Admin (lihat TC-06) |

| #   | Langkah                                    | Hasil yang Diharapkan                                                        |
| --- | ------------------------------------------ | ---------------------------------------------------------------------------- |
| 1   | Buka `http://localhost:7070/login`         | Halaman login tampil                                                         |
| 2   | Isi **Username:** `staff1`                 | Field terisi                                                                 |
| 3   | Isi **Password:** `(password yang dibuat)` | Field terisi                                                                 |
| 4   | Klik tombol **Login**                      | Redirect ke `/dashboard`                                                     |
| 5   | Periksa sidebar                            | Menu admin-only (**Approvals**, **Settings**, **Backup**) **TIDAK** terlihat |

**Status:** [ ]

---

### TC-03 — Login Gagal: Kredensial Salah

|               |           |
| ------------- | --------- |
| **ID**        | TC-03     |
| **Prioritas** | 🟡 Medium |
| **Prasyarat** | -         |

| #   | Langkah                            | Hasil yang Diharapkan                                                |
| --- | ---------------------------------- | -------------------------------------------------------------------- |
| 1   | Buka `http://localhost:7070/login` | Halaman login tampil                                                 |
| 2   | Isi **Username:** `admin`          | Field terisi                                                         |
| 3   | Isi **Password:** `wrong_password` | Field terisi                                                         |
| 4   | Klik tombol **Login**              | **Tetap** di halaman login, muncul pesan error "Invalid credentials" |

**Status:** [ ]

---

### TC-04 — Login Gagal: Akun Dinonaktifkan

|               |                                                                         |
| ------------- | ----------------------------------------------------------------------- |
| **ID**        | TC-04                                                                   |
| **Prioritas** | 🟡 Medium                                                               |
| **Prasyarat** | Akun staff sudah di-_disable_ oleh Admin (via Settings → Toggle Status) |

| #   | Langkah                                      | Hasil yang Diharapkan                                             |
| --- | -------------------------------------------- | ----------------------------------------------------------------- |
| 1   | Buka `http://localhost:7070/login`           | Halaman login tampil                                              |
| 2   | Isi kredensial akun yang sudah dinonaktifkan | Field terisi                                                      |
| 3   | Klik tombol **Login**                        | **Tetap** di halaman login, muncul pesan error "Account disabled" |

**Status:** [ ]

---

### TC-05 — Logout

|               |                  |
| ------------- | ---------------- |
| **ID**        | TC-05            |
| **Prioritas** | 🔴 High          |
| **Prasyarat** | User sudah login |

| #   | Langkah                                               | Hasil yang Diharapkan                                |
| --- | ----------------------------------------------------- | ---------------------------------------------------- |
| 1   | Klik tombol/link **Logout** di sidebar/topbar         | Redirect ke `/login`                                 |
| 2   | Coba akses `http://localhost:7070/dashboard` langsung | Redirect kembali ke `/login` (session sudah dihapus) |

**Status:** [ ]

---

## 2. Manajemen User (Admin)

### TC-06 — Admin Membuat Akun Staff Baru

|               |                     |
| ------------- | ------------------- |
| **ID**        | TC-06               |
| **Prioritas** | 🔴 High             |
| **Prasyarat** | Login sebagai Admin |

| #   | Langkah                                                              | Hasil yang Diharapkan                           |
| --- | -------------------------------------------------------------------- | ----------------------------------------------- |
| 1   | Navigasi ke **Settings** (`/settings`)                               | Halaman Settings terbuka                        |
| 2   | Klik tab **Kelola User**                                             | Tab user management aktif                       |
| 3   | Isi form: Username = `staff1`, Password = `staff123`, Role = `staff` | Field terisi                                    |
| 4   | Klik **Tambah User** / **Create User**                               | Muncul pesan sukses, user baru tampil di daftar |
| 5   | Logout dan login dengan `staff1` / `staff123`                        | Login berhasil, redirect ke dashboard           |

**Status:** [ ]

---

### TC-07 — Admin Menonaktifkan Akun Staff

|               |                                                    |
| ------------- | -------------------------------------------------- |
| **ID**        | TC-07                                              |
| **Prioritas** | 🟡 Medium                                          |
| **Prasyarat** | Login sebagai Admin, akun staff `staff1` sudah ada |

| #   | Langkah                                | Hasil yang Diharapkan                       |
| --- | -------------------------------------- | ------------------------------------------- |
| 1   | Navigasi ke **Settings → Kelola User** | Daftar user tampil                          |
| 2   | Klik toggle status pada user `staff1`  | Status berubah menjadi **inactive**         |
| 3   | Logout dan coba login dengan `staff1`  | Login gagal dengan pesan "Account disabled" |

**Status:** [ ]

---

## 3. Input Produk Baru

### TC-08 — Staff Input Produk Baru (Pending Verification)

|               |                     |
| ------------- | ------------------- |
| **ID**        | TC-08               |
| **Prioritas** | 🔴 High             |
| **Prasyarat** | Login sebagai Staff |

| #   | Langkah                                  | Hasil yang Diharapkan                                                     |
| --- | ---------------------------------------- | ------------------------------------------------------------------------- |
| 1   | Navigasi ke **Inventory** (`/inventory`) | Halaman daftar produk tampil                                              |
| 2   | Klik tombol **Tambah Produk**            | Modal/form input produk terbuka                                           |
| 3   | Isi data produk:                         |                                                                           |
|     | - Nama: `Paracetamol 500mg`              |                                                                           |
|     | - SKU: `PCT-500`                         |                                                                           |
|     | - Legal Category: `OTC`                  |                                                                           |
|     | - Therapeutic Class: `Analgesics`        |                                                                           |
|     | - Unit: `Box`                            |                                                                           |
|     | - Items per Unit: `10`                   |                                                                           |
|     | - Storage Location: `Rak A1`             |                                                                           |
|     | - Harga Beli: `15000`                    |                                                                           |
|     | - Harga Jual: `25000`                    |                                                                           |
|     | - Min Stock: `5`                         |                                                                           |
|     | - Batch Number: `B-2026-001`             |                                                                           |
|     | - Expiry Date: `2027-12-31`              |                                                                           |
|     | - Initial Stock: `50`                    |                                                                           |
| 4   | Klik **Simpan** / **Submit**             | Toast sukses muncul: "Product created successfully"                       |
| 5   | Periksa daftar produk                    | Produk `Paracetamol 500mg` muncul dengan **indikator pending/unverified** |

**Status:** [ ]

---

### TC-09 — Admin Input Produk Baru (Auto-Verified)

|               |                     |
| ------------- | ------------------- |
| **ID**        | TC-09               |
| **Prioritas** | 🟡 Medium           |
| **Prasyarat** | Login sebagai Admin |

| #   | Langkah                                  | Hasil yang Diharapkan                                                                                     |
| --- | ---------------------------------------- | --------------------------------------------------------------------------------------------------------- |
| 1   | Navigasi ke **Inventory** (`/inventory`) | Halaman daftar produk tampil                                                                              |
| 2   | Klik tombol **Tambah Produk**            | Modal/form input produk terbuka                                                                           |
| 3   | Isi data produk:                         |                                                                                                           |
|     | - Nama: `Amoxicillin 500mg`              |                                                                                                           |
|     | - SKU: `AMX-500`                         |                                                                                                           |
|     | - Legal Category: `Rx`                   |                                                                                                           |
|     | - Therapeutic Class: `Antibiotics`       |                                                                                                           |
|     | - Unit: `Strip`                          |                                                                                                           |
|     | - Items per Unit: `10`                   |                                                                                                           |
|     | - Harga Beli: `8000`                     |                                                                                                           |
|     | - Harga Jual: `15000`                    |                                                                                                           |
|     | - Min Stock: `10`                        |                                                                                                           |
|     | - Batch Number: `B-2026-002`             |                                                                                                           |
|     | - Expiry Date: `2027-06-30`              |                                                                                                           |
|     | - Initial Stock: `100`                   |                                                                                                           |
| 4   | Klik **Simpan**                          | Toast sukses muncul                                                                                       |
| 5   | Periksa daftar produk                    | Produk `Amoxicillin 500mg` muncul dengan status **verified** (langsung disetujui karena Admin yang input) |

**Status:** [ ]

---

### TC-10 — Input Produk Gagal: Nama Kosong

|               |                                |
| ------------- | ------------------------------ |
| **ID**        | TC-10                          |
| **Prioritas** | 🟡 Medium                      |
| **Prasyarat** | Login sebagai Admin atau Staff |

| #   | Langkah                                          | Hasil yang Diharapkan                                  |
| --- | ------------------------------------------------ | ------------------------------------------------------ |
| 1   | Buka form **Tambah Produk**                      | Modal terbuka                                          |
| 2   | Biarkan field **Nama** kosong, isi field lainnya | -                                                      |
| 3   | Klik **Simpan**                                  | Muncul pesan error / validasi, produk **tidak** dibuat |

**Status:** [ ]

---

### TC-11 — Input Produk Gagal: Unit Tidak Valid

|               |                                |
| ------------- | ------------------------------ |
| **ID**        | TC-11                          |
| **Prioritas** | 🟢 Low                         |
| **Prasyarat** | Login sebagai Admin atau Staff |

| #   | Langkah                                                                                          | Hasil yang Diharapkan                                                                                           |
| --- | ------------------------------------------------------------------------------------------------ | --------------------------------------------------------------------------------------------------------------- |
| 1   | Buka form **Tambah Produk**                                                                      | Modal terbuka                                                                                                   |
| 2   | Isi semua field dengan benar, tapi pilih/isi **Unit** dengan nilai tidak valid (misal: `Karung`) | -                                                                                                               |
| 3   | Klik **Simpan**                                                                                  | Muncul pesan error: "Invalid unit. Must be one of: Box, Strip, Pcs, Vial, Botol, Tube, Sachet, Ampul, Pot, Dus" |

**Status:** [ ]

---

## 4. Approval Workflow (Admin)

### TC-12 — Admin Melihat Daftar Pending Approval

|               |                                                            |
| ------------- | ---------------------------------------------------------- |
| **ID**        | TC-12                                                      |
| **Prioritas** | 🔴 High                                                    |
| **Prasyarat** | Login sebagai Admin, ada produk pending dari Staff (TC-08) |

| #   | Langkah                                                | Hasil yang Diharapkan                                                                                     |
| --- | ------------------------------------------------------ | --------------------------------------------------------------------------------------------------------- |
| 1   | Periksa sidebar/topbar                                 | Badge/counter pending items terlihat (jumlah > 0)                                                         |
| 2   | Navigasi ke **Admin → Approvals** (`/admin/approvals`) | Halaman approval dashboard tampil                                                                         |
| 3   | Periksa daftar pending items                           | Produk yang diinput Staff (misal: `Paracetamol 500mg`) muncul di daftar beserta batch dan stock entry-nya |

**Status:** [ ]

---

### TC-13 — Admin Approve Produk Satu-Satu

|               |                                                            |
| ------------- | ---------------------------------------------------------- |
| **ID**        | TC-13                                                      |
| **Prioritas** | 🔴 High                                                    |
| **Prasyarat** | Login sebagai Admin, ada pending item di halaman Approvals |

| #   | Langkah                                                     | Hasil yang Diharapkan                                         |
| --- | ----------------------------------------------------------- | ------------------------------------------------------------- |
| 1   | Buka halaman **Approvals** (`/admin/approvals`)             | Daftar pending tampil                                         |
| 2   | Klik tombol **Approve** pada satu item (produk/batch/stock) | Item menghilang dari daftar pending                           |
| 3   | Navigasi ke **Inventory** (`/inventory`)                    | Produk yang baru di-approve tampil dengan status **verified** |

**Status:** [ ]

---

### TC-14 — Admin Approve Group (Per Produk)

|               |                                                                             |
| ------------- | --------------------------------------------------------------------------- |
| **ID**        | TC-14                                                                       |
| **Prioritas** | 🟡 Medium                                                                   |
| **Prasyarat** | Login sebagai Admin, ada produk pending dengan multiple batch/stock entries |

| #   | Langkah                                         | Hasil yang Diharapkan                                                                     |
| --- | ----------------------------------------------- | ----------------------------------------------------------------------------------------- |
| 1   | Buka halaman **Approvals** (`/admin/approvals`) | Daftar pending tampil                                                                     |
| 2   | Klik tombol **Approve Group** pada satu produk  | Semua item terkait produk tersebut (produk + batch + stock entries) ter-approve sekaligus |
| 3   | Periksa kembali daftar pending                  | Produk dan semua item terkait sudah tidak muncul                                          |

**Status:** [ ]

---

### TC-15 — Admin Approve All

|               |                                                 |
| ------------- | ----------------------------------------------- |
| **ID**        | TC-15                                           |
| **Prioritas** | 🟡 Medium                                       |
| **Prasyarat** | Login sebagai Admin, ada beberapa pending items |

| #   | Langkah                                         | Hasil yang Diharapkan                          |
| --- | ----------------------------------------------- | ---------------------------------------------- |
| 1   | Buka halaman **Approvals** (`/admin/approvals`) | Daftar pending tampil                          |
| 2   | Klik tombol **Approve All**                     | Semua item langsung ter-approve                |
| 3   | Periksa kembali halaman                         | Daftar pending kosong, badge counter menjadi 0 |

**Status:** [ ]

---

### TC-16 — Admin Reject Produk

|               |                                       |
| ------------- | ------------------------------------- |
| **ID**        | TC-16                                 |
| **Prioritas** | 🟡 Medium                             |
| **Prasyarat** | Login sebagai Admin, ada pending item |

| #   | Langkah                               | Hasil yang Diharapkan                                                                                 |
| --- | ------------------------------------- | ----------------------------------------------------------------------------------------------------- |
| 1   | Buka halaman **Approvals**            | Daftar pending tampil                                                                                 |
| 2   | Klik tombol **Reject** pada satu item | Item menghilang dari daftar pending                                                                   |
| 3   | Navigasi ke **Inventory**             | Produk yang di-reject **tidak** tampil sebagai verified (atau statusnya rejected sesuai implementasi) |

**Status:** [ ]

---

### TC-17 — Staff Tidak Bisa Akses Halaman Approvals

|               |                     |
| ------------- | ------------------- |
| **ID**        | TC-17               |
| **Prioritas** | 🔴 High             |
| **Prasyarat** | Login sebagai Staff |

| #   | Langkah                                                           | Hasil yang Diharapkan                                                      |
| --- | ----------------------------------------------------------------- | -------------------------------------------------------------------------- |
| 1   | Akses langsung `http://localhost:7070/admin/approvals` di browser | Redirect atau pesan **Unauthorized / Forbidden** (bukan halaman approvals) |
| 2   | Periksa sidebar                                                   | Menu **Approvals** tidak tampil untuk role Staff                           |

**Status:** [ ]

---

## 5. POS / Transaksi Penjualan

### TC-18 — Staff Menjual Produk Approved (Happy Path)

|               |                                                                                               |
| ------------- | --------------------------------------------------------------------------------------------- |
| **ID**        | TC-18                                                                                         |
| **Prioritas** | 🔴 High                                                                                       |
| **Prasyarat** | Login sebagai Staff/Admin, produk `Paracetamol 500mg` sudah **approved** dengan stok tersedia |

| #   | Langkah                                                | Hasil yang Diharapkan                                                        |
| --- | ------------------------------------------------------ | ---------------------------------------------------------------------------- |
| 1   | Navigasi ke **POS** (`/pos`)                           | Halaman kasir tampil dengan daftar produk                                    |
| 2   | Cari produk `Paracetamol 500mg` menggunakan search bar | Produk muncul di hasil pencarian                                             |
| 3   | Klik/tambahkan produk ke keranjang                     | Produk masuk ke keranjang belanja                                            |
| 4   | Set **Quantity:** `5`                                  | Jumlah diperbarui, subtotal dihitung otomatis (5 × Rp25.000 = Rp125.000)     |
| 5   | Isi **Metode Pembayaran:** `Cash`                      | Field terisi                                                                 |
| 6   | (Opsional) Isi **Nama Pelanggan**                      | Field terisi                                                                 |
| 7   | Klik tombol **Checkout** / **Bayar**                   | Muncul respons sukses: "Transaction completed successfully" dengan `sale_id` |
| 8   | Navigasi ke **Inventory**                              | Stok produk `Paracetamol 500mg` berkurang 5 unit                             |

**Status:** [ ]

---

### TC-19 — Pencarian Produk di POS

|               |                                       |
| ------------- | ------------------------------------- |
| **ID**        | TC-19                                 |
| **Prioritas** | 🟡 Medium                             |
| **Prasyarat** | Login, beberapa produk sudah tersedia |

| #   | Langkah                                           | Hasil yang Diharapkan                                                 |
| --- | ------------------------------------------------- | --------------------------------------------------------------------- |
| 1   | Buka halaman **POS** (`/pos`)                     | Halaman kasir tampil                                                  |
| 2   | Ketik `para` di search bar                        | Produk yang mengandung kata "para" muncul (contoh: Paracetamol 500mg) |
| 3   | Filter berdasarkan kategori (misal: `Analgesics`) | Hanya produk dengan therapeutic class Analgesics yang muncul          |

**Status:** [ ]

---

### TC-20 — Checkout dengan Metode Pembayaran Transfer

|               |                                             |
| ------------- | ------------------------------------------- |
| **ID**        | TC-20                                       |
| **Prioritas** | 🟡 Medium                                   |
| **Prasyarat** | Login, produk approved dengan stok tersedia |

| #   | Langkah                                          | Hasil yang Diharapkan                                           |
| --- | ------------------------------------------------ | --------------------------------------------------------------- |
| 1   | Tambahkan produk ke keranjang di halaman **POS** | Produk masuk ke keranjang                                       |
| 2   | Set quantity                                     | Subtotal terupdate                                              |
| 3   | Pilih **Metode Pembayaran:** `Transfer`          | Field terisi                                                    |
| 4   | Klik **Checkout**                                | Transaksi sukses, `payment_method` tersimpan sebagai `Transfer` |

**Status:** [ ]

---

### TC-21 — Checkout dengan Metode Pembayaran QRIS

|               |                                             |
| ------------- | ------------------------------------------- |
| **ID**        | TC-21                                       |
| **Prioritas** | 🟡 Medium                                   |
| **Prasyarat** | Login, produk approved dengan stok tersedia |

| #   | Langkah                                          | Hasil yang Diharapkan                                       |
| --- | ------------------------------------------------ | ----------------------------------------------------------- |
| 1   | Tambahkan produk ke keranjang di halaman **POS** | Produk masuk ke keranjang                                   |
| 2   | Pilih **Metode Pembayaran:** `QRIS`              | Field terisi                                                |
| 3   | Klik **Checkout**                                | Transaksi sukses, `payment_method` tersimpan sebagai `QRIS` |

**Status:** [ ]

---

### TC-22 — Checkout dengan Diskon

|               |                                             |
| ------------- | ------------------------------------------- |
| **ID**        | TC-22                                       |
| **Prioritas** | 🟡 Medium                                   |
| **Prasyarat** | Login, produk approved dengan stok tersedia |

| #   | Langkah                                          | Hasil yang Diharapkan                                          |
| --- | ------------------------------------------------ | -------------------------------------------------------------- |
| 1   | Tambahkan produk ke keranjang di halaman **POS** | Produk masuk                                                   |
| 2   | Set quantity: `2`, harga jual Rp25.000           | Subtotal: Rp50.000                                             |
| 3   | Isi field **Diskon:** `5000`                     | Diskon terinput                                                |
| 4   | Klik **Checkout**                                | Transaksi sukses, total = Rp45.000 (Rp50.000 - Rp5.000 diskon) |

**Status:** [ ]

---

## 6. Validasi Stok & Status Produk

### TC-23 — Menjual Produk Melebihi Stok (Edge Case)

|               |                                                               |
| ------------- | ------------------------------------------------------------- |
| **ID**        | TC-23                                                         |
| **Prioritas** | 🔴 High                                                       |
| **Prasyarat** | Login, produk approved dengan stok terbatas (misal: stok = 3) |

| #   | Langkah                                       | Hasil yang Diharapkan                                           |
| --- | --------------------------------------------- | --------------------------------------------------------------- |
| 1   | Buka halaman **POS**                          | Kasir tampil                                                    |
| 2   | Tambahkan produk dengan stok = 3 ke keranjang | Produk masuk ke keranjang                                       |
| 3   | Set **Quantity:** `10` (melebihi stok)        | -                                                               |
| 4   | Klik **Checkout**                             | Muncul pesan error: stok tidak mencukupi. Transaksi **DITOLAK** |
| 5   | Periksa daftar produk di Inventory            | Stok tetap = 3 (tidak berubah)                                  |

**Status:** [ ]

---

### TC-24 — Menjual Produk yang Belum Approved (Edge Case)

|               |                                                                         |
| ------------- | ----------------------------------------------------------------------- |
| **ID**        | TC-24                                                                   |
| **Prioritas** | 🔴 High                                                                 |
| **Prasyarat** | Login sebagai Staff, input produk baru yang belum di-approve oleh Admin |

| #   | Langkah                                       | Hasil yang Diharapkan                                                                     |
| --- | --------------------------------------------- | ----------------------------------------------------------------------------------------- |
| 1   | Input produk baru sebagai Staff (lihat TC-08) | Produk dibuat dengan status **unverified**                                                |
| 2   | Buka halaman **POS** (`/pos`)                 | Kasir tampil                                                                              |
| 3   | Cari produk yang baru dibuat                  | Produk **tidak muncul** di hasil pencarian POS (hanya produk verified yang tampil di POS) |

> **Catatan:** POS hanya menampilkan produk dengan batch yang `is_verified = 1`, sehingga produk unverified secara otomatis tidak bisa dijual.

**Status:** [ ]

---

### TC-25 — Menjual Produk dengan Stok = 0

|               |                                                     |
| ------------- | --------------------------------------------------- |
| **ID**        | TC-25                                               |
| **Prioritas** | 🟡 Medium                                           |
| **Prasyarat** | Login, produk approved dengan stok yang sudah habis |

| #   | Langkah                                  | Hasil yang Diharapkan                                                              |
| --- | ---------------------------------------- | ---------------------------------------------------------------------------------- |
| 1   | Buka halaman **POS**                     | Kasir tampil                                                                       |
| 2   | Cari / pilih produk dengan stok = 0      | Produk mungkin tidak tampil, atau tampil tanpa opsi untuk ditambahkan ke keranjang |
| 3   | Coba tambahkan ke keranjang dan checkout | Muncul pesan error / transaksi ditolak                                             |

**Status:** [ ]

---

## 7. Void Transaksi

### TC-26 — Admin Void Transaksi Langsung

|               |                                          |
| ------------- | ---------------------------------------- |
| **ID**        | TC-26                                    |
| **Prioritas** | 🟡 Medium                                |
| **Prasyarat** | Login sebagai Admin, ada transaksi aktif |

| #   | Langkah                                          | Hasil yang Diharapkan                                             |
| --- | ------------------------------------------------ | ----------------------------------------------------------------- |
| 1   | Navigasi ke **Sales History** (`/sales-history`) | Daftar riwayat penjualan tampil                                   |
| 2   | Pilih transaksi yang ingin di-void               | Detail transaksi tampil                                           |
| 3   | Klik **Void** dan isi alasan                     | -                                                                 |
| 4   | Konfirmasi void                                  | Transaksi berubah status menjadi `void`, stok produk dikembalikan |

**Status:** [ ]

---

### TC-27 — Staff Request Void Transaksi

|               |                                          |
| ------------- | ---------------------------------------- |
| **ID**        | TC-27                                    |
| **Prioritas** | 🟡 Medium                                |
| **Prasyarat** | Login sebagai Staff, ada transaksi aktif |

| #   | Langkah                            | Hasil yang Diharapkan                                                    |
| --- | ---------------------------------- | ------------------------------------------------------------------------ |
| 1   | Navigasi ke **Sales History**      | Daftar riwayat penjualan tampil                                          |
| 2   | Pilih transaksi yang ingin di-void | Detail transaksi tampil                                                  |
| 3   | Klik **Void** dan isi alasan       | -                                                                        |
| 4   | Konfirmasi                         | Transaksi berubah status menjadi `pending_void`, **bukan langsung void** |
| 5   | Login sebagai **Admin**            | -                                                                        |
| 6   | Buka halaman **Approvals**         | Request void dari Staff tampil di daftar pending                         |
| 7   | **Approve** request void           | Status transaksi berubah menjadi `void`, stok dikembalikan               |

**Status:** [ ]

---

## 📊 E2E Full Workflow Test

### TC-28 — Alur Lengkap: Input Produk → Approve → Jual di Kasir

|               |                                                          |
| ------------- | -------------------------------------------------------- |
| **ID**        | TC-28                                                    |
| **Prioritas** | 🔴 Critical                                              |
| **Prasyarat** | Database bersih (opsional: `make reset-db && make seed`) |

Ini adalah test case paling kritis, menguji seluruh alur dari awal sampai akhir:

| #                                | Langkah                                                                                                                                                                                                                                              | Hasil yang Diharapkan                                  |
| -------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------ |
| **Fase 1: Setup**                |                                                                                                                                                                                                                                                      |                                                        |
| 1                                | Login sebagai **Admin**                                                                                                                                                                                                                              | Dashboard tampil                                       |
| 2                                | Buat akun Staff baru di **Settings → Kelola User** (username: `kasir1`, password: `kasir123`, role: `staff`)                                                                                                                                         | Akun berhasil dibuat                                   |
| 3                                | Logout                                                                                                                                                                                                                                               | Redirect ke login                                      |
| **Fase 2: Input Produk (Staff)** |                                                                                                                                                                                                                                                      |                                                        |
| 4                                | Login sebagai **Staff** (`kasir1` / `kasir123`)                                                                                                                                                                                                      | Dashboard tampil tanpa menu admin                      |
| 5                                | Buka **Inventory** → **Tambah Produk**                                                                                                                                                                                                               | Modal input terbuka                                    |
| 6                                | Isi: Nama=`Vitamin C 1000mg`, SKU=`VTC-1000`, Legal Category=`OTC`, Therapeutic Class=`Vitamins`, Unit=`Box`, Items per Unit=`12`, Harga Beli=`20000`, Harga Jual=`35000`, Min Stock=`10`, Batch=`VTC-B001`, Expiry=`2028-01-31`, Initial Stock=`30` | -                                                      |
| 7                                | Simpan                                                                                                                                                                                                                                               | Produk berhasil dibuat, status **unverified/pending**  |
| 8                                | Buka **POS** → cari `Vitamin C 1000mg`                                                                                                                                                                                                               | Produk **TIDAK muncul** di POS (belum approved)        |
| 9                                | Logout                                                                                                                                                                                                                                               | Redirect ke login                                      |
| **Fase 3: Approval (Admin)**     |                                                                                                                                                                                                                                                      |                                                        |
| 10                               | Login sebagai **Admin**                                                                                                                                                                                                                              | Dashboard tampil, badge pending > 0                    |
| 11                               | Buka **Admin → Approvals**                                                                                                                                                                                                                           | `Vitamin C 1000mg` muncul di daftar pending            |
| 12                               | Klik **Approve Group** atau approve satu per satu                                                                                                                                                                                                    | Semua item (product + batch + stock entry) ter-approve |
| 13                               | Logout                                                                                                                                                                                                                                               | Redirect ke login                                      |
| **Fase 4: Penjualan (Staff)**    |                                                                                                                                                                                                                                                      |                                                        |
| 14                               | Login kembali sebagai **Staff** (`kasir1`)                                                                                                                                                                                                           | Dashboard tampil                                       |
| 15                               | Buka **POS** → cari `Vitamin C 1000mg`                                                                                                                                                                                                               | Produk **MUNCUL** di POS (sudah approved!) ✅          |
| 16                               | Tambahkan ke keranjang, qty = `3`                                                                                                                                                                                                                    | Subtotal = 3 × Rp35.000 = Rp105.000                    |
| 17                               | Pilih pembayaran **Cash**, klik **Checkout**                                                                                                                                                                                                         | Transaksi sukses!                                      |
| **Fase 5: Verifikasi**           |                                                                                                                                                                                                                                                      |                                                        |
| 18                               | Buka **Inventory** → cari `Vitamin C 1000mg`                                                                                                                                                                                                         | Stok tersisa = **27** (30 - 3) ✅                      |
| 19                               | Buka **Sales History** (`/sales-history`)                                                                                                                                                                                                            | Transaksi terbaru tampil dengan detail yang benar      |

**Status:** [ ]

---

## 📝 Ringkasan Test Cases

| ID    | Deskripsi                            | Prioritas   | Status |
| ----- | ------------------------------------ | ----------- | ------ |
| TC-01 | Login Admin (Happy Path)             | 🔴 High     | [ ]    |
| TC-02 | Login Staff (Happy Path)             | 🔴 High     | [ ]    |
| TC-03 | Login Gagal: Kredensial Salah        | 🟡 Medium   | [ ]    |
| TC-04 | Login Gagal: Akun Dinonaktifkan      | 🟡 Medium   | [ ]    |
| TC-05 | Logout                               | 🔴 High     | [ ]    |
| TC-06 | Admin Membuat Akun Staff Baru        | 🔴 High     | [ ]    |
| TC-07 | Admin Menonaktifkan Akun Staff       | 🟡 Medium   | [ ]    |
| TC-08 | Staff Input Produk Baru (Pending)    | 🔴 High     | [ ]    |
| TC-09 | Admin Input Produk (Auto-Verified)   | 🟡 Medium   | [ ]    |
| TC-10 | Input Produk Gagal: Nama Kosong      | 🟡 Medium   | [ ]    |
| TC-11 | Input Produk Gagal: Unit Tidak Valid | 🟢 Low      | [ ]    |
| TC-12 | Admin Melihat Daftar Pending         | 🔴 High     | [ ]    |
| TC-13 | Admin Approve Produk Satu-Satu       | 🔴 High     | [ ]    |
| TC-14 | Admin Approve Group                  | 🟡 Medium   | [ ]    |
| TC-15 | Admin Approve All                    | 🟡 Medium   | [ ]    |
| TC-16 | Admin Reject Produk                  | 🟡 Medium   | [ ]    |
| TC-17 | Staff Tidak Bisa Akses Approvals     | 🔴 High     | [ ]    |
| TC-18 | Staff Menjual Produk Approved        | 🔴 High     | [ ]    |
| TC-19 | Pencarian Produk di POS              | 🟡 Medium   | [ ]    |
| TC-20 | Checkout Transfer                    | 🟡 Medium   | [ ]    |
| TC-21 | Checkout QRIS                        | 🟡 Medium   | [ ]    |
| TC-22 | Checkout dengan Diskon               | 🟡 Medium   | [ ]    |
| TC-23 | Jual Melebihi Stok                   | 🔴 High     | [ ]    |
| TC-24 | Jual Produk Belum Approved           | 🔴 High     | [ ]    |
| TC-25 | Jual Produk Stok Habis               | 🟡 Medium   | [ ]    |
| TC-26 | Admin Void Transaksi                 | 🟡 Medium   | [ ]    |
| TC-27 | Staff Request Void                   | 🟡 Medium   | [ ]    |
| TC-28 | Full Workflow E2E                    | 🔴 Critical | [ ]    |

---

> **Total Test Cases:** 28  
> **Passed:** ** / 28  
> **Failed:** ** / 28  
> **Blocked:** \_\_ / 28
