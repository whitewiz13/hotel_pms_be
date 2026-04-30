# Subscription Plans & Feature Gating

## Overview

Hotels are assigned a subscription plan that controls resource limits and feature access. Plans are managed by super admins.

---

## Plans

| | **Free** | **Basic** | **Pro** |
|--|----------|-----------|---------|
| **ID** | `free` | `basic` | `pro` |

### Resource Limits

| Resource | Free | Basic | Pro |
|----------|------|-------|-----|
| Rooms | 5 | 25 | Unlimited (-1) |
| Staff members | 3 | 10 | Unlimited (-1) |
| Reservations / month | 20 | 200 | Unlimited (-1) |
| File storage | 50 MB | 500 MB | 5 GB |

### Feature Access

| Feature | Free | Basic | Pro |
|---------|------|-------|-----|
| Reservations & Check-in/out | ✅ | ✅ | ✅ |
| Housekeeping | ✅ | ✅ | ✅ |
| Room management | ✅ | ✅ | ✅ |
| Amenities | ✅ | ✅ | ✅ |
| Billing | ✅ | ✅ | ✅ |
| Dashboard stats | ✅ | ✅ | ✅ |
| Menu & Room Service | ❌ | ✅ | ✅ |
| Activities & Bookings | ❌ | ✅ | ✅ |
| Push Notifications (FCM) | ❌ | ✅ | ✅ |
| Analytics (summary) | ❌ | ✅ | ✅ |
| Guest ID Uploads | ❌ | ✅ | ✅ |
| Guest Portal (self-service) | ❌ | ❌ | ✅ |
| Advanced Analytics (trends) | ❌ | ❌ | ✅ |
| Custom Roles & Permissions | ❌ | ❌ | ✅ |

---

## API Endpoints

### List All Plans

