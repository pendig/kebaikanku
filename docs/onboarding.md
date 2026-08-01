# Developer Onboarding & Local Runbook

This guide covers local environment setup, local staging validation with PostgreSQL, and mock API contract test fixtures for new contributors (Issues #1, #2, #3).

---

## 🚀 1. Developer Onboarding Checklist

Follow this checklist to get the project running locally.

### Prerequisites
- **Go:** version 1.21 or newer
- **Node.js:** version 18 or newer (with npm)
- **Docker:** for local PostgreSQL testing

### Local Runbook

#### A. Backend Setup
1. Navigate to the backend directory:
   ```bash
   cd backend
   ```
2. Copy the example environment file:
   ```bash
   cp .env.example .env
   ```
3. Run the Go backend server:
   ```bash
   go run cmd/api/main.go
   ```
4. Verify the server is running by hitting the healthcheck endpoint:
   ```bash
   curl http://localhost:8080/health
   ```
   *Expected Response:*
   ```json
   {"status":"healthy","time":"2026-06-09T23:44:00+07:00"}
   ```

#### B. Landing Page Setup
1. Navigate to the landing page directory:
   ```bash
   cd frontend/landing
   ```
2. Install Node dependencies:
   ```bash
   npm install
   ```
3. Run the local Vite dev server:
   ```bash
   npm run dev
   ```
4. Open your browser and navigate to `http://localhost:8383`.

### Known Limitations (Local SQLite)
- By default, the backend runs on **SQLite** (`DB_DRIVER=sqlite` and `DB_DSN=kebaikanku.db`).
- **Concurrency:** SQLite does not support high-concurrency writes. Do not use SQLite for load testing or production callbacks.
- **Database Lock:** If you see `database is locked` errors during test runs, it means multiple transactions are attempting to write to the SQLite file simultaneously. Migrate to PostgreSQL for concurrent workflows.

---

## 🗄️ 2. Local PostgreSQL & Staging Setup

For production-like local staging validation, you should run a local PostgreSQL instance.

### A. Run PostgreSQL Container
A pre-configured PostgreSQL service is available in the root directory's `docker-compose.yml`. Start it via:
```bash
# From the project root directory
docker compose up -d postgres
```

### B. Configure Environment Variables
Edit your `backend/.env` file to point to the local PostgreSQL instance:
```env
PORT=8080
APP_ENV=development

# Database Driver Selection: "postgres"
DB_DRIVER=postgres
DB_DSN="host=localhost user=kebaikanku password=kebaikanku dbname=kebaikanku port=5432 sslmode=disable TimeZone=Asia/Jakarta"

# Single-admin dashboard login
ADMIN_PASSWORD=replace-with-a-long-admin-password
ADMIN_SESSION_SECRET=replace-with-at-least-32-random-characters

# Midtrans Configuration (Sandbox)
MIDTRANS_ENV=sandbox
MIDTRANS_SERVER_KEY=SB-Mid-server-placeholder-key
MIDTRANS_CLIENT_KEY=SB-Mid-client-placeholder-key
```

### C. Healthcheck & Database Migration Verification
When you run `go run cmd/api/main.go` with `DB_DRIVER=postgres`, the console output must log:
```text
Database connected successfully using driver: postgres
Database schema migrated successfully.
Backend API server is running on http://localhost:8080 in development mode
```

---

## 📄 3. API Contract Test Fixtures

Use these mock JSON payloads to test backend API endpoints.

### A. Admin authentication

#### Login administrator (`POST /api/v1/admin/login`)
- **Request Body:**
  ```json
  {
    "password": "value-from-ADMIN_PASSWORD"
  }
  ```
- **Response:** `200 OK` plus a signed `HttpOnly` session cookie. Send subsequent dashboard requests with credentials enabled.
- Use `GET /api/v1/admin/session` to check the session and `POST /api/v1/admin/logout` to clear it.

---

### B. Campaigns

#### Create Campaign (`POST /api/v1/campaigns` - Authenticated)
- **Authentication:** admin session cookie
- **Request Body:**
  ```json
  {
    "title": "Bantuan Logistik Gempa Bumi Cianjur",
    "slug": "bantuan-logistik-gempa-cianjur",
    "description": "Penggalangan dana darurat untuk penyediaan tenda, makanan bayi, dan air bersih bagi pengungsi gempa.",
    "category": "kemanusiaan",
    "target_amount": 150000000,
    "end_date": "2026-08-31T23:59:59+07:00"
  }
  ```
- **Response Body (Success 201):**
  ```json
  {
    "success": true,
    "data": {
      "id": "d0e120bc-f89a-4c34-912b-368ab24de12c",
      "organization_id": "7ca6410c-60ff-4034-8d4e-d01dfc6a382e",
      "title": "Bantuan Logistik Gempa Bumi Cianjur",
      "slug": "bantuan-logistik-gempa-cianjur",
      "category": "kemanusiaan",
      "target_amount": 150000000,
      "collected_amount": 0,
      "status": "active"
    }
  }
  ```

---

### C. Donations & Payments

#### Create Donation (`POST /api/v1/donations`)
- **Request Body:**
  ```json
  {
    "campaign_id": "d0e120bc-f89a-4c34-912b-368ab24de12c",
    "donor": {
      "name": "Ahmad Budi",
      "phone_number": "+6281234567890",
      "email": "ahmad.budi@gmail.com"
    },
    "amount": 250000,
    "platform_tip": 5000,
    "payment_method": "midtrans_snap"
  }
  ```
- **Response Body (Success 201):**
  ```json
  {
    "success": true,
    "data": {
      "donation_id": "8f89c67a-12bc-401d-9e12-3a8bc02d41ab",
      "status": "pending",
      "payment": {
        "provider": "midtrans",
        "snap_token": "82a8f89b-9c78-40ef-8ab9-c0c8dfc48b2a",
        "redirect_url": "https://app.sandbox.midtrans.com/snap/v2/vtweb/82a8f89b-9c78-40ef-8ab9-c0c8dfc48b2a"
      }
    }
  }
  ```

#### Midtrans Callback (`POST /api/v1/payments/midtrans/notification`)
- **Request Body (from Midtrans webhook):**
  ```json
  {
    "transaction_time": "2026-06-09 23:45:00",
    "transaction_status": "settlement",
    "status_message": "midtrans payment notification",
    "status_code": "200",
    "signature_key": "40ef8ab9c0c8dfc48b2a51f28b7e2832a8f89b9c78c0c8dfc48b2a51f28b7e28...",
    "payment_type": "qris",
    "order_id": "8f89c67a-12bc-401d-9e12-3a8bc02d41ab",
    "gross_amount": "255000.00",
    "currency": "IDR"
  }
  ```
- **Response Body (Success 200):**
  ```json
  {
    "success": true,
    "message": "notification processed successfully"
  }
  ```
