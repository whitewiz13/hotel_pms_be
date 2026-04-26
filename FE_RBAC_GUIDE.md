# Role-Based Permission System — Frontend Implementation Guide

## Overview

The backend now supports **dynamic roles with granular permissions**. Roles are created per-hotel by hotel admins, and each role is assigned a set of permission codes. The JWT token includes the user's `permissions[]` array, so the FE can use it for both **route guarding** and **UI element visibility**.

---

## JWT Token Structure

After login, decode the JWT to get:

```json
{
  "user_id": "uuid",
  "email": "user@example.com",
  "role": "manager",
  "role_id": "uuid",
  "hotel_id": "uuid",
  "is_guest": false,
  "permissions": [
    "dashboard:view",
    "rooms:create",
    "rooms:read",
    "reservations:create",
    "reservations:read"
  ]
}
```

**Key fields:**
- `role` — slug of the assigned role (for display)
- `role_id` — UUID of the role record
- `permissions` — flat array of permission codes the user has
- `super_admin` users have **no permissions array** — they bypass all checks. Check `role === "super_admin"` to grant full access.

---

## API Endpoints

All endpoints are scoped under `/api/hotels/:hotel_id` and require **hotel_admin or super_admin** access.

### List All Available Permissions

```
GET /api/hotels/:hotel_id/permissions
```

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": "uuid",
      "code": "rooms:create",
      "module": "rooms",
      "action": "create",
      "description": "Create new rooms"
    }
  ]
}
```

### Create Role

```
POST /api/hotels/:hotel_id/roles
```

**Body:**
```json
{
  "name": "Night Manager",
  "description": "Handles night shift operations",
  "permissions": ["dashboard:view", "rooms:read", "reservations:read", "reservations:check_in"]
}
```

### List Roles

```
GET /api/hotels/:hotel_id/roles
```

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": "uuid",
      "name": "Night Manager",
      "slug": "night-manager",
      "description": "Handles night shift operations",
      "is_system": false,
      "permissions": [
        { "id": "uuid", "code": "dashboard:view", "module": "dashboard", "action": "view", "description": "..." }
      ]
    }
  ]
}
```

### Get Role by ID

```
GET /api/hotels/:hotel_id/roles/:id
```

### Update Role

```
PUT /api/hotels/:hotel_id/roles/:id
```

**Body (all fields optional):**
```json
{
  "name": "Senior Night Manager",
  "description": "Updated description",
  "permissions": ["dashboard:view", "rooms:read", "rooms:create"]
}
```

> **Note:** System roles (`is_system: true`) cannot be edited or deleted.

### Delete Role

```
DELETE /api/hotels/:hotel_id/roles/:id
```

> Fails if any users are currently assigned to this role.

### Create Staff (updated)

```
POST /api/hotels/:hotel_id/staff
```

**Body:**
```json
{
  "email": "john@hotel.com",
  "password": "securepass",
  "name": "John Doe",
  "phone": "+1234567890",
  "role_id": "uuid-of-the-role"
}
```

> The old `role` string field is replaced by `role_id`. Send the UUID of the role from the roles list.

---

## Permission Codes Reference

Permissions follow the pattern `module:action`. Group them by module in the UI for a clean checkbox grid.

| Module | Permissions |
|--------|------------|
| **dashboard** | `dashboard:view` |
| **rooms** | `rooms:create`, `rooms:read`, `rooms:update`, `rooms:delete`, `rooms:manage_pin` |
| **room_types** | `room_types:create`, `room_types:read`, `room_types:update`, `room_types:delete` |
| **reservations** | `reservations:create`, `reservations:read`, `reservations:check_in`, `reservations:check_out`, `reservations:cancel` |
| **amenities** | `amenities:create`, `amenities:read`, `amenities:update`, `amenities:delete` |
| **housekeeping** | `housekeeping:assign`, `housekeeping:read`, `housekeeping:update` |
| **menu** | `menu:create`, `menu:read`, `menu:update`, `menu:delete` |
| **orders** | `orders:create`, `orders:read`, `orders:update_status`, `orders:assign` |
| **activities** | `activities:create`, `activities:read`, `activities:update`, `activities:delete` |
| **activity_bookings** | `activity_bookings:create`, `activity_bookings:read`, `activity_bookings:update_status` |
| **billing** | `billing:generate`, `billing:read`, `billing:pay` |
| **staff** | `staff:create`, `staff:read`, `staff:update` |
| **guest_settings** | `guest_settings:read`, `guest_settings:update` |
| **roles** | `roles:create`, `roles:read`, `roles:update`, `roles:delete` |
| **hotels** | `hotels:read`, `hotels:update` |

---

## Frontend Implementation Steps

### 1. Store Permissions from JWT

After login, decode the JWT and store the permissions array in your auth store:

