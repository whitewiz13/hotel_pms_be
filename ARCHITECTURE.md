# HotelPMS Backend — Architecture Reference

> Auto-generated from codebase on 2026-04-25. Keep this file updated when making structural changes.

## Tech Stack

| Component      | Technology                          |
| -------------- | ----------------------------------- |
| Language       | Go 1.26.1                          |
| Web Framework  | Gin (`github.com/gin-gonic/gin`)   |
| ORM            | GORM (`gorm.io/gorm`)             |
| Database       | PostgreSQL (`gorm.io/driver/postgres`) |
| Auth           | JWT (`github.com/golang-jwt/jwt/v5`) |
| Password Hash  | bcrypt (`golang.org/x/crypto`)     |
| CORS           | `github.com/gin-contrib/cors`      |
| UUID           | `github.com/google/uuid`           |
| Env Loading    | `github.com/joho/godotenv`         |

**Module path:** `github.com/hotelpms/backend`

---

## Architecture Pattern

Clean Architecture with 4 layers:

```
HTTP Request
     │
     ▼
┌──────────────────────────────────────────────┐
│  ROUTER  (router/router.go)                  │
│  Gin engine + CORS + middleware chains       │
└───────────────────┬──────────────────────────┘
                    ▼
┌──────────────────────────────────────────────┐
│  HANDLERS  (handler/*.go)                    │
│  Parse request, validate input, call service │
│  Return JSON via utils.Respond*()            │
└───────────────────┬──────────────────────────┘
                    ▼
┌──────────────────────────────────────────────┐
│  SERVICES  (service/*.go)                    │
│  Business logic, state transitions,          │
│  transaction management, cross-repo ops      │
└───────────────────┬──────────────────────────┘
                    ▼
┌──────────────────────────────────────────────┐
│  REPOSITORIES  (repository/*.go)             │
│  GORM queries, preloading, pagination        │
└───────────────────┬──────────────────────────┘
                    ▼
┌──────────────────────────────────────────────┐
│  DATABASE  (PostgreSQL)                      │
│  Managed via GORM migrations                 │
└──────────────────────────────────────────────┘
```

**Dependency Injection:** All wired in `main.go` — repos get `*gorm.DB`, services get repos, handlers get services, router gets handlers + auth service.

---

## Directory Structure

```
main.go                    # Entry point, DI wiring, server start
config/config.go           # Env-based configuration struct
database/
  database.go              # PostgreSQL connection (GORM), connection pool
  migrate.go               # Ordered migration system with tracking table
models/                    # GORM models (DB schema)
dto/                       # Request/response structs (JSON binding)
repository/                # Data access layer (one per entity)
service/                   # Business logic layer (one per entity)
handler/                   # HTTP handlers (one per entity)
  handler.go               # Handler struct (empty, used for grouping)
middleware/
  auth.go                  # JWT extraction & validation middleware
  role.go                  # RBAC middleware + hotel access check
router/router.go           # All route definitions & middleware chains
seeds/seeder.go            # DB seeder (super admin user)
utils/
  password.go              # bcrypt hash, pin generation
  response.go              # Standardized JSON response helpers
```

---

## Configuration (`config/config.go`)

Loaded from `.env` file via `godotenv`.

| Env Variable             | Default   | Notes                    |
| ------------------------ | --------- | ------------------------ |
| `SERVER_PORT`            | `8080`    |                          |
| `GIN_MODE`               | `debug`   | `debug` / `release`     |
| `DB_HOST`                | —         | PostgreSQL host          |
| `DB_PORT`                | —         |                          |
| `DB_USER`                | —         |                          |
| `DB_PASSWORD`            | —         |                          |
| `DB_NAME`                | —         |                          |
| `DB_SSLMODE`             | `disable` |                          |
| `DB_TIMEZONE`            | `UTC`     |                          |
| `JWT_SECRET`             | —         | **Required**             |
| `JWT_EXPIRY_HOURS`       | `24`      | Staff token lifetime     |
| `JWT_GUEST_EXPIRY_HOURS` | `12`      | Guest token lifetime     |
| `RUN_MIGRATIONS`         | `true`    |                          |
| `SEED_DB`                | `false`   |                          |

---

