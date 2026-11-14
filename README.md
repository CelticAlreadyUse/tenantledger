# TenantLedger

Kerangka Go API modular monolith untuk aplikasi multi-tenant, dengan Authentik sebagai identity
provider lokal. Saat ini API menyediakan `GET /healthz`; login OIDC, RBAC,
membership organisasi, dan isolasi tenant belum diimplementasikan.

## Menjalankan Authentik

Prasyarat: Docker + Compose v2 atau Podman + podman-compose, OpenSSL, Make,
minimal 2 CPU dan RAM 2 GB. Script otomatis memilih Docker Compose jika tersedia,
kemudian podman-compose.

```sh
make setup              # buat secret lokal sekali, tanpa menimpa .env yang ada
make authentik-check    # validasi Compose tanpa menampilkan secret
make authentik-up
make authentik-status
```

Buka http://localhost:9000 dan selesaikan setup akun `akadmin`. Jika halaman
setup tidak terbuka otomatis, gunakan
http://localhost:9000/if/flow/initial-setup/ (termasuk slash terakhir).
Tunggu migrasi pertama selesai; lihat `make authentik-logs` jika belum siap.

```sh
make authentik-down     # hentikan service; volume database tetap tersimpan
```

Secret ada di `deployments/authentik/.env` (permission `600`, diabaikan Git).
Jangan mengganti secret/password setelah volume database terisi tanpa prosedur
rotasi. Simpan cadangan `.env` bersama backup database. Port HTTP `9000` dan
HTTPS `9443` hanya di-publish ke loopback. Deployment ini ditujukan untuk lokal;
gunakan domain dan TLS yang valid sebelum akses dari luar mesin.

Panduan provider OIDC, persistence, dan troubleshooting:
[docs/authentik.md](docs/authentik.md).

## Menjalankan API

Gunakan Go 1.26 atau lebih baru:

```sh
make run
curl http://127.0.0.1:8080/healthz
make check
make build              # bin/tenantledger-api
```

Alamat dapat diubah lewat `HTTP_ADDR`. Contoh konfigurasi aplikasi ada di
`configs/.env.example`; aplikasi membaca environment proses, bukan file dotenv.
Health endpoint hanya memeriksa proses API, bukan kesiapan Authentik.

HTTP router memakai [Gin](https://gin-gonic.com/en/docs/quickstart/), dengan
middleware request logging dan panic recovery. Query string tidak dicatat pada
access log. Forwarded headers tidak dipercaya secara default. Untuk menjalankan
dalam release mode, gunakan `GIN_MODE=release make run`.

## SQL dan code generation

sqlc dipin pada versi `v1.31.1`, dengan generated code PostgreSQL memakai `pgx/v5`.

```sh
make sqlc-install       # install tool lokal ke bin/tools
make sqlc-generate      # generate kode dari schema dan query
make sqlc-check         # compile SQL dan periksa generated code tetap sinkron
```

Konfigurasi ada di `sqlc.yaml`. Modul `tenancy` memiliki migration awal tabel
`tenancy.tenants`, query `CreateTenant`/`GetTenant`, dan generated package privat.
Generation tidak membutuhkan database hidup dan tidak menjalankan migration.
API belum membuka koneksi database atau mengekspos endpoint tenant.
Lihat [panduan sqlc](docs/sqlc.md) untuk alur perubahan SQL dan penambahan modul.

## Library aplikasi

| Library | Penggunaan |
| --- | --- |
| `samber/lo` | Default konfigurasi dan helper manipulasi koleksi saat diperlukan |
| `samber/do/v2` | Registrasi dependency, resolusi service, dan shutdown di composition root |
| `samber/oops` | Error dengan code, metadata, cause asli, dan stack trace |

`internal/app` mengelola container dan lifecycle HTTP server. `cmd/api` membaca
environment, menerima signal proses, dan mencatat error sebagai JSON melalui
`slog`. Panduan pemakaian ada di [docs/samber.md](docs/samber.md).

## Layout

```text
cmd/api/                 entry point dan lifecycle executable Go
internal/app/            composition root: wiring modul dan dependency
internal/modules/        modul bisnis dengan batas implementasi privat
  identity/              identitas eksternal, rencana login OIDC dan session
  organization/          organisasi, undangan, dan membership
  tenancy/               lifecycle tenant dan resolusi tenant aktif
  access/                role, permission, dan keputusan otorisasi
internal/platform/       adapter teknis bersama, termasuk Gin HTTP router
configs/                 contoh konfigurasi aplikasi
deployments/authentik/   Compose dan contoh environment Authentik
scripts/                 setup dan perintah operasional
docs/                    panduan pengembangan dan integrasi
progress/                catatan belajar yang sudah ada
```

Layout mengacu pada [golang-standards/project-layout](https://github.com/golang-standards/project-layout).
Referensi tersebut merupakan pola komunitas, bukan standar resmi tim Go.
Hanya direktori yang dipakai yang dibuat: belum ada kebutuhan untuk `pkg/`,
`web/`, atau `api/`. Konfigurasi Compose sebelumnya di `infra/compose/compose.yml`
sekarang berada di `deployments/authentik/compose.yml`.

Setiap modul memiliki facade pada root paket dan folder
`internal/{domain,application,infrastructure,transport/http}`. Struktur modul
masih berupa scaffold dengan query database awal pada tenancy; belum menyediakan
fitur bisnis melalui API. Satu `go.mod`, satu
executable, dan satu deployment API digunakan untuk seluruh modul.

Aturan dependency, kepemilikan data, dan contoh penempatan kode dijelaskan di
[arsitektur modular monolith](docs/architecture.md).