```ts
// authStore.ts
interface AuthState {
  user: User | null;
  token: string | null;
  permissions: string[];
}

// On login:
const decoded = jwtDecode<JWTPayload>(token);
authStore.permissions = decoded.permissions ?? [];
authStore.user = {
  id: decoded.user_id,
  email: decoded.email,
  role: decoded.role,
  roleId: decoded.role_id,
  hotelId: decoded.hotel_id,
};
```

### 2. Permission Check Utility

```ts
// lib/permissions.ts
import { useAuthStore } from '@/stores/authStore';

export function hasPermission(...codes: string[]): boolean {
  const { user, permissions } = useAuthStore.getState();

  // Super admin bypasses everything
  if (user?.role === 'super_admin') return true;

  return codes.some(code => permissions.includes(code));
}

export function hasAllPermissions(...codes: string[]): boolean {
  const { user, permissions } = useAuthStore.getState();

  if (user?.role === 'super_admin') return true;

  return codes.every(code => permissions.includes(code));
}
```

### 3. Route Guards

Replace hardcoded role checks with permission checks:

```tsx
// Before (hardcoded roles):
{ path: '/rooms', element: <Rooms />, roles: ['manager', 'front_desk'] }

// After (permission-based):
{ path: '/rooms', element: <ProtectedRoute permission="rooms:read"><Rooms /></ProtectedRoute> }
```

```tsx
// ProtectedRoute.tsx
function ProtectedRoute({ permission, children }: { permission: string; children: ReactNode }) {
  if (!hasPermission(permission)) {
    return <Navigate to="/unauthorized" />;
  }
  return <>{children}</>;
}
```

### 4. Sidebar/Nav Visibility

Show/hide navigation items based on permissions:

```tsx
const navItems = [
  { label: 'Dashboard',     path: '/dashboard',     permission: 'dashboard:view' },
  { label: 'Rooms',         path: '/rooms',          permission: 'rooms:read' },
  { label: 'Reservations',  path: '/reservations',   permission: 'reservations:read' },
  { label: 'Housekeeping',  path: '/housekeeping',   permission: 'housekeeping:read' },
  { label: 'Menu',          path: '/menu',           permission: 'menu:read' },
  { label: 'Orders',        path: '/orders',         permission: 'orders:read' },
  { label: 'Activities',    path: '/activities',     permission: 'activities:read' },
  { label: 'Billing',       path: '/billing',        permission: 'billing:read' },
  { label: 'Staff',         path: '/staff',          permission: 'staff:read' },
  { label: 'Roles',         path: '/roles',          permission: 'roles:read' },
  { label: 'Settings',      path: '/settings',       permission: 'hotels:update' },
];

// Filter nav items
const visibleItems = navItems.filter(item => hasPermission(item.permission));
```

### 5. UI Element Visibility (Buttons, Actions)

```tsx
// Show "Add Room" button only if user can create rooms
{hasPermission('rooms:create') && (
  <Button onClick={openCreateRoomDialog}>Add Room</Button>
)}

// Show delete button only if user can delete
{hasPermission('rooms:delete') && (
  <Button variant="destructive" onClick={() => deleteRoom(room.id)}>Delete</Button>
)}
```

### 6. Role Management Page

Build a page at `/roles` for hotel admins to manage roles:

1. **Fetch permissions** — `GET /permissions` — group by `module` field
2. **Fetch roles** — `GET /roles`
3. **Create/Edit role form:**
   - Name input
   - Description input
   - Permission grid: rows = modules, columns = actions, cells = checkboxes
4. **Save** — `POST /roles` or `PUT /roles/:id` with selected permission codes

**Permission grid UI example:**

| Module | Create | Read | Update | Delete | Other |
|--------|--------|------|--------|--------|-------|
| Rooms  | ☑      | ☑    | ☑      | ☐      | ☑ manage_pin |
| Menu   | ☐      | ☑    | ☐      | ☐      | — |
| Billing| —      | ☑    | —      | —      | ☑ generate, ☑ pay |

### 7. Update Staff Creation

Replace the role dropdown (hardcoded roles) with a dynamic list from the roles API:

```tsx
// Fetch roles for dropdown
const { data: roles } = useQuery(['roles', hotelId], () =>
  api.get(`/hotels/${hotelId}/roles`).then(r => r.data.data)
);

// In the form
<Select name="role_id" label="Role">
  {roles?.map(role => (
    <Option key={role.id} value={role.id}>{role.name}</Option>
  ))}
</Select>
```

---

## Migration Checklist

- [ ] Decode JWT and store `permissions[]` in auth state
- [ ] Create `hasPermission()` / `hasAllPermissions()` utility
- [ ] Replace all hardcoded role checks with permission checks
- [ ] Update sidebar/nav to filter by permissions
- [ ] Update route guards to use permissions
- [ ] Hide create/edit/delete buttons behind permission checks
- [ ] Build Role Management page (CRUD + permission grid)
- [ ] Update Staff Creation form to use `role_id` dropdown
- [ ] Handle `super_admin` as a bypass (no permissions needed)
- [ ] Show "Unauthorized" page for users without required permissions