## Startup Flow (`main.go`)

1. Load config from env
2. Connect to PostgreSQL (GORM) — pool: 10 idle, 100 max open
3. Run migrations if `RUN_MIGRATIONS=true`
4. Run seeders if `SEED_DB=true`
5. Initialize all 11 repositories
6. Initialize all 12 services
7. Initialize all 12 handlers
8. Setup router with handlers + auth service
9. Start Gin server on configured port

---

## Data Models (`models/`)

### Base Model (embedded in all entities)

```go
ID        uuid.UUID      // PK, auto-generated
CreatedAt time.Time
UpdatedAt time.Time
DeletedAt gorm.DeletedAt // Soft delete
```

### Entity Relationship Diagram

```
Hotel 1──* Room          (hotel_id FK)
Hotel 1──* User/Staff    (hotel_id FK)
Hotel 1──* Amenity       (hotel_id FK)
Hotel 1──* Reservation   (hotel_id FK)
Hotel 1──* HousekeepingTask (hotel_id FK)
Hotel 1──* MenuItem      (hotel_id FK)
Hotel 1──* Order         (hotel_id FK)
Hotel 1──* Activity      (hotel_id FK)
Hotel 1──* Bill          (hotel_id FK)

Room  *──* Amenity       (many-to-many join table)
Room  1──* Reservation   (room_id FK)
Room  1──* RoomInventory (room_id FK)
Room  1──* Order         (room_id FK)

Reservation 1──* RoomInventory  (reservation_id FK)
Reservation 1──* Order          (reservation_id FK)
Reservation 1──* ActivityBooking (reservation_id FK)
Reservation 1──1 Bill           (reservation_id FK, unique)

Activity 1──* ActivityBooking   (activity_id FK)

Order 1──* OrderItem            (order_id FK)
OrderItem *──1 MenuItem         (menu_item_id FK)

Bill 1──* BillLineItem          (bill_id FK)

HousekeepingTask *──1 Room      (room_id FK)
HousekeepingTask *──1 User      (assigned_to_id FK, nullable)
HousekeepingTask *──1 User      (assigned_by_id FK)
```

### Models Summary

| Model             | Key Fields                                                                 | Enums                                                    |
| ----------------- | -------------------------------------------------------------------------- | -------------------------------------------------------- |
| **User**          | email (unique), password_hash, name, phone, role, hotel_id?, is_active    | `UserRole`: see Roles section                           |
| **Hotel**         | name, address, city, state, country, zip_code, phone, email, description  | —                                                        |
| **Room**          | hotel_id, room_number (unique/hotel), room_type, floor, status, price_per_night, max_occupancy, access_pin | `RoomType`: single, double, suite, deluxe, penthouse; `RoomStatus`: available, occupied, dirty, cleaning, maintenance |
| **Amenity**       | hotel_id, name, description, icon, category, is_active                    | `AmenityCategory`: room, bathroom, general, dining, recreation |
| **Reservation**   | hotel_id, room_id, guest_name, guest_phone, check_in_date, check_out_date, status, notes | `ReservationStatus`: reserved, checked_in, checked_out, cancelled |
| **RoomInventory** | hotel_id, room_id, date, is_available, reservation_id? — Unique(room_id, date) | —                                                        |
| **HousekeepingTask** | hotel_id, room_id, assigned_to_id?, assigned_by_id, status, priority, notes | `HousekeepingStatus`: pending, in_progress, completed; `HousekeepingPriority`: low, normal, high, urgent |
| **MenuItem**      | hotel_id, name, description, price, category, is_available                | `MenuCategory`: appetizer, main_course, dessert, beverage, snack |
| **Order**         | hotel_id, room_id, reservation_id, guest_name, status, total_amount, notes, assigned_to_id? | `OrderStatus`: pending, preparing, ready, delivered, cancelled |
| **OrderItem**     | order_id, menu_item_id, quantity, unit_price, subtotal, notes             | —                                                        |
| **Activity**      | hotel_id, name, description, price, category, is_available                | `ActivityCategory`: cab, spa, tour, laundry, other       |
| **ActivityBooking** | hotel_id, room_id, reservation_id, activity_id, guest_name, scheduled_at?, status, amount, notes | `ActivityBookingStatus`: pending, confirmed, completed, cancelled |
| **Bill**          | hotel_id, reservation_id (unique), room_id, guest_name, room_charges, upfront_paid, room_service_total, activity_total, subtotal, tax_rate, tax_amount, total_amount, status, notes | `BillStatus`: pending, paid |
| **BillLineItem**  | bill_id, type, description, amount, reference_id?                         | `BillLineType`: room_charge, room_service, activity, upfront_payment |

