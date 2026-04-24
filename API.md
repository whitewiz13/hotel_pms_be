# Hotel PMS — API Reference

Base URL: `http://localhost:4001/api` | Content-Type: `application/json` | Auth: `Authorization: Bearer <token>`

---

## Architecture

Everything is **hotel-scoped**. A Super Admin creates hotels (each with a Hotel Admin). The Hotel Admin then manages their own hotel's staff, rooms, and amenities. Staff can only access their own hotel's data.

## Roles

| Role | Slug | Scope |
|------|------|-------|
| Super Admin | `super_admin` | Global — manages all hotels |
| Hotel Admin | `hotel_admin` | Single hotel — full management |
| Manager | `manager` | Single hotel — operational management |
| Front Desk | `front_desk` | Single hotel — room & guest ops |
| Housekeeping | `housekeeping` | Single hotel — view only |
| Staff | `staff` | Single hotel — view only |
| Guest | `guest` | Single room (virtual, JWT only) |

**Role groups used in permissions:**
- **Super Admin**: `super_admin`
- **Hotel Admin+**: `super_admin`, `hotel_admin`
- **Hotel Management**: `super_admin`, `hotel_admin`, `manager`
- **Front Desk+**: `super_admin`, `hotel_admin`, `manager`, `front_desk`
- **Any Staff**: all staff roles
- **Any Authenticated**: all staff + `guest`

## Response Structures

**Success:** `{"success": true, "data": {...}}`
**Success (message only):** `{"success": true, "message": "..."}`
**Paginated:** `{"success": true, "data": [...], "meta": {"page": 1, "per_page": 20, "total": 45, "total_pages": 3}}`
**Error:** `{"success": false, "error": "description"}`

List endpoints accept `?page=1&per_page=20` (max 100).

### Common Object Shapes

**User:** `{id, email, name, phone, role, hotel_id, is_active, created_at, updated_at}`
**Hotel:** `{id, name, address, city, state, country, zip_code, phone, email, description, is_active, created_at, updated_at}`
**Room:** `{id, hotel_id, room_number, room_type, floor, status, price_per_night, description, max_occupancy, is_active, amenities[], created_at, updated_at}`
**Amenity:** `{id, hotel_id, name, description, icon, category, is_active, created_at, updated_at}`
**Auth (staff):** `{token, user: User}`
**Auth (guest):** `{token, room_number, room_type}`

---

## Auth (Public)

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/api/auth/login` | No | Staff/admin login |
| POST | `/api/auth/guest/login` | No | Guest login via room PIN |

**POST /api/auth/login**
Body: `{email*, password*}` (min 6 chars). Returns `{token, user}`.
Works for super_admin, hotel_admin, manager, front_desk, housekeeping, staff.

**POST /api/auth/guest/login**
Body: `{room_number*, pin* (6 chars), hotel_id* (uuid)}`. Returns `{token, room_number, room_type}`.
Error 401 if invalid or room not occupied.

---

## Health

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/api/health` | No | Returns `{"status":"ok","service":"hotel-pms"}` |

---

## Super Admin APIs

*Login with super_admin credentials. These endpoints manage hotels globally.*

### Hotels (Global)

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/api/hotels` | Super Admin | Create hotel + hotel admin user |
| GET | `/api/hotels` | Super Admin | List all hotels (paginated) |
| GET | `/api/hotels/:hotel_id` | Super Admin | Get hotel details |
| PUT | `/api/hotels/:hotel_id` | Super Admin | Update hotel |
| DELETE | `/api/hotels/:hotel_id` | Super Admin | Delete hotel |

### Users (Global)

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/api/users` | Super Admin | List all users (paginated, `?role=`) |

**GET /api/users?role=hotel_admin** — List users, optionally filtered by role.
Accepts `?page=1&per_page=20&role=hotel_admin`. Returns paginated user list.

**POST /api/hotels** — Creates a hotel and its admin account in one transaction.
```json
{
  "name": "Grand Hotel",          // required, min 2
  "address": "123 Main St",       // required
  "city": "Mumbai",               // required
  "country": "India",             // required
  "state": "Maharashtra",
  "zip_code": "400001",
  "phone": "+91-22-12345678",
  "email": "info@grandhotel.com",
  "description": "5 star hotel",
  "admin_email": "admin@grand.com",    // required, unique
  "admin_password": "securepass123",   // required, min 8
  "admin_name": "John Doe"            // required, min 2
}
```
Returns 201: `{hotel: Hotel, admin: User}`

**PUT /api/hotels/:hotel_id** — Partial update.
Fields: `name, address, city, state, country, zip_code, phone, email, description, is_active` (all optional).

---

## Hotel Admin APIs

*Login with hotel_admin credentials. All routes are scoped to `/api/hotels/:hotel_id/...`*
*The `hotel_id` in the URL must match the admin's assigned hotel.*

