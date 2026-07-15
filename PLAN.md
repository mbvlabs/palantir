# Palantir Andurel v1 Port Plan

## Status

Approved by MBV.

This project ports the behavior in `../bak-palantir` to Andurel v1.4.0 using Bun, Inertia, Svelte 5, Tailwind CSS, and shadcn-svelte.

## Goals

- Keep sign-in and password reset as the only public application views.
- Remove registration and email confirmation completely.
- Keep public infrastructure needed by connected websites.
- Port the existing website, tracking, collection, and analytics behavior.
- Build a modern responsive dashboard with shadcn-svelte components and charts.
- Add deterministic development seed data suitable for demos and screenshots.
- Add tests before implementation.
- Do not add product features beyond the backup project.

A "connected site" means an existing Website record that can receive tracking data. Do not add a connection status, verification handshake, teams, alerts, exports, billing, funnels, or new analytics dimensions.

## Approved access policy

| Surface | Anonymous result | Authenticated result |
| --- | --- | --- |
| `GET /users/sign-in` | Allowed | Allowed, optionally redirect to websites |
| `POST /users/sign-in` | Allowed | Allowed |
| Password reset request, token edit, and update | Allowed | Allowed |
| `/` and all `/websites/*` views | 303 redirect to sign-in | Allowed |
| Sign-out | 303 redirect to sign-in | Allowed |
| Registration and email confirmation | 404 because routes and code are removed | 404 |
| `/api/collect` | Allowed | Allowed |
| `/t/script.js` | Allowed | Allowed |
| Health and assets | Allowed | Allowed |
| Unknown routes | 404 | 404 |

Password reset is the approved public-view exception beside sign-in.

Public collection needs a narrow CORS and CSRF exception. Do not weaken CORS or CSRF for unrelated API routes.

## Existing database baseline

The current migrations already define all required tables and indexes:

- `websites`
- `pageviews`
- `events`
- visitor hash and geo columns

No new migration is expected.

## Architecture

```text
External site                         Palantir application

  /t/script.js
       |
       +-- pageview/event --> POST /api/collect --> Collect controller
                                                     | validate input
                                                     v
                                         Pageview / Event Bun models
                                                     |
                                                     v
                                                 PostgreSQL
                                                     ^
                                                     | aggregate queries
Browser session --> AuthOnly --> Dashboard controller
       |                           |
       |                           v
       +---------------------- Inertia props --> Svelte + shadcn charts
```

### Layer placement

#### Models

Generate Bun models from the existing migrations:

- `models/website.go`
- `models/pageview.go`
- `models/event.go`

Generate matching factories under `models/factories/`.

Use the generated models as a starting point, then keep only behavior required by the existing product.

Website responsibilities:

- Entity and validation
- Create, update, destroy
- Ordered list by owner
- Owner-scoped lookup

Use one owner-scoped query for authorization:

```go
models.Website.FindOwned(ctx, db, websiteID, currentUserID)
```

Missing and foreign websites must both return `models.ErrNotFound`. Do not repeat fetch-then-owner-check logic in every controller action.

Pageview responsibilities:

- Create pageviews
- Total pageviews
- Unique visitors
- Bounce count and bounce rate inputs
- Views per visitor inputs
- Hour and day time buckets
- Top pages and referrers
- Browser, operating system, and device breakdowns
- Country and city breakdowns

Event responsibilities:

- Create events
- Top events
- Event time buckets

Remove unused generated generic methods such as Pageview update, delete, pagination, or upsert if no existing flow uses them.

#### Services

Port geolocation because the existing dashboard exposes country and city data.

Create a dashboard query service only if it genuinely coordinates Pageview and Event aggregate methods. Do not create pass-through services.

The geo client must have a short timeout and treat lookup failure as missing geo data rather than failing collection.

#### Controllers

Add:

- `Websites`: index, new, create, show, edit, update, destroy
- `Dashboard`: show with period parsing and Inertia props
- `Collect`: OPTIONS and POST
- `Tracking`: JavaScript response

Update `Pages` so `/` is guarded and leads to the websites area.

Keep controllers as HTTP adapters. Domain validation and persistence belong in models.

#### Routes

Add route definitions under `router/routes/` for:

- Website CRUD
- Website dashboard
- Collection
- Tracking script

Apply `middleware.AuthOnly` to root, all website routes, dashboard, and sign-out.

Do not register registration or confirmation routes. Remove their controllers, Svelte pages, and now-unused workflow code.

Keep all password reset routes public.

Regenerate frontend route helpers after route changes:

```sh
andurel generate routes --json
```

## Collection and tracking behavior

Port existing behavior without adding tracking features.

Tracking script:

- Serve from `/t/script.js`
- Read `data-website-id`
- Send an initial pageview
- Expose `window.palantir.track(name, data)` for events
- Use `sendBeacon` with the existing fallback
- Return JavaScript content type and cache headers

Collection endpoint:

