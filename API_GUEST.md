# Guest Portal API

All guest endpoints require a guest JWT token obtained via guest login.
No `hotel_id` is needed in the URL — everything is resolved from the token.

**Base URL:** `/api/guest`
**Auth:** `Authorization: Bearer <guest_token>`

---

## Authentication

### `POST /api/auth/guest/login`

Login as a guest using room number and PIN (provided at check-in).

**Request:**
```json
{
  "room_number": "101",
  "pin": "482619",
  "hotel_id": "uuid"
}
```

**Response `200`:**
```json
{
  "success": true,
  "data": {
    "token": "eyJhbGciOi...",
    "room_number": "101",
    "room_type": "deluxe"
  }
}
```

---

## Dashboard

### `GET /api/guest/dashboard`

Get an overview of the guest's current stay — order stats, activity stats, and spend totals.

**Response `200`:**
```json
{
  "success": true,
  "data": {
    "room_number": "101",
    "guest_name": "John Doe",
    "check_in_date": "2026-04-25",
    "check_out_date": "2026-04-28",
    "order_stats": {
      "pending": 1,
      "preparing": 1,
      "delivered": 3,
      "cancelled": 0
    },
    "total_orders": 5,
    "order_spend": 85.50,
    "activity_stats": {
      "pending": 1,
      "confirmed": 1,
      "completed": 2
    },
    "total_activities": 4,
    "activity_spend": 120.00
  }
}
```

**Notes:**
- `order_stats` / `activity_stats` — counts grouped by status (keys are only present if count > 0)
- `order_spend` / `activity_spend` — total for non-cancelled items only

---

## Reservation

### `GET /api/guest/reservation`

Get the guest's current checked-in reservation.

**Response `200`:**
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "hotel_id": "uuid",
    "room_id": "uuid",
    "guest_id": "uuid",
    "guest_name": "John Doe",
    "guest_phone": "+91-9876543210",
    "check_in_date": "2026-04-25T00:00:00Z",
    "check_out_date": "2026-04-28T00:00:00Z",
    "status": "checked_in",
    "notes": "",
    "room": { "id": "uuid", "room_number": "101", "room_type": "deluxe", "floor": 1 },
    "guest": { "id": "uuid", "name": "John Doe", "email": "john@example.com", "phone": "+91-9876543210" }
  }
}
```

---

## Menu

### `GET /api/guest/menu`

Browse the hotel's food & beverage menu. Paginated.

**Query params:** `page` (default 1), `per_page` (default 20)

**Response `200`:**
```json
{
  "success": true,
  "data": [
    {
      "id": "uuid",
      "name": "Club Sandwich",
      "description": "Triple-decker with fries",
      "price": 12.50,
      "category": "main_course",
      "is_available": true
    }
  ],
  "meta": { "page": 1, "per_page": 20, "total": 15, "total_pages": 1 }
}
```

**Categories:** `appetizer`, `main_course`, `dessert`, `beverage`, `snack`

---

## Orders (Room Service)

### `POST /api/guest/orders`

Place a food order from the menu. The order appears in the hotel staff's order queue.

**Request:**
```json
{
  "items": [
    { "menu_item_id": "uuid", "quantity": 2, "notes": "No onions" },
    { "menu_item_id": "uuid", "quantity": 1 }
  ],
  "notes": "Please deliver to the pool area"
}
```

**Response `201`:**
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "hotel_id": "uuid",
    "room_id": "uuid",
    "reservation_id": "uuid",
    "guest_id": "uuid",
    "guest_name": "John Doe",
    "status": "pending",
    "total_amount": 37.50,
    "notes": "Please deliver to the pool area",
    "items": [
      {
        "id": "uuid",
        "menu_item_id": "uuid",
        "quantity": 2,
        "unit_price": 12.50,
        "subtotal": 25.00,
        "notes": "No onions",
        "menu_item": { "id": "uuid", "name": "Club Sandwich", "price": 12.50, "category": "main_course" }
      }
    ],
    "room": { "id": "uuid", "room_number": "101" },
    "guest": { "id": "uuid", "name": "John Doe", "phone": "+91-9876543210" }
  }
}
```

