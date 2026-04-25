# Room Service, Activities & Billing API

Base URL: `/api/hotels/:hotel_id`

All endpoints require `Authorization: Bearer <token>` header.

---

## Roles

| Role | Description |
|------|-------------|
| `room_service` | **NEW** — Kitchen/delivery staff. Can view and update order statuses |
| Existing roles | `super_admin`, `hotel_admin`, `manager`, `front_desk`, `housekeeping`, `staff` |

---

## 1. Menu (Room Service Menu)

Hotels define menu items that guests can order.

### Create Menu Item

`POST /menu`

**Roles:** `hotel_admin`, `manager`

```json
{
  "name": "Chicken Caesar Salad",
  "description": "Fresh romaine lettuce with grilled chicken",
  "price": 15.99,
  "category": "main_course"
}
```

**Category options:** `appetizer`, `main_course`, `dessert`, `beverage`, `snack`

**Response:** `201 Created`
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "hotel_id": "uuid",
    "name": "Chicken Caesar Salad",
    "description": "Fresh romaine lettuce with grilled chicken",
    "price": 15.99,
    "category": "main_course",
    "is_available": true,
    "created_at": "2026-04-25T10:00:00Z",
    "updated_at": "2026-04-25T10:00:00Z"
  }
}
```

### List Menu Items

`GET /menu`

**Roles:** Any authenticated user (including guests)

**Query params:**
| Param | Type | Description |
|-------|------|-------------|
| `category` | string | Filter by category |
| `page` | int | Page number (default: 1) |
| `per_page` | int | Items per page (default: 20, max: 100) |

**Response:** `200 OK` — Paginated list

### Get Menu Item

`GET /menu/:id`

**Roles:** Any authenticated user

**Response:** `200 OK`

### Update Menu Item

`PUT /menu/:id`

**Roles:** `hotel_admin`, `manager`

```json
{
  "name": "Updated Name",
  "price": 18.99,
  "is_available": false
}
```

All fields optional.

**Response:** `200 OK`

### Delete Menu Item

`DELETE /menu/:id`

**Roles:** `hotel_admin`, `manager`

**Response:** `200 OK` — `{ "success": true, "message": "menu item deleted successfully" }`

---

## 2. Orders (Room Service Orders)

Guests place orders linked to their room and active reservation.

### Create Order

`POST /orders`

**Roles:** `front_desk` and above

**Prerequisite:** Reservation must be in `checked_in` status.

```json
{
  "room_id": "uuid",
  "reservation_id": "uuid",
  "guest_name": "John Doe",
  "items": [
    {
      "menu_item_id": "uuid",
      "quantity": 2,
      "notes": "No onions"
    },
    {
      "menu_item_id": "uuid",
      "quantity": 1
    }
  ],
  "notes": "Deliver to room"
}
```

**Response:** `201 Created`
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "hotel_id": "uuid",
    "room_id": "uuid",
    "reservation_id": "uuid",
    "guest_name": "John Doe",
    "status": "pending",
    "total_amount": 47.97,
    "notes": "Deliver to room",
    "items": [
      {
        "id": "uuid",
        "menu_item_id": "uuid",
        "quantity": 2,
        "unit_price": 15.99,
        "subtotal": 31.98,
        "notes": "No onions",
        "menu_item": { "name": "Chicken Caesar Salad", "..." : "..." }
      }
    ],
    "room": { "room_number": "101", "..." : "..." },
    "created_at": "2026-04-25T12:00:00Z"
  }
}
```

### List Orders

`GET /orders`

**Roles:** `front_desk` and above, `room_service`

**Query params:**
| Param | Type | Description |
|-------|------|-------------|
| `status` | string | `pending`, `preparing`, `ready`, `delivered`, `cancelled` |
| `room_id` | uuid | Filter by room |
| `reservation_id` | uuid | Filter by reservation |
| `assigned_to_id` | uuid | Filter by assigned staff |
| `page` | int | Page number |
| `per_page` | int | Items per page |