- Accept OPTIONS and POST at `/api/collect`
- Validate website ID, URL, type, and event name at the trust boundary
- Reject unknown websites
- Parse browser, OS, and device using the existing user-agent dependency rather than inventing a parser
- Compute the daily visitor hash from website ID, client IP, user agent, date, and configured salt
- Resolve geo data with a short timeout
- Create either a Pageview or Event through its model
- Return only appropriate status codes, without exposing internal errors

The visitor hash salt belongs in configuration and `.env.example`, not a direct `os.Getenv` call inside the controller.

## Dashboard behavior

Preserve the backup project's analytics:

- Unique visitors
- Total pageviews
- Views per visitor
- Bounce rate
- Percentage changes against the previous period
- Traffic time series
- Visitor time series
- Event time series
- Top pages
- Referrers
- Top countries
- Top cities
- Browsers
- Operating systems
- Devices
- Top events

Preserve period choices:

- Today
- Last 7 days, default
- Last 30 days
- This month
- Custom start and end dates

Use hour buckets for ranges up to 48 hours and day buckets otherwise.

Use Inertia polling every 15 seconds on the same dashboard route. Do not add a separate custom live endpoint unless Inertia polling proves insufficient.

## User interface

Initialize shadcn-svelte for the existing Vite and Tailwind setup.

Add only components that are actually used:

- sidebar
- button
- card
- input
- label
- table
- badge
- separator
- dropdown-menu
- alert-dialog
- breadcrumb
- chart

Use native `<input type="date">` controls for custom dates.

The shadcn chart component currently uses LayerChart v2 pre-release. MBV explicitly approved this tradeoff.

### Layout

```text
+-----------------------------------------------------------------------+
| Palantir  [site selector]                              [Sign out]      |
+---------------+-------------------------------------------------------+
| Websites      | example.com                     [7d] [30d] [Custom]   |
| * Example     |                                                       |
| * Business    | +----------+----------+----------+----------+         |
|               | | Visitors |Pageviews |Views/user|  Bounce  |         |
| + Add site    | +----------+----------+----------+----------+         |
|               | +-------------------------------------------+         |
| Settings      | | Traffic: pageviews + visitors             |         |
|               | |        /----\       /--------\            |         |
|               | | -------/    \-------/        \----        |         |
|               | +-------------------------------------------+         |
|               | +--------------+ +--------------+ +--------------+   |
|               | | Top pages    | | Referrers    | | Events       |   |
|               | +--------------+ +--------------+ +--------------+   |
|               | | Geo          | | Browsers/OS  | | Devices      |   |
|               | +--------------+ +--------------+ +--------------+   |
+---------------+-------------------------------------------------------+
```

Retain these pages and states:

- Sign-in with password reset link, but no registration link
- Password reset request and edit
- Website list and empty state
- Add website
- Website details with tracking snippet
- Edit website
- Delete confirmation
- Dashboard with metrics, charts, breakdowns, filters, loading, empty, and error states

The interface must be responsive, keyboard accessible, and dark themed.

## Development seed data

Extend the development seed so a fresh checkout immediately shows a useful dashboard.

Seed:

- The existing validated `admin@example.com` user
- The existing validated regular user
- Two clearly labeled demo websites owned by the admin
- A deterministic rolling 30-day set of pageviews and events

The dataset must exercise:

- Direct and referral traffic
- Multiple browsers and operating systems
- Desktop, mobile, and tablet devices
- Multiple countries and cities
- Repeat visitors and bounce visitors
- Several URLs
- Several named events
- Empty and active periods where useful for graph shape

Use generated Website, Pageview, and Event factories. Use standard-library deterministic generation and add no seed-only dependency.

Reruns must replace records belonging to fixed demo website IDs while preserving unrelated developer data. Chart totals and shapes must remain stable across reruns.

Expected files:

- `database/seeds/seeds.go`
- A focused Palantir development seed file under `database/seeds/`
- Factory options required by the seed

## Test-first contract

MBV approved all test cases and explicitly allowed focused `go test` commands for red/green development.

Write tests before implementation and confirm the expected red state.

### 1. HTTP access matrix

A table-driven handler test proves:

- Sign-in GET and POST are public
- Password reset routes are public
- Every private view redirects anonymous users to sign-in
- Registration and confirmation routes return 404
- Collection, tracking script, health, and assets remain public

### 2. Owner isolation

Using two users and two websites in migrated PostgreSQL:

- User A can read and mutate user A's website
- User A receives 404 for user B's details, edit, update, delete, and dashboard routes

### 3. Collection to dashboard

Through HTTP handlers:

1. POST a pageview.
2. POST an event.
3. Request the owner's dashboard.
4. Assert Inertia props include expected totals, series, event, device, and geo data.

### 4. Website lifecycle and validation

Through HTTP handlers:

- Create
- List
- Update
- Delete
- Invalid name/domain returns validation errors and writes no row

### 5. Date range edge cases

Use a compact table for:

- Default 7 days
- Today
- 30 days
- Current month
- Valid custom range
- Invalid custom fallback
- Hour/day bucket selection
- Previous-period boundaries

