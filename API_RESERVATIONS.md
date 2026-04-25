# Hotel PMS — Reservation & Booking API Reference

Base URL: `http://localhost:4001/api` | Content-Type: `application/json` | Auth: `Authorization: Bearer <token>`

All reservation endpoints are scoped to `/api/hotels/:hotel_id/...` and require **Front Desk+** role (`super_admin`, `hotel_admin`, `manager`, `front_desk`).

---

## Object Shapes

**Reservation:** `{id, hotel_id, room_id, guest_name, guest_phone, check_in_date, check_out_date, status, notes, room: Room, created_at, updated_at}`

**Status values:** `reserved`, `checked_in`, `checked_out`, `cancelled`

**State transitions:**
```
reserved → checked_in → checked_out
reserved → cancelled
```

---

## Date Rules

- **check_in_date**: inclusive (guest stays this night)
- **check_out_date**: exclusive (guest leaves this morning)
- Example: `2026-05-01` → `2026-05-03` = 2 nights (May 1, May 2)
- Back-to-back bookings allowed: (May 1–3) and (May 3–5) do not conflict
- Dates must be in `YYYY-MM-DD` format
- check_in_date cannot be in the past
- check_out_date must be after check_in_date

---

## Endpoints

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/api/hotels/:hotel_id/availability` | Front Desk+ | Search available rooms |
| POST | `/api/hotels/:hotel_id/reservations` | Front Desk+ | Create reservation |
| GET | `/api/hotels/:hotel_id/reservations` | Front Desk+ | List reservations (filtered) |
| GET | `/api/hotels/:hotel_id/reservations/:id` | Front Desk+ | Get reservation details |
| POST | `/api/hotels/:hotel_id/reservations/:id/check-in` | Front Desk+ | Check in guest |
| POST | `/api/hotels/:hotel_id/reservations/:id/check-out` | Front Desk+ | Check out guest |
| POST | `/api/hotels/:hotel_id/reservations/:id/cancel` | Front Desk+ | Cancel reservation |

---

## GET /api/hotels/:hotel_id/availability

Search for rooms available across the entire date range.

**Query Parameters:**

| Param | Required | Description |
|-------|----------|-------------|
| `check_in` | Yes | Check-in date (`YYYY-MM-DD`) |
| `check_out` | Yes | Check-out date (`YYYY-MM-DD`) |

**Example:** `GET /api/hotels/:hotel_id/availability?check_in=2026-05-01&check_out=2026-05-03`

**Response 200:**
```json
{
  "success": true,
  "data": [
    {
      "id": "uuid",
      "hotel_id": "uuid",
      "room_number": "101",
      "room_type": "deluxe",
      "floor": 1,
      "status": "available",
      "price_per_night": 5000.00,
      "max_occupancy": 2,
      "description": "Sea view room",
      "is_active": true,
      "amenities": [],
      "created_at": "2026-04-25T10:00:00Z",
      "updated_at": "2026-04-25T10:00:00Z"
    }
  ]
}
```

A room appears only if it is available for **all** dates in the range.

---

## POST /api/hotels/:hotel_id/reservations

Create a new reservation. Uses DB transactions with row-level locking to prevent double booking.

**Request Body:**
```json
{
  "room_id": "uuid",              // required, must belong to hotel
  "guest_name": "John Doe",       // required, min 2
  "guest_phone": "+91-9876543210", // required
  "check_in_date": "2026-05-01",  // required, YYYY-MM-DD
  "check_out_date": "2026-05-03", // required, YYYY-MM-DD
  "notes": "Late arrival"         // optional
}
```

**Response 201:**
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "hotel_id": "uuid",
    "room_id": "uuid",
    "guest_name": "John Doe",
    "guest_phone": "+91-9876543210",
    "check_in_date": "2026-05-01T00:00:00Z",
    "check_out_date": "2026-05-03T00:00:00Z",
    "status": "reserved",
    "notes": "Late arrival",
    "room": { ... },
    "created_at": "2026-04-25T10:00:00Z",
    "updated_at": "2026-04-25T10:00:00Z"
  }
}
```

