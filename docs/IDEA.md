# Veyra

Local Reverse Proxy Control Plane

- Veyra bukan web server
- Veyra bukan reverse proxy engine

Veyra adalah control plane yang mengelola Caddy sebagai data plane

## Bagaimana Cara Kerjanya?

Veyra mengatur "siapa yang menerima domain apa, diteruskan ke mana, dan bagaimana request/response
diubah/di-rewrite", sedangkan Caddy menjalankan traffic HTTP sebenarnya dengan TLS, HTTP/2, dan
reverse proxy yang battle-tested.

## Masalah yang Diselesaikan

Contoh kasus nyata:

```
Internet
    |
    |
zahra.syntopia.dev
    |
 Pangolin Tunnel
    |
  Veyra
    |
  Caddy
    |
 studio.makeupartis.test
    |
   Valet
    |
 Laravel + Filament
```

Laravel percaya:

```
Host: studio.makeupartis.test
```

Sehingga:

- `request()->getHost()` benar
- Filament `->domain()` tetap berjalan
- session/cookie/domain logic tetap konsisten

Namun browser tetap melihat:

```
https://zahra.syntopia.dev
```

Karena response seperti:

```
Location: https://studio.makeupartis.test/login
```

diubah menjadi:

```
Location: https://zahra.syntopia.dev/login
```

## Arsitektur

### High-level

```
                  Veyra CLI
                      |
                      |
                  Veyra Core
                      |
          +-----------+-----------+
          |                       |
       BoltDB              Caddy Config Generator
          |                       |
          +-----------+-----------+
                      |
               Caddy Controller
                      |
              reload / reconfigure
                      |
                  Embedded Caddy
                      |
                 Reverse Proxy
                      |
                    Valet
                      |
                  Laravel
```

## Komponen

### 1. CLI

Interface utama pengguna.

Contoh:

```
veyra add zahra.syntopia.dev studio.makeupartis.test
veyra list
veyra remove zahra.syntopia.dev
veyra start
```

### 2. BoltDB

Database internal Veyra.

Fungsi:

- menyimpan semua aturan (rules)
- menjadi source of truth
- dipakai untuk regenerate konfigurasi Caddy

File:

```
veyra.db
```

### 3. Rule Engine

Veyra tidak menyimpan hanya:

```
public -> internal
```

tetapi:

```
pattern -> target
```

Contoh model:

```
Rule {
    PublicPattern
    InternalTarget
}
```

## Jenis Rule

### 1. Static Mapping

Command:

```
veyra add zahra.syntopia.dev studio.makeupartis.test
```

Hasil:

```
zahra.syntopia.dev
        |
        v
studio.makeupartis.test
```

### 2. Wildcard Capture (Backreference)

Command:

```
veyra add *.zahra.syntopia.dev *.zahra.test
```

Arti:

```
$1.zahra.syntopia.dev
          |
          v
$1.zahra.test
```

Contoh:

```
api.zahra.syntopia.dev
          |
          v
api.zahra.test
```

```
admin.zahra.syntopia.dev
          |
          v
admin.zahra.test
```

### 3. Wildcard Flatten (Static Target)

Command:

```
veyra add *.zahra.syntopia.dev plane.zahra.test
```

Artinya:

Semua subdomain diarahkan ke satu host.

```
api.zahra.syntopia.dev
              |
              v
        plane.zahra.test
```

```
anything.zahra.syntopia.dev
              |
              v
        plane.zahra.test
```

## Lifecycle `add`

Ketika user menjalankan:

```
veyra add zahra.syntopia.dev studio.makeupartis.test
```

Flow:

```
CLI
 |
 | validate rule
 |
 v
BoltDB
 |
 | save rule
 |
 v
Generate Caddyfile
 |
 v
Reload Embedded Caddy
 |
 v
Rule aktif tanpa restart
```

## Caddy sebagai Data Plane

Caddy bertanggung jawab untuk:

### HTTP Server

- HTTP/1.1
- HTTP/2
- HTTP/3 (opsional)

