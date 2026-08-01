# Belcanto API — invitation and activation slice

This service implements Belcanto's closed-access first vertical slice. There is
no public sign-up route. A school Owner, or an existing Administrator to whom
the Owner explicitly grants the fixed `StudentOnboardingManager.v1` profile,
creates an adult Student record. The assigned Teacher publishes the First
Belcanto Minute. Only then can the Owner issue a seven-day, one-time activation
invitation. The Student activates the existing account, chooses their own
password, and signs in separately.

## Runtime

- Go language directive: `1.26.0`; pinned toolchain and container builder:
  `1.26.5`.
- PostgreSQL: `18.4` is the supported development and deployment baseline.
- HTTP: opaque bearer access tokens plus rotating opaque refresh tokens.

Required environment:

```text
APP_ENV=production
DATABASE_URL=postgres://belcanto:belcanto@localhost:5432/belcanto?sslmode=disable
TOKEN_HMAC_KEY_BASE64=<at-least-32-random-bytes-as-standard-base64>
PUBLIC_ACTIVATION_BASE_URL=https://app.example.com/activate
LISTEN_ADDRESS=:8080
AUTO_MIGRATE=false
```

Generate the token master key with a secret-management system or, for local
development only, `openssl rand -base64 32`. The same protected key must be
available to all API replicas. Do not commit it.

Apply the non-destructive migration command and create the first Owner exactly
once:

```sh
go run ./cmd/migrate up

go run ./cmd/bootstrap-owner \
  --tenant-id belcanto-school \
  --tenant-name "Belcanto Vocal School" \
  --full-name "Owner Name" \
  --phone +77001234567 \
  --operator "ops@example.com" \
  --reason "initial production bootstrap"
```

The command prints the non-secret Owner `accountId` followed by the Owner
activation link once. Preserve the account ID in the operational record and
treat the link/stdout as a secret.
The database stores only a keyed digest of the invitation token. It has no raw,
encrypted, or otherwise recoverable token column. After activating, the Owner
calls `POST /v1/sessions` to obtain a session.

After the Owner activates and obtains their `accountId` from
`GET /v1/me/bootstrap`, provision the initial staff through the controlled CLI
bridge (never through HTTP):

```sh
go run ./cmd/bootstrap-staff \
  --tenant-id belcanto-school \
  --owner-account-id acct_<owner-id> \
  --role Administrator \
  --full-name "Administrator Name" \
  --phone +77001234568 \
  --operator "ops@example.com" \
  --reason "approved initial Administrator"

go run ./cmd/bootstrap-staff \
  --tenant-id belcanto-school \
  --owner-account-id acct_<owner-id> \
  --role Teacher \
  --full-name "Teacher Name" \
  --phone +77001234569 \
  --operator "ops@example.com" \
  --reason "approved initial Teacher"
```

The transaction verifies that `--owner-account-id` is an active Owner in the
same tenant and rejects Owner/Student roles and repeated phone identifiers.
Each bootstrap prints the non-secret staff `accountId` followed by the one-time
link. The staff member activates it, chooses their password, and then signs in
normally. `GET /v1/staff` provides the resulting opaque
account IDs to the Owner and delegated onboarding UI.

If an unconsumed Owner or staff activation link is lost, reissue it without
creating another account or role grant:

```sh
go run ./cmd/reissue-bootstrap-invitation \
  --tenant-id belcanto-school \
  --account-id acct_<pending-account-id> \
  --operator "ops@example.com" \
  --reason "recipient lost the original activation link"
```

The recovery command accepts only an existing `pending_activation` Owner,
Administrator, or Teacher account with an original controlled invitation. It
atomically supersedes every still-issued link, preserves the identity and role
grant, and records the operator and reason. It rejects an already-active
account. Like the bootstrap commands, it prints the credential once.

Start the API after applying migrations (or set `AUTO_MIGRATE=true` only for a
controlled single-instance deployment):

```sh
go run ./cmd/api
```

Run local tests:

```sh
go test ./...
go test ./... -race
```

PostgreSQL integration tests are enabled only when `TEST_DATABASE_URL` points
to a disposable PostgreSQL 18.4 database. They apply and tear down the initial
schema.

## HTTP contract

All JSON write requests use `Content-Type: application/json`. Mutation routes
that participate in idempotency require a visible-ASCII `Idempotency-Key`
header of 1–128 bytes.
Authenticated routes use `Authorization: Bearer <accessToken>`.

| Method | Path | Access | Request body | Success |
|---|---|---|---|---|
| `GET` | `/healthz` | Public | none | Process liveness |
| `GET` | `/readyz` | Public | none | PostgreSQL-backed readiness |
| `POST` | `/v1/activations/preview` | Public, rate-limited | `{ "token": "..." }` | Activation preview |
| `POST` | `/v1/activations/complete` | Public, rate-limited | `{ "token": "...", "phone": "+...", "password": "..." }` | `204`; no session is created |
| `POST` | `/v1/sessions` | Public, rate-limited | `{ "phone": "+...", "password": "..." }` | Access and refresh tokens |
| `POST` | `/v1/sessions/refresh` | Public, rate-limited | `{ "refreshToken": "..." }` | Rotated token pair |
| `DELETE` | `/v1/sessions/current` | Authenticated | none | `204` |
| `GET` | `/v1/me/bootstrap` | Authenticated | none | Roles, effective access, and Student First Minute |
| `GET` | `/v1/staff?role=Administrator\|Teacher` | Owner; delegated Administrator may query Teacher only | none | Active same-tenant staff discovery |
| `GET` | `/v1/student-onboarding` | Owner, delegated Administrator, or assigned Teacher | none | Deterministic onboarding queue |
| `POST` | `/v1/access/delegations` | Owner + current password | `{ "administratorAccountId": "...", "reason": "...", "expiresAt": null, "currentPassword": "..." }` | Delegation |
| `POST` | `/v1/access/delegations/{delegationId}/revoke` | Owner + current password | `{ "reason": "...", "currentPassword": "..." }` | `204` |
| `POST` | `/v1/students` | Owner or delegated Administrator | `{ "fullName": "...", "phone": "+...", "enrollmentReference": "...", "teacherAccountId": "...", "locale": "ru-KZ", "timezone": "Asia/Almaty", "adultConfirmed": true }` | Student identity chain |
| `PUT` | `/v1/students/{studentId}/first-minute` | Assigned Teacher | `{ "whatWorked": "...", "currentFocus": "...", "nextStep": "...", "expectedVersion": 0 }` | Published revision |
| `POST` | `/v1/students/{studentId}/activation-invitations` | Owner only | none | Invitation and one-time activation link |
| `POST` | `/v1/students/{studentId}/activation-invitations/reissue` | Owner only | none | Replaces a currently active invitation; old link is invalid immediately |
| `POST` | `/v1/activation-invitations/{invitationId}/revoke` | Owner only | none | `204` |

Errors use one envelope:

```json
{
  "error": {
    "code": "FORBIDDEN",
    "message": "student onboarding permission is required",
    "requestId": "request_..."
  }
}
```

`GET /v1/me/bootstrap` exposes both persona roles and effective authorization.
Clients must use `permissions`, not infer capabilities from `Administrator`:

- Owner: `students.create`, `student_onboarding.read`,
  `student_invitations.issue`, `student_invitations.reissue`,
  `student_invitations.revoke`, and the nondelegable
  `student_onboarding.delegate` permission.
- delegated Administrator: access profile `StudentOnboardingManager.v1` and
  exactly `students.create` and `student_onboarding.read`.
- ordinary Administrator: no onboarding profile and none of those permissions.

The staff discovery response contains only non-secret data:
`accountId`, `fullName`, `roles`, and `accessProfiles`. When the Owner lists
Administrators with an active v1 grant, the item also contains
`onboardingDelegationId` and optional `onboardingDelegationExpiresAt`, so the
Owner can revoke the persisted grant after a restart.