**Error 400:** `"room is not available for the selected dates"` — if any date in the range is already booked.

---

## GET /api/hotels/:hotel_id/reservations

List reservations with optional filters.

**Query Parameters:**

| Param | Required | Description |
|-------|----------|-------------|
| `status` | No | Filter by status: `reserved`, `checked_in`, `checked_out`, `cancelled` |
| `room_id` | No | Filter by room UUID |
| `date_from` | No | Check-in date >= this (`YYYY-MM-DD`) |
| `date_to` | No | Check-out date <= this (`YYYY-MM-DD`) |
| `page` | No | Page number (default: 1) |
| `per_page` | No | Items per page (default: 20, max: 100) |

**Example:** `GET /api/hotels/:hotel_id/reservations?status=reserved&page=1&per_page=10`

**Response 200:**
```json
{
  "success": true,
  "data": [
    {
      "id": "uuid",
      "hotel_id": "uuid",
      "room_id": "uuid",
      "guest_name": "John Doe",
      "guest_phone": "+91-9876543210",
      "check_in_date": "2026-05-01T00:00:00Z",
      "check_out_date": "2026-05-03T00:00:00Z",
      "status": "reserved",
      "notes": "",
      "room": { ... },
      "created_at": "2026-04-25T10:00:00Z",
      "updated_at": "2026-04-25T10:00:00Z"
    }
  ],
  "meta": {
    "page": 1,
    "per_page": 10,
    "total": 25,
    "total_pages": 3
  }
}
```

Results are ordered by check-in date descending.

---

## GET /api/hotels/:hotel_id/reservations/:id

Get a single reservation with room details.

**Response 200:**
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "hotel_id": "uuid",
    "room_id": "uuid",
    "guest_name": "John Doe",
    "guest_phone": "+91-9876543210",
    "check_in_date": "2026-05-01T00:00:00Z",
    "check_out_date": "2026-05-03T00:00:00Z",
    "status": "reserved",
    "notes": "",
    "room": { ... },
    "created_at": "2026-04-25T10:00:00Z",
    "updated_at": "2026-04-25T10:00:00Z"
  }
}
```

**Error 404:** `"reservation not found"`

---

## POST /api/hotels/:hotel_id/reservations/:id/check-in

Transition reservation from `reserved` → `checked_in`.

**Rules:**
- Status must be `reserved`
- Current date must be >= check-in date

**Request Body:** None

**Response 200:**
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "status": "checked_in",
    ...
  }
}
```

**Error 400:**
- `"only reserved bookings can be checked in"` — wrong status
- `"cannot check in before the check-in date"` — too early

---

## POST /api/hotels/:hotel_id/reservations/:id/check-out

Transition reservation from `checked_in` → `checked_out`.

**Rules:**
- Status must be `checked_in`

**Request Body:** None

**Response 200:**
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "status": "checked_out",
    ...
  }
}
```

**Error 400:** `"only checked-in bookings can be checked out"`

---

## POST /api/hotels/:hotel_id/reservations/:id/cancel

Transition reservation from `reserved` → `cancelled`. Frees up inventory so the room becomes bookable again.

**Rules:**
- Status must be `reserved` (cannot cancel after check-in)

**Request Body:** None

**Response 200:**
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "status": "cancelled",
    ...
  }
}
```

**Error 400:** `"only reserved bookings can be cancelled"`

---

## Double Booking Prevention

The system uses two layers to prevent double booking:

1. **Application layer:** `SELECT ... FOR UPDATE` row-level locking within a DB transaction checks for conflicts before inserting inventory
2. **Database layer:** `UNIQUE(room_id, date)` constraint on `room_inventories` table prevents duplicate entries even under race conditions

If a conflict is detected, the booking fails with: `"room is not available for the selected dates"`
