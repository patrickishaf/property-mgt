# Property Management API

Human-readable reference for the HTTP API. The machine-readable contract is [`openapi.yaml`](./openapi.yaml).

**Browse it:** with the server running and `APP_ENV=development`, open `http://localhost:<PORT>/docs` for an interactive Swagger UI of `openapi.yaml` (the raw spec is at `/docs/openapi.yaml`). These routes are not registered in other environments.

> **Keeping this current:** any change to a route, request field, response field, status code, error message or access rule must update **both** this file and `openapi.yaml` in the same change.

## Contents

- [Conventions](#conventions)
  - [Authentication](#authentication)
  - [Response envelope](#response-envelope)
  - [Status codes](#status-codes)
  - [Pagination](#pagination)
- [Auth](#auth)
  - [`POST /auth/accounts`](#post-authaccounts) – create owner account
  - [`POST /auth/login`](#post-authlogin) – log in with password
  - [`GET /auth/user`](#get-authuser) – current user
  - [`POST /auth/token/refresh`](#post-authtokenrefresh) – new access token
  - [`POST /auth/password/set`](#post-authpasswordset) – invited tenant or manager sets password
- [Tenants](#tenants)
  - [Access rules](#access-rules)
  - [Onboarding flow](#onboarding-flow)
  - [`POST /tenants`](#post-tenants) – create tenant and lease
  - [`GET /tenants`](#get-tenants) – list tenants
  - [`GET /tenants/{id}`](#get-tenantsid) – get tenant
  - [`PATCH /tenants/{id}`](#patch-tenantsid) – update tenant
- [Properties](#properties)
  - [Access rules](#access-rules-1)
  - [Objects](#property-object)
  - [`POST /properties`](#post-properties) – create property (with units)
  - [`GET /properties`](#get-properties) – list properties
  - [`GET /properties/{id}`](#get-propertiesid) – get property
  - [`PATCH /properties/{id}`](#patch-propertiesid) – update property
  - [`POST /properties/{id}/archive`](#post-propertiesidarchive) – archive property
  - [`PUT /properties/{id}/manager`](#put-propertiesidmanager) – assign manager
  - [`DELETE /properties/{id}/manager`](#delete-propertiesidmanager) – unassign manager
  - [`GET /properties/{id}/units`](#get-propertiesidunits) – list units
  - [`POST /properties/{id}/units`](#post-propertiesidunits) – add unit
  - [`PATCH /properties/{id}/units/{unitId}`](#patch-propertiesidunitsunitid) – update unit
  - [`DELETE /properties/{id}/units/{unitId}`](#delete-propertiesidunitsunitid) – delete unit
- [Property managers](#property-managers)
  - [`POST /properties/managers`](#post-propertiesmanagers) – create and invite manager
  - [`GET /properties/managers`](#get-propertiesmanagers) – list managers
  - [`GET /properties/managers/{id}`](#get-propertiesmanagersid) – get manager
  - [`PATCH /properties/managers/{id}`](#patch-propertiesmanagersid) – update manager
  - [`DELETE /properties/managers/{id}`](#delete-propertiesmanagersid) – remove manager
- [Payments](#payments)
  - [How payments work](#how-payments-work)
  - [Objects](#payment-object)
  - [`POST /payments`](#post-payments) – record an offline payment
  - [`GET /payments`](#get-payments) – list and filter payments
  - [`GET /payments/{id}`](#get-paymentsid) – payment with allocations and receipt
  - [`GET /payments/stats`](#get-paymentsstats) – portfolio totals
  - [`POST /payments/checkout`](#post-paymentscheckout) – tenant starts a checkout
  - [`POST /payments/checkout/{reference}/complete`](#post-paymentscheckoutreferencecomplete) – complete a checkout (development only)
  - [`GET /payments/installments`](#get-paymentsinstallments) – list installments
  - [`POST /payments/installments`](#post-paymentsinstallments) – set a lease's installment schedule
  - [`PATCH /payments/installments/{id}`](#patch-paymentsinstallmentsid) – change an installment's due date
  - [`PATCH /payments/leases/{id}`](#patch-paymentsleasesid) – partial-payment settings

---

## Conventions

- **Base URL:** `http://localhost:<PORT>` in development (`PORT` env var). There is no version prefix.
- **Content type:** request and response bodies are JSON (`Content-Type: application/json`).
- **Dates** in request bodies are `YYYY-MM-DD`.

### Authentication

Protected endpoints require an access token:

```
Authorization: Bearer <accessToken>
```

There are three token types. All are HS256 JWTs carrying the user (`actor`) and a `typ` claim; each is only accepted where its type is expected.

| Type | Issued by | Accepted by | Lifetime |
|---|---|---|---|
| `access` | login, account creation, refresh, set password | `Authorization` header on protected routes | `JWT_AT_EXPIRATION_MINS` minutes |
| `refresh` | login, account creation, set password | `POST /auth/token/refresh` | `JWT_RT_EXPIRATION_HOURS` hours |
| `invite` | `POST /tenants`, `POST /properties/managers` (sent in the invite email link) | `POST /auth/password/set` | `PASSWORD_RESET_EXPIRATION_SECONDS` seconds |

Auth middleware failures on protected routes:

| Status | `message` | When |
|---|---|---|
| 401 | `invalid request` | `Authorization` header missing or not `Bearer <token>` |
| 401 | `invalid token` | Token invalid, expired, or not an access token |

### Response envelope

Every response uses the same envelope.

```json
{
  "status": "success",
  "message": "tenant created",
  "data": { }
}
```

| Field | Type | Description |
|---|---|---|
| `status` | `"success"` \| `"error"` | Outcome |
| `message` | string | Human-readable message. On errors, one of the messages listed per endpoint. May be empty on success. |
| `data` | object \| null | Payload on success; always `null` on error |

Error example:

```json
{
  "status": "error",
  "message": "invalid credentials",
  "data": null
}
```

### Status codes

| Status | Meaning |
|---|---|
| 200 | Success |
| 201 | Resource created |
| 400 | Body, query or path failed validation (`message: "invalid request"`), or a business rule on the input failed |
| 401 | Missing/invalid token or wrong credentials |
| 403 | Authenticated, but the role is not allowed to perform this action |
| 404 | Resource does not exist **or** the caller is not allowed to see it (the API does not distinguish) |
| 409 | Conflict with existing data (duplicate email, occupied unit, …) |
| 500 | Unexpected server error (`message: "something went wrong"`) |

Validation failures always return `400` with `message: "invalid request"`; field-level details are not returned.

### Pagination

List endpoints accept `page` (default `1`, min `1`) and `limit` (default `20`, min `1`, max `100`) query parameters and return:

```json
{
  "items": [],
  "page": 1,
  "limit": 20,
  "total": 0
}
```

---

## Auth

### `POST /auth/accounts`

Create a property **owner** account and log in. **Auth:** none.

**Request body**

| Field | Type | Required | Rules |
|---|---|---|---|
| `email` | string | yes | valid email; unique among owners |
| `fname` | string | yes | first name |
| `lname` | string | yes | last name |
| `password` | string | yes | min 8 characters |
| `phoneNumber` | string | yes | |

```json
{
  "email": "owner@example.com",
  "fname": "Ngozi",
  "lname": "Eze",
  "password": "password123",
  "phoneNumber": "08012345678"
}
```

**Response `200`** – `message: "account created"`

```json
{
  "status": "success",
  "message": "account created",
  "data": {
    "user": {
      "id": "6",
      "createdAt": "2026-10-31 13:35:59.123456 +0100 WAT",
      "email": "owner@example.com",
      "firstName": "Ngozi",
      "lastName": "Eze",
      "phoneNumber": "08012345678",
      "role": "owner"
    },
    "accessToken": "eyJhbGciOi...",
    "refreshToken": "eyJhbGciOi..."
  }
}
```

`user` is an [AdminUser](#adminuser) (all values are strings).

**Errors**

| Status | `message` | When |
|---|---|---|
| 400 | `invalid request` | Body fails validation |
| 409 | `email already taken` | An owner with this email exists |
| 500 | `something went wrong` | |

---

### `POST /auth/login`

Log in with email and password. **Auth:** none.

**Request body**

| Field | Type | Required | Rules |
|---|---|---|---|
| `email` | string | yes | |
| `password` | string | yes | |
| `accountType` | string | yes | `tenant` or `admin` (owners and property managers use `admin`) |

```json
{
  "email": "owner@example.com",
  "password": "password123",
  "accountType": "admin"
}
```

**Response `200`** – `message: "login successful"`

```json
{
  "status": "success",
  "message": "login successful",
  "data": {
    "accessToken": "eyJhbGciOi...",
    "refreshToken": "eyJhbGciOi..."
  }
}
```

**Errors**

| Status | `message` | When |
|---|---|---|
| 400 | `invalid request` | Body fails validation (including an unknown `accountType`) |
| 401 | `invalid credentials` | Unknown email, wrong password, or an invited tenant/manager who has not set a password yet |
| 500 | `something went wrong` | |

---

### `GET /auth/user`

Return the logged-in user's profile. **Auth:** access token (any role).

**Response `200`** – `data` is an [AdminUser](#adminuser) for owners/managers or a [TenantUser](#tenantuser) for tenants.

```json
{
  "status": "success",
  "message": "",
  "data": {
    "id": "6",
    "createdAt": "2026-10-31 13:35:59.123456 +0100 WAT",
    "email": "owner@example.com",
    "firstName": "Ngozi",
    "lastName": "Eze",
    "phoneNumber": "08012345678",
    "role": "owner"
  }
}
```

**Errors**

| Status | `message` | When |
|---|---|---|
| 400 | `invalid request` | Token carries an unrecognised role |
| 401 | `invalid request` / `invalid token` / `invalid credentials` | See [Authentication](#authentication) |
| 404 | `record not found in the db` | The user no longer exists |
| 500 | `something went wrong` | |

#### AdminUser

All values are strings.

| Field | Description |
|---|---|
| `id` | Admin ID |
| `createdAt` | Server-local timestamp, Go `time.String()` format |
| `email`, `firstName`, `lastName`, `phoneNumber` | Profile |
| `role` | `owner` or `property_manager` |

#### TenantUser

All values are strings.

| Field | Description |
|---|---|
| `id` | Tenant ID |
| `createdAt` | Server-local timestamp, Go `time.String()` format |
| `email`, `firstName`, `lastName`, `phoneNumber` | Profile |
| `expectedAnnualRent`, `outstandingAnnualRent` | Rent amounts, as decimal strings |
| `leaseId`, `propertyId`, `unitNumber` | Current lease and unit |

---

### `POST /auth/token/refresh`

Exchange a refresh token (plus the current, possibly expired, access token) for a new access token. **Auth:** none (tokens are in the body). The refresh token is not rotated.

**Request body**

| Field | Type | Required | Rules |
|---|---|---|---|
| `accessToken` | string | yes | The user's current access token; may be expired but must have a valid signature |
| `refreshToken` | string | yes | Unexpired refresh token for the **same** user |

```json
{
  "accessToken": "eyJhbGciOi...",
  "refreshToken": "eyJhbGciOi..."
}
```

**Response `200`** – `message: "token refreshed"`

```json
{
  "status": "success",
  "message": "token refreshed",
  "data": {
    "accessToken": "eyJhbGciOi..."
  }
}
```

**Errors**

| Status | `message` | When |
|---|---|---|
| 400 | `invalid request` | Body fails validation |
| 401 | `invalid token` | Either token invalid or of the wrong type, refresh token expired, or the tokens belong to different users |
| 401 | `invalid credentials` | The user no longer exists |
| 500 | `something went wrong` | |

---

### `POST /auth/password/set`

A newly invited tenant or property manager sets their password using the token from the invite link, and is logged in. Each invite can be used once. **Auth:** none (invite token is in the body).

**Request body**

| Field | Type | Required | Rules |
|---|---|---|---|
| `token` | string | yes | The `token` query parameter from the invite link |
| `password` | string | yes | min 8 characters |

```json
{
  "token": "eyJhbGciOi...",
  "password": "tenantpass1"
}
```

**Response `200`** – `message: "password set"`

```json
{
  "status": "success",
  "message": "password set",
  "data": {
    "accessToken": "eyJhbGciOi...",
    "refreshToken": "eyJhbGciOi..."
  }
}
```

**Errors**

| Status | `message` | When |
|---|---|---|
| 400 | `invalid request` | Body fails validation |
| 401 | `invalid token` | Token invalid, expired, not an invite token, or the invited user no longer exists |
| 409 | `password already set` | The invite was already used |
| 500 | `something went wrong` | |

---

## Tenants

All tenant endpoints require an access token.

### Access rules

| Role | Create | List | Get / update |
|---|---|---|---|
| `owner` | On properties they own | Tenants on properties they own | Tenants on properties they own |
| `property_manager` | On properties they manage | Tenants on properties they manage | Tenants on properties they manage |
| `tenant` | `403` | `403` | Their own record only |

Tenants and properties the caller cannot access return `404`, the same as ones that don't exist.

### Onboarding flow

1. An owner or manager calls `POST /tenants` for an existing, vacant unit. This creates the tenant (with no password) and an active lease, increments the property's `numberOfOccupiedUnits`, and queues an invite email.
2. The tenant opens the link in the email: `<WEB_APP_URL>/set-password?token=<inviteToken>`.
3. The web app calls `POST /auth/password/set` with that token and the chosen password, and receives tokens.
4. From then on the tenant logs in with `POST /auth/login` and `accountType: "tenant"`.

> Invite emails are queued as jobs; delivery depends on the email worker. In development (`APP_ENV=development`) the link is also written to the server log.

### Tenant object

Returned by every tenant endpoint.

| Field | Type | Description |
|---|---|---|
| `id` | integer | Tenant ID |
| `email` | string | |
| `firstName`, `lastName`, `phoneNumber` | string | |
| `propertyId` | integer | Property the tenant lives in |
| `unitNumber` | string | |
| `leaseId` | integer | Current lease |
| `expectedRent` | number | Expected rent for the current lease |
| `outstandingRent` | number | Rent still owed |

---

### `POST /tenants`

Create a tenant and their lease, and send them an invite. **Auth:** owner or property manager of the property.

**Request body**

| Field | Type | Required | Rules |
|---|---|---|---|
| `email` | string | yes | valid email; unique among tenants |
| `firstName` | string | yes | |
| `lastName` | string | yes | |
| `phoneNumber` | string | yes | |
| `propertyId` | integer | yes | Property the caller owns or manages; must not be archived |
| `unitNumber` | string | yes | `number` of an existing [unit](#unit-object) of the property; must not have an active lease |
| `lease` | object | yes | See below |
| `lease.startDate` | string | yes | `YYYY-MM-DD` |
| `lease.endDate` | string | yes | `YYYY-MM-DD`, after `startDate` |
| `lease.durationType` | string | yes | `annual` or `monthly` |
| `lease.expectedRent` | number | yes | > 0 |
| `lease.allowPartialPayment` | boolean | no | default `false`; lets the tenant pay part of the balance at checkout |
| `lease.minimumPartialAmount` | number | no | ≥ 0 and ≤ `expectedRent`; smallest partial payment allowed |

```json
{
  "email": "ada@example.com",
  "firstName": "Ada",
  "lastName": "Obi",
  "phoneNumber": "09012345678",
  "propertyId": 1,
  "unitNumber": "1A",
  "lease": {
    "startDate": "2026-11-01",
    "endDate": "2027-10-31",
    "durationType": "annual",
    "expectedRent": 1200
  }
}
```

**Response `201`** – `message: "tenant created"`, `data` is a [Tenant](#tenant-object).

```json
{
  "status": "success",
  "message": "tenant created",
  "data": {
    "id": 1,
    "email": "ada@example.com",
    "expectedRent": 1200,
    "firstName": "Ada",
    "lastName": "Obi",
    "leaseId": 1,
    "outstandingRent": 1200,
    "phoneNumber": "09012345678",
    "propertyId": 1,
    "unitNumber": "1A"
  }
}
```

**Errors**

| Status | `message` | When |
|---|---|---|
| 400 | `invalid request` | Body fails validation |
| 400 | `lease end date must be after start date` | |
| 400 | `minimum partial amount cannot exceed the lease's expected rent` | |
| 403 | `invalid credentials` | Caller is a tenant |
| 404 | `property not found` | Property doesn't exist or caller doesn't own/manage it |
| 404 | `unit not found` | The property has no unit with this `unitNumber` |
| 409 | `email already taken` | A tenant with this email exists |
| 409 | `property is archived` | |
| 409 | `unit already has an active lease` | |
| 500 | `something went wrong` | |

---

### `GET /tenants`

List tenants on properties the caller owns or manages, newest first. **Auth:** owner or property manager.

**Query parameters**

| Param | Type | Required | Rules |
|---|---|---|---|
| `propertyId` | integer | no | Only tenants of this property (must be one the caller can access) |
| `page` | integer | no | default `1`, min `1` |
| `limit` | integer | no | default `20`, min `1`, max `100` |

**Response `200`** – `data` is a paginated list of [Tenant](#tenant-object).

```json
{
  "status": "success",
  "message": "",
  "data": {
    "items": [
      {
        "id": 1,
        "email": "ada@example.com",
        "expectedRent": 1200,
        "firstName": "Ada",
        "lastName": "Obi",
        "leaseId": 1,
        "outstandingRent": 1200,
        "phoneNumber": "09012345678",
        "propertyId": 1,
        "unitNumber": "1A"
      }
    ],
    "page": 1,
    "limit": 20,
    "total": 1
  }
}
```

**Errors**

| Status | `message` | When |
|---|---|---|
| 400 | `invalid request` | Query fails validation |
| 403 | `invalid credentials` | Caller is a tenant |
| 404 | `property not found` | `propertyId` doesn't exist or caller can't access it |
| 500 | `something went wrong` | |

---

### `GET /tenants/{id}`

Get one tenant. **Auth:** the tenant themself, or the owner/manager of the tenant's property.

**Response `200`** – `data` is a [Tenant](#tenant-object).

**Errors**

| Status | `message` | When |
|---|---|---|
| 400 | `invalid request` | `id` is not a positive integer |
| 404 | `tenant not found` | Tenant doesn't exist or caller can't access it |
| 500 | `something went wrong` | |

---

### `PATCH /tenants/{id}`

Update a tenant's profile. Only the fields sent are changed. **Auth:** the tenant themself, or the owner/manager of the tenant's property.

**Request body** – all fields optional; if present they must be non-empty.

| Field | Type | Required | Rules |
|---|---|---|---|
| `firstName` | string | no | min 1 character |
| `lastName` | string | no | min 1 character |
| `phoneNumber` | string | no | min 1 character |

```json
{
  "phoneNumber": "09087654321"
}
```

**Response `200`** – `message: "tenant updated"`, `data` is the updated [Tenant](#tenant-object).

**Errors**

| Status | `message` | When |
|---|---|---|
| 400 | `invalid request` | `id` or body fails validation |
| 404 | `tenant not found` | Tenant doesn't exist or caller can't access it |
| 500 | `something went wrong` | |

---

## Properties

All property endpoints require an access token.

### Access rules

| Action | `owner` | `property_manager` | `tenant` |
|---|---|---|---|
| Create property | Yes (they become the owner) | `403` | `403` |
| List / get properties | Properties they own | Properties assigned to them | `403` |
| Update / archive property, assign or unassign manager | Properties they own | `403` on assigned properties | `403` |
| List / add / update / delete units | Properties they own | Properties assigned to them | `403` |

Properties the caller neither owns nor manages return `404 property not found`.

### Property object

| Field | Type | Description |
|---|---|---|
| `id` | integer | |
| `name` | string | |
| `type` | string | Free text, e.g. `residential` |
| `activityStatus` | string | Free text; may be empty |
| `location` | string | Street address; may be empty |
| `city`, `state`, `country` | string | |
| `ownerId` | integer | Owner's admin ID |
| `managerId` | integer \| null | Assigned manager's admin ID, or `null` |
| `archivedAt` | string \| null | RFC 3339 timestamp, or `null` if not archived |
| `numberOfUnits` | integer | Number of units on the property |
| `numberOfOccupiedUnits` | integer | Units with an active lease |
| `rentExpected` | number | Sum of `expectedRent` over the property's active leases |
| `rentCollected` | number | Sum of amounts paid on active leases |
| `outstandingRent` | number | Sum of rent still owed on active leases |

The rent fields are calculated from leases on every read. The unit counts are stored and kept in step with unit and lease changes in the same database transaction.

### Unit object

| Field | Type | Description |
|---|---|---|
| `id` | integer | |
| `number` | string | Unique within the property; tenants reference units by this value. Cannot be changed. |
| `floor` | integer | `0` if not set |
| `label` | string | Optional description |
| `occupied` | boolean | Whether the unit has an active lease |

---

### `POST /properties`

Create a property, optionally with its units. **Auth:** owner.

**Request body**

| Field | Type | Required | Rules |
|---|---|---|---|
| `name` | string | yes | |
| `type` | string | yes | |
| `activityStatus` | string | no | |
| `location` | string | no | |
| `city` | string | yes | |
| `state` | string | yes | |
| `country` | string | yes | |
| `units` | array | no | Each item: `number` (string, required), `floor` (integer), `label` (string). Numbers must be unique within the request. |

```json
{
  "name": "Palm Court",
  "type": "residential",
  "location": "12 Admiralty Way",
  "city": "Lagos",
  "state": "Lagos",
  "country": "NG",
  "units": [
    { "number": "1A", "floor": 1 },
    { "number": "1B", "floor": 1, "label": "Corner" }
  ]
}
```

**Response `201`** – `message: "property created"`, `data` is a [Property](#property-object).

```json
{
  "status": "success",
  "message": "property created",
  "data": {
    "id": 4,
    "activityStatus": "",
    "archivedAt": null,
    "city": "Lagos",
    "country": "NG",
    "location": "12 Admiralty Way",
    "managerId": null,
    "name": "Palm Court",
    "numberOfOccupiedUnits": 0,
    "numberOfUnits": 2,
    "outstandingRent": 0,
    "ownerId": 6,
    "rentCollected": 0,
    "rentExpected": 0,
    "state": "Lagos",
    "type": "residential"
  }
}
```

**Errors**

| Status | `message` | When |
|---|---|---|
| 400 | `invalid request` | Body fails validation |
| 400 | `duplicate unit number` | Two units in `units` share a `number` |
| 403 | `invalid credentials` | Caller is not an owner |
| 500 | `something went wrong` | |

---

### `GET /properties`

List the caller's properties (owned for owners, assigned for managers), newest first. **Auth:** owner or property manager.

**Query parameters**

| Param | Type | Required | Rules |
|---|---|---|---|
| `includeArchived` | boolean | no | default `false` |
| `page` | integer | no | default `1`, min `1` |
| `limit` | integer | no | default `20`, min `1`, max `100` |

**Response `200`** – `data` is a [paginated](#pagination) list of [Property](#property-object).

**Errors**

| Status | `message` | When |
|---|---|---|
| 400 | `invalid request` | Query fails validation |
| 403 | `invalid credentials` | Caller is a tenant |
| 500 | `something went wrong` | |

---

### `GET /properties/{id}`

Get one property, including archived ones. **Auth:** owner or assigned manager.

**Response `200`** – `data` is a [Property](#property-object).

**Errors**

| Status | `message` | When |
|---|---|---|
| 400 | `invalid request` | `id` is not a positive integer |
| 403 | `invalid credentials` | Caller is a tenant |
| 404 | `property not found` | |
| 500 | `something went wrong` | |

---

### `PATCH /properties/{id}`

Update property details. Only fields sent are changed; unit counts and rent fields cannot be set. **Auth:** owner.

**Request body** – all optional.

| Field | Type | Rules |
|---|---|---|
| `name`, `type`, `city`, `state`, `country` | string | non-empty if present |
| `activityStatus`, `location` | string | may be empty |

```json
{ "location": "14 Admiralty Way" }
```

**Response `200`** – `message: "property updated"`, `data` is the updated [Property](#property-object).

**Errors**

| Status | `message` | When |
|---|---|---|
| 400 | `invalid request` | `id` or body fails validation |
| 403 | `invalid credentials` | Caller is a tenant, or the property's manager |
| 404 | `property not found` | |
| 500 | `something went wrong` | |

---

### `POST /properties/{id}/archive`

Archive a property. Archived properties are hidden from `GET /properties` unless `includeArchived=true`, and reject new units and tenants. There is no unarchive yet. **Auth:** owner.

**Response `200`** – `message: "property archived"`, `data` is the [Property](#property-object) with `archivedAt` set.

**Errors**

| Status | `message` | When |
|---|---|---|
| 400 | `invalid request` | `id` is not a positive integer |
| 403 | `invalid credentials` | Caller is a tenant, or the property's manager |
| 404 | `property not found` | |
| 409 | `property is already archived` | |
| 409 | `property has active leases` | |
| 500 | `something went wrong` | |

---

### `PUT /properties/{id}/manager`

Assign a manager to a property, replacing any current one. **Auth:** owner.

**Request body**

| Field | Type | Required | Rules |
|---|---|---|---|
| `managerId` | integer | yes | A manager created by the caller |

```json
{ "managerId": 19 }
```

**Response `200`** – `message: "manager assigned"`, `data` is the updated [Property](#property-object).

**Errors**

| Status | `message` | When |
|---|---|---|
| 400 | `invalid request` | `id` or body fails validation |
| 403 | `invalid credentials` | Caller is a tenant, or the property's manager |
| 404 | `property not found` | |
| 404 | `manager not found` | No manager with this ID belongs to the caller |
| 500 | `something went wrong` | |

---

### `DELETE /properties/{id}/manager`

Remove the property's manager. Succeeds even if none is assigned. **Auth:** owner.

**Response `200`** – `message: "manager unassigned"`, `data` is the updated [Property](#property-object) with `managerId: null`.

**Errors**

| Status | `message` | When |
|---|---|---|
| 400 | `invalid request` | `id` is not a positive integer |
| 403 | `invalid credentials` | Caller is a tenant, or the property's manager |
| 404 | `property not found` | |
| 500 | `something went wrong` | |

---

### `GET /properties/{id}/units`

List the property's units, ordered by `number`. **Auth:** owner or assigned manager.

**Response `200`** – `data` is an array of [Unit](#unit-object) (not paginated).

```json
{
  "status": "success",
  "message": "",
  "data": [
    { "id": 7, "floor": 1, "label": "", "number": "1A", "occupied": true },
    { "id": 8, "floor": 1, "label": "Corner", "number": "1B", "occupied": false }
  ]
}
```

**Errors**

| Status | `message` | When |
|---|---|---|
| 400 | `invalid request` | `id` is not a positive integer |
| 403 | `invalid credentials` | Caller is a tenant |
| 404 | `property not found` | |
| 500 | `something went wrong` | |

---

### `POST /properties/{id}/units`

Add a unit. Increments `numberOfUnits` in the same transaction. **Auth:** owner or assigned manager.

**Request body**

| Field | Type | Required | Rules |
|---|---|---|---|
| `number` | string | yes | Unique within the property |
| `floor` | integer | no | |
| `label` | string | no | |

```json
{ "number": "2A", "floor": 2 }
```

**Response `201`** – `message: "unit created"`, `data` is a [Unit](#unit-object).

**Errors**

| Status | `message` | When |
|---|---|---|
| 400 | `invalid request` | `id` or body fails validation |
| 403 | `invalid credentials` | Caller is a tenant |
| 404 | `property not found` | |
| 409 | `property is archived` | |
| 409 | `unit number already exists` | |
| 500 | `something went wrong` | |

---

### `PATCH /properties/{id}/units/{unitId}`

Update a unit's floor or label. The `number` cannot be changed because leases reference it. **Auth:** owner or assigned manager.

**Request body** – all optional.

| Field | Type |
|---|---|
| `floor` | integer |
| `label` | string |

**Response `200`** – `message: "unit updated"`, `data` is the updated [Unit](#unit-object).

**Errors**

| Status | `message` | When |
|---|---|---|
| 400 | `invalid request` | Path or body fails validation |
| 403 | `invalid credentials` | Caller is a tenant |
| 404 | `property not found` | |
| 404 | `unit not found` | No unit with this ID on the property |
| 500 | `something went wrong` | |

---

### `DELETE /properties/{id}/units/{unitId}`

Delete a vacant unit. Decrements `numberOfUnits` in the same transaction; the unit number can then be reused. **Auth:** owner or assigned manager.

**Response `200`** – `message: "unit deleted"`, `data` is `null`.

**Errors**

| Status | `message` | When |
|---|---|---|
| 400 | `invalid request` | Path fails validation |
| 403 | `invalid credentials` | Caller is a tenant |
| 404 | `property not found` | |
| 404 | `unit not found` | |
| 409 | `unit is occupied` | The unit has an active lease |
| 500 | `something went wrong` | |

---

## Property managers

Managers are admin accounts with role `property_manager`, created by an owner and belonging to that owner. They log in with `POST /auth/login` and `accountType: "admin"`.

Onboarding works like tenants: `POST /properties/managers` queues an invite email linking to `<WEB_APP_URL>/set-password?token=<inviteToken>`; the manager sets a password with `POST /auth/password/set`. In development the link is also logged.

### Manager object

| Field | Type | Description |
|---|---|---|
| `id` | integer | Admin ID (use as `managerId`) |
| `email`, `firstName`, `lastName`, `phoneNumber` | string | |
| `status` | string | `invited` until a password is set, then `active` |

---

### `POST /properties/managers`

Create a manager and send them an invite. **Auth:** owner.

**Request body**

| Field | Type | Required | Rules |
|---|---|---|---|
| `email` | string | yes | valid email; not used by any admin (owner or manager) |
| `firstName` | string | yes | |
| `lastName` | string | yes | |
| `phoneNumber` | string | yes | |

```json
{
  "email": "mia@example.com",
  "firstName": "Mia",
  "lastName": "Okoro",
  "phoneNumber": "07012345678"
}
```

**Response `201`** – `message: "manager created"`, `data` is a [Manager](#manager-object) with `status: "invited"`.

**Errors**

| Status | `message` | When |
|---|---|---|
| 400 | `invalid request` | Body fails validation |
| 403 | `invalid credentials` | Caller is not an owner |
| 409 | `email already taken` | |
| 500 | `something went wrong` | |

---

### `GET /properties/managers`

List the caller's managers, newest first. **Auth:** owner.

**Query parameters:** `page`, `limit` (see [Pagination](#pagination)).

**Response `200`** – `data` is a paginated list of [Manager](#manager-object).

**Errors**

| Status | `message` | When |
|---|---|---|
| 400 | `invalid request` | Query fails validation |
| 403 | `invalid credentials` | Caller is not an owner |
| 500 | `something went wrong` | |

---

### `GET /properties/managers/{id}`

Get a manager. **Auth:** the owner the manager belongs to, or the manager themself.

**Response `200`** – `data` is a [Manager](#manager-object).

**Errors**

| Status | `message` | When |
|---|---|---|
| 400 | `invalid request` | `id` is not a positive integer |
| 403 | `invalid credentials` | Caller is a tenant |
| 404 | `manager not found` | No such manager, or it isn't the caller's (or the caller themself) |
| 500 | `something went wrong` | |

---

### `PATCH /properties/managers/{id}`

Update a manager's profile. Only fields sent are changed. **Auth:** owner.

**Request body** – all optional; non-empty if present: `firstName`, `lastName`, `phoneNumber`.

**Response `200`** – `message: "manager updated"`, `data` is the updated [Manager](#manager-object).

**Errors**

| Status | `message` | When |
|---|---|---|
| 400 | `invalid request` | `id` or body fails validation |
| 403 | `invalid credentials` | Caller is not an owner |
| 404 | `manager not found` | |
| 500 | `something went wrong` | |

---

### `DELETE /properties/managers/{id}`

Remove a manager. In one transaction, unassigns them from every property and deletes their account. They can no longer log in or refresh tokens; an access token they already hold stays valid until it expires but grants no property access. **Auth:** owner.

**Response `200`** – `message: "manager removed"`, `data` is `null`.

**Errors**

| Status | `message` | When |
|---|---|---|
| 400 | `invalid request` | `id` is not a positive integer |
| 403 | `invalid credentials` | Caller is not an owner |
| 404 | `manager not found` | |
| 500 | `something went wrong` | |

---

## Payments

All payment endpoints require an access token, except the development-only checkout completion route.

### How payments work

- A **lease** is what a tenant owes: `expectedRent`, `amountPaid` and `outstandingRent`. Its `allowPartialPayment` and `minimumPartialAmount` settings control partial checkouts.
- **Installments** optionally split a lease's outstanding rent into scheduled amounts. The installments' outstanding amounts always add up exactly to the lease's `outstandingRent`.
- A **payment** is money received, either recorded offline by an owner or manager, or paid by the tenant through a checkout. Every completed payment is **allocated**:
  - payments for specific installments go to those installments first;
  - lease and partial payments go to installments in due-date order (oldest first), or directly to the lease if it has no installments;
  - the lease's `amountPaid` and `outstandingRent`, the tenant's outstanding rent, and the affected installments are updated in one transaction, and a receipt is created.
- All amounts are in a single currency (`PAYMENT_CURRENCY`, default `NGN`) and are rounded to 2 decimal places. A payment can never exceed the lease's outstanding rent.

**Access:** owners and managers act on leases in properties they own or manage; tenants only see their own payments and installments. Records the caller can't access return `404`.

**Payment statuses**

| Status | Meaning |
|---|---|
| `unpaid` | Checkout started but not completed |
| `paid` | Fully covered what it was for: the whole lease balance (`lease`/`partial`), or every chosen installment (`installments`) |
| `partially_paid` | Covered part of what it was for |
| `expired` | Checkout was not completed before `expiresAt` |
| `failed` | The payment provider could not create the checkout |

### Payment object

| Field | Type | Description |
|---|---|---|
| `id` | integer | |
| `reference` | string | Unique reference, e.g. `PAY-3F9A…`; offline payments may supply their own |
| `amount` | number | Money received |
| `targetAmount` | number | What the payment was meant to cover when it was created |
| `currency` | string | |
| `status` | string | See statuses above |
| `purpose` | string | `lease`, `installments` or `partial` |
| `source` | string | `offline` or `checkout` |
| `method` | string | Offline: `cash`, `bank_transfer`, `pos`, `cheque`, `other`. Checkout: `checkout` |
| `leaseId`, `tenantId`, `propertyId` | integer | |
| `tenantFullName`, `propertyName`, `unitNumber` | string | Snapshot taken when the payment was created |
| `paidAt` | string \| null | RFC 3339; set when the payment is applied |
| `createdAt` | string | RFC 3339 |
| `note` | string | Offline payments only |
| `recordedBy` | integer \| null | Admin who recorded an offline payment |
| `checkoutUrl` | string | Checkout payments only (omitted otherwise) |
| `expiresAt` | string | Checkout payments only (omitted otherwise) |

**Payment detail** (`GET /payments/{id}`, `POST /payments`, checkout completion) adds:

| Field | Type | Description |
|---|---|---|
| `allocations` | array | `{ installmentId: integer \| null, amount: number }` per installment the payment reached; `installmentId` is `null` when applied to a lease without installments |
| `receipt` | object \| null | See below; `null` until the payment is applied |

**Receipt** – all values are strings: `id` (integer), `senderFullName`, `senderEmail`, `paymentDate` (RFC 3339), `paymentMethod`, `amountPaid`, `previousPaymentAmount` (lease amount paid before this payment), `installmentAmount`, `installmentAmountPaid`, `installmentStatus`, `installmentDueDate`, `outstandingInstallmentAmount`. The installment fields describe the last installment the payment reached; for leases without installments they are empty and `outstandingInstallmentAmount` is the lease's remaining balance.

### Installment object

| Field | Type | Description |
|---|---|---|
| `id` | integer | |
| `leaseId`, `tenantId`, `propertyId` | integer | |
| `amountExpected` | number | |
| `amountPaid` | number | |
| `outstandingAmount` | number | |
| `dueDate` | string | `YYYY-MM-DD` |
| `lastPaymentDate` | string \| null | RFC 3339 |
| `status` | string | `unpaid`, `partially_paid` or `paid` |
| `overdue` | boolean | `dueDate` has passed and money is still owed |

### Lease object

Returned by `PATCH /payments/leases/{id}`.

| Field | Type |
|---|---|
| `id`, `tenantId`, `propertyId` | integer |
| `unitNumber`, `durationType`, `status` | string |
| `startDate`, `endDate` | string (`YYYY-MM-DD`) |
| `expectedRent`, `amountPaid`, `outstandingRent` | number |
| `allowPartialPayment` | boolean |
| `minimumPartialAmount` | number |

---

### `POST /payments`

Record money received outside the app (cash, bank transfer, …) and apply it to the lease. Partial-payment settings don't apply because the money has already been received. **Auth:** owner or manager of the lease's property.

**Request body**

| Field | Type | Required | Rules |
|---|---|---|---|
| `leaseId` | integer | yes | Active lease the caller can access |
| `amount` | number | yes | > 0 and ≤ the lease's outstanding rent |
| `method` | string | yes | `cash`, `bank_transfer`, `pos`, `cheque` or `other` |
| `paidAt` | string | no | `YYYY-MM-DD`; defaults to now |
| `reference` | string | no | Must be unique; generated if omitted |
| `note` | string | no | |
| `installmentIds` | integer[] | no | Unpaid installments of the lease to pay first; without it the payment is applied as `lease` |

```json
{
  "leaseId": 4,
  "amount": 300,
  "method": "bank_transfer",
  "paidAt": "2026-10-30",
  "reference": "TRF-88213",
  "note": "Final balance"
}
```

**Response `201`** – `message: "payment recorded"`, `data` is a [payment detail](#payment-object).

```json
{
  "status": "success",
  "message": "payment recorded",
  "data": {
    "id": 12,
    "amount": 300,
    "createdAt": "2026-10-31T15:02:11.412+01:00",
    "currency": "NGN",
    "leaseId": 4,
    "method": "bank_transfer",
    "note": "Final balance",
    "paidAt": "2026-10-30T00:00:00Z",
    "propertyId": 6,
    "propertyName": "Palm Court",
    "purpose": "lease",
    "recordedBy": 22,
    "reference": "TRF-88213",
    "source": "offline",
    "status": "paid",
    "targetAmount": 300,
    "tenantFullName": "Ada Obi",
    "tenantId": 5,
    "unitNumber": "1A",
    "allocations": [{ "amount": 300, "installmentId": 3 }],
    "receipt": {
      "id": 7,
      "amountPaid": "300.00",
      "installmentAmount": "400.00",
      "installmentAmountPaid": "400.00",
      "installmentDueDate": "2027-01-01",
      "installmentStatus": "paid",
      "outstandingInstallmentAmount": "0.00",
      "paymentDate": "2026-10-30T00:00:00Z",
      "paymentMethod": "bank_transfer",
      "previousPaymentAmount": "900.00",
      "senderEmail": "ada@example.com",
      "senderFullName": "Ada Obi"
    }
  }
}
```

**Errors**

| Status | `message` | When |
|---|---|---|
| 400 | `invalid request` | Body fails validation |
| 400 | `amount exceeds outstanding balance` | |
| 400 | `invalid installment` | An ID isn't an unpaid installment of the lease, or is repeated |
| 403 | `invalid credentials` | Caller is a tenant |
| 404 | `lease not found` | Lease doesn't exist or caller can't access it |
| 409 | `lease is not active` | |
| 409 | `reference already exists` | |
| 500 | `something went wrong` | |

---

### `GET /payments`

List payments, newest first. **Auth:** any role; owners and managers see payments on their properties, tenants see only their own.

**Query parameters** – all optional.

| Param | Type | Rules |
|---|---|---|
| `propertyId` | integer | |
| `tenantId` | integer | Ignored for tenants (always themselves) |
| `leaseId` | integer | |
| `status` | string | `unpaid`, `partially_paid`, `paid`, `failed`, `expired` |
| `source` | string | `offline` or `checkout` |
| `purpose` | string | `lease`, `installments` or `partial` |
| `method` | string | e.g. `cash`, `checkout` |
| `from` | string | `YYYY-MM-DD`; payments on or after this date (uses `paidAt`, or `createdAt` if unpaid) |
| `to` | string | `YYYY-MM-DD`; payments on or before this date |
| `page` | integer | default `1`, min `1` |
| `limit` | integer | **default `10`**, min `1`, max `100` |

**Response `200`** – `data` is a [paginated](#pagination) list of [Payment](#payment-object) (without allocations or receipt). Unpaid checkouts past their expiry are marked `expired` before listing.

**Errors**

| Status | `message` | When |
|---|---|---|
| 400 | `invalid request` | Query fails validation |
| 500 | `something went wrong` | |

---

### `GET /payments/{id}`

Get one payment with its allocations and receipt. **Auth:** the payment's tenant, or the owner/manager of its property.

**Response `200`** – `data` is a [payment detail](#payment-object).

**Errors**

| Status | `message` | When |
|---|---|---|
| 400 | `invalid request` | `id` is not a positive integer |
| 404 | `payment not found` | |
| 500 | `something went wrong` | |

---

### `GET /payments/stats`

Rent totals across the caller's properties, for the dashboard. Calculated from active leases and installments. **Auth:** owner or manager.

**Query parameters:** `propertyId` (optional) – limit to one property the caller can access.

**Response `200`**

```json
{
  "status": "success",
  "message": "",
  "data": {
    "activeTenantCount": 1,
    "openInstallmentCount": 3,
    "outstandingRentBalance": 1200,
    "overdueInstallmentCount": 1,
    "percentageRentCollected": 0,
    "totalOverdueAmount": 400,
    "totalRentCollected": 0,
    "totalRentExpected": 1200
  }
}
```

| Field | Description |
|---|---|
| `totalRentExpected` | Sum of `expectedRent` over active leases |
| `totalRentCollected` | Sum of `amountPaid` over active leases |
| `percentageRentCollected` | `totalRentCollected / totalRentExpected × 100`, 2 decimals; `0` when nothing is expected |
| `outstandingRentBalance` | Sum of `outstandingRent` over active leases |
| `activeTenantCount` | Tenants with an active lease |
| `openInstallmentCount` | Installments with money still owed |
| `overdueInstallmentCount`, `totalOverdueAmount` | Open installments past their due date, and what's owed on them |

**Errors**

| Status | `message` | When |
|---|---|---|
| 400 | `invalid request` | Query fails validation |
| 403 | `invalid credentials` | Caller is a tenant |
| 404 | `property not found` | `propertyId` doesn't exist or caller can't access it |
| 500 | `something went wrong` | |

---

### `POST /payments/checkout`

Start a checkout for the tenant's active lease. Creates an `unpaid` payment and returns the provider's checkout URL. Only one open checkout per lease is allowed at a time. **Auth:** tenant.

> The payment provider (Bachs) is currently stubbed: `checkoutUrl` points to `<WEB_APP_URL>/checkout/stub?reference=…`, and checkouts are completed with the [development-only route](#post-paymentscheckoutreferencecomplete).

**Request body**

| Field | Type | Required | Rules |
|---|---|---|---|
| `type` | string | yes | `lease` (whole outstanding balance), `installments` or `partial` |
| `installmentIds` | integer[] | for `installments` | Unpaid installments of the tenant's lease; the amount is their total outstanding |
| `amount` | number | for `partial` | Lease must allow partial payments; `minimumPartialAmount` ≤ amount ≤ outstanding rent |

```json
{ "type": "partial", "amount": 500 }
```

**Response `201`** – `message: "checkout created"`, `data` is a [Payment](#payment-object) with `status: "unpaid"`, `checkoutUrl` and `expiresAt` (default 30 minutes, `CHECKOUT_EXPIRY_MINUTES`).

**Errors**

| Status | `message` | When |
|---|---|---|
| 400 | `invalid request` | Body fails validation, or `installmentIds`/`amount` missing for the chosen type |
| 400 | `invalid installment` | |
| 400 | `partial payments are not allowed for this lease` | |
| 400 | `amount is below the minimum partial amount of <n>` | |
| 400 | `amount exceeds outstanding balance` | |
| 403 | `invalid credentials` | Caller is not a tenant |
| 404 | `no active lease` | |
| 409 | `lease has no outstanding balance` | |
| 409 | `a checkout is already in progress for this lease` | |
| 502 | `payment provider unavailable` | The provider failed; the payment is marked `failed` |
| 500 | `something went wrong` | |

---

### `POST /payments/checkout/{reference}/complete`

**Development only** (registered when `APP_ENV=development`; no token required). Simulates the payment provider confirming a checkout: allocates the payment to the lease and installments and creates a receipt. A future provider webhook will run the same logic.

**Response `200`** – `message: "checkout completed"`, `data` is a [payment detail](#payment-object) with status `paid` or `partially_paid`.

**Errors**

| Status | `message` | When |
|---|---|---|
| 404 | `payment not found` | No payment with this reference |
| 409 | `payment is not a checkout` | It's an offline payment |
| 409 | `payment is not awaiting completion` | Already completed, expired or failed |
| 409 | `checkout has expired` | Past `expiresAt`; the payment is marked `expired` |
| 409 | `amount exceeds outstanding balance` | The lease balance dropped below the checkout amount since it started |
| 500 | `something went wrong` | |

---

### `GET /payments/installments`

List installments, ordered by due date. **Auth:** any role; scoped like payments.

**Query parameters** – all optional.

| Param | Type | Rules |
|---|---|---|
| `leaseId`, `tenantId`, `propertyId` | integer | `tenantId` is ignored for tenants |
| `status` | string | `unpaid`, `partially_paid`, `paid` or `overdue` |
| `dueBefore` | string | `YYYY-MM-DD`; installments due before this date |
| `page` | integer | default `1` |
| `limit` | integer | default `10`, max `100` |

**Response `200`** – `data` is a paginated list of [Installment](#installment-object).

**Errors**

| Status | `message` | When |
|---|---|---|
| 400 | `invalid request` | Query fails validation |
| 500 | `something went wrong` | |

---

### `POST /payments/installments`

Set a lease's installment schedule. Replaces the lease's installments that have received no money; installments that have been partly or fully paid are kept. The new items plus the outstanding amounts of kept installments must add up **exactly** to the lease's outstanding rent. **Auth:** owner or manager of the lease's property.

**Request body**

| Field | Type | Required | Rules |
|---|---|---|---|
| `leaseId` | integer | yes | Active lease the caller can access |
| `items` | array | yes | At least one `{ amount: number > 0, dueDate: "YYYY-MM-DD" }` |

```json
{
  "leaseId": 4,
  "items": [
    { "amount": 400, "dueDate": "2026-10-15" },
    { "amount": 400, "dueDate": "2026-12-01" },
    { "amount": 400, "dueDate": "2027-01-01" }
  ]
}
```

**Response `201`** – `message: "installments set"`, `data` is the lease's full schedule as an array of [Installment](#installment-object), in due-date order.

**Errors**

| Status | `message` | When |
|---|---|---|
| 400 | `invalid request` | Body fails validation |
| 400 | `installments must add up to the lease's outstanding rent of <n>` | |
| 403 | `invalid credentials` | Caller is a tenant |
| 404 | `lease not found` | |
| 409 | `lease is not active` | |
| 500 | `something went wrong` | |

---

### `PATCH /payments/installments/{id}`

Change an installment's due date. Amounts can only change by setting a new schedule, so the total stays equal to the lease's outstanding rent. **Auth:** owner or manager of the property.

**Request body**

| Field | Type | Required | Rules |
|---|---|---|---|
| `dueDate` | string | yes | `YYYY-MM-DD` |

**Response `200`** – `message: "installment updated"`, `data` is the updated [Installment](#installment-object).

**Errors**

| Status | `message` | When |
|---|---|---|
| 400 | `invalid request` | `id` or body fails validation |
| 403 | `invalid credentials` | Caller is a tenant |
| 404 | `installment not found` | |
| 500 | `something went wrong` | |

---

### `PATCH /payments/leases/{id}`

Set whether the tenant may make partial payments, and the minimum. Only fields sent are changed. **Auth:** owner or manager of the lease's property.

**Request body** – all optional.

| Field | Type | Rules |
|---|---|---|
| `allowPartialPayment` | boolean | |
| `minimumPartialAmount` | number | ≥ 0 and ≤ the lease's `expectedRent` |

```json
{ "allowPartialPayment": true, "minimumPartialAmount": 100 }
```

**Response `200`** – `message: "lease updated"`, `data` is the [Lease](#lease-object).

**Errors**

| Status | `message` | When |
|---|---|---|
| 400 | `invalid request` | `id` or body fails validation |
| 400 | `minimum partial amount cannot exceed the lease's expected rent` | |
| 403 | `invalid credentials` | Caller is a tenant |
| 404 | `lease not found` | |
| 500 | `something went wrong` | |
