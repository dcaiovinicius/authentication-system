# Authentication System API

RESTful API for user authentication with JWT and refresh tokens.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Configuration](#configuration)
- [Environment Variables](#environment-variables)
- [Running the API](#running-the-api)
- [Endpoints](#endpoints)
- [CURL Examples](#curl-examples)
- [Authentication Flow](#authentication-flow)
- [Status Codes](#status-codes)

---

## Prerequisites

- Go 1.21+
- PostgreSQL
- Docker (optional)

---

## Configuration

### 1. Clone the repository

```bash
git clone <repo-url>
cd authentication-system
```

### 2. Configure environment variables

Create a `.env` file in the project root:

```bash
# Database
DatabaseURL=postgres://user:password@localhost:5432/authdb?sslmode=disable

# JWT (generate a secure key: openssl rand -base64 32)
JWT_SECRET=your-secret-key-minimum-32-characters-here
JWT_ISSUER=authentication-system

# Server
DefaultPort=:8080
```

### 3. Run the migrations

```bash
# Create users table
CREATE TABLE users (
    id UUID PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    username VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

-- Create refresh tokens table
CREATE TABLE refresh_tokens (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES users(id),
    token VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL,
    revoked BOOLEAN DEFAULT false
);
```

### 4. Start the database (Docker)

```bash
make up
```

---

## Environment Variables

| Variable | Description | Required |
|----------|-------------|----------|
| `DatabaseURL` | PostgreSQL connection string | Yes |
| `JWT_SECRET` | Secret key for signing JWT tokens (min 32 chars) | Yes |
| `JWTIssuer` | JWT token issuer | No (default: `authentication-system`) |
| `DefaultPort` | Server port | No (default: `:8080`) |

---

## Running the API

```bash
# Development
go run cmd/server/main.go

# or using Makefile
make run
```

The server will start at `http://localhost:8080`.

---

## Endpoints

| Method | Endpoint | Authentication | Description |
|--------|----------|----------------|-------------|
| `POST` | `/api/v1/register` | No | Create new user |
| `POST` | `/api/v1/login` | No | Authenticate user |
| `POST` | `/api/v1/refresh` | No | Renew access token |
| `POST` | `/api/v1/logout` | Yes | Revoke refresh tokens |
| `GET` | `/api/v1/users` | Yes | Get current user data |

---

## CURL Examples

### 1. Register new user

```bash
curl -X POST http://localhost:8080/api/v1/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "username": "myuser",
    "password": "password123456"
  }'
```

**Response:**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "a1b2c3d4e5f6..."
}
```

> **Note:** Password must be at least 8 characters.

---

### 2. Login (Authenticate)

```bash
curl -X POST http://localhost:8080/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123456"
  }'
```

**Response:**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "a1b2c3d4e5f6..."
}
```

---

### 3. Get current user (Protected)

```bash
curl -X GET http://localhost:8080/api/v1/user \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN_HERE"
```

**Response:**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "user@example.com",
  "username": "myuser"
}
```

---

### 4. Renew access token

When the access token expires (24 hours), use the refresh token to get a new pair of tokens:

```bash
curl -X POST http://localhost:8080/api/v1/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "YOUR_REFRESH_TOKEN_HERE"
  }'
```

**Response:**
```json
{
  "access_token": "new-access-token...",
  "refresh_token": "new-refresh-token..."
}
```

> **Note:** The previous refresh token is automatically revoked.

---

### 5. Logout

```bash
curl -X POST http://localhost:8080/api/v1/logout \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN_HERE"
```

**Response:**
- Status `204 No Content` on success.

---

### Token Details

| Token | Duration | Usage |
|-------|----------|-------|
| `access_token` | 24 hours | Authenticate protected requests |
| `refresh_token` | 7 days | Renew access token |

---

## Status Codes

| Code | Description |
|------|-------------|
| `200` | Success |
| `201` | Created successfully |
| `204` | No content (success) |
| `400` | Bad request |
| `401` | Unauthorized (invalid/expired token) |
| `404` | Not found |
| `500` | Internal server error |