```
GET /api/plans
Authorization: Bearer <token>
```

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": "free",
      "name": "Free",
      "max_rooms": 5,
      "max_staff": 3,
      "max_reservations_month": 20,
      "max_storage_mb": 50,
      "feature_room_service": false,
      "feature_activities": false,
      "feature_guest_portal": false,
      "feature_notifications": false,
      "feature_analytics": false,
      "feature_adv_analytics": false,
      "feature_custom_roles": false,
      "feature_guest_uploads": false
    },
    {
      "id": "basic",
      "name": "Basic",
      "max_rooms": 25,
      "max_staff": 10,
      "max_reservations_month": 200,
      "max_storage_mb": 500,
      "feature_room_service": true,
      "feature_activities": true,
      "feature_guest_portal": false,
      "feature_notifications": true,
      "feature_analytics": true,
      "feature_adv_analytics": false,
      "feature_custom_roles": false,
      "feature_guest_uploads": true
    },
    {
      "id": "pro",
      "name": "Pro",
      "max_rooms": -1,
      "max_staff": -1,
      "max_reservations_month": -1,
      "max_storage_mb": 5120,
      "feature_room_service": true,
      "feature_activities": true,
      "feature_guest_portal": true,
      "feature_notifications": true,
      "feature_analytics": true,
      "feature_adv_analytics": true,
      "feature_custom_roles": true,
      "feature_guest_uploads": true
    }
  ]
}
```

---

### Get Hotel Subscription

```
GET /api/hotels/:hotel_id/subscription
Authorization: Bearer <token>
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "hotel_id": "uuid",
    "plan_id": "basic",
    "plan": {
      "id": "basic",
      "name": "Basic",
      "max_rooms": 25,
      "max_staff": 10,
      "max_reservations_month": 200,
      "max_storage_mb": 500,
      "feature_room_service": true,
      "feature_activities": true,
      "feature_guest_portal": false,
      "feature_notifications": true,
      "feature_analytics": true,
      "feature_adv_analytics": false,
      "feature_custom_roles": false,
      "feature_guest_uploads": true
    },
    "status": "active",
    "expires_at": null,
    "created_at": "2026-04-30T10:00:00Z",
    "updated_at": "2026-04-30T10:00:00Z"
  }
}
```

---

### Get Hotel Plan Usage

```
GET /api/hotels/:hotel_id/subscription/usage
Authorization: Bearer <token>
```

**Response:**
```json
{
  "success": true,
  "data": {
    "plan": {
      "id": "free",
      "name": "Free",
      "max_rooms": 5,
      "max_staff": 3,
      "max_reservations_month": 20,
      "max_storage_mb": 50,
      "feature_room_service": false,
      "feature_activities": false,
      "feature_guest_portal": false,
      "feature_notifications": false,
      "feature_analytics": false,
      "feature_adv_analytics": false,
      "feature_custom_roles": false,
      "feature_guest_uploads": false
    },
    "usage": {
      "rooms": 3,
      "staff": 2,
      "reservations_month": 8
    },
    "limits": {
      "rooms": 5,
      "staff": 3,
      "reservations_month": 20,
      "storage_mb": 50
    }
  }
}
```

---

### Change Hotel Plan (Super Admin Only)

```
PUT /api/hotels/:hotel_id/subscription
Authorization: Bearer <super_admin_token>
Content-Type: application/json
```

**Request:**
```json
{
  "plan_id": "pro"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "hotel_id": "uuid",
    "plan_id": "pro",
    "plan": { ... },
    "status": "active",
    "expires_at": null,
    "created_at": "2026-04-30T10:00:00Z",
    "updated_at": "2026-04-30T12:00:00Z"
  }
}
```

---

## Error Responses

### Resource Limit Reached (400)

When a hotel hits a plan limit (rooms, staff, reservations):

```json
{
  "success": false,
  "error": "room limit reached (5). Upgrade your plan to add more rooms"
}
```

```json
{
  "success": false,
  "error": "staff limit reached (3). Upgrade your plan to add more staff"
}
```

```json
{
  "success": false,
  "error": "monthly reservation limit reached (20). Upgrade your plan for more reservations"
}
```

### Feature Not Available (403)

When accessing a gated feature not included in the plan:

```json
{
  "success": false,
  "error": "this feature is not available on your current plan. Please upgrade to access it"
}
```

---

## Feature Gate Mapping

These routes are gated by plan features:

| Feature Key | Gated Routes |
|-------------|--------------|
| `room_service` | `/menu/*`, `/orders/*` |
| `activities` | `/activities/*`, `/activity-bookings/*` |
| `guest_portal` | `/guest/*` (entire guest self-service) |
| `analytics` | `/analytics/*` |
| `adv_analytics` | `/analytics/occupancy`, `/analytics/revenue`, `/analytics/reservations`, `/analytics/room-types` |
| `custom_roles` | `POST/PUT/DELETE /roles` (read is always allowed) |
| `guest_uploads` | `POST /uploads` |

---

## Behavior Notes

- New hotels are **automatically assigned the Free plan** on creation.
- Super admin can change any hotel's plan via `PUT /hotels/:hotel_id/subscription`.
- Plan changes take effect **immediately** — no grace period.
- If a hotel is downgraded, **existing data is preserved** but they cannot create new resources beyond the limit.
- `-1` means unlimited for that resource.
- Feature gates return 403 — the FE should hide UI elements based on the plan features from the subscription endpoint.
- The `/plans` endpoint can be used to build a pricing/comparison page.
- The `/subscription/usage` endpoint is useful for showing progress bars (e.g., "3/5 rooms used").

---

## FE Implementation Notes

1. **Fetch plan on login** — After login, call `GET /hotels/:id/subscription` and store the plan in app state.
2. **Conditionally render navigation** — Hide menu items for features not in plan (e.g., hide "Room Service" tab if `feature_room_service` is false).
3. **Show upgrade prompts** — When a 403 with the upgrade message comes back, show an upgrade modal/banner.
4. **Usage indicators** — Use the `/subscription/usage` endpoint to show "3/5 rooms used" type progress bars in settings.
5. **Super admin panel** — Add a plan management UI where super admin can view all hotels and change their plans.
6. **Plan comparison page** — Use `GET /plans` to dynamically render the feature comparison table.
