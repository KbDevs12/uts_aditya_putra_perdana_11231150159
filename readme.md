# UTS Mobile Semester 6 Backend & E-Commerce App

Repo ini isinya backend buat dua aplikasi mobile: e-commerce (parfum) dan Kantongin (e-wallet). Backend-nya satu, dipakai bareng-bareng sama dua app itu lewat Firebase project dan database yang sama. Di sini yang ada cuma backend Go-nya sama frontend e-commerce-nya, Silahkan cek [Repository Kantongin](https://github.com/KbDevs12/wallet-nih) untuk bagian kantongin.

Kalau bingung kenapa ada logic wallet/topup/transfer di backend padahal ini "repo e-commerce", jawabannya karena checkout di app e-commerce bisa dibayar pakai saldo Kantongin. Jadi pas checkout, backend bikin payment intent, terus e-commerce app buka Kantongin lewat deep link buat bayar di sana.

## Struktur folder

```
.
├── config/            # koneksi DB, redis, firebase admin, jwt, smtp
├── internal/
│   ├── domain/        # entity + interface repository
│   ├── repository/    # implementasi query ke postgres (gorm)
│   ├── usecase/       # business logic
│   ├── delivery/http/ # handler gin
│   └── middleware/     # jwt auth middleware
├── main.go            # entrypoint, routing
└── frontend/           # app Flutter e-commerce (parfum)
```

Backend-nya pakai Clean Architecture ala-ala: domain → repository → usecase → delivery. Bukan yang strict banget, disesuaikan sama kebutuhan tugas aja.

## Tech stack

Backend:

- Go 1.25 + Gin
- GORM + PostgreSQL
- Firebase Admin SDK (verifikasi token auth)
- Redis Cloud (nyimpen OTP sementara kalau `REDIS_ADDR` kosong, OTP-nya cuma keluar di log, jadi tetap bisa jalan tanpa redis pas development)
- JWT buat session setelah login

Frontend (`/frontend`):

- Flutter, Provider buat state management
- Firebase Auth
- Dio buat HTTP client, disimpen tokennya pakai flutter_secure_storage
- QR & url_launcher (buat handle deep link ke Kantongin pas checkout)

## Cara jalanin

### 1. Backend

Butuh Go, PostgreSQL, sama service account Firebase.

```bash
go mod download
```

Bikin file `.env` di root:

```env
PORT=8080
DB_DSN=host=localhost user=postgres password=postgres dbname=uts_mobile port=5432 sslmode=disable
JWT_SECRET=ganti-ini-dengan-string-random

# firebase, download service account json dari firebase console
FIREBASE_CREDENTIALS=firebase.json

# redis pakai Redis Cloud, bukan redis lokal
# ambil host:port sama password-nya dari dashboard Redis Cloud
# kalau REDIS_ADDR dikosongin, OTP cuma ditulis ke log (tetep bisa jalan tanpa redis, tapi buat production ini wajib diisi)
REDIS_ADDR=redis-xxxxx.c000.us-east-1-2.ec2.redns.redis-cloud.com:xxxxx
REDIS_USERNAME=default
REDIS_PASSWORD=
REDIS_TLS=true

# smtp opsional buat kirim OTP beneran, kalau kosong OTP muncul di terminal aja
SMTP_HOST=
SMTP_PORT=
SMTP_USERNAME=
SMTP_PASSWORD=
SMTP_FROM=
```

Taruh file service account Firebase sebagai `firebase.json` di root (jangan di-commit, sudah ada di `.gitignore`).

Migrasi tabel jalan otomatis pas start (pakai `AutoMigrate` dari GORM), jadi tinggal:

```bash
go run main.go
```

Server nyala di `http://localhost:8080`. Cek `/health` buat mastiin DB dan redis-nya nyambung.

### 2. Frontend (e-commerce)

```bash
cd frontend
flutter pub get
```

Base URL API-nya di-hardcode di `lib/core/constant/app_constant.dart` — kalau backend jalan lokal atau lewat ngrok, ganti dulu value `baseUrl`-nya sebelum build.

```bash
flutter run
```

## Endpoint

Public (nggak perlu token):

| Method | Path                          | Buat apa                                         |
| ------ | ----------------------------- | ------------------------------------------------ |
| POST   | `/auth/register`              | daftar pakai firebase token                      |
| POST   | `/auth/login`                 | login, balikin JWT                               |
| POST   | `/otp/send-email`             | kirim ulang OTP verifikasi email                 |
| POST   | `/auth/verify-email-otp`      | verifikasi OTP                                   |
| GET    | `/api/payment-intents/:token` | detail payment intent (dipanggil dari Kantongin) |

Sisanya butuh header `Authorization: Bearer <jwt>`, prefix `/api`:

- `GET/POST /products`, `/products/:id` — katalog produk
- `GET/POST/DELETE /cart` — cart
- `POST /orders/checkout`, `GET /orders`, `GET /orders/:id`, `POST /orders/:id/payment-intent` — order & checkout
- `GET /wallet`, `POST /wallet/topup`, `POST /wallet/transfer`, `GET /wallet/transactions`, `POST /wallet/pin`, `POST /wallet/pin/verify`, `POST /payment-intents/:token/pay` — semua yang wallet related (dipakai Kantongin, bukan e-commerce)
- `POST /auth/setup-2fa`, `POST /auth/verify-2fa`, `POST /auth/notification-token` — 2FA & push token

## Hal yang perlu diinget

- Satu backend dipakai dua app, jadi kalau ubah struktur response di `domain/` atau `usecase/`, cek dulu apa Kantongin ikut kepakai field itu juga.
- OTP email dibedain "brand"-nya berdasarkan field `app` yang dikirim pas register (`kantongin` atau `ecommerce`), itu yang nentuin nama pengirim di email OTP-nya beda.
- Nggak ada test otomatis buat sekarang, jadi tiap ubah endpoint mending dicek manual pakai Postman/Thunder Client dulu sebelum push.
