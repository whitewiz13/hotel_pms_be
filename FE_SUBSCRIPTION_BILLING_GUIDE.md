# Frontend Subscription Billing Guide

## Goal

The subscription screen now needs to show both billing state and actual hotel access state.

- Billing state comes from `status`, `access_until`, `days_left`, `is_overdue`, and `overdue_days`.
- Access state comes from `hotel_is_active` and `can_access`.

## API Endpoints

- `GET /api/hotels/:hotel_id/subscription`
- `PUT /api/hotels/:hotel_id/subscription`

The route did not change. Only the request/response shape expanded.

## New Response Fields

From `GET /subscription`, use these fields:

- `status`: billing state badge
- `assigned_at`: when the current plan was assigned
- `renewed_at`: last manual renewal timestamp
- `access_until`: billing reminder date to show the next due date
- `days_left`: nullable integer, use when `access_until` exists and is not overdue
- `is_overdue`: boolean flag for overdue styling
- `overdue_days`: integer count when overdue
- `suspended_at`: when the subscription was manually suspended
- `suspension_reason`: optional admin note
- `hotel_is_active`: hard hotel access switch
- `can_access`: final effective access state

`expires_at` is still returned for compatibility, but new UI should read `access_until`.

## Recommended UI

Show four blocks on the hotel subscription page:

1. Plan summary
2. Billing timeline
3. Access control
4. Admin actions

Recommended fields to show:

- Current plan name
- Billing status badge
- Access status badge
- Assigned date
- Renewed date
- Access-until date
- Days left or overdue days
- Suspension reason when present

Recommended badges:

- `status=active` and `can_access=true`: Active
- `status=past_due` and `can_access=true`: Past Due
- `status=suspended` or `can_access=false`: Suspended
- `status=cancelled`: Cancelled

Recommended urgency styling:

- `days_left <= 7`: warning
- `is_overdue = true`: danger

## Update Payloads

### Renew after payment

Send:

```json
{
  "status": "active",
  "renewed_at": "2026-06-05T09:00:00Z",
  "access_until": "2026-07-05T23:59:59Z",
  "hotel_is_active": true,
  "suspension_reason": ""
}
```

### Mark overdue but do not block

Send:

```json
{
  "status": "past_due",
  "access_until": "2026-06-01T23:59:59Z",
  "hotel_is_active": true
}
```

### Suspend for non-payment

Send:

```json
{
  "status": "suspended",
  "suspended_at": "2026-06-02T10:00:00Z",
  "suspension_reason": "Payment overdue",
  "hotel_is_active": false
}
```

### Change plan and renew together

Send:

```json
{
  "plan_id": "basic",
  "status": "active",
  "renewed_at": "2026-06-05T09:00:00Z",
  "access_until": "2026-07-05T23:59:59Z",
  "hotel_is_active": true
}
```

## Frontend Behavior Notes

- Treat all date fields as RFC3339 timestamps.
- Prefer rendering local time in the admin UI.
- If `days_left` is `null`, show `No due date set`.
- If `is_overdue` is `true`, show `Overdue by X days` using `overdue_days`.
- If `can_access` is `false`, disable normal hotel-management actions and show the suspension reason if present.
- If `hotel_is_active` is `false`, show a stronger lock message than a simple billing reminder.

## Suggested FE State Mapping

- Billing badge source: `status`
- Access badge source: `can_access`
- Due-date text source: `access_until`
- Countdown source: `days_left`
- Overdue banner source: `is_overdue` and `overdue_days`
- Lock banner source: `hotel_is_active`, `can_access`, `suspension_reason`

## Backward Compatibility

- Existing code reading `expires_at` will still work.
- New code should migrate to `access_until`.
- Existing `PUT /subscription` calls that only send `plan_id` still work.