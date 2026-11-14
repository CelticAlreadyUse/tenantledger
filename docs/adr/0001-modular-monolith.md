# ADR-0001: Modular monolith dengan batas implementasi per modul

Status: diterapkan berdasarkan permintaan pengguna untuk struktur modular monolith.

## Konteks

Proyek masih pada tahap kerangka API dan deployment Authentik. Fitur berikutnya
mencakup IAM, RBAC, organisasi, membership, dan tenant isolation. Struktur perlu
memisahkan tanggung jawab bisnis, tetap mudah dijalankan secara lokal, serta
mempertahankan satu build Go dan satu deployment aplikasi.

## Keputusan

Gunakan `internal/modules/{identity,organization,tenancy,access}`. Setiap modul
memiliki facade root dan implementasi pada nested `internal/`, dibagi menjadi
domain, application, infrastructure, serta transport HTTP. Composition root ada
di `internal/app`; adapter teknis bersama ada di `internal/platform`.
Gunakan satu `go.mod`, satu executable `cmd/api`, dan komunikasi Go dalam proses.
Modul memiliki data dan repository masing-masing saat persistence ditambahkan.

## Alternatif

- Paket global `handlers/services/repositories`: mudah dimulai, tetapi ownership
  fitur dan akses data antar domain menjadi kurang jelas.
- Microservices atau `go.mod` per modul: dapat memisahkan release, tetapi menambah
  kontrak jaringan, deployment, dan koordinasi versi yang belum dibutuhkan.

## Konsekuensi

Go menegakkan batas nested `internal`, sedangkan kontrak lintas modul tetap perlu
ditinjau agar tidak membentuk ketergantungan siklik. Struktur memiliki lebih
banyak folder pada fase awal; layer dibiarkan sebagai scaffold sampai diperlukan.
Semua modul berbagi lifecycle proses dan deployment, sehingga tidak dapat
diskalakan atau dirilis secara independen. Struktur ini tidak otomatis menjamin
tenant isolation; pemeriksaan akses dan pengujian isolasi harus ditambahkan
bersama implementasi bisnis.
