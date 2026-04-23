# Fragrance App

Aplikasi katalog dan checkout parfum berbasis **Flutter** dengan backend **Go (Gin)** dan autentikasi **Firebase**. Dibangun sebagai proyek Praktikum UTS dengan menerapkan Clean Architecture dan Provider sebagai state management.

> **Demo Video**: [https://youtu.be/rxKkao6aBZg](https://youtu.be/rxKkao6aBZg)

---

## Fitur Utama

| Fitur              | Keterangan                                  |
| ------------------ | ------------------------------------------- |
| Authentication     | Register & login via Firebase Auth          |
| Email Verification | Login diblokir sebelum email diverifikasi   |
| JWT Integration    | Firebase Token ditukar ke JWT backend       |
| Catalog Product    | List & detail produk dari backend API       |
| Cart               | Tambah, hapus item, hitung total harga      |
| Checkout           | Simulasi checkout & riwayat order           |
| State Management   | Provider + ChangeNotifier + notifyListeners |
| Clean Architecture | Pemisahan core, features, domain, usecase   |

---

## 🗂️ Struktur Proyek

```
fragrance-app/                    # ← Root = Go Backend
│
├── config/
│   ├── db.go                          # PostgreSQL connection (GORM)
│   ├── firebase.go                    # Firebase Admin SDK init
│   └── jwt.go                         # JWT generate & verify
│
├── internal/
│   ├── domain/                        # Entities & repository interfaces
│   │   ├── user.go
│   │   ├── product.go
│   │   ├── cart.go / cart_item.go
│   │   └── order.go / repository.go
│   ├── repository/                    # Data access layer (PostgreSQL)
│   │   ├── user_repo.go
│   │   ├── product_repo.go
│   │   ├── cart_repo.go
│   │   └── order_repo.go
│   ├── usecase/                       # Business logic
│   │   ├── auth_usecase.go
│   │   ├── product_usecase.go
│   │   ├── cart_usecase.go
│   │   └── order_usecase.go
│   ├── delivery/http/                 # HTTP handlers (Gin)
│   │   ├── handler.go
│   │   ├── auth_handler.go
│   │   ├── product_handler.go
│   │   ├── cart_handler.go
│   │   └── order_handler.go
│   └── middleware/
│       └── jwt.go                     # JWT auth middleware
│
├── main.go                            # Entry point + Gin router + Ngrok
├── go.mod
├── go.sum
├── firebase.json                      # Firebase service account key
├── .env                               # Environment variables
│
└── frontend/                     # ← Sub-folder Flutter
    ├── lib/
    │   ├── core/
    │   │   ├── api/
    │   │   │   └── api_client.dart        # Dio + JWT interceptor
    │   │   ├── constant/
    │   │   │   └── app_constant.dart      # Base URL & endpoint
    │   │   └── utils/
    │   │       └── currency_formatter.dart
    │   │
    │   ├── features/
    │   │   ├── auth/
    │   │   │   ├── providers/auth_provider.dart
    │   │   │   └── screens/
    │   │   │       ├── login_screen.dart
    │   │   │       └── register_screen.dart
    │   │   ├── products/
    │   │   │   ├── providers/product_provider.dart
    │   │   │   └── screens/
    │   │   │       ├── product_list_screen.dart
    │   │   │       └── product_detail_screen.dart
    │   │   ├── cart/
    │   │   │   ├── providers/cart_provider.dart
    │   │   │   └── screens/cart_screen.dart
    │   │   └── orders/
    │   │       ├── providers/order_provider.dart
    │   │       └── screens/
    │   │           ├── orders_screen.dart
    │   │           └── order_detail_screen.dart
    │   │
    │   ├── shared/
    │   │   └── models/models.dart         # ProductModel, CartItemModel, OrderModel
    │   │
    │   ├── firebase_options.dart
    │   └── main.dart                      # MultiProvider setup + AuthGate
    │
    └── pubspec.yaml
```

---

## Alur API

```
[Flutter]                    [Go Backend]              [Firebase / DB]
   │                              │                          │
   │── Register ──────────────────►                          │
   │   (email, password)          │── createUser ───────────►│ Firebase Auth
   │                              │◄── ID Token ─────────────│
   │   sendEmailVerification ─────────────────────────────── │
   │                              │── save user ────────────►│ PostgreSQL
   │                              │                          │
   │── Login ───────────────────► │                          │
   │   (ID Token)                 │── verifyIDToken ────────►│ Firebase Auth
   │                              │── check emailVerified    │
   │                              │── findByUID ────────────►│ PostgreSQL
   │◄── JWT (access_token) ───────│                          │
   │                              │                          │
   │── GET /api/products ────────►│ [JWT Middleware]         │
   │   Bearer: JWT                │── query products ───────►│ PostgreSQL
   │◄── [ ] ProductList ──────────│                          │
   │                              │                          │
   │── POST /api/cart ───────────►│ [JWT Middleware]         │
   │── POST /api/orders/checkout ►│── create order ─────────►│ PostgreSQL
   │◄── Order created ────────────│                          │
```

---

## Tech Stack

### Frontend (Flutter)

| Package                           | Kegunaan                              |
| --------------------------------- | ------------------------------------- |
| `firebase_core` & `firebase_auth` | Autentikasi Firebase                  |
| `provider`                        | State management                      |
| `dio`                             | HTTP client dengan interceptor        |
| `flutter_secure_storage`          | Simpan JWT secara aman                |
| `google_fonts`                    | Tipografi (Cormorant Garamond + Jost) |

### Backend (Go)

| Package                     | Kegunaan                        |
| --------------------------- | ------------------------------- |
| `gin-gonic/gin`             | HTTP router & framework         |
| `firebase.google.com/go`    | Firebase Admin SDK              |
| `golang-jwt/jwt/v5`         | Generate & verify JWT           |
| `gorm.io/gorm`              | ORM untuk PostgreSQL            |
| `joho/godotenv`             | Load environment variables      |
| `golang.ngrok.com/ngrok/v2` | Public tunnel untuk development |

---

## Setup & Menjalankan

### Prasyarat

- Flutter SDK ≥ 3.x
- Go ≥ 1.21
- PostgreSQL
- Firebase project (dengan Email/Password provider aktif)
- Ngrok account (untuk backend tunnel)

### Backend

```bash
# 1. Clone repo dan masuk ke root project
cd fragrance-app

# 2. Salin environment file
cp .env.example .env
# Isi: DB_DSN, JWT_SECRET, NGROK_AUTHTOKEN

# 3. Letakkan service account Firebase
# Rename file JSON credential ke: firebase.json

# 4. Jalankan backend
go run main.go

# Output:
#  NGROK_AUTHTOKEN ditemukan: xxxxxxxx...
#  Public URL: https://xxxx.ngrok-free.app
```

### Frontend (Flutter)

```bash
# 1. Masuk ke folder Flutter
cd frontend

# 2. Update base URL di:
# frontend/lib/core/constant/app_constant.dart
# static const String baseUrl = "https://xxxx.ngrok-free.app";

# 3. Install dependencies
flutter pub get

# 4. Jalankan aplikasi
flutter run
```

### Environment Variables (`.env`)

```env
DB_DSN=host=localhost user=postgres password=secret dbname=fragrance port=5432 sslmode=disable
JWT_SECRET=your-super-secret-key
NGROK_AUTHTOKEN=your-ngrok-token
```

---

## State Management

Proyek ini menggunakan **Provider** sebagai state management sesuai requirement. Setiap fitur memiliki provider sendiri yang extends `ChangeNotifier`.

```dart
// main.dart — semua provider didaftarkan di root
MultiProvider(
  providers: [
    ChangeNotifierProvider(create: (_) => AuthProvider()..checkAuth()),
    ChangeNotifierProvider(create: (_) => ProductProvider()),
    ChangeNotifierProvider(create: (_) => CartProvider()),
    ChangeNotifierProvider(create: (_) => OrderProvider()),
  ],
  ...
)
```

Pattern yang digunakan konsisten di seluruh provider:

```dart
class CartProvider extends ChangeNotifier {
  bool _isLoading = false;
  List<CartItemModel> _items = [];

  // Computed getter — otomatis recalculate
  double get totalPrice =>
    _items.fold(0, (sum, item) => sum + (item.price * item.quantity));

  Future<void> fetchCart() async {
    _isLoading = true;
    notifyListeners(); // ← UI rebuild: tampilkan loading

    // ... fetch data dari API ...

    _isLoading = false;
    notifyListeners(); // ← UI rebuild: tampilkan data
  }
}
```

---

## Authentication Flow

```
Register:
  Flutter → Firebase.createUser → sendEmailVerification
         → kirim ID Token ke backend → backend simpan user ke DB

Login:
  Flutter → Firebase.signIn → cek emailVerified (WAJIB)
         → kirim ID Token ke backend → backend verifyIDToken
         → backend generate JWT → Flutter simpan JWT (SecureStorage)

Protected API:
  Flutter → Dio interceptor otomatis attach "Bearer JWT"
         → backend JWT middleware verify → proses request
```

> **Catatan**: User yang belum memverifikasi email akan langsung di-`signOut` dan mendapatkan pesan error _"Please verify your email first"_ tanpa bisa masuk ke aplikasi.

---

## API Endpoints

### Public

| Method | Endpoint         | Keterangan           |
| ------ | ---------------- | -------------------- |
| `POST` | `/auth/register` | Register user baru   |
| `POST` | `/auth/login`    | Login & dapatkan JWT |

### Protected (butuh `Authorization: Bearer <jwt>`)

| Method   | Endpoint               | Keterangan           |
| -------- | ---------------------- | -------------------- |
| `GET`    | `/api/products`        | List semua produk    |
| `GET`    | `/api/products/:id`    | Detail produk        |
| `GET`    | `/api/cart`            | Lihat isi cart       |
| `POST`   | `/api/cart`            | Tambah item ke cart  |
| `DELETE` | `/api/cart/:id`        | Hapus item dari cart |
| `DELETE` | `/api/cart`            | Kosongkan cart       |
| `POST`   | `/api/orders/checkout` | Proses checkout      |
| `GET`    | `/api/orders`          | Riwayat order        |
| `GET`    | `/api/orders/:id`      | Detail order         |

---

## Demo

Tonton demo lengkap aplikasi di YouTube:

[![Fragrance App Demo](https://img.shields.io/badge/YouTube-Demo%20Video-red?style=for-the-badge&logo=youtube)](https://youtu.be/rxKkao6aBZg)

**[▶ https://youtu.be/rxKkao6aBZg](https://youtu.be/rxKkao6aBZg)**

---

## Lisensi

Proyek ini dibuat untuk keperluan akademik — Praktikum UTS.
