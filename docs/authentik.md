# Authentik lokal

## Arsitektur dan batas akses

Compose menjalankan PostgreSQL 16, Authentik server, dan Authentik worker.
Server dan worker dipin ke versi `2026.8.5`, mengikuti Compose resmi saat setup.
Database ini khusus Authentik; database bisnis TenantLedger harus terpisah.
PostgreSQL tidak mem-publish port ke host.

Volume bernama `database`, `data`, `certs`, dan `templates` memakai project
`tenantledger-authentik`. Server dan worker berbagi data serta template.
Worker berjalan sebagai user bawaan image tanpa mount Docker socket; pengelolaan
outpost otomatis melalui Docker tidak diaktifkan. OIDC tidak memerlukan outpost.
Tidak ada bootstrap password/token admin yang ditanam di repository.

`scripts/authentik.sh` selalu memakai file Compose dan `.env` dari direktori
deployment, sehingga perintah tidak bergantung pada working directory pemanggil.
Environment host dengan nama sama tetap dapat mengoverride nilai `.env` menurut
aturan Compose. Hindari menjalankan `config` tanpa redirection karena hasilnya
berisi secret; `make authentik-check` menyembunyikan hasil render.

## Setup provider OIDC untuk TenantLedger

Setelah `make authentik-up`, selesaikan initial setup melalui browser lalu masuk
ke Admin interface. Pembuatan provider ini dilakukan setelah Authentik berjalan;
Compose tidak membuat aplikasi/provider secara otomatis.

1. Pilih **Applications → Applications → Create with Provider**.
2. Isi nama `TenantLedger`, slug `tenantledger`, dan pilih OAuth2/OpenID Provider.
3. Gunakan client confidential jika kode ditukar oleh backend; gunakan public
   client + PKCE untuk frontend yang tidak dapat menyimpan secret.
4. Pilih authentication/authorization flow bawaan yang sesuai, signing key
   asimetris, serta scope mappings `openid`, `profile`, dan `email` seperlunya.
5. Tetapkan redirect URI dengan mode **Strict** sesuai callback aplikasi yang
   nanti diimplementasikan, misalnya `http://localhost:8080/auth/callback`.
   Callback tersebut belum tersedia pada kerangka API ini.
6. Gunakan issuer mode per-provider bawaan. Simpan client ID/secret ke environment
   aplikasi secara lokal; jangan masukkan secret ke Git atau log.

Untuk slug di atas, discovery document dapat diperiksa dengan:

```sh
curl -fsS http://localhost:9000/application/o/tenantledger/.well-known/openid-configuration
```

Issuer: `http://localhost:9000/application/o/tenantledger/`. Gunakan hostname yang
konsisten: `localhost` dan `127.0.0.1` menghasilkan issuer berbeda. Jika API nanti
berjalan dalam container, `localhost` menunjuk container itu sendiri; siapkan
hostname bersama yang dapat dijangkau browser dan API.

Kerangka ini belum melakukan pertukaran authorization code, validasi token,
atau menyediakan endpoint bisnis. Saat integrasi API dibuat, verifikasi signature
melalui JWKS, issuer, audience, expiry, serta state/nonce dan PKCE untuk login.
Gunakan identitas `(issuer, subject)` sebagai referensi user; otorisasi tenant,
role, dan membership tetap harus diverifikasi oleh TenantLedger di setiap
endpoint bisnis. Jangan mempercayai tenant ID yang dikirim klien tanpa cek membership.

## Operasional

```sh
make authentik-status
make authentik-logs
make authentik-down
```

`down` mempertahankan volume. Jangan memakai `down -v` kecuali sengaja menghapus
seluruh data Authentik. Setup script mempertahankan `.env` yang sudah ada; jika
membuat file dengan menyalin `.env.example`, isi `PG_PASS` dan
`AUTHENTIK_SECRET_KEY` sendiri dengan nilai acak sebelum menjalankan Compose.

Jika port terpakai, ubah `COMPOSE_PORT_HTTP`/`COMPOSE_PORT_HTTPS` di `.env` lalu
jalankan `make authentik-up` lagi dan sesuaikan URL issuer/callback.
Jika startup gagal, periksa log PostgreSQL dan worker dengan
`sh scripts/authentik.sh logs --tail=100 postgresql worker`.
Jangan memasang timezone host ke container Authentik; operasi internal memakai UTC.

Upgrade dilakukan secara sengaja: backup PostgreSQL, data, dan `.env`, baca
release notes, ubah `AUTHENTIK_TAG`, lalu pull dan jalankan kembali service.
Jangan memakai tag `latest` untuk upgrade otomatis.

Referensi:

- [Instalasi Docker Compose resmi](https://docs.goauthentik.io/install-config/install/docker-compose/)
- [Provider OAuth2/OIDC dan endpoint](https://docs.goauthentik.io/add-secure-apps/providers/oauth2/)
- [Pola layout proyek Go](https://github.com/golang-standards/project-layout)