---

## Authentication & Authorization

### JWT Claims Structure

```go
type JWTClaims struct {
    UserID     string
    Email      string
    Role       UserRole  // or "guest"
    HotelID    string
    RoomID     string    // guest only
    RoomNumber string    // guest only
    IsGuest    bool
    jwt.RegisteredClaims
}
```

### User Roles (8 total)

| Role             | Constant              | Stored in DB | Notes                   |
| ---------------- | --------------------- | ------------ | ----------------------- |
| `super_admin`    | `RoleSuperAdmin`      | Yes          | Global access           |
| `hotel_admin`    | `RoleHotelAdmin`      | Yes          | Single hotel full access |
| `manager`        | `RoleManager`         | Yes          | Hotel operational mgmt  |
| `front_desk`     | `RoleFrontDesk`       | Yes          | Room & guest operations |
| `housekeeping`   | `RoleHousekeeping`    | Yes          | Cleaning tasks          |
| `room_service`   | `RoleRoomService`     | Yes          | Order management        |
| `staff`          | `RoleStaff`           | Yes          | Basic view access       |
| `guest`          | `RoleGuest`           | **No**       | Virtual, JWT-only       |

### Role Hierarchy Middleware

| Middleware Function              | Allowed Roles                                                |
| -------------------------------- | ------------------------------------------------------------ |
| `RequireSuperAdmin()`            | super_admin                                                  |
| `RequireHotelAdminOrAbove()`     | super_admin, hotel_admin                                     |
| `RequireHotelManagement()`       | super_admin, hotel_admin, manager                            |
| `RequireHotelFrontDeskOrAbove()` | super_admin, hotel_admin, manager, front_desk                |
| `RequireHousekeepingOrAbove()`   | super_admin, hotel_admin, manager, front_desk, housekeeping  |
| `RequireRoomServiceOrAbove()`    | super_admin, hotel_admin, manager, front_desk, room_service  |
| `RequireAnyStaff()`              | All `ValidRoles` (all non-guest)                             |
| `RequireAnyAuthenticated()`      | All `ValidRoles` + guest                                     |

### Middleware Chain

Every protected route goes through:
1. **AuthMiddleware** — Extracts `Bearer <token>`, validates JWT, sets claims in context
2. **HotelAccessMiddleware** — Ensures URL `hotel_id` matches user's hotel (super_admin bypasses)
3. **Role Middleware** — One of the `Require*()` functions per endpoint

Helper: `GetClaims(c *gin.Context) *JWTClaims` — extracts claims from Gin context

---

## Complete API Route Map

### Public Routes

| Method | Path                    | Handler              | Notes           |
| ------ | ----------------------- | -------------------- | --------------- |
| GET    | `/api/health`           | inline               | Health check    |
| POST   | `/api/auth/login`       | Auth.Login           | Staff login     |
| POST   | `/api/auth/guest/login` | Auth.GuestLogin      | Guest room PIN  |

### Super Admin Routes

| Method | Path           | Handler        | Role         |
| ------ | -------------- | -------------- | ------------ |
| POST   | `/api/hotels`  | Hotel.Create   | super_admin  |
| GET    | `/api/hotels`  | Hotel.GetAll   | super_admin  |
| GET    | `/api/users`   | User.GetAll    | super_admin  |

### Hotel-Scoped Routes (`/api/hotels/:hotel_id/...`)

All require `AuthMiddleware` + `HotelAccessMiddleware`.

#### Hotel Management

| Method | Path    | Handler       | Min Role     |
| ------ | ------- | ------------- | ------------ |
| GET    | `/`     | Hotel.GetByID | any staff    |
| PUT    | `/`     | Hotel.Update  | hotel_admin+ |
| DELETE | `/`     | Hotel.Delete  | super_admin  |

