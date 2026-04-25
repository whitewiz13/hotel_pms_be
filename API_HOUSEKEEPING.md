# Hotel PMS — Housekeeping API Reference

Base URL: `http://localhost:4001/api` | Content-Type: `application/json` | Auth: `Authorization: Bearer <token>`

All housekeeping endpoints are scoped to `/api/hotels/:hotel_id/housekeeping`.

---

## Object Shape

**HousekeepingTask:**
```json
{
  "id": "uuid",
  "hotel_id": "uuid",
  "room_id": "uuid",
  "assigned_to_id": "uuid | null",
  "assigned_by_id": "uuid",
  "status": "pending",
  "priority": "normal",
  "notes": "",
  "room": { ... },
  "assigned_to": { "id": "uuid", "name": "...", "email": "...", "role": "housekeeping" },
  "assigned_by": { "id": "uuid", "name": "...", "email": "...", "role": "hotel_admin" },
  "created_at": "2026-04-25T10:00:00Z",
  "updated_at": "2026-04-25T10:00:00Z"
}
```

**Status values:** `pending`, `in_progress`, `completed`

**Priority values:** `low`, `normal`, `high`, `urgent`

**State transitions:**
```
pending → completed
in_progress → completed
```

---

## Access Roles

| Action | Minimum Role |
|--------|-------------|
| Assign task | Front Desk+ (`super_admin`, `hotel_admin`, `manager`, `front_desk`) |
| List / Get / Complete task | Housekeeping+ (`super_admin`, `hotel_admin`, `manager`, `front_desk`, `housekeeping`) |

---

## Endpoints

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/api/hotels/:hotel_id/housekeeping` | Front Desk+ | Assign room to housekeeping |
| GET | `/api/hotels/:hotel_id/housekeeping` | Housekeeping+ | List tasks (priority-ordered) |
| GET | `/api/hotels/:hotel_id/housekeeping/:id` | Housekeeping+ | Get task details |
| POST | `/api/hotels/:hotel_id/housekeeping/:id/complete` | Housekeeping+ | Mark task as done |

---

## POST /api/hotels/:hotel_id/housekeeping

Create a housekeeping task for a room. Sets the room status to `cleaning`.

**Request Body:**
```json
{
  "room_id": "uuid",          // required, must belong to hotel
  "assigned_to_id": "uuid",   // optional, must belong to hotel
  "priority": "high",         // required: low | normal | high | urgent
  "notes": "Deep clean needed" // optional
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
    "assigned_to_id": "uuid",
    "assigned_by_id": "uuid",
    "status": "pending",
    "priority": "high",
    "notes": "Deep clean needed",
    "room": { ... },
    "assigned_to": { ... },
    "assigned_by": { ... },
    "created_at": "2026-04-25T10:00:00Z",
    "updated_at": "2026-04-25T10:00:00Z"
  }
}
```

**Error 400:**
- `"room not found"` — invalid room or wrong hotel
- `"room already has an active housekeeping task"` — pending/in-progress task exists
- `"assigned user not found"` — invalid assigned_to_id
- `"assigned user does not belong to this hotel"` — user from another hotel

---

## GET /api/hotels/:hotel_id/housekeeping

List housekeeping tasks ordered by priority (urgent → high → normal → low), then by creation date ascending.

**Query Parameters:**

| Param | Required | Description |
|-------|----------|-------------|
| `status` | No | Filter: `pending`, `in_progress`, `completed` |
| `priority` | No | Filter: `low`, `normal`, `high`, `urgent` |
| `assigned_to_id` | No | Filter by housekeeper UUID |
| `room_id` | No | Filter by room UUID |
| `page` | No | Page number (default: 1) |
| `per_page` | No | Items per page (default: 20, max: 100) |

**Example:** `GET /api/hotels/:hotel_id/housekeeping?status=pending&priority=urgent&page=1`

**Response 200:**
```json
{
  "success": true,
  "data": [
    {
      "id": "uuid",
      "hotel_id": "uuid",
      "room_id": "uuid",
      "assigned_to_id": null,
      "assigned_by_id": "uuid",
      "status": "pending",
      "priority": "urgent",
      "notes": "VIP guest arriving soon",
      "room": { "id": "uuid", "room_number": "301", "room_type": "suite", ... },
      "assigned_to": null,
      "assigned_by": { "id": "uuid", "name": "Admin", ... },
      "created_at": "2026-04-25T10:00:00Z",
      "updated_at": "2026-04-25T10:00:00Z"
    }
  ],
  "meta": {
    "page": 1,
    "per_page": 20,
    "total": 5,
    "total_pages": 1
  }
}
```

---

## GET /api/hotels/:hotel_id/housekeeping/:id

Get a single housekeeping task with room and user details.

**Response 200:**
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "hotel_id": "uuid",
    "room_id": "uuid",
    "assigned_to_id": "uuid",
    "assigned_by_id": "uuid",
    "status": "pending",
    "priority": "high",
    "notes": "Deep clean needed",
    "room": { ... },
    "assigned_to": { ... },
    "assigned_by": { ... },
    "created_at": "2026-04-25T10:00:00Z",
    "updated_at": "2026-04-25T10:00:00Z"
  }
}
```

**Error 404:** `"housekeeping task not found"`

---

## POST /api/hotels/:hotel_id/housekeeping/:id/complete

Mark a housekeeping task as completed. Sets the room status to `available`.

**Rules:**
- Task must not already be `completed`
- If task is assigned to a specific housekeeper, only that user can complete it
- If task is unassigned, any housekeeping+ user can complete it (and becomes the assignee)

**Request Body (optional):**
```json
{
  "notes": "Room cleaned and restocked"  // optional, replaces existing notes
}
```

**Response 200:**
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "status": "completed",
    "room": { "status": "available", ... },
    ...
  }
}
```

**Error 400:**
- `"housekeeping task not found"` — invalid id or wrong hotel
- `"task is already completed"` — cannot complete twice
- `"task is assigned to another housekeeper"` — wrong user

---

## Typical Workflow

```
1. Guest checks out
   └── Room status → "dirty" (automatic)

2. Admin/front desk assigns housekeeping
   POST /housekeeping { room_id, priority: "high" }
   └── Room status → "cleaning"
   └── Task created with status "pending"

3. Housekeeper views their tasks
   GET /housekeeping?status=pending
   └── Tasks sorted: urgent > high > normal > low

4. Housekeeper completes the task
   POST /housekeeping/:id/complete
   └── Task status → "completed"
   └── Room status → "available"
```
