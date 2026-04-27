# Analytics API

All analytics endpoints are hotel-scoped and require the **`analytics:view`** permission.

**Base URL:** `/api/hotels/:hotel_id/analytics`

---

## Permission

| Code | Module | Action | Description |
|------|--------|--------|-------------|
| `analytics:view` | analytics | view | View analytics and reports |

Assign this permission to roles that should access the analytics page (e.g. `hotel_admin`, `manager`). Users without this permission will get a `403 Forbidden`.

---

## Common Query Parameters

All endpoints accept these optional query params:

| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `from` | `YYYY-MM-DD` | 30 days ago | Start date (inclusive) |
| `to` | `YYYY-MM-DD` | Today | End date (inclusive) |

Trend endpoints also accept:

| Param | Type | Default | Options |
|-------|------|---------|---------|
| `period` | string | `daily` | `daily`, `monthly`, `yearly` |

---

## 1. Summary (KPI Cards)

```
GET /analytics/summary?from=2026-01-01&to=2026-04-27
```

Use for top-level stat cards on the analytics page.

**Response:**
```json
{
  "success": true,
  "data": {
    "total_revenue": 125000.00,
    "room_revenue": 95000.00,
    "service_revenue": 20000.00,
    "activity_revenue": 10000.00,
    "total_reservations": 320,
    "total_check_ins": 280,
    "total_check_outs": 265,
    "total_cancelled": 15,
    "total_guests": 210,
    "avg_daily_rate": 339.29,
    "rev_par": 225.00,
    "avg_stay_length": 2.5,
    "occupancy_rate": 66.3
  }
}
```

| Field | Description | Suggested UI |
|-------|-------------|--------------|
| `total_revenue` | Sum of all paid bills | Card |
| `room_revenue` | Room charges only | Card / pie chart slice |
| `service_revenue` | Room service revenue | Card / pie chart slice |
| `activity_revenue` | Activity booking revenue | Card / pie chart slice |
| `total_reservations` | Reservations created in range | Card |
| `total_check_ins` | Guests checked in | Card |
| `total_check_outs` | Guests checked out | Card |
| `total_cancelled` | Cancelled reservations | Card |
| `total_guests` | Unique guests | Card |
| `avg_daily_rate` | ADR = room revenue / rooms sold | Card |
| `rev_par` | RevPAR = room revenue / total room-nights | Card |
| `avg_stay_length` | Average nights stayed | Card |
| `occupancy_rate` | % of room-nights occupied | Card / gauge |

---

## 2. Occupancy Trend (Line/Area Chart)

```
GET /analytics/occupancy?from=2026-01-01&to=2026-04-27&period=daily
```

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "date": "2026-01-01",
      "occupied_rooms": 18,
      "total_rooms": 25,
      "rate": 72.0
    }
  ]
}
```

**Suggested UI:** Line or area chart with `date` on X-axis, `rate` on Y-axis. Tooltip can show `occupied_rooms / total_rooms`.

---

## 3. Revenue Trend (Stacked Bar/Line Chart)

```
GET /analytics/revenue?from=2026-01-01&to=2026-04-27&period=monthly
```

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "date": "2026-01",
      "room_revenue": 32000.00,
      "service_revenue": 6500.00,
      "activity_revenue": 3200.00,
      "total": 41700.00
    }
  ]
}
```

**Suggested UI:** Stacked bar chart (room / service / activity) or multi-line chart. X-axis = date, Y-axis = revenue.

---

## 4. Reservation Stats (Grouped Bar Chart)

```
GET /analytics/reservations?from=2026-01-01&to=2026-04-27&period=daily
```

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "date": "2026-01-01",
      "reservations": 12,
      "check_ins": 10,
      "check_outs": 8,
      "cancellations": 1
    }
  ]
}
```

**Suggested UI:** Grouped bar chart or multi-line chart showing all four series over time.

---

## 5. Room Type Performance (Table)

```
GET /analytics/room-types?from=2026-01-01&to=2026-04-27
```

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "room_type": "Deluxe",
      "total_rooms": 10,
      "reservations": 85,
      "revenue": 42500.00,
      "occupancy_rate": 72.5,
      "avg_rate": 500.00
    }
  ]
}
```

**Suggested UI:** Data table with columns for each field. Can also add a horizontal bar chart for occupancy comparison across types.

---

## Frontend Implementation Notes

### Date Range Picker
Add a shared date range picker component at the top of the analytics page. All endpoints use the same `from`/`to` params, so a single picker can drive all API calls.

### Period Selector
For trend charts (occupancy, revenue, reservations), add a toggle for `daily | monthly | yearly`. The `room-types` endpoint doesn't use `period`.

### RBAC
Gate the analytics page/route behind the `analytics:view` permission, same pattern as other features:

```ts
// Check permission before rendering
const canViewAnalytics = userPermissions.includes("analytics:view");
```

Hide the analytics nav item if the user doesn't have `analytics:view`.

### API Calls Example

```ts
const params = { from: "2026-01-01", to: "2026-04-27", period: "monthly" };

const [summary, occupancy, revenue, reservations, roomTypes] = await Promise.all([
  api.get(`/hotels/${hotelId}/analytics/summary`, { params }),
  api.get(`/hotels/${hotelId}/analytics/occupancy`, { params }),
  api.get(`/hotels/${hotelId}/analytics/revenue`, { params }),
  api.get(`/hotels/${hotelId}/analytics/reservations`, { params }),
  api.get(`/hotels/${hotelId}/analytics/room-types`, { params }),
]);
```