Each student-onboarding queue item contains `studentId`, `fullName`,
`enrollmentReference`, `teacherAccountId`, `studentVersion`, and exactly one
state: `awaiting_first_minute`, `ready_to_invite`, `invited`, or `activated`.
Only `invited` includes the current `invitationId` and
`invitationExpiresAt`. No phone is returned. Owner and delegated Administrator
see active Students in their tenant; a Teacher sees only active assignments.

## Security and operational boundaries

- Invitation and session tokens are 256-bit opaque values. The invitation ID
  is random; the invitation token is derived with a domain-separated server
  HMAC. PostgreSQL persists only the token HMAC digest. Audit, outbox, and
  idempotency records contain no raw or recoverable invitation token.
- Passwords are normalized to NFC and stored with Argon2id using OWASP's
  19 MiB / two-iteration minimum. The API accepts 15–128 Unicode characters,
  rejects a bounded embedded common-password corpus, and applies one bounded
  process-wide Argon2 semaphore across every hasher instance.
- Invitation tokens are single-use, expire after seven days, can be revoked,
  and are invalidated immediately on reissue. Reissue requires an existing
  active invitation; after expiry/revocation the caller uses ordinary issue.
- Refresh tokens rotate. Reuse of a replaced token revokes the full session
  family. Logout also revokes the full tenant/account-scoped family, including
  any token concurrently produced by refresh rotation.
- Owner grant/revoke requires both an Owner session and fresh verification of
  the current password. Those Argon2 checks share a strict per-IP and
  per-Owner rate limit. The fixed v1 profile never gains future permissions
  automatically.
- Persisted names, reasons, enrollment references, locale, timezone, IDs, and
  idempotency keys have service and database bounds. Locale is parsed as BCP
  47; timezone must resolve as an IANA location.
- Phone-only sign-in makes a normalized phone a global identifier in this B.0
  authentication realm, so the database intentionally retains global
  uniqueness rather than allowing ambiguous cross-tenant credentials. Phone
  reservation is available only to authenticated/controlled onboarding paths,
  and every duplicate returns the same neutral `login identifier is
  unavailable` conflict without revealing an account or tenant.
- `APP_ENV` is exactly `development`, `test`, or `production` and defaults to
  `development`. `PUBLIC_ACTIVATION_BASE_URL` has no default. In production it
  must be exactly an HTTPS origin without userinfo or an explicit port, plus
  the path `/activate`, with no query or fragment. The exact custom-scheme
  value `belcanto://activate` is allowed only in development and test.
- The built-in rate limiter has per-subject and coarse per-IP layers plus a
  hard bucket-cardinality cap. The subject bucket is independent of source IP,
  so rotating IPs cannot reset its allowance. It is still per-process: a production
  multi-replica rollout must add an edge or distributed limiter and must not
  trust arbitrary forwarded-IP headers at this layer.
- Mutation idempotency is scoped by tenant, authenticated actor, operation,
  and key. Current authorization is re-read before any replay is returned, so
  revoking an Administrator delegation or changing a Teacher assignment takes
  effect even for an exact retry.
- `GET /healthz` is process liveness. `GET /readyz` pings PostgreSQL and returns
  `503 UNAVAILABLE` when the dependency is unavailable.
- Migrations use a transaction-scoped advisory lock and a version/checksum
  ledger. Concurrent or repeated application is safe; checksum drift fails
  closed.
- TLS termination, secret rotation, backups, outbox delivery, retention, and
  monitoring are deployment responsibilities outside this binary.

### Controlled staff bootstrap boundary

This slice deliberately does not invent public or self-service staff sign-up.
`cmd/bootstrap-staff` is the initial operational bridge: it requires database
operator access, verifies an active same-tenant Owner inside the transaction,
and can create only pending Administrator or Teacher accounts. Production code
exposes no seed endpoint. A later staff-management product slice can replace
this CLI without introducing public registration.