#### Staff

| Method | Path     | Handler          | Min Role     |
| ------ | -------- | ---------------- | ------------ |
| POST   | `/staff` | Auth.CreateStaff | hotel_admin+ |
| GET    | `/staff` | User.GetByHotelID | hotel_admin+ |

#### Rooms

| Method | Path              | Handler            | Min Role     |
| ------ | ----------------- | ------------------ | ------------ |
| POST   | `/rooms`          | Room.Create        | front_desk+  |
| GET    | `/rooms`          | Room.GetByHotelID  | any staff    |
| GET    | `/rooms/:id`      | Room.GetByID       | any staff    |
| PUT    | `/rooms/:id`      | Room.Update        | front_desk+  |
| DELETE | `/rooms/:id`      | Room.Delete        | hotel_admin+ |
| POST   | `/rooms/:id/pin`  | Room.GeneratePin   | front_desk+  |
| DELETE | `/rooms/:id/pin`  | Room.ClearPin      | front_desk+  |

#### Reservations

| Method | Path                            | Handler                  | Min Role    |
| ------ | ------------------------------- | ------------------------ | ----------- |
| GET    | `/availability`                 | Reservation.GetAvailability | front_desk+ |
| POST   | `/reservations`                 | Reservation.Create       | front_desk+ |
| GET    | `/reservations`                 | Reservation.List         | front_desk+ |
| GET    | `/reservations/:id`             | Reservation.GetByID      | front_desk+ |
| POST   | `/reservations/:id/check-in`    | Reservation.CheckIn      | front_desk+ |
| POST   | `/reservations/:id/check-out`   | Reservation.CheckOut     | front_desk+ |
| POST   | `/reservations/:id/cancel`      | Reservation.Cancel       | front_desk+ |

#### Amenities

| Method | Path              | Handler         | Min Role     |
| ------ | ----------------- | --------------- | ------------ |
| POST   | `/amenities`      | Amenity.Create  | hotel_admin+ |
| GET    | `/amenities`      | Amenity.GetAll  | any staff    |
| GET    | `/amenities/:id`  | Amenity.GetByID | any staff    |
| PUT    | `/amenities/:id`  | Amenity.Update  | hotel_admin+ |
| DELETE | `/amenities/:id`  | Amenity.Delete  | hotel_admin+ |

#### Housekeeping

| Method | Path                        | Handler               | Min Role        |
| ------ | --------------------------- | --------------------- | --------------- |
| POST   | `/housekeeping`             | Housekeeping.Assign   | front_desk+     |
| GET    | `/housekeeping`             | Housekeeping.List     | housekeeping+   |
| GET    | `/housekeeping/:id`         | Housekeeping.GetByID  | housekeeping+   |
| POST   | `/housekeeping/:id/start`   | Housekeeping.Start    | housekeeping+   |
| POST   | `/housekeeping/:id/complete`| Housekeeping.Complete | housekeeping+   |

#### Menu (Room Service)

| Method | Path          | Handler      | Min Role      |
| ------ | ------------- | ------------ | ------------- |
| POST   | `/menu`       | Menu.Create  | management    |
| GET    | `/menu`       | Menu.List    | any auth      |
| GET    | `/menu/:id`   | Menu.GetByID | any auth      |
| PUT    | `/menu/:id`   | Menu.Update  | management    |
| DELETE | `/menu/:id`   | Menu.Delete  | management    |

#### Orders (Room Service)

| Method | Path                     | Handler             | Min Role       |
| ------ | ------------------------ | ------------------- | -------------- |
| POST   | `/orders`                | Order.Create        | front_desk+    |
| GET    | `/orders`                | Order.List          | room_service+  |
| GET    | `/orders/:id`            | Order.GetByID       | room_service+  |
| POST   | `/orders/:id/status`     | Order.UpdateStatus  | room_service+  |
| POST   | `/orders/:id/assign`     | Order.Assign        | front_desk+    |

#### Activities

