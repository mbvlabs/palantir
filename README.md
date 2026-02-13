# palantir

A full-stack web application built with [Andurel](https://github.com/mbvlabs/andurel), a Rails-like web framework for Go that prioritizes development speed.

## Project Structure

```
palantir/
├── assets/              # Static assets (compiled CSS, images)
├── bin/                 # Command-line tools
│   ├── app              # Main application binary
│   ├── console          # Database console
│   ├── migration        # Migration runner
│   └── shadowfax        # Development server
├── cmd/                 # Command entry points
│   └── app/             # Main web application
├── clients/             # External service clients
├── config/              # Application configuration
├── controllers/         # HTTP request handlers
├── css/                 # Source CSS files (Tailwind)
├── database/
│   ├── migrations/      # SQL migration files
│   └── queries/         # SQLC query definitions
├── email/               # Email templates and sending
├── models/              # Data models and business logic
│   └── internal/db/     # Generated SQLC code (don't edit)
├── queue/               # Background job processing
│   ├── jobs/            # Job definitions
│   └── workers/         # Worker implementations
├── router/              # Routes and middleware
│   ├── routes/          # Route definitions
│   ├── cookies/         # Session helpers
│   └── middleware/      # Custom middleware
├── pkg/
│ 	└──telemetry/        # Observability (logs, traces, metrics)
├── views/               # Templ templates
├── .env.example         # Example environment configuration
└── go.mod               # Go dependencies
```

## Quick Start

### Prerequisites

- Go 1.24.4 or higher
- PostgreSQL database
- Andurel CLI: `go install github.com/mbvlabs/andurel@latest`

### Setup

1. **Configure environment**
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

2. **Create database**
   ```bash
   createdb palantir_development
   ```

3. **Run migrations**
   ```bash
   andurel migration up
   ```

4. **Start the development server**
   ```bash
   andurel run
   ```

Your application is now running at `http://localhost:8080` with live reload for Go, Templ, and CSS changes!

## Available Commands

### Development Server

```bash
# Run development server with hot reload for Go, Templ, and CSS
andurel run
```

This orchestrates Air (Go), Templ watch, and Tailwind CSS compilation.

### Database Console

```bash
# Open interactive database console
andurel app console
```

Provides a SQL console connected to your database for ad-hoc queries and exploration.

### Migration Management

```bash
# Create a new migration
andurel migration new create_users_table

# Run all pending migrations
andurel migration up

# Rollback last migration
andurel migration down

# Rollback to specific version
andurel migration down-to [version]

# Apply up to specific version
andurel migration up-to [version]

# Reset database (rollback all, then reapply)
andurel migration reset

# Fix migration version gaps
andurel migration fix
```

## How-To Guides

### Generate a New Resource

The Andurel generator creates complete CRUD resources with models, controllers, views, and routes.

**Prerequisites**: You need a database table first. Create a migration:

```bash
# 1. Create a migration for your table
andurel migration new create_products_table
```

Edit the generated migration file in `database/migrations/` to define your table schema:

```sql
-- +goose Up
CREATE TABLE products (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    price DECIMAL(10, 2) NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

-- +goose Down
DROP TABLE products;
```

Apply the migration:

```bash
andurel migration up
```

**Generate the resource**:

```bash
# Generate model + controller + views + routes
andurel generate resource Product

# Or use shorthand
andurel g resource Product
```

This creates:
- `models/product.go` - Data model with CRUD methods
- `controllers/products.go` - HTTP handlers for CRUD operations
- `views/products_*.templ` - Template files for all CRUD views
- Routes automatically registered in `router/routes/products.go`

The generator also:
- Updates `database/queries/` with SQLC queries
- Regenerates SQLC code
- Creates a complete CRUD interface at `/products`

**Custom table names**: If your table doesn't follow Rails naming conventions (model `Product` → table `products`):

```bash
# Map Product model to a custom table name
andurel g resource Product --table products_catalog
```

**Individual components**:

```bash
# Generate only the model
andurel g model Product

# Generate controller with views
andurel g controller Product --with-views

# Generate views with controller
andurel g view Product --with-controller

# Refresh model after schema changes
andurel g model Product --refresh
```

### Setup Background Jobs

This project uses [River](https://riverqueue.com/) for background job processing with PostgreSQL.

**1. Define a job**

Create a new job type in `queue/jobs/`:

```go
// queue/jobs/my_job.go
package jobs

type MyJobArgs struct {
    UserID   string
    Action   string
}

func (MyJobArgs) Kind() string { return "my_job" }
```

**2. Implement a worker**

Create the worker in `queue/workers/`:

```go
// queue/workers/my_job.go
package workers

import (
    "context"
    "palantir/queue/jobs"
)

func ProcessMyJob(ctx context.Context, msg []byte) error {
    // Your job logic here
    // Unmarshal msg to jobs.MyJobArgs and process
    return nil
}
```

**3. Register the worker**

Add your worker to `queue/workers/workers.go`:

```go
// Register in your queue setup
```

**4. Enqueue jobs**

From anywhere in your application:

```go
import "palantir/queue/jobs"

// Enqueue a job through your queue client
err := queue.Enqueue(ctx, jobs.MyJobArgs{
    UserID: "123",
    Action: "send_welcome_email",
})
```


**Job Options**

Customize job behavior:

```go
// Configure retry behavior and priorities in your queue setup
```

### Send Emails

This project includes built-in email functionality with Mailpit for development testing.

**1. Create an email template**

Add your template in `email/`:

```go
// email/welcome.templ
package email

templ WelcomeEmail(userName string) {
    @BaseLayout() {
        <h1>Welcome, { userName }!</h1>
        <p>Thank you for joining us.</p>
    }
}
```

**2. Send the email**

```go
import "palantir/email"

// Send an email
data := email.TransactionalData{
    From:    "noreply@example.com",
    To:      []string{"user@example.com"},
    Subject: "Welcome!",
    Body:    WelcomeEmail("John Doe"),
}

err := email.SendTransactional(ctx, data, sender)
```

**3. Background email jobs**

For better performance, send emails asynchronously:

```go
// Enqueue email job through your queue
```

**Development Testing**

Emails are sent to Mailpit in development. Access the web UI at `http://localhost:8025` to view sent emails.

### Working with the Database

**Add queries**

Create SQL queries in `database/queries/`:

```sql
-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1 LIMIT 1;

-- name: ListActiveUsers :many
SELECT * FROM users WHERE active = true ORDER BY created_at DESC;
```

**Generate type-safe code**

```bash
andurel sqlc generate
```

**Use generated code**

```go
import "palantir/models/internal/db"

user, err := queries.GetUserByEmail(ctx, "user@example.com")
users, err := queries.ListActiveUsers(ctx)
```

### Schema Changes

When modifying your database schema:

```bash
# 1. Create a migration
andurel migration new add_email_to_users

# 2. Edit the migration file
# Add your ALTER TABLE statements

# 3. Apply the migration
andurel migration up

# 4. Update queries if needed (in database/queries/)

# 5. Regenerate SQLC code
andurel sqlc generate

# 6. Refresh affected models
andurel g model User --refresh
```

### Customize Styling

This project uses Tailwind CSS. Customize your theme in `css/theme.css`:

```css
@layer theme {
  :root {
    --color-primary: theme('colors.blue.600');
    --color-secondary: theme('colors.gray.600');
  }
}
```

Add custom utilities in `css/base.css`. The development server automatically rebuilds CSS on changes.


## Environment Configuration

Key environment variables (see `.env.example` for all options):

```bash
# Application
ENVIRONMENT=development
HOST=localhost
PORT=8080
PROJECT_NAME=palantir
DOMAIN=localhost:8080
PROTOCOL=http

# Database
DB_KIND=postgres
DB_HOST=127.0.0.1
DB_PORT=5432
DB_NAME=palantir_development
DB_USER=postgres
DB_PASSWORD=postgres
DB_SSL_MODE=disable

# Email (Mailpit for development)
MAILPIT_HOST=0.0.0.0
MAILPIT_PORT=1025
DEFAULT_SENDER_SIGNATURE=info@palantir.com

# Security (auto-generated during scaffolding)
SESSION_KEY=<auto-generated>
SESSION_ENCRYPTION_KEY=<auto-generated>
TOKEN_SIGNING_KEY=<auto-generated>
PEPPER=<auto-generated>

# CSRF (Rails-style)
CSRF_STRATEGY=header_only
CSRF_TRUSTED_ORIGINS=

# Telemetry (optional)
TELEMETRY_SERVICE_NAME=palantir
TELEMETRY_SERVICE_VERSION=1.0.0
OTLP_LOGS_ENDPOINT=
OTLP_METRICS_ENDPOINT=
OTLP_TRACES_ENDPOINT=
TRACE_SAMPLE_RATE=1.0
```

## CSRF Protection

CSRF protection is always enabled for HTML routes and uses Fetch Metadata to align with Rails behavior. API and asset routes skip CSRF checks.

**Strategies** (`CSRF_STRATEGY`):
- `header_only` (default): Unsafe requests must include the `Sec-Fetch-Site` header. Requests missing this header are rejected with `403`.
- `header_or_legacy_token`: Allows legacy form tokens when `Sec-Fetch-Site` is missing. Forms must submit `_csrf` or send `X-CSRF-Token` header.

**Trusted origins**:
- The base URL (`PROTOCOL` + `DOMAIN`) is always trusted automatically.
- `CSRF_TRUSTED_ORIGINS` accepts a comma-separated list of additional origins (e.g., `https://api.example.com,https://admin.example.com`).

**Client/testing tips**:
- For unsafe requests in tests or custom clients, include `Sec-Fetch-Site: same-origin`.
- When using `header_or_legacy_token`, submit `_csrf` with forms or send `X-CSRF-Token` header.

## Development Tips

1. **Live Reload**: Use `andurel run` during development for automatic reloading
2. **Type Safety**: Let SQLC and Templ catch errors at compile time
3. **Database Console**: Use `andurel app console` for quick database queries
4. **Hot Reload**: Changes to Go, Templ, or CSS automatically trigger rebuilds
5. **Tailwind**: Use Tailwind's utility classes in your Templ templates

## Common Tasks

```bash
# Start development
andurel run

# Create a new resource
andurel g resource Product

# Add a migration
andurel migration new add_field_to_products

# Run migrations
andurel migration up

# Regenerate SQLC code
andurel sqlc generate

# Access database console
andurel app console

# Run tests
go test ./...
```

## Integration Testing

This project includes a built-in integration testing framework that makes it easy to test controllers and models with real database interactions.

### Test Infrastructure

The framework provides:
- **Automatic test database setup**: Uses [testcontainers](https://golang.testcontainers.org/) to spin up PostgreSQL in Docker
- **Transaction-based tests**: Each test runs in a transaction that automatically rolls back, keeping tests isolated
- **Factory pattern**: Simple builders for creating test data with sensible defaults

### Writing Controller Tests

**1. Create a test file** (e.g., `controllers/products_controller_test.go`):

```go
package controllers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/jackc/pgx/v5"
	"palantir/controllers"
	"palantir/database"
	"palantir/models"
	"palantir/models/factories"
)

var testDB *database.TestDB

func TestMain(m *testing.M) {
	var err error
	testDB, err = database.NewTestDB()
	if err != nil {
		panic(err)
	}
	defer testDB.Close()

	m.Run()
}

func TestProducts_Create(t *testing.T) {
	testDB.WithTx(t, func(tx pgx.Tx) {
		controller := controllers.NewProducts(testDB.DB)

		// Create test request
		e := echo.New()
		req := httptest.NewRequest(http.MethodPost, "/products", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// Test the controller action
		err := controller.Create(c)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}

		// Assert database state
		products, err := models.AllProducts(c.Request().Context(), tx)
		if err != nil {
			t.Fatalf("failed to query products: %v", err)
		}

		if len(products) != 1 {
			t.Errorf("expected 1 product, got %d", len(products))
		}
	})
}
```

### Creating Test Factories

**1. Create a factory** in `models/factories/product_factory.go`:

```go
package factories

import (
	"palantir/models"
)

type ProductBuilder struct {
	data models.CreateProductData
}

func Product() *ProductBuilder {
	return &ProductBuilder{
		data: models.CreateProductData{
			Name:        "Test Product",
			Description: "Test description",
			Price:       "29.99",
		},
	}
}

func (b *ProductBuilder) WithName(name string) *ProductBuilder {
	b.data.Name = name
	return b
}

func (b *ProductBuilder) WithPrice(price string) *ProductBuilder {
	b.data.Price = price
	return b
}

func (b *ProductBuilder) Create(dbtx DBTX) models.Product {
	product, err := models.CreateProduct(ctx, dbtx, b.data)
	if err != nil {
		panic(err)
	}
	return product
}

func (b *ProductBuilder) Build() models.CreateProductData {
	return b.data
}
```

**2. Use factories in tests**:

```go
func TestProducts_Show(t *testing.T) {
	testDB.WithTx(t, func(tx pgx.Tx) {
		// Create test data with default values
		product := factories.Product().Create(tx)

		// Or customize specific fields
		premiumProduct := factories.Product().
			WithName("Premium Product").
			WithPrice("99.99").
			Create(tx)

		// Test your controller with the created data
		// ...
	})
}
```

### Testing Patterns

**Test database queries**:

```go
func TestFindProduct(t *testing.T) {
	testDB.WithTx(t, func(tx pgx.Tx) {
		product := factories.Product().Create(tx)

		found, err := models.FindProduct(context.Background(), tx, product.ID)
		if err != nil {
			t.Fatalf("FindProduct failed: %v", err)
		}

		if found.Name != product.Name {
			t.Errorf("expected name %s, got %s", product.Name, found.Name)
		}
	})
}
```

**Test with multiple records**:

```go
func TestPaginateProducts(t *testing.T) {
	testDB.WithTx(t, func(tx pgx.Tx) {
		// Create test data
		for i := 0; i < 25; i++ {
			factories.Product().Create(tx)
		}

		// Test pagination
		result, err := models.PaginateProducts(context.Background(), tx, 1, 10)
		if err != nil {
			t.Fatalf("PaginateProducts failed: %v", err)
		}

		if len(result.Products) != 10 {
			t.Errorf("expected 10 products, got %d", len(result.Products))
		}

		if result.TotalCount != 25 {
			t.Errorf("expected total count 25, got %d", result.TotalCount)
		}
	})
}
```

**Test with related data**:

```go
func TestCreateOrder(t *testing.T) {
	testDB.WithTx(t, func(tx pgx.Tx) {
		// Create dependencies
		user := factories.User().Create(tx)
		product := factories.Product().Create(tx)

		// Test order creation
		order := factories.Order().
			WithUserID(user.ID).
			WithProductID(product.ID).
			Create(tx)

		if order.UserID != user.ID {
			t.Errorf("order user_id mismatch")
		}
	})
}
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests in a specific package
go test ./controllers

# Run tests with coverage
go test -cover ./...

# Run a specific test
go test ./controllers -run TestProducts_Create

# Verbose output
go test -v ./...
```

### Test Database Setup

**Prerequisites**: Docker must be running to use testcontainers.

The test helper automatically:
1. Starts a PostgreSQL container with `postgres:17-alpine`
2. Runs all migrations from `database/migrations/`
3. Provides a clean database for each test suite
4. Cleans up containers when tests complete

**Note**: The first test run will download the PostgreSQL Docker image, which may take a moment.

### Best Practices

1. **Use transactions**: Always wrap test logic in `testDB.WithTx()` for automatic cleanup
2. **Use factories**: Create test data with factories instead of manual model creation
3. **Test isolation**: Each test should be independent and not rely on other tests
4. **Descriptive names**: Name tests clearly (e.g., `TestProducts_Create_WithInvalidData`)
5. **Assert clearly**: Check both success cases and expected database state
6. **Don't test frameworks**: Focus on your business logic, not Echo or SQLC behavior

## Learn More

- [Andurel Documentation](https://github.com/mbvlabs/andurel)
- [Echo Framework](https://echo.labstack.com/)
- [SQLC](https://sqlc.dev/)
- [Templ](https://templ.guide/)
- [Datastar](https://data-star.dev/)
- [goqite](https://github.com/maragudk/goqite)
- [OpenTelemetry](https://opentelemetry.io/)

## Getting Help

For Andurel-specific questions and issues:
- GitHub Issues: https://github.com/mbvlabs/andurel/issues
- Documentation: https://github.com/mbvlabs/andurel

## License

This project is licensed under the MIT License.