**Validations:**
- At least 1 item required
- Each `menu_item_id` must be a valid UUID of an available menu item
- `quantity` must be ≥ 1

**Order lifecycle (managed by hotel staff):**
`pending` → `preparing` → `ready` → `delivered`
Any state except `delivered` can move to `cancelled`.

---

### `GET /api/guest/orders`

List all orders for the current stay. Paginated, newest first.

**Query params:** `page`, `per_page`

**Response `200`:**
```json
{
  "success": true,
  "data": [
    {
      "id": "uuid",
      "status": "delivered",
      "total_amount": 37.50,
      "notes": "",
      "guest_name": "John Doe",
      "items": [...],
      "room": {...},
      "guest": {...},
      "created_at": "2026-04-25T14:30:00Z"
    }
  ],
  "meta": { "page": 1, "per_page": 20, "total": 5, "total_pages": 1 }
}
```

---

## Activities

### `GET /api/guest/activities`

Browse available hotel activities/services. Paginated.

**Query params:** `page`, `per_page`

**Response `200`:**
```json
{
  "success": true,
  "data": [
    {
      "id": "uuid",
      "name": "Airport Cab",
      "description": "One-way airport transfer",
      "price": 35.00,
      "category": "cab",
      "is_available": true
    }
  ],
  "meta": { "page": 1, "per_page": 20, "total": 8, "total_pages": 1 }
}
```

**Categories:** `cab`, `spa`, `tour`, `laundry`, `other`

---

### `POST /api/guest/activity-bookings`

Book a hotel activity/service.

**Request:**
```json
{
  "activity_id": "uuid",
  "scheduled_at": "2026-04-26T10:00:00Z",
  "notes": "Pick up from lobby"
}
```

- `scheduled_at` is optional (RFC 3339 format)

**Response `201`:**
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "hotel_id": "uuid",
    "room_id": "uuid",
    "reservation_id": "uuid",
    "activity_id": "uuid",
    "guest_id": "uuid",
    "guest_name": "John Doe",
    "scheduled_at": "2026-04-26T10:00:00Z",
    "status": "pending",
    "amount": 35.00,
    "notes": "Pick up from lobby",
    "activity": { "id": "uuid", "name": "Airport Cab", "price": 35.00, "category": "cab" },
    "room": { "id": "uuid", "room_number": "101" },
    "guest": { "id": "uuid", "name": "John Doe", "phone": "+91-9876543210" }
  }
}
```

**Booking lifecycle (managed by hotel staff):**
`pending` → `confirmed` → `completed`
`pending` or `confirmed` can move to `cancelled`.

---

### `GET /api/guest/activity-bookings`

List all activity bookings for the current stay. Paginated, newest first.

**Query params:** `page`, `per_page`

**Response `200`:**
```json
{
  "success": true,
  "data": [
    {
      "id": "uuid",
      "status": "confirmed",
      "amount": 35.00,
      "scheduled_at": "2026-04-26T10:00:00Z",
      "guest_name": "John Doe",
      "activity": {...},
      "room": {...},
      "guest": {...},
      "created_at": "2026-04-25T18:00:00Z"
    }
  ],
  "meta": { "page": 1, "per_page": 20, "total": 4, "total_pages": 1 }
}
```

---

## Error Responses

All errors follow the same format:

```json
{
  "success": false,
  "error": "description of what went wrong"
}
```

| Status | Meaning |
|--------|---------|
| `400`  | Bad request — validation error or business rule violation |
| `401`  | Unauthorized — missing/invalid/expired guest token |
| `404`  | Not found — no active reservation for this room |

---

## Quick Reference

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/auth/guest/login` | Guest login with room number + PIN |
| `GET` | `/api/guest/dashboard` | Stay overview with order/activity stats |
| `GET` | `/api/guest/reservation` | Current reservation details |
| `GET` | `/api/guest/menu` | Browse food menu |
| `GET` | `/api/guest/activities` | Browse available activities |
| `POST` | `/api/guest/orders` | Place a food order |
| `GET` | `/api/guest/orders` | List my orders |
| `POST` | `/api/guest/activity-bookings` | Book an activity |
| `GET` | `/api/guest/activity-bookings` | List my activity bookings |
