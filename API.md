# Hotel PMS — API Reference

Base URL: `http://localhost:4001/api` | Content-Type: `application/json` | Auth: `Authorization: Bearer <token>`

## Roles

- **Super Admin**: `super_admin`
- **Management**: `super_admin`, `manager`
- **Front Desk+**: `super_admin`, `manager`, `front_desk`
- **Any Staff**: all staff roles (`super_admin`, `manager`, `front_desk`, `housekeeping`, `staff`)
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
**Amenity:** `{id, name, description, icon, category, is_active, created_at, updated_at}`
**Auth (staff):** `{token, user: User}`
**Auth (guest):** `{token, room_number, room_type}`

## Endpoints

### Health

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/api/health` | No | Returns `{"status":"ok","service":"hotel-pms"}` |

### Auth

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/api/auth/login` | No | Staff login |
| POST | `/api/auth/guest/login` | No | Guest login via room PIN |

**POST /api/auth/login** — Body: `{email*, password*}` (min 6 chars). Returns `{token, user}`.

**POST /api/auth/guest/login** — Body: `{room_number*, pin* (6 chars), hotel_id* (uuid)}`. Returns `{token, room_number, room_type}`. Error 401 if invalid or room not occupied.

### Staff

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/api/staff` | Management | Create staff member |

**POST /api/staff** — Body: `{email* (unique), password* (min 8), name* (min 2), phone, role* (manager|front_desk|housekeeping|staff), hotel_id* (uuid)}`. Cannot create `super_admin`. Returns 201 with user object.

### Hotels

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/api/hotels` | Super Admin | Create hotel |
| GET | `/api/hotels` | Any Staff | List hotels (paginated) |
| GET | `/api/hotels/:id` | Any Staff | Get hotel |
| PUT | `/api/hotels/:id` | Management | Update hotel (partial) |
| DELETE | `/api/hotels/:id` | Super Admin | Delete hotel |
| GET | `/api/hotels/:id/rooms` | Any Staff | List hotel rooms (paginated, includes amenities) |

**Hotel fields**: `name* (min 2), address*, city*, state, country*, zip_code, phone, email (valid if provided), description, is_active (update only)`.

### Rooms

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/api/rooms` | Front Desk+ | Create room |
| GET | `/api/rooms/:id` | Any Authenticated | Get room (with amenities) |
| PUT | `/api/rooms/:id` | Front Desk+ | Update room (partial) |
| DELETE | `/api/rooms/:id` | Management | Delete room |
| POST | `/api/rooms/:id/pin` | Front Desk+ | Generate 6-digit guest PIN |
| DELETE | `/api/rooms/:id/pin` | Front Desk+ | Clear guest PIN |

**Room fields**: `hotel_id* (uuid), room_number* (unique per hotel), room_type* (single|double|suite|deluxe|penthouse), floor* (min 0), price_per_night* (>0), description, max_occupancy* (min 1), amenity_ids (uuid[])`.

**Update-only fields**: `status (available|occupied|maintenance|cleaning), is_active`. Note: `amenity_ids` replaces the full list (not append).

### Amenities

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/api/amenities` | Management | Create amenity |
| GET | `/api/amenities` | Any Authenticated | List amenities (paginated, filter by `?category=`) |
| GET | `/api/amenities/:id` | Any Authenticated | Get amenity |
| PUT | `/api/amenities/:id` | Management | Update amenity (partial) |
| DELETE | `/api/amenities/:id` | Super Admin | Delete amenity |

**Amenity fields**: `name* (min 2), description, icon, category* (room|bathroom|general|dining|recreation), is_active (update only)`.

## HTTP Status Codes

200 Success | 201 Created | 400 Bad Request | 401 Unauthorized | 403 Forbidden | 404 Not Found | 500 Server Error