### 6. Development seed repeatability

Run the development seed twice against a migrated test database and assert:

- Demo users and sites are not duplicated
- Unrelated rows survive
- Dashboard aggregate counts remain identical

### Test mechanics

Use the existing:

- `internal/storage.NewTestCluster`
- Embedded migrations
- Generated factories

Prefer controller and handler tests. Use focused pure tests only for date arithmetic that is awkward to prove through HTTP.

Do not add Vitest or Playwright in this pass. Go HTTP tests validate server contracts and Inertia props. The Vite production build validates Svelte compilation. Browser screenshot tests can be added later only if explicitly requested.

## Implementation sequence

### 1. Tests first

- Add all approved handler integration tests.
- Add minimal date range tests.
- Add the seed repeatability test.
- Run focused `go test` commands and capture the red state.

### 2. Access lockdown and removal

- Remove registration and confirmation route registration.
- Remove their controllers, Svelte pages, and now-unused workflow code.
- Keep password reset public.
- Guard root, website routes, dashboard, and sign-out.
- Add exact public exceptions for sign-in and non-view infrastructure.

### 3. Generate and tailor models

Run serially because every generation updates `models/model.go`:

```sh
andurel project info --json
andurel generate model Website --dry-run --diff --json
andurel generate model Website --json
andurel generate model Pageview --json
andurel generate model Event --json
```

Inspect generated artifacts before editing. Generate or sync factories through Andurel.

Add only required validation, ownership, persistence, and aggregate behavior. Remove unused generic generated methods.

### 4. Port controllers and collection

Preview before mutation:

```sh
andurel generate controller Website --inertia --dry-run --diff --json
andurel generate controller Dashboard show --model-name Website --inertia --dry-run --diff --json
```

Then generate and tailor Website and Dashboard controllers.

Port Collect and Tracking manually because CRUD generation would create unused actions and views.

Port geolocation, user-agent parsing, visitor hashes, date periods, and dashboard props.

### 5. Add development seed

- Extend generated factories only where needed.
- Ensure fixed demo users and sites.
- Replace analytics for fixed demo site IDs with deterministic 30-day data.
- Verify safe repeatability.

### 6. Build shadcn interface

- Initialize shadcn-svelte for Vite.
- Add only the approved components.
- Create auth and dashboard layouts.
- Create website and dashboard pages.
- Add Inertia polling.
- Remove registration links and retain password reset.

### 7. Regenerate and verify

```sh
andurel generate routes --json
andurel generate factories --check --json
andurel routes --json
andurel doctor --json
gofmt -w <changed-go-files>
go vet ./...
pnpm exec vite build
```

Focused and broader `go test` commands are permitted for this task.

Never use:

- `npm run`
- `go build`

Manually inspect:

- Desktop and mobile layouts
- Keyboard focus
- Empty dashboard
- Seeded dashboard
- Website CRUD
- Password reset access
- Anonymous redirects
- Foreign website 404 behavior
- Cross-origin collection

## Expected file footprint

- `models/website.go`
- `models/pageview.go`
- `models/event.go`
- `models/model.go`
- `models/factories/website.go`
- `models/factories/pageview.go`
- `models/factories/event.go`
- `controllers/websites.go`
- `controllers/dashboard.go`
- `controllers/collect.go`
- `controllers/tracking.go`
- `controllers/controller.go`
- `router/routes/websites.go`
- `router/routes/collect.go`
- `router/routes/tracking.go`
- Relevant auth and CSRF middleware files
- `services/geolocation.go` and service registration
- `database/seeds/seeds.go`
- Focused development seed file
- `resources/js/Layouts/AuthLayout.svelte`
- `resources/js/Layouts/DashboardLayout.svelte`
- `resources/js/Pages/Websites/*.svelte`
- `resources/js/Pages/Dashboard/Show.svelte`
- App-level components under `resources/js/Components/`
- Generated shadcn files under `resources/js/components/ui/`
- `components.json`
- `css/base.css`
- `package.json` and `pnpm-lock.yaml`
- Generated `resources/js/routes.ts`
- Controller/model/seed test files
- `.env.example` visitor hash salt setting

Exact shadcn-generated UI files depend on the current CLI. No migration files are expected.

## Acceptance criteria

- Sign-in and password reset are the only public application views.
- Registration and confirmation code is removed.
- Anonymous requests cannot reach root, website, dashboard, or sign-out handlers.
- Public tracking infrastructure works cross-origin without broadly weakening security.
- Users cannot discover or mutate another user's website.
- Website CRUD and tracking snippet work.
- Pageviews and events flow from collection into dashboard aggregates.
- All analytics from the backup project are represented.
- Period filters and 15 second refresh work.
- The dashboard uses shadcn-svelte and approved charts.
- Development seed data creates a stable, complete demo and is safe to rerun.
- Approved tests pass.
- Andurel doctor, route generation, factory checks, Go vet, and Vite build pass.
