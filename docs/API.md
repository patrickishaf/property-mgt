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
  - [`POST /auth/password/set`](#post-authpasswordset) – tenant sets password from invite
- [Tenants](#tenants)
  - [Access rules](#access-rules)
  - [Onboarding flow](#onboarding-flow)
  - [`POST /tenants`](#post-tenants) – create tenant and lease
  - [`GET /tenants`](#get-tenants) – list tenants
  - [`GET /tenants/{id}`](#get-tenantsid) – get tenant
  - [`PATCH /tenants/{id}`](#patch-tenantsid) – update tenant

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
| `invite` | `POST /tenants` (sent in the invite email link) | `POST /auth/password/set` | `PASSWORD_RESET_EXPIRATION_SECONDS` seconds |

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
| 401 | `invalid credentials` | Unknown email, wrong password, or a tenant who has not set a password yet |
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

A newly invited tenant sets their password using the token from the invite link, and is logged in. Each invite can be used once. **Auth:** none (invite token is in the body).

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
| 401 | `invalid token` | Token invalid, expired, not an invite token, or the tenant no longer exists |
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

1. An owner or manager calls `POST /tenants`. This creates the tenant (with no password) and an active lease, marks one more unit occupied, and queues an invite email.
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
| `propertyId` | integer | yes | Property the caller owns or manages |
| `unitNumber` | string | yes | Must not have an active lease |
| `lease` | object | yes | See below |
| `lease.startDate` | string | yes | `YYYY-MM-DD` |
| `lease.endDate` | string | yes | `YYYY-MM-DD`, after `startDate` |
| `lease.durationType` | string | yes | `annual` or `monthly` |
| `lease.expectedRent` | number | yes | > 0 |

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
| 403 | `invalid credentials` | Caller is a tenant |
| 404 | `property not found` | Property doesn't exist or caller doesn't own/manage it |
| 409 | `email already taken` | A tenant with this email exists |
| 409 | `unit already has an active lease` | |
| 409 | `property has no vacant units` | Occupied units already equal total units |
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