### HTTPs

Auto:

- TLS certificate
- certificate renewal
- challenge handling

### Reverse Proxy

Mengubah request (misalnya):

```
Incoming:
Host: zahra.syntopia.dev
```

Menjadi:

```
Upstream:
Host: studio.makeupartis.test
```

### Response Rewriting

Mengubah (misalnya):

```
Location:
https://studio.makeupartis.test/login
```

Menjadi:

```
Location:
https://zahra.syntopia.dev/login
```

## Embedded Caddy

Veyra bukan menjalankan Caddy external.

Binary Veyra berisi:

```
+----------------+
|     Veyra      |
|                |
| CLI            |
| BoltDB         |
| Rule Engine    |
| Config Builder |
| Caddy Runtime  |
+----------------+
```

Sehingga user hanya memiliki:

```
veyra
```

Tidak perlu:

- Install Caddy
- Menjalankan Docker Compose
- Mengelola Caddyfile secara manual

## Docker Ready

Walaupun single binary, Veyra tetap dapat dijalankan dalam container.

Contoh:

```
Docker
 |
 | run veyra
 |
 +-- veyra binary
       |
       +-- embedded Caddy
       |
       +-- veyra.db
```

Volume dapat dipakai untuk:

```
/data/veyra.db
/data/Caddyfile
/data/caddy
```

## Daftar Command

### `add`

Menambahkan rule baru.

```
veyra add <public> <internal>
```

Contoh:

```
veyra add zahra.syntopia.dev studio.makeupartis.test
```

### `remove`

Menghapus rule.

```
veyra remove <public>
```

Flow:

```
delete BoltDB
       |
regenerate Caddy config
       |
reload Caddy
```

### `list`

Melihat seluruh rule.

Contoh output:

```
zahra.syntopia.dev       -> studio.makeupartis.test
*.zahra.syntopia.dev     -> *.zahra.test
*.demo.syntopia.dev      -> plane.zahra.test
```

### `get`

Melihat detail satu rule.

```
veyra get zahra.syntopia.dev
```

Output:

```
Type: static
Public:
  zahra.syntopia.dev

Internal:
  studio.makeupartis.test
```

### `start`

Menjalankan Veyra daemon.

Fungsi:

- load BoltDB
- generate Caddy config awal
- start embedded Caddy
- membuka API kontrol (opsional)

Contoh:

```
veyra start
```

### `reload`

Reload manual konfigurasi Caddy.

Biasanya tidak diperlukan karena add/remove otomatis reload.

### `validate`

Mengecek seluruh rule:

- format host
- konflik wildcard
- konflik static rule
- validitas Caddy config

### `export`

Export konfigurasi:

- Caddyfile
- JSON
- backup rule

Contoh:

```
veyra export caddy
```

## Struktur Internal (Arah Pengembangan)

```
veyra/
├── cmd
│   ├── add.go
│   ├── remove.go
│   ├── list.go
│   ├── start.go
│   └── root.go
│
├── core
│   ├── rules.go
│   ├── matcher.go
│   ├── validator.go
│   └── generator.go
│
├── caddy
│   ├── runtime.go
│   ├── config.go
│   └── reload.go
│
├── model
│   └── rule.go
│
├── store
│   └── bolt.go
│
├── main.go
│
└── veyra.db
```

## Visi Akhir

Veyra adalah:

> Single-binary local ingress controller berbasis Caddy yang menyediakan rule-based domain translation, request/response header rewriting, dan dynamic reverse proxy management untuk environment development maupun self-hosted.

## Prinsip Desain Utama

- Single binary
- Zero dependency
- Docker-ready
- Embedded Caddy
- BoltDB sebagai source of truth
- Caddy sebagai data plane
- Veyra sebagai control plane
- Rule-based routing
- Wildcard capture (\* → $1)
- Wildcard flatten (\* → static)
- Auto TLS / HTTP/2 / HTTP/3
- Hot reload tanpa downtime
