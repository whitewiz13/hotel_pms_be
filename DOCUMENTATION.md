# Hotel PMS Backend — Documentation

> **Tech Stack:** Go 1.26 · Gin · GORM · PostgreSQL · JWT (HS256) · bcrypt

---

## Table of Contents

- [Architecture Overview](#architecture-overview)
- [Project Structure](#project-structure)
- [Application Startup Flow](#application-startup-flow)
- [Configuration & Environment](#configuration--environment)
- [Database Layer](#database-layer)
  - [Connection](#connection)
  - [Migrations](#migrations)
  - [Schema & Tables](#schema--tables)
  - [Relationships](#relationships)
  - [Seeding](#seeding)
- [Authentication & Authorization](#authentication--authorization)
  - [JWT Tokens](#jwt-tokens)
  - [Roles & Permissions](#roles--permissions)
  - [Auth Middleware Pipeline](#auth-middleware-pipeline)
  - [Staff Login Flow](#staff-login-flow)
  - [Guest Login Flow](#guest-login-flow)
- [API Routes](#api-routes)
  - [Public Routes](#public-routes)
  - [Protected Routes](#protected-routes)
  - [Full Route Table](#full-route-table)
- [Request / Response Format](#request--response-format)
  - [Standard Response](#standard-response)
  - [Paginated Response](#paginated-response)
  - [Error Response](#error-response)
- [Domain Flows](#domain-flows)
  - [Hotel Management Flow](#hotel-management-flow)
  - [Room Management Flow](#room-management-flow)
  - [Room PIN / Guest Access Flow](#room-pin--guest-access-flow)
  - [Amenity Management Flow](#amenity-management-flow)
  - [Staff Creation Flow](#staff-creation-flow)
- [Layer Interaction Pattern](#layer-interaction-pattern)
- [Models & DTOs Reference](#models--dtos-reference)

---

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                          CLIENT                                 │
│                  (Frontend @ localhost:5173)                     │
└───────────────────────────┬─────────────────────────────────────┘
                            │  HTTP + JSON
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│                     GIN HTTP SERVER                             │
│                                                                 │
│  ┌──────────┐  ┌────────────────────┐  ┌────────────────────┐  │
│  │   CORS   │→ │  Auth Middleware    │→ │  Role Middleware    │  │
│  └──────────┘  └────────────────────┘  └────────────────────┘  │
│                            │                                    │
│                            ▼                                    │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │                      HANDLERS                             │  │
│  │  AuthHandler · HotelHandler · RoomHandler · AmenityHandler│  │
│  └──────────────────────────┬───────────────────────────────┘  │
│                              │                                  │
│                              ▼                                  │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │                      SERVICES                             │  │
│  │  AuthService · HotelService · RoomService · AmenityService│  │
│  └──────────────────────────┬───────────────────────────────┘  │
│                              │                                  │
│                              ▼                                  │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │                    REPOSITORIES                           │  │
│  │  UserRepo · HotelRepo · RoomRepo · AmenityRepo           │  │
│  └──────────────────────────┬───────────────────────────────┘  │
│                              │  GORM                            │
└──────────────────────────────┼──────────────────────────────────┘
                               ▼
                  ┌──────────────────────┐
                  │     PostgreSQL       │
                  │     (hotel_pms)      │
                  └──────────────────────┘
```

---

## Project Structure

```
backend/
├── main.go                  # Entry point — wires everything together
├── go.mod                   # Go module & dependencies
│
├── config/
│   └── config.go            # Loads .env, exposes Config struct
│
├── database/
│   ├── database.go          # PostgreSQL connection via GORM
│   └── migrate.go           # AutoMigrate + custom indexes
│
├── models/                  # GORM models (= DB tables)
│   ├── base.go              # BaseModel (UUID PK, timestamps, soft delete)
│   ├── user.go              # User (staff accounts)
│   ├── hotel.go             # Hotel
│   ├── room.go              # Room
│   └── amenity.go           # Amenity
│
├── dto/                     # Data Transfer Objects (request validation)
│   ├── auth_dto.go          # Login, GuestLogin, CreateStaff
│   ├── hotel_dto.go         # CreateHotel, UpdateHotel
│   ├── room_dto.go          # CreateRoom, UpdateRoom, SetRoomPin
│   └── amenity_dto.go       # CreateAmenity, UpdateAmenity
│
├── repository/              # Data access layer (DB queries)
│   ├── user_repo.go
│   ├── hotel_repo.go
│   ├── room_repo.go
│   └── amenity_repo.go
│
├── service/                 # Business logic
│   ├── auth_service.go      # Login, guest login, JWT, staff creation
│   ├── hotel_service.go     # Hotel CRUD
│   ├── room_service.go      # Room CRUD + PIN management
│   └── amenity_service.go   # Amenity CRUD
│
├── handler/                 # HTTP handlers (controllers)
│   ├── handler.go           # Package declaration
│   ├── auth_handler.go      # /auth endpoints
│   ├── hotel_handler.go     # /hotels endpoints
│   ├── room_handler.go      # /rooms endpoints
│   └── amenity_handler.go   # /amenities endpoints
│
├── middleware/
│   ├── auth.go              # JWT extraction & validation
│   └── role.go              # Role-based access control
│
├── router/
│   └── router.go            # Route definitions, CORS, middleware wiring
│
├── seeds/
│   └── seeder.go            # Default super_admin + sample amenities
│
├── utils/
│   ├── password.go          # bcrypt hash/check, PIN generation
│   └── response.go          # Standardized JSON response helpers
│
└── tmp/                     # Air hot-reload temp files (gitignored)
```

---

## Application Startup Flow

```
main.go
  │
  ├─ 1. config.Load()            Load .env → Config struct
  │
  ├─ 2. database.Connect()       Open PostgreSQL connection (GORM)
  │
  ├─ 3. database.RunMigrations() AutoMigrate models → create/update tables
  │     (if RUN_MIGRATIONS=true)  + create pgcrypto extension
  │                                + composite unique index on rooms
  │
  ├─ 4. seeds.Run()              Create default super_admin + amenities
  │     (if SEED_DB=true)         (idempotent — skips if data exists)
  │
  ├─ 5. Initialize Repositories  NewUserRepo, NewHotelRepo, NewRoomRepo, NewAmenityRepo
  │
  ├─ 6. Initialize Services      NewAuthService, NewHotelService, NewRoomService, NewAmenityService
  │
  ├─ 7. Initialize Handlers      NewAuthHandler, NewHotelHandler, NewRoomHandler, NewAmenityHandler
  │
  ├─ 8. router.Setup()           Register routes, middleware, CORS
  │
  └─ 9. r.Run(":8080")           Start HTTP server
```

---

## Configuration & Environment

All config is loaded from environment variables (or `.env` file via `godotenv`).

| Variable | Default | Description |
|---|---|---|
| `SERVER_PORT` | `8080` | HTTP listen port |
| `GIN_MODE` | `debug` | Gin mode (`debug` / `release` / `test`) |
| `DB_HOST` | `localhost` | PostgreSQL host |
| `DB_PORT` | `5432` | PostgreSQL port |
| `DB_USER` | `postgres` | Database user |
| `DB_PASSWORD` | `postgres` | Database password |
| `DB_NAME` | `hotel_pms` | Database name |
| `DB_SSLMODE` | `disable` | SSL mode |
| `DB_TIMEZONE` | `UTC` | Timezone |
| `JWT_SECRET` | _(required)_ | HMAC signing key for JWT tokens |
| `JWT_EXPIRY_HOURS` | `24` | Staff token lifetime (hours) |
| `JWT_GUEST_EXPIRY_HOURS` | `12` | Guest token lifetime (hours) |
| `RUN_MIGRATIONS` | `true` | Run AutoMigrate on startup |
| `SEED_DB` | `false` | Seed default data on startup |

---

## Database Layer

### Connection

- **Driver:** `gorm.io/driver/postgres`
- **DSN format:** `host=... user=... password=... dbname=... port=... sslmode=... TimeZone=...`
- **Pool:** 10 idle / 100 max open connections
- Extension created: `pgcrypto` (for `gen_random_uuid()`)

### Migrations

Migrations run via `db.AutoMigrate()` on startup when `RUN_MIGRATIONS=true`. GORM inspects the model structs and creates/alters tables to match.

After AutoMigrate, a manual index is created:

```sql
CREATE UNIQUE INDEX IF NOT EXISTS idx_rooms_hotel_room_number
  ON rooms(hotel_id, room_number)
  WHERE deleted_at IS NULL;
```

This ensures room numbers are unique **per hotel** (excluding soft-deleted rows).

### Schema & Tables

All tables share a common base:

| Column | Type | Notes |
|---|---|---|
| `id` | `uuid` | PK, default `gen_random_uuid()` |
| `created_at` | `timestamp` | Auto-set on create |
| `updated_at` | `timestamp` | Auto-set on update |
| `deleted_at` | `timestamp` | Soft delete (nullable, indexed) |

#### `users`

| Column | Type | Constraints | Description |
|---|---|---|---|
| `email` | `varchar(255)` | `UNIQUE NOT NULL` | Login email |
| `password_hash` | `text` | `NOT NULL` | bcrypt hash |
| `name` | `varchar(255)` | `NOT NULL` | Display name |
| `phone` | `varchar(20)` | | Phone number |
| `role` | `varchar(30)` | `NOT NULL DEFAULT 'staff'` | One of: `super_admin`, `manager`, `front_desk`, `housekeeping`, `staff` |
| `hotel_id` | `uuid` | `INDEX`, nullable | FK → `hotels.id` (NULL for super_admin) |
| `is_active` | `boolean` | `DEFAULT true` | Soft active flag |

#### `hotels`

| Column | Type | Constraints | Description |
|---|---|---|---|
| `name` | `varchar(255)` | `NOT NULL` | Hotel name |
| `address` | `varchar(500)` | `NOT NULL` | Street address |
| `city` | `varchar(100)` | `NOT NULL` | City |
| `state` | `varchar(100)` | | State/province |
| `country` | `varchar(100)` | `NOT NULL` | Country |
| `zip_code` | `varchar(20)` | | Postal code |
| `phone` | `varchar(20)` | | Contact phone |
| `email` | `varchar(255)` | | Contact email |
| `description` | `text` | | Description |
| `is_active` | `boolean` | `DEFAULT true` | Active flag |

#### `rooms`

| Column | Type | Constraints | Description |
|---|---|---|---|
| `hotel_id` | `uuid` | `NOT NULL INDEX` | FK → `hotels.id` |
| `room_number` | `varchar(20)` | `NOT NULL` | e.g. "101", "A-12" |
| `room_type` | `varchar(20)` | `NOT NULL DEFAULT 'single'` | `single` / `double` / `suite` / `deluxe` / `penthouse` |
| `floor` | `int` | `NOT NULL DEFAULT 1` | Floor number |
| `status` | `varchar(20)` | `NOT NULL DEFAULT 'available'` | `available` / `occupied` / `maintenance` / `cleaning` |
| `price_per_night` | `float` | `NOT NULL DEFAULT 0` | Nightly rate |
| `description` | `text` | | Room description |
| `max_occupancy` | `int` | `NOT NULL DEFAULT 2` | Max guests |
| `access_pin` | `varchar(6)` | | 6-digit guest access PIN |
| `is_active` | `boolean` | `DEFAULT true` | Active flag |

**Unique constraint:** `(hotel_id, room_number)` where `deleted_at IS NULL`

#### `amenities`

| Column | Type | Constraints | Description |
|---|---|---|---|
| `name` | `varchar(255)` | `NOT NULL` | Amenity name |
| `description` | `text` | | Description |
| `icon` | `varchar(100)` | | Icon identifier |
| `category` | `varchar(30)` | `NOT NULL DEFAULT 'general'` | `room` / `bathroom` / `general` / `dining` / `recreation` |
| `is_active` | `boolean` | `DEFAULT true` | Active flag |

#### `room_amenities` (join table)

| Column | Type | Description |
|---|---|---|
| `room_id` | `uuid` | FK → `rooms.id` |
| `amenity_id` | `uuid` | FK → `amenities.id` |

Auto-managed by GORM's `many2many:room_amenities` tag.

### Relationships

```
hotels 1──────┤< rooms
hotels 1──────┤< users (staff)
rooms  >┤─────┤< amenities     (via room_amenities join table)
```

- A **Hotel** has many **Rooms** (`foreignKey:HotelID`)
- A **Hotel** has many **Users** (staff) (`foreignKey:HotelID`)
- A **Room** belongs to a **Hotel**
- A **Room** has many **Amenities** (many-to-many via `room_amenities`)
- An **Amenity** has many **Rooms** (inverse many-to-many)

### Seeding

When `SEED_DB=true`, two seeders run (idempotent — skip if data exists):

**1. Super Admin**

| Field | Value |
|---|---|
| Email | `admin@hotelpms.com` |
| Password | `admin@123` |
| Role | `super_admin` |
| Hotel | _(none)_ |

**2. Default Amenities** (15 items)

| Name | Category |
|---|---|
| WiFi | room |
| Air Conditioning | room |
| TV | room |
| Mini Bar | room |
| Safe | room |
| Bathtub | bathroom |
| Rain Shower | bathroom |
| Hair Dryer | bathroom |
| Swimming Pool | recreation |
| Gym | recreation |
| Spa | recreation |
| Restaurant | dining |
| Room Service | dining |
| Parking | general |
| Laundry | general |

---

## Authentication & Authorization

### JWT Tokens

Tokens are signed with **HS256** using the `JWT_SECRET` env variable.

**Staff token claims:**

| Claim | Type | Description |
|---|---|---|
| `user_id` | string | User UUID |
| `email` | string | User email |
| `role` | string | User role |
| `hotel_id` | string | Assigned hotel UUID |
| `is_guest` | bool | Always `false` |
| `exp` | timestamp | Expiry (default 24h) |
| `iat` | timestamp | Issued at |

**Guest token claims:**

| Claim | Type | Description |
|---|---|---|
| `role` | string | Always `"guest"` |
| `hotel_id` | string | Hotel UUID |
| `room_id` | string | Room UUID |
| `room_number` | string | Room number |
| `is_guest` | bool | Always `true` |
| `exp` | timestamp | Expiry (default 12h) |
| `iat` | timestamp | Issued at |

### Roles & Permissions

| Role | Stored in DB | Description |
|---|---|---|
| `super_admin` | Yes | Full system access, manages hotels |
| `manager` | Yes | Hotel-level management |
| `front_desk` | Yes | Check-in/out, room ops |
| `housekeeping` | Yes | Room status updates |
| `staff` | Yes | Basic staff access |
| `guest` | No (JWT only) | Room-scoped guest access via PIN |

**Permission hierarchy used by middleware:**

| Middleware Helper | Allowed Roles |
|---|---|
| `RequireSuperAdmin()` | `super_admin` |
| `RequireManagement()` | `super_admin`, `manager` |
| `RequireFrontDeskOrAbove()` | `super_admin`, `manager`, `front_desk` |
| `RequireAnyStaff()` | All 5 staff roles |
| `RequireAnyAuthenticated()` | All 5 staff roles + `guest` |

### Auth Middleware Pipeline

```
Request
  │
  ▼
┌──────────────────────────────────────────────┐
│ AuthMiddleware                                │
│  1. Extract "Authorization: Bearer <token>"  │
│  2. Validate JWT signature + expiry          │
│  3. Parse claims → store in gin.Context      │
│     (key: "claims")                          │
│  4. Abort 401 on failure                     │
└──────────────────────┬───────────────────────┘
                       │
                       ▼
┌──────────────────────────────────────────────┐
│ RequireRole(...)                              │
│  1. Read claims from context                 │
│  2. Check if user role ∈ allowed roles       │
│  3. Abort 403 on mismatch                   │
└──────────────────────┬───────────────────────┘
                       │
                       ▼
                   Handler
```

### Staff Login Flow

```
Client                          Server
  │                               │
  │  POST /api/auth/login         │
  │  { email, password }          │
  │──────────────────────────────▶│
  │                               │
  │                 AuthHandler.Login()
  │                    │
  │                    ▼
  │              AuthService.Login()
  │                    │
  │                    ├─ UserRepo.FindByEmail(email)
  │                    │     └─ SELECT * FROM users WHERE email = ?
  │                    │
  │                    ├─ Check user.IsActive
  │                    │
  │                    ├─ bcrypt.Compare(password, hash)
  │                    │
  │                    └─ Generate JWT (HS256, 24h expiry)
  │                               │
  │  { token, user }              │
  │◀──────────────────────────────│
```

### Guest Login Flow

```
Client                            Server
  │                                 │
  │  POST /api/auth/guest/login     │
  │  { room_number, pin, hotel_id } │
  │────────────────────────────────▶│
  │                                 │
  │                   AuthHandler.GuestLogin()
  │                      │
  │                      ▼
  │                AuthService.GuestLogin()
  │                      │
  │                      ├─ RoomRepo.FindByRoomNumberAndHotel(room_number, hotel_id)
  │                      │     └─ SELECT * FROM rooms WHERE room_number = ? AND hotel_id = ?
  │                      │
  │                      ├─ Verify room.AccessPin matches provided pin
  │                      │
  │                      ├─ Verify room.Status == "occupied"
  │                      │
  │                      └─ Generate guest JWT (HS256, 12h expiry)
  │                                 │
  │  { token, room_number,          │
  │    room_type }                  │
  │◀────────────────────────────────│
```

---

## API Routes

**Base URL:** `http://localhost:8080/api`

### Public Routes

| Method | Path | Handler | Description |
|---|---|---|---|
| `GET` | `/api/health` | inline | Health check |
| `POST` | `/api/auth/login` | `AuthHandler.Login` | Staff login |
| `POST` | `/api/auth/guest/login` | `AuthHandler.GuestLogin` | Guest login via room PIN |

### Protected Routes

All routes below require `Authorization: Bearer <token>` header.

#### Staff Management

| Method | Path | Role Required | Handler | Description |
|---|---|---|---|---|
| `POST` | `/api/staff` | Management | `AuthHandler.CreateStaff` | Create a staff account |

#### Hotels

| Method | Path | Role Required | Handler | Description |
|---|---|---|---|---|
| `POST` | `/api/hotels` | Super Admin | `HotelHandler.Create` | Create a hotel |
| `GET` | `/api/hotels` | Any Staff | `HotelHandler.GetAll` | List hotels (paginated) |
| `GET` | `/api/hotels/:id` | Any Staff | `HotelHandler.GetByID` | Get hotel by ID |
| `PUT` | `/api/hotels/:id` | Management | `HotelHandler.Update` | Update hotel |
| `DELETE` | `/api/hotels/:id` | Super Admin | `HotelHandler.Delete` | Soft delete hotel |
| `GET` | `/api/hotels/:hotelId/rooms` | Any Staff | `RoomHandler.GetByHotelID` | List rooms in a hotel |

#### Rooms

| Method | Path | Role Required | Handler | Description |
|---|---|---|---|---|
| `POST` | `/api/rooms` | Front Desk+ | `RoomHandler.Create` | Create a room |
| `GET` | `/api/rooms/:id` | Any Authenticated | `RoomHandler.GetByID` | Get room by ID |
| `PUT` | `/api/rooms/:id` | Front Desk+ | `RoomHandler.Update` | Update room |
| `DELETE` | `/api/rooms/:id` | Management | `RoomHandler.Delete` | Soft delete room |
| `POST` | `/api/rooms/:id/pin` | Front Desk+ | `RoomHandler.GeneratePin` | Generate 6-digit access PIN |
| `DELETE` | `/api/rooms/:id/pin` | Front Desk+ | `RoomHandler.ClearPin` | Clear room access PIN |

#### Amenities

| Method | Path | Role Required | Handler | Description |
|---|---|---|---|---|
| `POST` | `/api/amenities` | Management | `AmenityHandler.Create` | Create an amenity |
| `GET` | `/api/amenities` | Any Authenticated | `AmenityHandler.GetAll` | List amenities (paginated, filterable) |
| `GET` | `/api/amenities/:id` | Any Authenticated | `AmenityHandler.GetByID` | Get amenity by ID |
| `PUT` | `/api/amenities/:id` | Management | `AmenityHandler.Update` | Update amenity |
| `DELETE` | `/api/amenities/:id` | Super Admin | `AmenityHandler.Delete` | Soft delete amenity |

### Full Route Table

```
PUBLIC
  GET    /api/health
  POST   /api/auth/login
  POST   /api/auth/guest/login

PROTECTED (Bearer token required)
  POST   /api/staff                    [super_admin, manager]
  POST   /api/hotels                   [super_admin]
  GET    /api/hotels                   [any staff]
  GET    /api/hotels/:id               [any staff]
  PUT    /api/hotels/:id               [super_admin, manager]
  DELETE /api/hotels/:id               [super_admin]
  GET    /api/hotels/:hotelId/rooms    [any staff]
  POST   /api/rooms                    [super_admin, manager, front_desk]
  GET    /api/rooms/:id                [any staff + guest]
  PUT    /api/rooms/:id                [super_admin, manager, front_desk]
  DELETE /api/rooms/:id                [super_admin, manager]
  POST   /api/rooms/:id/pin           [super_admin, manager, front_desk]
  DELETE /api/rooms/:id/pin           [super_admin, manager, front_desk]
  POST   /api/amenities               [super_admin, manager]
  GET    /api/amenities               [any staff + guest]
  GET    /api/amenities/:id           [any staff + guest]
  PUT    /api/amenities/:id           [super_admin, manager]
  DELETE /api/amenities/:id           [super_admin]
```

---

## Request / Response Format

### Standard Response

```json
{
  "success": true,
  "data": { ... }
}
```

### Paginated Response

Returned by list endpoints. Query params: `?page=1&per_page=20`

```json
{
  "success": true,
  "data": [ ... ],
  "meta": {
    "page": 1,
    "per_page": 20,
    "total": 45,
    "total_pages": 3
  }
}
```

### Error Response

```json
{
  "success": false,
  "error": "description of what went wrong"
}
```

| HTTP Code | Meaning | Used When |
|---|---|---|
| `200` | OK | Successful read / update / delete |
| `201` | Created | Successful create |
| `400` | Bad Request | Validation error, business rule violation |
| `401` | Unauthorized | Missing / invalid / expired token |
| `403` | Forbidden | Valid token but insufficient role |
| `404` | Not Found | Entity not found |
| `500` | Internal Error | Unexpected server error |

---

## Domain Flows

### Hotel Management Flow

```
                    ┌──────────────┐
                    │ Super Admin  │
                    └──────┬───────┘
                           │
              ┌────────────┼────────────┐
              ▼            ▼            ▼
        POST /hotels  PUT /hotels/:id  DELETE /hotels/:id
              │            │            │
              ▼            ▼            ▼
        HotelService  HotelService  HotelService
         .Create()     .Update()     .Delete()
              │            │            │
              ▼            ▼            ▼
        HotelRepo     HotelRepo     HotelRepo
         .Create()     .Save()      .Delete() ← soft delete
              │            │            │
              ▼            ▼            ▼
          INSERT INTO   UPDATE        UPDATE SET
          hotels        hotels        deleted_at = NOW()
```

**Listing hotels** (`GET /api/hotels`) and **getting by ID** (`GET /api/hotels/:id`) are available to any staff member.

### Room Management Flow

```
Front Desk+ staff
       │
       ├─ POST /api/rooms ──────────────────────────────────────┐
       │   Body: { hotel_id, room_number, room_type,            │
       │           floor, price_per_night, max_occupancy,       │
       │           description, amenity_ids[] }                 │
       │                                                        ▼
       │                                              RoomService.Create()
       │                                                   │
       │                                    ┌──────────────┼──────────────┐
       │                                    ▼              ▼              │
       │                              RoomRepo       AmenityRepo         │
       │                              .Create()      .FindByIDs()        │
       │                                    │              │              │
       │                                    │              ▼              │
       │                                    │        RoomRepo             │
       │                                    │        .UpdateAmenities()   │
       │                                    │     (replace join table)    │
       │                                    └──────────────┬──────────────┘
       │                                                   ▼
       │                                            RoomRepo.FindByID()
       │                                           (return with amenities)
       │
       ├─ PUT /api/rooms/:id ──▶ Partial update (pointer fields)
       │                          + optional amenity_ids replacement
       │
       └─ DELETE /api/rooms/:id ──▶ Soft delete (Management role only)
```

### Room PIN / Guest Access Flow

```
┌────────────┐                                  ┌────────────┐
│ Front Desk │                                  │   Guest    │
└─────┬──────┘                                  └─────┬──────┘
      │                                               │
      │ POST /api/rooms/:id/pin                       │
      │──────────────────────────▶                    │
      │                                               │
      │ RoomService.SetAccessPin()                    │
      │   ├─ GeneratePin(6)  ← crypto/rand            │
      │   └─ Save pin to room.access_pin              │
      │                                               │
      │ ◀── { "pin": "384921" }                       │
      │                                               │
      │  (staff gives PIN to guest at check-in)       │
      │ ─────────────────────────────────────────────▶│
      │                                               │
      │                          POST /api/auth/guest/login
      │                          { room_number, pin, hotel_id }
      │                                               │
      │                          ◀── { token, room_number, room_type }
      │                                               │
      │                          Guest can now call:   │
      │                            GET /api/rooms/:id  │
      │                            GET /api/amenities  │
      │                                               │
      │ DELETE /api/rooms/:id/pin                      │
      │──────────────────────────▶                    │
      │  (clears PIN at checkout)                     │
```

### Amenity Management Flow

```
Manager / Super Admin
       │
       ├─ POST /api/amenities
       │   { name, description, icon, category }
       │   category ∈ [room, bathroom, general, dining, recreation]
       │
       ├─ PUT /api/amenities/:id
       │   (partial update via pointer fields)
       │
       └─ DELETE /api/amenities/:id  (Super Admin only)

Any Authenticated User (staff + guest)
       │
       ├─ GET /api/amenities?page=1&per_page=20
       │   Optional filter: ?category=room
       │
       └─ GET /api/amenities/:id
```

Amenities are linked to rooms via the **room_amenities** join table. When creating or updating a room, pass `amenity_ids[]` to set the room's amenities (full replacement, not append).

### Staff Creation Flow

```
Manager / Super Admin
       │
       │  POST /api/staff
       │  { email, password, name, phone, role, hotel_id }
       │
       ▼
  AuthService.CreateStaff()
       │
       ├─ Validate role ∈ ValidRoles
       ├─ Reject if role == "super_admin"
       ├─ Check email uniqueness
       ├─ Hash password (bcrypt)
       └─ Create user record
              │
              ▼
          INSERT INTO users
```

**Constraints:**
- Cannot create another `super_admin` via API
- Email must be unique
- Password minimum 8 characters
- `hotel_id` is required (staff must belong to a hotel)

---

## Layer Interaction Pattern

Every domain follows this exact pattern:

```
HTTP Request
     │
     ▼
  Handler         ← Parses request body (DTO), calls service, formats response
     │
     ▼
  Service         ← Business logic, validation, orchestrates repo calls
     │
     ▼
  Repository      ← Raw database queries (GORM)
     │
     ▼
  PostgreSQL
```

**Key conventions:**
- **Handlers** never touch the DB directly
- **Services** never access `gin.Context`
- **Repositories** have no business logic — just CRUD
- **DTOs** use Gin's binding tags for input validation
- **Update DTOs** use pointer fields (`*string`, `*int`) so only provided fields are updated (partial updates)
- All entities use **soft delete** (`deleted_at` column)
- All IDs are **UUIDs** generated by PostgreSQL (`gen_random_uuid()`)

---

## Models & DTOs Reference

### User Model

```go
type User struct {
    BaseModel
    Email        string    // unique, required
    PasswordHash string    // bcrypt, hidden from JSON
    Name         string    // required
    Phone        string
    Role         UserRole  // super_admin | manager | front_desk | housekeeping | staff
    HotelID      *string   // nullable (super_admin has no hotel)
    IsActive     bool      // default true
}
```

### Hotel Model

```go
type Hotel struct {
    BaseModel
    Name, Address, City, State, Country, ZipCode, Phone, Email, Description string
    IsActive bool
    Rooms    []Room  // has many
    Staff    []User  // has many
}
```

### Room Model

```go
type Room struct {
    BaseModel
    HotelID       string      // FK → hotels
    RoomNumber    string
    RoomType      RoomType    // single | double | suite | deluxe | penthouse
    Floor         int
    Status        RoomStatus  // available | occupied | maintenance | cleaning
    PricePerNight float64
    Description   string
    MaxOccupancy  int
    AccessPin     string      // hidden from JSON, 6-digit
    IsActive      bool
    Amenities     []Amenity   // many-to-many
}
```

### Amenity Model

```go
type Amenity struct {
    BaseModel
    Name        string
    Description string
    Icon        string
    Category    AmenityCategory  // room | bathroom | general | dining | recreation
    IsActive    bool
    Rooms       []Room           // many-to-many (inverse)
}
```

### DTO Validation Rules

| DTO | Field | Validation |
|---|---|---|
| `LoginRequest` | `email` | required, valid email |
| | `password` | required, min 6 chars |
| `GuestLoginRequest` | `room_number` | required |
| | `pin` | required, exactly 6 chars |
| | `hotel_id` | required, valid UUID |
| `CreateStaffRequest` | `email` | required, valid email |
| | `password` | required, min 8 chars |
| | `name` | required, min 2 chars |
| | `role` | required |
| | `hotel_id` | required, valid UUID |
| `CreateHotelRequest` | `name` | required, min 2 chars |
| | `address` | required |
| | `city` | required |
| | `country` | required |
| `CreateRoomRequest` | `hotel_id` | required, valid UUID |
| | `room_number` | required |
| | `room_type` | required, oneof: single/double/suite/deluxe/penthouse |
| | `floor` | required, min 0 |
| | `price_per_night` | required, > 0 |
| | `max_occupancy` | required, min 1 |
| `SetRoomPinRequest` | `pin` | required, exactly 6 chars, numeric |
| `CreateAmenityRequest` | `name` | required, min 2 chars |
| | `category` | required, oneof: room/bathroom/general/dining/recreation |

---

## CORS Configuration

| Setting | Value |
|---|---|
| Allowed Origins | `http://localhost:5173` |
| Allowed Methods | `GET, POST, PUT, DELETE, OPTIONS` |
| Allowed Headers | `Origin, Content-Type, Authorization` |
| Allow Credentials | `true` |
