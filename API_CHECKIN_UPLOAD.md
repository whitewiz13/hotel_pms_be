# Hotel PMS — File Upload & Guest Identity Verification API

Base URL: `http://localhost:4001/api` | Auth: `Authorization: Bearer <token>`

All endpoints are scoped to `/api/hotels/:hotel_id/...`.

---

## Overview

These APIs enable:
1. **File uploads** — Upload ID documents (or any file) to the server
2. **Guest email at reservation** — Capture email when creating a reservation
3. **Identity verification at check-in** — Collect guest email, ID type, ID number, and uploaded document URL during check-in

---

## Endpoints

| Method | Path | Content-Type | Auth | Description |
|--------|------|-------------|------|-------------|
| POST | `/api/hotels/:hotel_id/uploads` | `multipart/form-data` | Front Desk+ | Upload a file |
| POST | `/api/hotels/:hotel_id/reservations` | `application/json` | Front Desk+ | Create reservation (now with `guest_email`) |
| POST | `/api/hotels/:hotel_id/reservations/:id/check-in` | `application/json` | Front Desk+ | Check in with identity details |

---

## POST /api/hotels/:hotel_id/uploads

Upload a file to the server. Returns a public URL.

**Permission:** `reservations:check_in`

**Content-Type:** `multipart/form-data`

**Form Fields:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `file` | file | Yes | The file to upload |
| `folder` | string | No | Subfolder name (default: `"uploads"`). Use `"id-documents"` for ID docs |

**Constraints:**
- Max file size: **5 MB**
- Allowed types: `image/jpeg`, `image/png`, `image/webp`, `application/pdf`

**Response 200:**
```json
{
  "success": true,
  "data": {
    "url": "http://localhost:4001/uploads/id-documents/550e8400-e29b-41d4-a716-446655440000.jpg"
  }
}
```

**Error 400:**
- `"file is required"` — no file in request
- `"file size exceeds maximum allowed size of 5MB"`
- `"file type image/gif is not allowed, accepted: jpeg, png, webp, pdf"`

---

## POST /api/hotels/:hotel_id/reservations (updated)

Create a reservation. Now accepts an optional `guest_email` field.

**Request Body:**
```json
{
  "room_id": "uuid",
  "guest_name": "John Doe",
  "guest_phone": "+91-9876543210",
  "guest_email": "john@example.com",
  "check_in_date": "2026-05-01",
  "check_out_date": "2026-05-03",
  "notes": "Late arrival"
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `room_id` | string (uuid) | Yes | Room to book |
| `guest_name` | string | Yes | Guest name (min 2 chars) |
| `guest_phone` | string | Yes | Guest phone number |
| `guest_email` | string | No | Guest email address |
| `check_in_date` | string | Yes | `YYYY-MM-DD` format |
| `check_out_date` | string | Yes | `YYYY-MM-DD` format |
| `notes` | string | No | Booking notes |

**Response 201:**
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
    "guest_email": "john@example.com",
    "check_in_date": "2026-05-01T00:00:00Z",
    "check_out_date": "2026-05-03T00:00:00Z",
    "status": "reserved",
    "notes": "Late arrival",
    "room": { "..." },
    "guest": {
      "id": "uuid",
      "name": "John Doe",
      "email": "john@example.com",
      "phone": "+91-9876543210",
      "id_type": "",
      "id_number": "",
      "id_document_url": ""
    },
    "created_at": "2026-04-29T10:00:00Z",
    "updated_at": "2026-04-29T10:00:00Z"
  }
}
```

---

## POST /api/hotels/:hotel_id/reservations/:id/check-in (updated)

Transition reservation from `reserved` → `checked_in`. Now accepts an **optional** JSON body with guest email and identity verification details.

**Permission:** `reservations:check_in`

**Rules:**
- Status must be `reserved`
- Current date must be >= check-in date
- Body is optional — check-in still works with an empty body (backward compatible)

**Request Body (optional):**
```json
{
  "guest_email": "john@example.com",
  "id_type": "aadhaar",
  "id_number": "1234-5678-9012",
  "id_document_url": "http://localhost:4001/uploads/id-documents/550e8400-e29b-41d4-a716-446655440000.jpg"
}
```

| Field | Type | Required | Validation |
|-------|------|----------|------------|
| `guest_email` | string | No | Must be valid email format |
| `id_type` | string | No | One of: `aadhaar`, `pan`, `passport`, `driving_license`, `voter_id` |
| `id_number` | string | No | The document number |
| `id_document_url` | string | No | URL returned from the upload endpoint |

**Response 200:**
```json
{
  "success": true,
  "data": {
    "reservation": {
      "id": "uuid",
      "hotel_id": "uuid",
      "room_id": "uuid",
      "guest_id": "uuid",
      "guest_name": "John Doe",
      "guest_phone": "+91-9876543210",
      "guest_email": "john@example.com",
      "check_in_date": "2026-05-01T00:00:00Z",
      "check_out_date": "2026-05-03T00:00:00Z",
      "status": "checked_in",
      "notes": "",
      "room": { "..." },
      "guest": {
        "id": "uuid",
        "name": "John Doe",
        "email": "john@example.com",
        "phone": "+91-9876543210",
        "id_type": "aadhaar",
        "id_number": "1234-5678-9012",
        "id_document_url": "http://localhost:4001/uploads/id-documents/550e8400.jpg"
      }
    },
    "access_pin": "482917"
  }
}
```

**Error 400:**
- `"only reserved bookings can be checked in"` — wrong status
- `"cannot check in before the check-in date"` — too early

---

## Frontend Integration Flow

### During Reservation
1. Collect `guest_email` in the reservation form
2. Send it in `POST /reservations` body

### During Check-In
1. User clicks **Check In** → open a modal/dialog
2. Show form with:
   - **Email** — pre-fill from `reservation.guest_email` or `reservation.guest.email` if available
   - **ID Type** — dropdown: `Aadhaar | PAN | Passport | Driving License | Voter ID`
   - **ID Number** — text input
   - **ID Document** — file picker (image or PDF)
3. On submit:
   - If file selected → call `POST /hotels/:hotel_id/uploads` with `folder=id-documents` → get URL
   - Call `POST /hotels/:hotel_id/reservations/:id/check-in` with all fields
4. Show the returned `access_pin` to the user

### Guest Object Shape (returned in reservation responses)
```json
{
  "id": "uuid",
  "hotel_id": "uuid",
  "name": "John Doe",
  "email": "john@example.com",
  "phone": "+91-9876543210",
  "id_type": "aadhaar",
  "id_number": "1234-5678-9012",
  "id_document_url": "http://localhost:4001/uploads/id-documents/550e8400.jpg",
  "created_at": "2026-04-29T10:00:00Z",
  "updated_at": "2026-04-29T10:00:00Z"
}
```

### ID Type Values

| Value | Display Label |
|-------|--------------|
| `aadhaar` | Aadhaar |
| `pan` | PAN |
| `passport` | Passport |
| `driving_license` | Driving License |
| `voter_id` | Voter ID |