### Hotel Details

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/api/hotels/:hotel_id` | Hotel Admin+ | View own hotel |
| PUT | `/api/hotels/:hotel_id` | Hotel Admin+ | Update own hotel |

### Staff Management

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/api/hotels/:hotel_id/staff` | Hotel Admin+ | Create staff member |
| GET | `/api/hotels/:hotel_id/staff` | Hotel Admin+ | List hotel staff (paginated) |

**GET /api/hotels/:hotel_id/staff** — List all staff members of the hotel (paginated).
Accepts `?page=1&per_page=20`. Returns paginated user list.

**POST /api/hotels/:hotel_id/staff**
```json
{
  "email": "staff@hotel.com",     // required, unique
  "password": "password123",      // required, min 8
  "name": "Jane Smith",           // required, min 2
  "phone": "+91-9876543210",
  "role": "manager"               // required: manager|front_desk|housekeeping|staff
}
```
Cannot create `super_admin` or `hotel_admin`. Returns 201 with user object.

### Rooms

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/api/hotels/:hotel_id/rooms` | Front Desk+ | Create room |
| GET | `/api/hotels/:hotel_id/rooms` | Any hotel staff | List rooms (paginated) |
| GET | `/api/hotels/:hotel_id/rooms/:id` | Any hotel staff | Get room (with amenities) |
| PUT | `/api/hotels/:hotel_id/rooms/:id` | Front Desk+ | Update room (partial) |
| DELETE | `/api/hotels/:hotel_id/rooms/:id` | Hotel Admin+ | Delete room |
| POST | `/api/hotels/:hotel_id/rooms/:id/pin` | Front Desk+ | Generate 6-digit guest PIN |
| DELETE | `/api/hotels/:hotel_id/rooms/:id/pin` | Front Desk+ | Clear guest PIN |

**POST /api/hotels/:hotel_id/rooms**
```json
{
  "room_number": "101",           // required, unique per hotel
  "room_type": "deluxe",          // required: single|double|suite|deluxe|penthouse
  "floor": 1,                     // required, min 0
  "price_per_night": 5000.00,     // required, > 0
  "max_occupancy": 2,             // required, min 1
  "description": "Sea view room",
  "amenity_ids": ["uuid1", "uuid2"]  // optional, must belong to same hotel
}
```

**PUT /api/hotels/:hotel_id/rooms/:id** — Partial update.
Fields: `room_number, room_type, floor, status (available|occupied|maintenance|cleaning), price_per_night, description, max_occupancy, is_active, amenity_ids`.
Note: `amenity_ids` replaces the full list (not append).

**POST /api/hotels/:hotel_id/rooms/:id/pin** — Returns `{pin: "123456"}`.

### Amenities

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/api/hotels/:hotel_id/amenities` | Hotel Admin+ | Create amenity |
| GET | `/api/hotels/:hotel_id/amenities` | Any hotel staff | List amenities (paginated, `?category=`) |
| GET | `/api/hotels/:hotel_id/amenities/:id` | Any hotel staff | Get amenity |
| PUT | `/api/hotels/:hotel_id/amenities/:id` | Hotel Admin+ | Update amenity (partial) |
| DELETE | `/api/hotels/:hotel_id/amenities/:id` | Hotel Admin+ | Delete amenity |

**POST /api/hotels/:hotel_id/amenities**
```json
{
  "name": "WiFi",                    // required, min 2
  "description": "High-speed internet",
  "icon": "wifi",
  "category": "room"                 // required: room|bathroom|general|dining|recreation
}
```

**GET /api/hotels/:hotel_id/amenities?category=room** — Filter by category.

---

## Staff APIs

*Login with manager/front_desk/housekeeping/staff credentials.*
*All routes scoped to the staff member's assigned hotel.*

Staff members can access the same hotel-scoped endpoints listed above, limited by their role:

| Role | Can Do |
|------|--------|
| **Manager** | Manage staff, rooms, amenities, pins |
| **Front Desk** | Create/update rooms, manage guest pins, view amenities |
| **Housekeeping** | View rooms, view amenities |
| **Staff** | View rooms, view amenities |

---

## Guest APIs

*Login with room PIN. Access is limited to the single room.*

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/api/auth/guest/login` | No | Login with room pin |

**POST /api/auth/guest/login**
```json
{
  "room_number": "101",
  "pin": "123456",
  "hotel_id": "uuid"
}
```
Returns: `{token, room_number, room_type}`

Guest token contains `hotel_id`, `room_id`, `room_number`, and `is_guest: true`.

---

## Hotel Access Control

All `/api/hotels/:hotel_id/*` routes enforce hotel scoping:
- **Super Admin** — can access any hotel
- **Hotel Admin / Staff** — `hotel_id` in URL must match their JWT `hotel_id`, otherwise 403
- **Guest** — limited to guest login only

---

## HTTP Status Codes

| Code | Meaning |
|------|---------|
| 200 | Success |
| 201 | Created |
| 400 | Bad Request (validation error) |
| 401 | Unauthorized (no/invalid token) |
| 403 | Forbidden (wrong role or wrong hotel) |
| 404 | Not Found |
| 500 | Server Error |