**Response:** `200 OK` — Paginated list

### Get Order

`GET /orders/:id`

**Roles:** `front_desk` and above, `room_service`

**Response:** `200 OK`

### Update Order Status

`POST /orders/:id/status`

**Roles:** `front_desk` and above, `room_service`

```json
{
  "status": "preparing"
}
```

**Valid transitions:**
| From | To |
|------|----|
| `pending` | `preparing`, `cancelled` |
| `preparing` | `ready`, `cancelled` |
| `ready` | `delivered` |

**Response:** `200 OK`

### Assign Order to Staff

`POST /orders/:id/assign`

**Roles:** `front_desk` and above

```json
{
  "assigned_to_id": "uuid"
}
```

**Response:** `200 OK`

---

## 3. Activities

Hotels define activities (cab, spa, tour, etc.) that can be booked for guests.

### Create Activity

`POST /activities`

**Roles:** `hotel_admin`, `manager`

```json
{
  "name": "Airport Cab",
  "description": "One-way cab to airport",
  "price": 45.00,
  "category": "cab"
}
```

**Category options:** `cab`, `spa`, `tour`, `laundry`, `other`

**Response:** `201 Created`

### List Activities

`GET /activities`

**Roles:** Any authenticated user (including guests)

**Query params:**
| Param | Type | Description |
|-------|------|-------------|
| `category` | string | Filter by category |
| `page` | int | Page number |
| `per_page` | int | Items per page |

**Response:** `200 OK` — Paginated list

### Get Activity

`GET /activities/:id`

**Roles:** Any authenticated user

**Response:** `200 OK`

### Update Activity

`PUT /activities/:id`

**Roles:** `hotel_admin`, `manager`

```json
{
  "price": 50.00,
  "is_available": false
}
```

**Response:** `200 OK`

### Delete Activity

`DELETE /activities/:id`

**Roles:** `hotel_admin`, `manager`

**Response:** `200 OK`

---

## 4. Activity Bookings

Book activities for guests linked to their room and reservation.

### Create Activity Booking

`POST /activity-bookings`

**Roles:** `front_desk` and above

**Prerequisite:** Reservation must be in `checked_in` status.

```json
{
  "room_id": "uuid",
  "reservation_id": "uuid",
  "activity_id": "uuid",
  "guest_name": "John Doe",
  "scheduled_at": "2026-04-26T14:00:00Z",
  "notes": "Pick up from lobby"
}
```

`scheduled_at` is optional, format: RFC3339.