| Method | Path                              | Handler                      | Min Role      |
| ------ | --------------------------------- | ---------------------------- | ------------- |
| POST   | `/activities`                     | Activity.Create              | management    |
| GET    | `/activities`                     | Activity.List                | any auth      |
| GET    | `/activities/:id`                 | Activity.GetByID             | any auth      |
| PUT    | `/activities/:id`                 | Activity.Update              | management    |
| DELETE | `/activities/:id`                 | Activity.Delete              | management    |
| POST   | `/activity-bookings`              | Activity.CreateBooking       | front_desk+   |
| GET    | `/activity-bookings`              | Activity.ListBookings        | front_desk+   |
| GET    | `/activity-bookings/:id`          | Activity.GetBookingByID      | front_desk+   |
| POST   | `/activity-bookings/:id/status`   | Activity.UpdateBookingStatus | front_desk+   |

#### Billing

| Method | Path                          | Handler              | Min Role    |
| ------ | ----------------------------- | -------------------- | ----------- |
| POST   | `/reservations/:id/bill`      | Bill.Generate        | front_desk+ |
| GET    | `/reservations/:id/bill`      | Bill.GetByReservation | front_desk+ |
| GET    | `/bills`                      | Bill.List            | front_desk+ |
| GET    | `/bills/:id`                  | Bill.GetByID         | front_desk+ |
| POST   | `/bills/:id/pay`              | Bill.MarkPaid        | front_desk+ |

#### Dashboard

| Method | Path                | Handler              | Min Role    |
| ------ | ------------------- | -------------------- | ----------- |
| GET    | `/dashboard/stats`  | Dashboard.GetStats   | front_desk+ |
| GET    | `/activity`         | Dashboard.GetActivity | front_desk+ |

---

## Key Business Logic

### Reservation Flow

```
reserved ──→ checked_in ──→ checked_out
    │                            │
    └──→ cancelled               └──→ room.status = "dirty"
```

- **Create:** Validates date range → `SELECT FOR UPDATE` on RoomInventory → creates inventory records (marks dates unavailable) — all within a transaction
- **CheckIn:** Status `reserved` → `checked_in`, room status → `occupied`
- **CheckOut:** Status `checked_in` → `checked_out`, room status → `dirty`, frees inventory
- **Cancel:** Status `reserved` → `cancelled`, frees inventory

### Order Status Transitions

```
pending ──→ preparing ──→ ready ──→ delivered
   │            │
   └──→ cancelled ←──┘
```

### Activity Booking Status Transitions

```
pending ──→ confirmed ──→ completed
   │            │
   └──→ cancelled ←──┘
```

### Bill Generation

1. Calculate room charges: `price_per_night × number_of_nights`
2. Sum room service orders (non-cancelled)
3. Sum activity bookings (non-cancelled)
4. Apply upfront payment as negative line item
5. `subtotal = room + service + activity - upfront`
6. `tax = subtotal × tax_rate`
7. `total = subtotal + tax`

One bill per reservation (unique constraint on `reservation_id`).

### Guest Access (PIN-based)

1. Front desk generates 6-digit PIN for a room (`POST /rooms/:id/pin`)
2. Guest logs in with room_number + pin + hotel_id (`POST /auth/guest/login`)
3. Guest gets JWT with `IsGuest=true`, can access menu & activities (read-only)

---

## Database

### Connection Pool

- Max idle connections: 10
- Max open connections: 100

### Migration System (`database/migrate.go`)

Ordered migrations tracked in `schema_migrations` table:

| Version                          | Creates                                    |
| -------------------------------- | ------------------------------------------ |
| `20260424_001_initial_schema`    | users, hotels, rooms, amenities, join table |
| `20260425_002_reservations`      | reservations, room_inventory               |
| `20260425_003_housekeeping`      | housekeeping_tasks                         |
| `20260425_004_room_service`      | menu_items, orders, order_items            |
| `20260425_005_activities`        | activities, activity_bookings              |
| `20260425_006_billing`           | bills, bill_line_items                     |

### Seed Data (`seeds/seeder.go`)

Tracked in `schema_seeders` table:

| Seeder                          | Creates                                            |
| ------------------------------- | -------------------------------------------------- |
| `20260424_001_super_admin`      | Super admin: `admin@hotelpms.com` / `admin@123`   |

---

## DTOs (`dto/`)

### Naming Convention

