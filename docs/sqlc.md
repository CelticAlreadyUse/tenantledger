# sqlc dan PostgreSQL

sqlc menghasilkan kode Go bertipe dari SQL. Versi tool dipin ke `v1.31.1` di
Makefile dan dipasang melalui `go install` ke `bin/tools/sqlc/v1.31.1/sqlc`.
Dependency compiler sqlc tidak masuk ke `go.mod` aplikasi; generated code memakai
driver `github.com/jackc/pgx/v5` yang tercatat pada `go.mod`/`go.sum`.

## Struktur per modul

```text
sqlc.yaml
internal/modules/tenancy/internal/infrastructure/postgres/
  migrations/
    000001_create_tenants.up.sql
    000001_create_tenants.down.sql
  queries/
    tenants.sql
  db/
    db.go
    models.go
    querier.go
    tenants.sql.go
```

Tabel awal `tenancy.tenants` berisi UUID tenant, nama, dan waktu pembuatan.
UUID disediakan pemanggil; schema tidak membutuhkan extension UUID PostgreSQL.
Query awal hanya `CreateTenant` dan `GetTenant` berdasarkan ID. Validasi input,
cek akses, dan lifecycle tenant akan ditambahkan pada application/domain sebelum
query ini dipakai oleh endpoint bisnis.

Generated model merupakan tipe persistence, bukan entity domain atau DTO HTTP.
Repository adapter milik tenancy nantinya mengonversi generated model ke entity
domain. Modul lain mengakses kontrak tenancy melalui facade, bukan mengimpor
generated package atau membaca tabel tenancy secara langsung.

## Alur pengembangan

```sh
make sqlc-install
make sqlc-generate
make sqlc-check
make check
```

Ubah SQL di `queries/` atau tambahkan migration bernomor berikutnya di
`migrations/`, kemudian generate ulang. File `.down.sql` diabaikan saat sqlc
menganalisis schema. Simpan SQL sumber dan generated Go bersama di Git; jangan
mengedit file generated secara manual. `sqlc-check` menjalankan `compile` dan
`diff`, sehingga perubahan SQL yang belum di-generate menyebabkan kegagalan.
Pengecekan ini tidak memerlukan server database atau sqlc Cloud.

Untuk modul baru, tambahkan block tersendiri pada daftar `sql` di `sqlc.yaml`
dengan schema, queries, nama package, dan output directory milik modul tersebut.
Jangan memakai satu generated package global untuk semua modul.

## Menghubungkan database nanti

Schema ini untuk **database bisnis TenantLedger yang terpisah dari Authentik**.
Setup sqlc tidak membuat database, tidak menjalankan migration, dan tidak
mengubah database Authentik. File migration memakai pasangan `.up.sql` dan
`.down.sql`; runner migration otomatis belum dipasang.

Untuk database bisnis lokal yang masih kosong, migration pertama dapat
dijalankan sekali dengan `psql` setelah `DATABASE_URL` diatur ke database bisnis:

```sh
psql "$DATABASE_URL" -v ON_ERROR_STOP=1 --single-transaction \
  -f internal/modules/tenancy/internal/infrastructure/postgres/migrations/000001_create_tenants.up.sql
```

Perintah manual ini tidak melacak versi migration. Pasang migration runner
sebelum memakai banyak migration atau deployment bersama. Jangan menjalankan
file down pada database berisi data kecuali memang bermaksud menghapusnya.

Saat adapter repository diimplementasikan, gunakan `tenancydb.New(pool)` untuk
`*pgxpool.Pool` atau `tenancydb.New(tx)`/`WithTx(tx)` untuk transaksi pgx. Buat
dan tutup pool di lifecycle aplikasi lalu injeksikan ke facade modul. File
generated berada di nested `internal`, jadi hanya implementasi modul tenancy
yang boleh mengimpornya. API saat ini belum membuat pool atau memanggil query.

Query pada tabel data bisnis mendatang harus menyertakan tenant ID yang telah
diverifikasi dan cek membership/permission sebelum akses. Code generation
memastikan kecocokan SQL dan tipe, tetapi tidak menyediakan tenant isolation
atau otorisasi secara otomatis.

Referensi: [tutorial PostgreSQL sqlc](https://docs.sqlc.dev/en/v1.31.1/tutorials/getting-started-postgresql.html)
dan [konfigurasi sqlc](https://docs.sqlc.dev/en/v1.31.1/reference/config.html).