**Response:** `201 Created`
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "hotel_id": "uuid",
    "room_id": "uuid",
    "reservation_id": "uuid",
    "activity_id": "uuid",
    "guest_name": "John Doe",
    "scheduled_at": "2026-04-26T14:00:00Z",
    "status": "pending",
    "amount": 45.00,
    "notes": "Pick up from lobby",
    "activity": { "name": "Airport Cab", "..." : "..." },
    "room": { "room_number": "101", "..." : "..." }
  }
}
```

### List Activity Bookings

`GET /activity-bookings`

**Roles:** `front_desk` and above

**Query params:**
| Param | Type | Description |
|-------|------|-------------|
| `status` | string | `pending`, `confirmed`, `completed`, `cancelled` |
| `room_id` | uuid | Filter by room |
| `reservation_id` | uuid | Filter by reservation |
| `activity_id` | uuid | Filter by activity |
| `page` | int | Page number |
| `per_page` | int | Items per page |

**Response:** `200 OK` — Paginated list

### Get Activity Booking

`GET /activity-bookings/:id`

**Roles:** `front_desk` and above

**Response:** `200 OK`

### Update Activity Booking Status

`POST /activity-bookings/:id/status`

**Roles:** `front_desk` and above

```json
{
  "status": "confirmed"
}
```

**Valid transitions:**
| From | To |
|------|----|
| `pending` | `confirmed`, `cancelled` |
| `confirmed` | `completed`, `cancelled` |

**Response:** `200 OK`

---

## 5. Billing

Bills are generated for reservations at checkout time. They aggregate room charges, room service orders, activity bookings, and upfront payments.

### Generate Bill

`POST /reservations/:id/bill`

**Roles:** `front_desk` and above

**Prerequisite:** Reservation must be `checked_in` or `checked_out`. Only one bill per reservation.

```json
{
  "upfront_paid": 500.00,
  "tax_rate": 18,
  "notes": "Corporate booking discount applied"
}
```

All fields optional. `tax_rate` is a percentage (e.g., `18` = 18%).

**Response:** `201 Created`
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "hotel_id": "uuid",
    "reservation_id": "uuid",
    "room_id": "uuid",
    "guest_name": "John Doe",
    "room_charges": 600.00,
    "upfront_paid": 500.00,
    "room_service_total": 47.97,
    "activity_total": 45.00,
    "subtotal": 192.97,
    "tax_rate": 18,
    "tax_amount": 34.73,
    "total_amount": 227.70,
    "status": "pending",
    "notes": "Corporate booking discount applied",
    "line_items": [
      {
        "id": "uuid",
        "type": "room_charge",
        "description": "Room 101 - deluxe (3 night(s) × $200.00)",
        "amount": 600.00
      },
      {
        "id": "uuid",
        "type": "room_service",
        "description": "Room Service Order #a1b2c3d4",
        "amount": 47.97,
        "reference_id": "order-uuid"
      },
      {
        "id": "uuid",
        "type": "activity",
        "description": "Activity: Airport Cab",
        "amount": 45.00,
        "reference_id": "booking-uuid"
      },
      {
        "id": "uuid",
        "type": "upfront_payment",
        "description": "Upfront Payment",
        "amount": -500.00
      }
    ],
    "room": { "room_number": "101" },
    "reservation": { "check_in_date": "2026-04-23", "check_out_date": "2026-04-26" }
  }
}
```

### Get Bill by Reservation

`GET /reservations/:id/bill`

**Roles:** `front_desk` and above

**Response:** `200 OK`

### List Bills

`GET /bills`

**Roles:** `front_desk` and above

**Query params:**
| Param | Type | Description |
|-------|------|-------------|
| `status` | string | `pending`, `paid` |
| `reservation_id` | uuid | Filter by reservation |
| `page` | int | Page number |
| `per_page` | int | Items per page |

**Response:** `200 OK` — Paginated list

### Get Bill by ID

`GET /bills/:id`

**Roles:** `front_desk` and above

**Response:** `200 OK`

### Mark Bill as Paid

`POST /bills/:id/pay`

**Roles:** `front_desk` and above

**Response:** `200 OK`

---

## Typical Checkout Flow

1. **During stay** — Staff creates orders (`POST /orders`) and activity bookings (`POST /activity-bookings`) as guest requests them
2. **Before checkout** — Generate bill: `POST /reservations/:id/bill` with `upfront_paid` and `tax_rate`
3. **Review bill** — `GET /reservations/:id/bill` to view full breakdown
4. **Collect payment** — `POST /bills/:id/pay` to mark as paid
5. **Complete checkout** — `POST /reservations/:id/check-out`

---

## Bill Calculation Formula

```
room_charges     = price_per_night × number_of_nights
room_service     = sum of all non-cancelled order totals
activity_total   = sum of all non-cancelled activity booking amounts
subtotal         = room_charges + room_service + activity_total - upfront_paid
tax_amount       = subtotal × (tax_rate / 100)
total_amount     = subtotal + tax_amount
```

---

## Error Responses

All errors follow the standard format:

```json
{
  "success": false,
  "error": "descriptive error message"
}
```

| Status | Meaning |
|--------|---------|
| `400` | Bad request / validation error |
| `401` | Unauthorized |
| `403` | Forbidden (insufficient role) |
| `404` | Resource not found |
| `500` | Internal server error |
