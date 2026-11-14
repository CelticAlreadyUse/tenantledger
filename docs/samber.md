# Library Samber

Versi dependency yang dipasang: `lo v1.53.0`, `do/v2 v2.1.0`, dan
`oops v1.23.2`. Versi persis dan checksum tersimpan pada `go.mod` dan `go.sum`.

## lo: helper data

`internal/app/config.go` memakai `lo.CoalesceOrEmpty` untuk default `HTTP_ADDR`.
Untuk manipulasi koleksi di application atau adapter, gunakan helper seperti
`lo.Map`, `lo.Filter`, dan `lo.Uniq` jika membuat maksud kode lebih jelas.
Tidak perlu membuat wrapper di `platform` untuk setiap fungsi `lo`.

Contoh menghapus duplikasi daftar permission yang sudah divalidasi:

```go
permissions := lo.Uniq([]string{"tenant:read", "tenant:read", "tenant:write"})
```

Gunakan versi sekuensial secara default; penggunaan helper bukan alasan untuk
menambah concurrency atau meniadakan validasi dan pemeriksaan akses.

## do v2: dependency dan lifecycle

`app.New` membuat container privat, mendaftarkan `Config`, `http.Handler`, dan
`*http.Server`, kemudian me-resolve server sebelum aplikasi mulai menerima
request. Provider memakai `do.Invoke` yang mengembalikan error agar kegagalan
initialization dapat dilaporkan tanpa panic.

`Application.Run(ctx)` membuka listener, melayani request, lalu melakukan
`ShutdownWithContext` dengan timeout 10 detik saat context dibatalkan atau
server gagal. `http.Server` sudah memiliki method `Shutdown(context.Context)
error`, sehingga dikenali oleh do tanpa wrapper lifecycle tambahan. Jika
graceful shutdown gagal, aplikasi menutup koneksi tersisa dan mengembalikan error.
Setiap instance aplikasi hanya dipakai untuk satu kali `Run`.

Untuk modul mendatang, tambahkan provider di composition root dan injeksikan
dependency facade melalui constructor. Jangan meneruskan injector ke handler,
domain, use case, atau repository; hindari service locator di logika bisnis.
Detail nested `internal` modul tetap privat. Container hanya mengelola service
yang didaftarkan; database dan Authentik client belum dihubungkan pada runtime.

Saat menambah resource seperti database pool, pastikan facade atau adapter
pemiliknya memiliki kontrak shutdown yang sesuai. `pgxpool.Pool.Close()` tidak
langsung memenuhi interface shutdown do; adapter perlu mendelegasikan cleanup
ke `Close()` lewat method `Shutdown`.

## oops: error dengan konteks

Startup dan lifecycle memakai error code seperti `dependency_init_failed`,
`listen_failed`, `serve_failed`, dan `shutdown_failed`. Kegagalan listener
menyimpan alamat sebagai metadata. `Wrap` mempertahankan cause asli untuk
`errors.Is` dan `errors.As`, serta menangkap stack trace.

Contoh pola di batas adapter:

```go
return oops.In("tenancy.postgres").
    Code("tenant_lookup_failed").
    Wrap(err)
```

Tambahkan metadata yang diperlukan dan aman. Jangan memasukkan DSN dengan
password, token, cookie, authorization header, atau body request secara utuh.
Jangan mengganti error terduga seperti not-found menjadi error internal tanpa
memetakan maknanya pada application/transport.

`cmd/api` memakai JSON handler `slog`; error oops menyediakan representasi
terstruktur termasuk code, metadata, dan stack trace. Catat error sekali pada
batas pemilik request/proses. Error startup menghentikan proses dengan exit code
1 setelah cleanup selesai; SIGINT/SIGTERM normal menghasilkan exit code 0.
Metadata dan stack trace ditujukan untuk log internal. Handler HTTP mendatang
harus memetakan error ke status dan pesan publik yang aman, bukan mengirim
serialisasi error mentah ke client.

Referensi resmi:

- [lo](https://github.com/samber/lo)
- [do v2](https://github.com/samber/do)
- [oops](https://github.com/samber/oops)