- `Create*Request` — POST body for creation
- `Update*Request` — PUT body (pointer fields for partial update)
- `List*Query` — GET query params for filtering/pagination
- `*Response` — Explicit response shapes (used sparingly; most endpoints return model directly)

### Common Query Params

All list endpoints support: `page` (default 1), `per_page` (default 20).

### Key DTOs

| File                  | Structs                                                                              |
| --------------------- | ------------------------------------------------------------------------------------ |
| `auth_dto.go`         | LoginRequest, GuestLoginRequest, LoginResponse, CreateStaffRequest                   |
| `hotel_dto.go`        | CreateHotelRequest (includes admin_email/password/name), UpdateHotelRequest           |
| `room_dto.go`         | CreateRoomRequest (includes amenity_ids[]), UpdateRoomRequest, SetRoomPinRequest      |
| `reservation_dto.go`  | CreateReservationRequest, ListReservationsQuery, AvailabilityQuery                    |
| `amenity_dto.go`      | CreateAmenityRequest, UpdateAmenityRequest                                            |
| `housekeeping_dto.go` | AssignHousekeepingRequest, UpdateHousekeepingTaskRequest, ListHousekeepingQuery       |
| `menu_dto.go`         | CreateMenuItemRequest, UpdateMenuItemRequest, ListMenuQuery                           |
| `order_dto.go`        | CreateOrderRequest (items[]), OrderItemRequest, UpdateOrderStatusRequest, AssignOrderRequest, ListOrdersQuery |
| `activity_dto.go`     | CreateActivityRequest, UpdateActivityRequest, CreateActivityBookingRequest, UpdateActivityBookingStatusRequest, ListActivitiesQuery, ListActivityBookingsQuery |
| `bill_dto.go`         | GenerateBillRequest (upfront_paid, tax_rate), UpdateBillStatusRequest, ListBillsQuery |
| `dashboard_dto.go`    | DashboardStatsResponse (room counts, occupancy_rate, check-ins/outs, pending housekeeping), ActivityItem |

---

## Utils (`utils/`)

### `response.go` — Standardized JSON Responses

```go
type APIResponse       { Success, Message, Data, Error }
type PaginatedResponse { Success, Data, Meta{Page, PerPage, Total, TotalPages} }

RespondOK(c, data)                    // 200
RespondCreated(c, data)               // 201
RespondMessage(c, message)            // 200 message-only
RespondError(c, status, message)      // custom status
RespondBadRequest(c, message)         // 400
RespondUnauthorized(c, message)       // 401
RespondForbidden(c, message)          // 403
RespondNotFound(c, message)           // 404
RespondInternalError(c, message)      // 500
RespondPaginated(c, data, page, perPage, total) // paginated
```

### `password.go`

```go
HashPassword(password string) (string, error)    // bcrypt
CheckPassword(password, hash string) bool        // bcrypt compare
GeneratePin(length int) (string, error)          // random digits
```

---

## CORS

Configured in `router/router.go`:
- Allowed origin: `http://localhost:5173`
- Allowed methods: GET, POST, PUT, DELETE, OPTIONS
- Allowed headers: Origin, Content-Type, Authorization
- Credentials: true

---

## Conventions & Patterns

1. **Hotel scoping** — Almost all data is scoped to a hotel via `hotel_id`. URL param `hotel_id` is enforced by `HotelAccessMiddleware`
2. **Soft deletes** — All models use GORM's `DeletedAt` for soft deletion
3. **UUID primary keys** — All entities use `uuid.UUID` as PK, auto-generated
4. **Pagination** — All list endpoints accept `page`/`per_page`, return `PaginatedResponse`
5. **Partial updates** — Update DTOs use pointer fields (`*string`, `*float64`) so zero values vs absent values are distinguishable
6. **Date format** — Reservation dates use `YYYY-MM-DD` string format in DTOs, parsed to `time.Time`
7. **Transaction safety** — Reservation create/cancel/checkout use `gorm.DB.Transaction()` with `SELECT FOR UPDATE` for inventory locking
8. **No interfaces** — Repos and services are concrete structs, not interfaces (no mocking layer)
9. **Error propagation** — Services return `error`, handlers translate to HTTP status codes
10. **Preloading** — Repositories use GORM `.Preload()` for related data (e.g., Room.Amenities, Order.Items)
