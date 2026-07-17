# Palantir

Palantir is a self-hosted, privacy-focused web analytics application. It provides useful traffic and event insights without cookies or invasive visitor profiles.

A small tracking script records pageviews and explicit events. The authenticated dashboard presents trends, referrers, popular pages, geography, browser details, devices, and operating systems across multiple websites.

## Features

- Cookie-free pageview tracking
- Daily salted visitor hashes for unique visitor metrics
- Pageviews, unique visitors, views per visitor, and bounce rate
- Interactive metric charts with today, 7-day, 30-day, monthly, and custom ranges
- Top pages and referrer breakdowns
- Country heat map plus country and city tables
- Browser, operating system, and device breakdowns
- Automatic click events through HTML attributes
- Manual JavaScript events with optional structured data
- Event trends and top-event summaries
- Multiple websites per account
- Live dashboard refresh
- Cross-origin collection with configurable CORS controls

## Tracking setup

Create a website in Palantir and open its **Tracking setup** page. Palantir provides the correct script URL and website ID.

Add the generated snippet before the closing `</head>` tag:

```html
<script
  defer
  data-website-id="YOUR_WEBSITE_ID"
  src="https://analytics.example.com/t/script.js"
></script>
```

The script automatically records a pageview when it loads.

### Track click events

Add `data-palantir-event` to any clickable element:

```html
<button data-palantir-event="pricing-cta">
  Start free trial
</button>
```

Clicking the element records `pricing-cta` and makes it available in the event dashboard.

### Track events manually

Use the browser API for events triggered from application code:

```js
window.palantir.track('purchase', { plan: 'pro' })
```

The first argument is the event name. The optional second argument is stored as JSON event data.

## Run locally

### Requirements

- Go 1.26.5 or newer
- PostgreSQL
- pnpm
- [Andurel CLI](https://github.com/mbvlabs/andurel)

Install Andurel if needed:

```bash
go install github.com/mbvlabs/andurel@latest
```

Set up and run Palantir:

```bash
cp .env.example .env
pnpm install
andurel database create
andurel database migrate up
andurel run
```

The application runs at `http://localhost:8080` by default. Update `.env` for your database, application URL, session secrets, visitor hash salt, and allowed collection origins.

Useful commands:

```bash
andurel database migrate status
andurel database migrate new migration_name
andurel database migrate up
andurel database seed --list
andurel run
```

## Configuration

See [`.env.example`](.env.example) for all available settings. Important values include:

| Variable | Purpose |
| --- | --- |
| `DOMAIN` and `PROTOCOL` | Public application URL |
| `DB_*` | PostgreSQL connection |
| `SESSION_KEY` | Session signing key |
| `SESSION_ENCRYPTION_KEY` | Session encryption key |
| `VISITOR_HASH_SALT` | Salt used for daily visitor hashes |
| `CORS_ALLOWED_ORIGINS` | Additional sites allowed to submit analytics |
| `CSRF_STRATEGY` | CSRF enforcement mode |

Use fresh random values for all secrets in production.

## Built with Andurel

Palantir is built with [Andurel](https://github.com/mbvlabs/andurel), a full-stack Go web framework. Andurel provides the project structure, dependency wiring, routing conventions, database tooling, migrations, development server, and code generation workflows used by this repository.

The application combines:

- Go and Echo for HTTP handling
- Bun and PostgreSQL for persistence
- Inertia and Svelte for the dashboard
- Tailwind CSS for styling
- LayerChart for analytics charts
- River for background jobs

For framework commands, project conventions, or generator documentation, refer to the [Andurel repository](https://github.com/mbvlabs/andurel).
