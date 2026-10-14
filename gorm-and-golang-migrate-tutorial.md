# GORM and golang-migrate: A Complete Tutorial

This tutorial teaches you how to build the database layer of a real Go application. It uses PostgreSQL for storage, GORM for reading and writing data, and golang-migrate for changing the database structure over time.

Everything is built around one small but realistic project: an online shop with users, products, categories, orders, and wallets. The code follows the habits you would want in a production service: explicit SQL migrations, context on every query, proper error handling, and transactions wherever money or stock is involved.

## Before you start

You need:

- A current Go release installed
- Docker with Docker Compose, to run PostgreSQL locally
- Basic Go: structs, interfaces, errors, packages
- Basic SQL: SELECT, INSERT, UPDATE, JOIN, primary keys, foreign keys

Tools used: GORM v2 (`gorm.io/gorm`) with its PostgreSQL driver, golang-migrate v4, and PostgreSQL 17. Newer versions will likely be out when you read this. The ideas stay the same, but check each project's release notes before upgrading.

## Contents

1. [The big picture](#lesson-1-the-big-picture)
2. [Project setup](#lesson-2-project-setup)
3. [Connecting with GORM](#lesson-3-connecting-with-gorm)
4. [golang-migrate fundamentals](#lesson-4-golang-migrate-fundamentals)
5. [Writing the shop migrations](#lesson-5-writing-the-shop-migrations)
6. [Running migrations from Go code](#lesson-6-running-migrations-from-go-code)
7. [Models and conventions](#lesson-7-models-and-conventions)
8. [CRUD with GORM](#lesson-8-crud-with-gorm)
9. [Associations](#lesson-9-associations)
10. [Transactions, locking, and money](#lesson-10-transactions-locking-and-money)
11. [The query toolbox](#lesson-11-the-query-toolbox)
12. [The generics API](#lesson-12-the-generics-api)
13. [Testing the database layer](#lesson-13-testing-the-database-layer)
14. [Changing a live schema safely](#lesson-14-changing-a-live-schema-safely)
15. [Production checklist](#lesson-15-production-checklist)
16. [Common mistakes](#lesson-16-common-mistakes)
- [Appendix A: AutoMigrate](#appendix-a-automigrate)
- [Appendix B: Practice tasks](#appendix-b-practice-tasks)
- [Appendix C: Command cheat sheet](#appendix-c-command-cheat-sheet)

## Lesson 1: The big picture

### What a migration is

Think of a building. The database schema is the floor plan: which tables exist, which columns they have, which rules they enforce. A migration is one numbered revision of that floor plan. Revision 1 pours the foundation. Revision 2 adds the first floor. Revision 3 removes a wall and adds a door.

Because every revision is saved in order, anyone can build an identical copy of the building by applying the revisions from the start. Anyone can also tell exactly which revision a particular building is on.

Without migrations, schema changes happen when someone runs SQL by hand on their laptop and then tries to remember to repeat it on staging and production. Sooner or later one environment drifts from the others, and you meet errors like "column does not exist" at the worst possible time.

Migrations give your database what Git gives your code: history, order, and the ability to reproduce any state.

### What GORM is

An ORM (object-relational mapper) translates between rows in tables and structs in your program. You write `db.Create(&user)` and GORM writes the INSERT statement for you.

Think of a professional interpreter at a meeting. The interpreter saves effort and the conversation reads naturally, but you still need to know roughly what is being said in both languages. If you do not understand the SQL GORM produces, you will write slow or wrong queries without noticing. Throughout this tutorial you will see the SQL behind the GORM calls.

### Who owns the schema?

There are two common ways to combine GORM with a database structure.

**GORM owns the schema (AutoMigrate).** Your structs are the source of truth. At startup GORM compares them with the database and alters tables to match. This is fast for prototypes.

**SQL files own the schema (golang-migrate).** You write numbered SQL files. They are the source of truth. GORM only reads and writes rows.

This tutorial uses the second approach, for these reasons:

- Every change is explicit SQL that a teammate can review in a pull request.
- You can use database features that struct tags cannot express: partial unique indexes, CHECK constraints, triggers, and so on.
- You can write down paths, backfill data, and rename things. AutoMigrate never drops or renames columns.
- The exact same files run on a laptop, in CI, and in production.

The cost is that your structs and your SQL are two places that must agree. Lesson 13 shows a small test that catches disagreement.

One rule matters more than the rest: **pick one owner per table.** Never let AutoMigrate and SQL migrations both change the same table.

### Division of labor

- golang-migrate creates and changes tables, columns, indexes, and constraints.
- GORM inserts, reads, updates, and deletes rows, and runs transactions.
- Your structs describe the shape of rows, and nothing else.

[Back to contents](#contents)

## Lesson 2: Project setup

### Create the module

```bash
mkdir shop && cd shop
go mod init shop
go get gorm.io/gorm gorm.io/driver/postgres
go get github.com/golang-migrate/migrate/v4
mkdir -p cmd/api cmd/migrate db/migrations internal/database internal/models internal/store internal/testdb
```

### Run PostgreSQL with Docker

Create `docker-compose.yml`:

```yaml
services:
  db:
    image: postgres:17
    environment:
      POSTGRES_USER: app
      POSTGRES_PASSWORD: secret
      POSTGRES_DB: shop
    ports:
      - "5432:5432"
    volumes:
      - pgdata:/var/lib/postgresql/data

volumes:
  pgdata:
```

Start it, then create a second database for tests:

```bash
docker compose up -d
docker compose exec db createdb -U app shop_test
docker compose exec db psql -U app -d shop -c 'select version();'
```

### Connection strings

Both GORM and golang-migrate can read the same URL, as long as it only contains standard options:

```bash
export DATABASE_URL='postgres://app:secret@localhost:5432/shop?sslmode=disable'
export TEST_DATABASE_URL='postgres://app:secret@localhost:5432/shop_test?sslmode=disable'
```

Read the URL from the environment and never commit real credentials. If you use a `.env` file, add it to `.gitignore` and commit a `.env.example` instead.

### Folder layout

```
shop/
  cmd/
    api/main.go            the application entry point
    migrate/main.go        a small migration runner
  db/
    migrations/            numbered SQL files, plus embed.go
  internal/
    database/              connection and migrator setup
    models/                structs that mirror table rows
    store/                 queries, grouped by area (users, products, orders)
    testdb/                helper for tests that need a real database
  docker-compose.yml
```

The `store` package is a thin repository layer. HTTP handlers call stores. Stores call GORM. Nothing outside `store` should build queries. That keeps SQL knowledge in one place and makes handlers easy to test.

[Back to contents](#contents)

## Lesson 3: Connecting with GORM

### The connection code

Create `internal/database/database.go`:

```go
package database

import (
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Config struct {
	URL          string
	MaxOpenConns int
	MaxIdleConns int
	ConnMaxLife  time.Duration
	LogLevel     logger.LogLevel
}

func Connect(cfg Config) (*gorm.DB, error) {
	gormLogger := logger.New(
		log.New(os.Stdout, "", log.LstdFlags),
		logger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  cfg.LogLevel,
			IgnoreRecordNotFoundError: true,
			ParameterizedQueries:      true,
		},
	)

	db, err := gorm.Open(postgres.Open(cfg.URL), &gorm.Config{
		Logger:         gormLogger,
		TranslateError: true,
		NowFunc:        func() time.Time { return time.Now().UTC() },
	})
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql.DB: %w", err)
	}
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLife)

	return db, nil
}
```

`gorm.Open` pings the database by default, so a wrong password or a stopped container fails right here rather than on the first query.

### What each option does

**Logger.** GORM logs every SQL statement at `logger.Info`, only slow queries and errors at `logger.Warn`, and nothing at `logger.Silent`. Use Info while learning and Warn in production. `SlowThreshold` flags queries slower than the limit you pick. `IgnoreRecordNotFoundError` stops the normal "no rows" case from filling your logs with errors. `ParameterizedQueries` prints `$1, $2` instead of real values, which keeps emails and other personal data out of your logs.

**TranslateError.** Without it, a duplicate email comes back as a raw driver error that you would have to inspect by error code. With it, GORM converts common database errors into its own: `gorm.ErrDuplicatedKey` for unique violations and `gorm.ErrForeignKeyViolated` for foreign key violations. Your code can then use `errors.Is`.

**NowFunc.** GORM uses it for `CreatedAt` and `UpdatedAt`. Forcing UTC avoids surprises when servers run in different time zones.

### The connection pool

`*gorm.DB` wraps Go's standard `database/sql` pool. Think of a supermarket with checkout counters:

- `MaxOpenConns` is the total number of counters. Once all are busy, new customers wait in line.
- `MaxIdleConns` is how many counters stay staffed when it is quiet, so the next customer does not wait for someone to arrive.
- `ConnMaxLifetime` retires a counter after a while, so nothing stays open forever and gets silently cut by a firewall or load balancer.

PostgreSQL defaults to 100 total connections. If you run five copies of your app with 25 open connections each, you are already over. A sensible start is 20 open, 10 idle, and a 30 minute lifetime. Then measure.

### A first program

Create `cmd/api/main.go`:

```go
package main

import (
	"log"
	"os"
	"time"

	"gorm.io/gorm/logger"

	"shop/internal/database"
)

func main() {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	db, err := database.Connect(database.Config{
		URL:          url,
		MaxOpenConns: 20,
		MaxIdleConns: 10,
		ConnMaxLife:  30 * time.Minute,
		LogLevel:     logger.Info,
	})
	if err != nil {
		log.Fatal(err)
	}

	var one int
	if err := db.Raw("SELECT 1").Scan(&one).Error; err != nil {
		log.Fatal(err)
	}
	log.Println("connected, SELECT 1 returned", one)
}
```

Run it with `go run ./cmd/api`.

### Two rules about *gorm.DB

**Always pass a context.** Start queries with `db.WithContext(ctx)`. When a request is cancelled or times out, the query is cancelled too, and the connection returns to the pool.

**Do not reuse a half-built query.** The `db` you get from `Connect` is safe to share. But once you chain onto it, the result remembers its conditions:

```go
q := db.Model(&models.Product{}).Where("stock > 0")
q.Where("price_cents < 1000").Find(&cheap)
q.Where("price_cents >= 1000").Find(&pricey) // wrong: also carries price_cents < 1000
```

The second query silently has both conditions. Either build each query from `db` fresh, or call `.Session(&gorm.Session{})` on the shared part, which makes every later chain start from a clean copy:

```go
base := db.Model(&models.Product{}).Where("stock > 0").Session(&gorm.Session{})
base.Where("price_cents < 1000").Find(&cheap)
base.Where("price_cents >= 1000").Find(&pricey) // correct
```

This is the most common source of mysterious GORM bugs. Remember it.

[Back to contents](#contents)

## Lesson 4: golang-migrate fundamentals

### Install the CLI

Pick one:

```bash
# macOS or Linux with Homebrew
brew install golang-migrate

# Any platform with Go installed (the tag selects the database driver)
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

Check it with `migrate -version`. Prebuilt binaries and a Docker image (`migrate/migrate`) also exist. The library you added with `go get` is what your Go code will use. The CLI is a convenience for day to day work.

### How it works

A migration is a pair of files:

```
000001_create_users_and_profiles.up.sql
000001_create_users_and_profiles.down.sql
```

- The number is the **version**. Migrations always run in numeric order.
- `up` moves the database forward. `down` undoes it.
- The text after the number is only a label for humans.

golang-migrate keeps its state in one table that it creates by itself, `schema_migrations`. It holds exactly one row with two columns: the current `version`, and a `dirty` flag. It does not store a list of every applied migration. "Up" means: run every file with a version higher than the stored one, in order.

### The commands you will use

```bash
# create an empty pair of files with the next sequence number
migrate create -ext sql -dir db/migrations -seq create_users_and_profiles

# apply all pending migrations, or just the next N
migrate -path db/migrations -database "$DATABASE_URL" up
migrate -path db/migrations -database "$DATABASE_URL" up 1

# undo the last migration, or the last N
migrate -path db/migrations -database "$DATABASE_URL" down 1

# jump to a specific version (goes up or down as needed)
migrate -path db/migrations -database "$DATABASE_URL" goto 3

# show the current version
migrate -path db/migrations -database "$DATABASE_URL" version

# repair command, explained below
migrate -path db/migrations -database "$DATABASE_URL" force 2

# delete everything in the database (local development only)
migrate -path db/migrations -database "$DATABASE_URL" drop
```

`down` with no number is meant for undoing everything. Avoid it outside local development.

### Sequence numbers or timestamps

`-seq` gives you `000001`, `000002`, and so on. Without it, the file names get a timestamp prefix.

Sequence numbers are easy to read, but two people who each create `000005` on separate branches will collide at merge time. Timestamps almost never collide, but the names are long. A small team can use sequence numbers and let the merge conflict act as a reminder to renumber. A larger team usually prefers timestamps. Either is fine. Pick one and keep it.

### The rules

1. **Never edit a migration that has been applied anywhere shared.** Someone else's database already ran the old version. Write a new migration instead.
2. **One purpose per migration.** "Create orders tables" is fine. "Create orders and also fix some user columns" is not.
3. **Wrap each file in `BEGIN;` and `COMMIT;`.** PostgreSQL can undo schema changes inside a transaction, so a failure halfway through leaves nothing half applied. There is one exception, covered in Lesson 14.
4. **Always write the down file, and test it.** Even if production never uses it, development and CI will.
5. **Keep migrations in version control** next to the code that needs them.

### The dirty state

Before running a migration, golang-migrate records its version with `dirty = true`. If the migration succeeds, it flips the flag back to false. If the migration fails, the flag stays true and every later command refuses to run.

Think of a renovation that stopped halfway. The site gets a "do not enter" sign until a person inspects the damage. The tool cannot know whether your SQL did half its work, so it makes you decide.

To recover:

1. Read the error and look at the database. If you used `BEGIN` and `COMMIT`, nothing from the failed file was kept.
2. Fix the SQL file (this is allowed because the migration never completed).
3. Run `force N`, where N is the last version that fully succeeded. `force` only rewrites the version row and clears the flag. It runs no SQL.
4. Run `up` again.

### Concurrent runs

The PostgreSQL driver takes an advisory lock while migrating. If two copies of your app start at the same moment and both try to migrate, one waits for the other. You do not need to build your own locking.

[Back to contents](#contents)

## Lesson 5: Writing the shop migrations

Create the first pair of files:

```bash
migrate create -ext sql -dir db/migrations -seq create_users_and_profiles
migrate create -ext sql -dir db/migrations -seq create_catalog
migrate create -ext sql -dir db/migrations -seq create_orders
migrate create -ext sql -dir db/migrations -seq create_wallets
```

Then fill them in. These files use block comments (`/* ... */`) for notes.

### 000001: users and profiles

`000001_create_users_and_profiles.up.sql`:

```sql
BEGIN;

CREATE TABLE users (
    id            BIGSERIAL   PRIMARY KEY,
    email         TEXT        NOT NULL,
    name          TEXT        NOT NULL,
    password_hash TEXT        NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at    TIMESTAMPTZ
);

/* Unique only among rows that are not soft deleted, so a deleted
   account does not block someone from registering the same email later. */
CREATE UNIQUE INDEX users_email_active_key ON users (email) WHERE deleted_at IS NULL;

CREATE TABLE profiles (
    id         BIGSERIAL PRIMARY KEY,
    user_id    BIGINT    NOT NULL UNIQUE REFERENCES users (id) ON DELETE CASCADE,
    bio        TEXT      NOT NULL DEFAULT '',
    avatar_url TEXT      NOT NULL DEFAULT ''
);

COMMIT;
```

`000001_create_users_and_profiles.down.sql`:

```sql
BEGIN;

DROP TABLE IF EXISTS profiles;
DROP TABLE IF EXISTS users;

COMMIT;
```

The down file drops tables in the reverse order of creation, because `profiles` depends on `users`.

### 000002: the catalog

`000002_create_catalog.up.sql`:

```sql
BEGIN;

CREATE TABLE categories (
    id   BIGSERIAL PRIMARY KEY,
    name TEXT      NOT NULL UNIQUE
);

CREATE TABLE products (
    id          BIGSERIAL   PRIMARY KEY,
    sku         TEXT        NOT NULL,
    name        TEXT        NOT NULL,
    price_cents BIGINT      NOT NULL CHECK (price_cents >= 0),
    stock       INTEGER     NOT NULL DEFAULT 0 CHECK (stock >= 0),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at  TIMESTAMPTZ
);

CREATE UNIQUE INDEX products_sku_active_key ON products (sku) WHERE deleted_at IS NULL;

/* Join table for the many-to-many link between products and categories. */
CREATE TABLE product_categories (
    product_id  BIGINT NOT NULL REFERENCES products (id)   ON DELETE CASCADE,
    category_id BIGINT NOT NULL REFERENCES categories (id) ON DELETE CASCADE,
    PRIMARY KEY (product_id, category_id)
);

CREATE INDEX product_categories_category_id_idx ON product_categories (category_id);

COMMIT;
```

`000002_create_catalog.down.sql`:

```sql
BEGIN;

DROP TABLE IF EXISTS product_categories;
DROP TABLE IF EXISTS products;
DROP TABLE IF EXISTS categories;

COMMIT;
```

### 000003: orders

`000003_create_orders.up.sql`:

```sql
BEGIN;

CREATE TABLE orders (
    id          BIGSERIAL   PRIMARY KEY,
    user_id     BIGINT      NOT NULL REFERENCES users (id),
    status      TEXT        NOT NULL DEFAULT 'pending'
                CHECK (status IN ('pending', 'paid', 'shipped', 'cancelled')),
    total_cents BIGINT      NOT NULL CHECK (total_cents >= 0),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX orders_user_id_idx ON orders (user_id);

CREATE TABLE order_items (
    id               BIGSERIAL PRIMARY KEY,
    order_id         BIGINT    NOT NULL REFERENCES orders (id) ON DELETE CASCADE,
    product_id       BIGINT    NOT NULL REFERENCES products (id),
    quantity         INTEGER   NOT NULL CHECK (quantity > 0),
    unit_price_cents BIGINT    NOT NULL CHECK (unit_price_cents >= 0)
);

CREATE INDEX order_items_order_id_idx   ON order_items (order_id);
CREATE INDEX order_items_product_id_idx ON order_items (product_id);

COMMIT;
```

`000003_create_orders.down.sql`:

```sql
BEGIN;

DROP TABLE IF EXISTS order_items;
DROP TABLE IF EXISTS orders;

COMMIT;
```

### 000004: wallets and transfers

`000004_create_wallets.up.sql`:

```sql
BEGIN;

CREATE TABLE wallets (
    id            BIGSERIAL   PRIMARY KEY,
    user_id       BIGINT      NOT NULL UNIQUE REFERENCES users (id),
    balance_cents BIGINT      NOT NULL DEFAULT 0 CHECK (balance_cents >= 0),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE transfers (
    id             BIGSERIAL   PRIMARY KEY,
    from_wallet_id BIGINT      NOT NULL REFERENCES wallets (id),
    to_wallet_id   BIGINT      NOT NULL REFERENCES wallets (id),
    amount_cents   BIGINT      NOT NULL CHECK (amount_cents > 0),
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (from_wallet_id <> to_wallet_id)
);

CREATE INDEX transfers_from_wallet_id_idx ON transfers (from_wallet_id);
CREATE INDEX transfers_to_wallet_id_idx   ON transfers (to_wallet_id);

COMMIT;
```

`000004_create_wallets.down.sql`:

```sql
BEGIN;

DROP TABLE IF EXISTS transfers;
DROP TABLE IF EXISTS wallets;

COMMIT;
```

### Design choices worth noticing

- **Money is stored as whole cents in a BIGINT.** Floating point numbers cannot represent most decimal amounts exactly, so `0.1 + 0.2` is not `0.3`. Whole numbers of the smallest currency unit never have that problem.
- **CHECK constraints protect the data even if your Go code has a bug.** A negative stock or a negative balance is rejected by the database itself.
- **Foreign keys are indexed by hand.** PostgreSQL indexes primary keys and unique constraints automatically, but not the referencing side of a foreign key. Without an index, deleting a parent row or loading children by parent id scans the whole child table.
- **Soft delete uses a partial unique index.** A plain `UNIQUE` on `email` would still count deleted rows. Restricting the index to `deleted_at IS NULL` fixes that.
- **Orders have no `deleted_at`.** Financial records should not disappear. Use a status instead, such as `cancelled`.
- **`order_items` copies the unit price.** If a product's price changes next month, old orders must still show what the customer actually paid.

### Apply and inspect

```bash
migrate -path db/migrations -database "$DATABASE_URL" up
migrate -path db/migrations -database "$DATABASE_URL" version
docker compose exec db psql -U app -d shop -c '\dt'
docker compose exec db psql -U app -d shop -c 'select * from schema_migrations;'
```

Now practice the reverse direction. Undo the last migration, check that `wallets` is gone, and apply it again:

```bash
migrate -path db/migrations -database "$DATABASE_URL" down 1
migrate -path db/migrations -database "$DATABASE_URL" up
```

Running up, down, up on every new migration proves that both files work.

[Back to contents](#contents)

## Lesson 6: Running migrations from Go code

The CLI is perfect on your laptop. In a deployed service you often want the migration files baked into your program, so that the deployed binary carries its own migrations and no one needs to install a separate tool.

### Embed the files

Go can embed files into the compiled binary. Create `db/migrations/embed.go`:

```go
package migrations

import "embed"

//go:embed *.sql
var FS embed.FS
```

### A shared migrator

Create `internal/database/migrate.go`:

```go
package database

import (
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"

	"shop/db/migrations"
)

func NewMigrator(databaseURL string) (*migrate.Migrate, error) {
	src, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return nil, fmt.Errorf("load migrations: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", src, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("create migrator: %w", err)
	}
	return m, nil
}
```

The blank import of the `postgres` package registers the driver for `postgres://` URLs. Without it you get "unknown driver". The migrator opens its own connection from the URL. That is deliberate: closing the migrator then never affects your application's pool.

### A small migration command

Create `cmd/migrate/main.go`:

```go
package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/golang-migrate/migrate/v4"

	"shop/internal/database"
)

func main() {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		log.Fatal("DATABASE_URL is not set")
	}
	if len(os.Args) < 2 {
		log.Fatal("usage: migrate <up|down [n]|version|force <v>>")
	}

	m, err := database.NewMigrator(url)
	if err != nil {
		log.Fatal(err)
	}
	defer m.Close()

	err = run(m, os.Args[1], os.Args[2:])
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatal(err)
	}
	if errors.Is(err, migrate.ErrNoChange) {
		fmt.Println("nothing to do")
	}
}

func run(m *migrate.Migrate, cmd string, args []string) error {
	switch cmd {
	case "up":
		return m.Up()
	case "down":
		n := 1
		if len(args) > 0 {
			v, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid step count: %w", err)
			}
			n = v
		}
		return m.Steps(-n)
	case "version":
		v, dirty, err := m.Version()
		if errors.Is(err, migrate.ErrNilVersion) {
			fmt.Println("version: none")
			return nil
		}
		if err != nil {
			return err
		}
		fmt.Printf("version: %d dirty: %v\n", v, dirty)
		return nil
	case "force":
		if len(args) < 1 {
			return errors.New("force needs a version number")
		}
		v, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid version: %w", err)
		}
		return m.Force(v)
	default:
		return fmt.Errorf("unknown command %q", cmd)
	}
}
```

Notes:

- `migrate.ErrNoChange` is returned when there is nothing to apply. It is not a failure, so treat it as success.
- `migrate.ErrNilVersion` means no migration has ever been applied.
- `m.Steps(-1)` rolls back one migration. `m.Steps(2)` applies two.

Run it:

```bash
go run ./cmd/migrate up
go run ./cmd/migrate version
go run ./cmd/migrate down 1
```

### When should migrations run?

There are two common patterns.

**A separate step before the app starts.** Your deploy pipeline runs the migration command once (as a pre-deploy job, or an init container in Kubernetes), and only then starts the new app version. This is the safer choice for production. A failed migration stops the deploy, and ten app instances do not race to migrate.

**Automatically at app startup.** The app calls `m.Up()` before it begins serving. This is simple and fine for small services, side projects, and tests. The advisory lock keeps concurrent starts safe, but a slow migration delays startup, and a failed one crash loops every instance.

Whichever you choose, apply migrations to a copy of production data before you trust them in production.

[Back to contents](#contents)

## Lesson 7: Models and conventions

A GORM model is a plain Go struct that describes one table's rows. Because golang-migrate owns the schema, models stay small: no size limits, no index hints, no `not null`. Those belong in SQL.

### The shop models

Create `internal/models/models.go`:

```go
package models

import (
	"time"

	"gorm.io/gorm"
)

const (
	OrderPending   = "pending"
	OrderPaid      = "paid"
	OrderShipped   = "shipped"
	OrderCancelled = "cancelled"
)

type User struct {
	gorm.Model
	Email        string
	Name         string
	PasswordHash string
	Profile      *Profile
	Orders       []Order
}

type Profile struct {
	ID        uint
	UserID    uint
	Bio       string
	AvatarURL string
}

type Category struct {
	ID   uint
	Name string
}

type Product struct {
	gorm.Model
	SKU        string
	Name       string
	PriceCents int64
	Stock      int
	Categories []Category `gorm:"many2many:product_categories;"`
}

type Order struct {
	ID         uint
	UserID     uint
	Status     string
	TotalCents int64
	CreatedAt  time.Time
	UpdatedAt  time.Time
	Items      []OrderItem
}

type OrderItem struct {
	ID             uint
	OrderID        uint
	ProductID      uint
	Product        Product
	Quantity       int
	UnitPriceCents int64
}

type Wallet struct {
	ID           uint
	UserID       uint
	BalanceCents int64
	UpdatedAt    time.Time
}

type Transfer struct {
	ID           uint
	FromWalletID uint
	ToWalletID   uint
	AmountCents  int64
	CreatedAt    time.Time
}
```

### The conventions GORM follows

GORM guesses names so you do not have to repeat them:

- A struct named `User` maps to a table named `users`. The name is snake_case and plural. `OrderItem` becomes `order_items`.
- A field `PasswordHash` maps to a column `password_hash`. `AvatarURL` becomes `avatar_url`.
- A field named `ID` is the primary key.
- `CreatedAt` and `UpdatedAt` are filled in automatically on insert and update.
- A field of type `gorm.DeletedAt` turns on soft delete, described below.
- `gorm.Model` is just a shortcut for `ID`, `CreatedAt`, `UpdatedAt`, and `DeletedAt`.

If a table name does not follow the convention, add a method:

```go
func (Category) TableName() string { return "product_categories_lookup" }
```

If your whole database uses singular table names, set `NamingStrategy: schema.NamingStrategy{SingularTable: true}` in `gorm.Config` instead of writing a method per model.

### Soft delete

When a struct has a `gorm.DeletedAt` field, `Delete` does not remove the row. It sets `deleted_at` to the current time, and every normal query adds `WHERE deleted_at IS NULL`. It works like the recycle bin on a computer: the file looks gone, but you can bring it back.

Use it for things people expect to restore or audit, like accounts and products. Do not use it everywhere. Orders and transfers should not be deletable at all, which is why those structs do not use `gorm.Model`.

Soft delete has two traps. First, unique indexes must ignore deleted rows (we handled that with partial indexes). Second, a join or preload onto a soft deleted row silently returns nothing. Lesson 9 shows how to handle it for order history.

### Nullable columns

Go's zero values cannot tell "empty" from "missing". If a column allows NULL, use a pointer:

```go
Phone *string      // NULL becomes nil
ShippedAt *time.Time
```

Alternatively use `sql.NullString` and friends. Pointers are simpler to work with and serialize to JSON as `null`.

### Money and time

- Store money as `int64` cents. If you need real decimals, use a decimal library such as `shopspring/decimal`, which works with GORM and PostgreSQL's `NUMERIC` type.
- Store timestamps as `TIMESTAMPTZ` and use `time.Time`. Convert to a local zone only when displaying.

### Do not expose models directly

`User` contains `PasswordHash`. If you serialize it straight to JSON, you will eventually leak it. Define separate response structs for your API and copy only the fields you intend to show. Models describe the database. Response structs describe the API. They change for different reasons.

[Back to contents](#contents)

## Lesson 8: CRUD with GORM

CRUD means create, read, update, delete. This lesson builds the `store` package step by step.

### Domain errors

Callers of the store should not need to import GORM to understand a failure. The store translates database errors into its own. Create `internal/store/errors.go`:

```go
package store

import "errors"

var (
	ErrNotFound          = errors.New("not found")
	ErrEmailTaken        = errors.New("email already taken")
	ErrInvalidInput      = errors.New("invalid input")
	ErrInsufficientStock = errors.New("insufficient stock")
	ErrInsufficientFunds = errors.New("insufficient funds")
)
```

### Pagination as a scope

A scope is a function that adjusts a query. It lets you reuse the same logic across queries. Create `internal/store/scopes.go`:

```go
package store

import "gorm.io/gorm"

func Paginate(page, size int) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if page < 1 {
			page = 1
		}
		if size < 1 {
			size = 20
		}
		if size > 100 {
			size = 100
		}
		return db.Offset((page - 1) * size).Limit(size)
	}
}
```

The cap on `size` matters. Without it, a client can ask for a million rows.

### Create

Create `internal/store/users.go`:

```go
package store

import (
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"

	"shop/internal/models"
)

type UserStore struct {
	db *gorm.DB
}

func NewUserStore(db *gorm.DB) *UserStore {
	return &UserStore{db: db}
}

func (s *UserStore) Create(ctx context.Context, u *models.User) error {
	u.Email = strings.ToLower(strings.TrimSpace(u.Email))

	err := s.db.WithContext(ctx).Create(u).Error
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return ErrEmailTaken
	}
	return err
}
```

`Create` runs `INSERT ... RETURNING id`, so `u.ID`, `u.CreatedAt`, and `u.UpdatedAt` are filled in when it returns. If `u.Profile` is set, GORM inserts the profile too, inside the same transaction.

To insert many rows at once:

```go
err := db.WithContext(ctx).CreateInBatches(products, 100).Error
```

`CreateInBatches` sends 100 rows per INSERT. One enormous INSERT can exceed PostgreSQL's limit of 65,535 parameters per statement.

### Read

Continue `users.go`:

```go
func (s *UserStore) GetByID(ctx context.Context, id uint) (*models.User, error) {
	var u models.User
	err := s.db.WithContext(ctx).Preload("Profile").First(&u, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (s *UserStore) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	var u models.User
	err := s.db.WithContext(ctx).
		Where("email = ?", strings.ToLower(strings.TrimSpace(email))).
		First(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}
```

Differences between the single-row methods:

- `First` orders by primary key and takes one row. It returns `gorm.ErrRecordNotFound` if nothing matches.
- `Take` takes one row with no ordering and also returns `ErrRecordNotFound`.
- `Last` orders by primary key descending.
- `Find` loads zero or more rows into a slice and **does not** return an error for an empty result.

A few more patterns you will use all the time:

```go
// Conditions: always use ? placeholders so values are sent as parameters.
db.Where("price_cents BETWEEN ? AND ?", 500, 2000).Find(&products)
db.Where("sku IN ?", []string{"A1", "B2"}).Find(&products)
db.Where("name ILIKE ?", "%"+q+"%").Find(&products)

// Choose columns, order, limit.
db.Select("id", "name", "price_cents").Order("price_cents DESC").Limit(10).Find(&products)

// One column into a slice.
var emails []string
db.Model(&models.User{}).Pluck("email", &emails)

// Counting.
var total int64
db.Model(&models.Product{}).Where("stock = 0").Count(&total)

// Grouping into a custom struct.
type StatusCount struct {
	Status string
	Count  int64
}
var counts []StatusCount
db.Model(&models.Order{}).
	Select("status, COUNT(*) AS count").
	Group("status").
	Scan(&counts)
```

Two warnings about conditions:

**Struct conditions skip zero values.** `db.Where(&models.Product{Stock: 0})` adds no condition at all, because `0` is Go's zero value. Use string conditions or a map when zero is a meaningful value.

**Never build SQL with string concatenation.** `Where("name = '" + name + "'")` is an SQL injection hole. The `?` form is safe. Be careful with `Order(userInput)` too, because column names cannot be parameterized. Check the input against a list of allowed sort fields first.

### Search with total count

Create `internal/store/products.go`:

```go
package store

import (
	"context"

	"gorm.io/gorm"

	"shop/internal/models"
)

type ProductStore struct {
	db *gorm.DB
}

func NewProductStore(db *gorm.DB) *ProductStore {
	return &ProductStore{db: db}
}

func (s *ProductStore) Search(ctx context.Context, q string, page, size int) ([]models.Product, int64, error) {
	tx := s.db.WithContext(ctx).Model(&models.Product{})
	if q != "" {
		tx = tx.Where("name ILIKE ?", "%"+q+"%")
	}
	/* Session makes tx safe to reuse for both Count and Find. */
	tx = tx.Session(&gorm.Session{})

	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var products []models.Product
	err := tx.Scopes(Paginate(page, size)).Order("id DESC").Find(&products).Error
	return products, total, err
}
```

This is the legitimate case for reusing a partly built query, and `Session(&gorm.Session{})` is what makes it safe (see Lesson 3).

Offset pagination gets slower as the page number grows, because the database still walks past all the skipped rows. For feeds and very large tables, use keyset pagination: `WHERE id < ? ORDER BY id DESC LIMIT ?`, passing the last id from the previous page.

### Update

Continue `users.go`:

```go
func (s *UserStore) Rename(ctx context.Context, id uint, name string) error {
	res := s.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", id).
		Update("name", name)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
```

Always check `RowsAffected`. An UPDATE that matches no rows is not an error to the database, so without the check you would report success for a user that does not exist.

GORM has several ways to update, and they behave differently:

```go
// Update one column. Also sets updated_at.
db.Model(&product).Update("stock", 20)

// Updates with a struct: zero values are IGNORED.
db.Model(&product).Updates(models.Product{Name: "Mug", Stock: 0})
// Only name changes. Stock stays as it was.

// Updates with a map: zero values ARE written.
db.Model(&product).Updates(map[string]any{"name": "Mug", "stock": 0})

// Select picks exactly which struct fields to write, zero or not.
db.Model(&product).Select("Name", "Stock").Updates(models.Product{Name: "Mug", Stock: 0})

// Save writes every field of the struct, and inserts if the primary key is zero.
db.Save(&product)

// Update a column using its current value, in one atomic statement.
db.Model(&product).Update("stock", gorm.Expr("stock - ?", 3))

// Skip hooks and the updated_at change.
db.Model(&product).UpdateColumn("stock", 20)
```

Guidance:

- Use a map or `Select` when zero values matter. The struct-skips-zero rule is the most common GORM surprise.
- `Save` overwrites every column with whatever is in your struct. If another request changed the row after you loaded it, `Save` quietly erases that change. Prefer targeted `Update` or `Updates` calls.
- For counters and balances, use `gorm.Expr`. Reading a value into Go, changing it, and writing it back loses updates when two requests overlap.
- GORM refuses to run an update or delete with no conditions at all and returns `gorm.ErrMissingWhereClause`. That is a safety net, not a bug.

### Delete

```go
func (s *UserStore) Delete(ctx context.Context, id uint) error {
	res := s.db.WithContext(ctx).Delete(&models.User{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
```

Because `User` has `DeletedAt`, this runs `UPDATE users SET deleted_at = now() WHERE id = ? AND deleted_at IS NULL`. Variations:

```go
// Really delete the row.
db.Unscoped().Delete(&models.User{}, id)

// Include soft deleted rows in a query.
db.Unscoped().Where("email = ?", email).Find(&users)

// Restore a soft deleted row.
db.Unscoped().Model(&models.User{}).Where("id = ?", id).Update("deleted_at", nil)
```

### Creating a product with categories

Back in `products.go`:

```go
func (s *ProductStore) Create(ctx context.Context, p *models.Product, categoryIDs []uint) error {
	var cats []models.Category
	if len(categoryIDs) > 0 {
		err := s.db.WithContext(ctx).Where("id IN ?", categoryIDs).Find(&cats).Error
		if err != nil {
			return err
		}
		if len(cats) != len(categoryIDs) {
			return ErrNotFound
		}
	}
	p.Categories = cats

	err := s.db.WithContext(ctx).Create(p).Error
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return ErrInvalidInput
	}
	return err
}
```

Add `"errors"` to the imports. Loading the categories first confirms they exist before anything is written. Lesson 9 explains what GORM does with `p.Categories` during the insert.

[Back to contents](#contents)

## Lesson 9: Associations

Associations describe how tables relate. GORM supports four kinds, and all of them appear in the shop.

### The four relationships

**Has one.** A user has one profile. The foreign key (`user_id`) lives in the other table.

```go
type User struct {
	gorm.Model
	Profile *Profile
}
type Profile struct {
	ID     uint
	UserID uint // foreign key
}
```

**Has many.** A user has many orders. An order has many items.

```go
type User struct {
	gorm.Model
	Orders []Order
}
type Order struct {
	ID     uint
	UserID uint // foreign key
}
```

**Belongs to.** An order item belongs to a product. The struct holds the foreign key and the related struct together.

```go
type OrderItem struct {
	ID        uint
	ProductID uint    // foreign key
	Product   Product // the related row
}
```

**Many to many.** A product has many categories, and a category has many products. A third table links them.

```go
type Product struct {
	gorm.Model
	Categories []Category `gorm:"many2many:product_categories;"`
}
```

GORM finds foreign keys by name: the field `UserID` in `Order` points at `User`. If you name things differently, add `gorm:"foreignKey:OwnerID"` to say which field to use.

### Loading related rows: Preload

By default GORM loads only the row you asked for. Related rows stay empty until you ask:

```go
var u models.User
db.Preload("Profile").Preload("Orders").First(&u, 1)
```

`Preload` runs one extra query per association:

```
SELECT * FROM users WHERE id = 1 ORDER BY id LIMIT 1;
SELECT * FROM profiles WHERE user_id = 1;
SELECT * FROM orders WHERE user_id = 1;
```

Nested and conditional preloads:

```go
db.Preload("Items.Product").Find(&orders)

db.Preload("Orders", "status = ?", models.OrderPaid).First(&u, 1)

db.Preload("Orders", func(db *gorm.DB) *gorm.DB {
	return db.Order("created_at DESC").Limit(5)
}).First(&u, 1)
```

### Why preload matters: the N+1 problem

Imagine you need ten grocery items. One option is ten separate trips to the store, one item each. The other is one trip with a list.

This code makes the ten trips:

```go
db.Find(&orders) // 1 query
for i := range orders {
	db.Where("order_id = ?", orders[i].ID).Find(&orders[i].Items) // N more queries
}
```

For 100 orders that is 101 queries. `Preload("Items")` does it in two, using `WHERE order_id IN (...)`. When a page feels slow, turn the logger to Info and count the statements.

### Loading with a join: Joins

For has-one and belongs-to relations, `Joins` fetches everything in a single query:

```go
db.Joins("Profile").Find(&users)
```

This runs a `LEFT JOIN profiles` and fills in each user's profile. Use `Preload` for has-many and many-to-many, because a join would repeat the parent row once per child.

To filter by a related table, write the join yourself:

```go
// All products in one category.
db.Joins("JOIN product_categories pc ON pc.product_id = products.id").
	Where("pc.category_id = ?", categoryID).
	Find(&products)
```

### Order history needs Unscoped

Suppose a product is soft deleted after someone bought it. Their old order should still show the product name. But `Preload("Items.Product")` adds `deleted_at IS NULL`, so the product silently comes back empty.

Tell the preload to include deleted rows:

```go
func (s *OrderStore) ListForUser(ctx context.Context, userID uint, page, size int) ([]models.Order, error) {
	var orders []models.Order
	err := s.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Preload("Items.Product", func(db *gorm.DB) *gorm.DB {
			return db.Unscoped()
		}).
		Scopes(Paginate(page, size)).
		Order("created_at DESC, id DESC").
		Find(&orders).Error
	return orders, err
}
```

### Writing associations

When you create a parent that has related structs filled in, GORM writes them too. Creating a product with `Categories` set runs, in one transaction:

1. `INSERT INTO products ...`
2. `INSERT INTO categories ... ON CONFLICT DO NOTHING` for the categories in the slice
3. `INSERT INTO product_categories ... ON CONFLICT DO NOTHING` for the links

Step 2 is easy to miss. GORM tries to upsert related rows. If you pass a `Category{ID: 99}` that does not exist, GORM may insert a nameless category. This is why `ProductStore.Create` loaded the real categories first.

To save only the parent and skip all related structs:

```go
db.Omit(clause.Associations).Create(&order)
```

For editing links after the fact, use association mode:

```go
var p models.Product
db.First(&p, productID)

assoc := db.Model(&p).Association("Categories")

assoc.Append(&cats)     // add links
assoc.Replace(&cats)    // make the links exactly this set
assoc.Delete(&oneCat)   // remove a link (the category row stays)
assoc.Clear()           // remove all links
n := assoc.Count()      // how many are linked
```

Each of these returns an error in real code. Check it.

### Foreign key behavior

Deleting a parent row interacts with your foreign keys. In our migrations, `profiles.user_id` uses `ON DELETE CASCADE`, so removing a user row (a hard delete) removes the profile too. `orders.user_id` has no cascade, so the database refuses to hard delete a user who has orders, and you get `gorm.ErrForeignKeyViolated`. That protects your financial records. Soft delete avoids the whole problem, since the row stays.

[Back to contents](#contents)

## Lesson 10: Transactions, locking, and money

### What a transaction is

When you pay at a shop, money leaves your account and arrives in the shop's account. If the system crashed between those two steps, you would be charged and the shop would never be paid, or the reverse. Nobody accepts that. A transaction groups several steps so that either all of them happen or none do.

Databases describe this with four words:

- **Atomic.** All steps happen, or none.
- **Consistent.** Rules like foreign keys and CHECK constraints hold before and after.
- **Isolated.** Other transactions running at the same time do not see your half finished work.
- **Durable.** Once committed, the data survives a crash.

Use a transaction whenever one logical action changes more than one row, or when a change depends on a value you just read. Checkout and money transfers are textbook cases.

### The three ways to use transactions in GORM

**Transaction helper (use this by default).** Return an error and everything rolls back. Return nil and it commits. A panic also rolls back.

```go
err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
	if err := tx.Create(&order).Error; err != nil {
		return err
	}
	return tx.Model(&product).Update("stock", gorm.Expr("stock - ?", 1)).Error
})
```

**Manual control.** Needed only when begin and commit must happen in different places:

```go
tx := db.WithContext(ctx).Begin()
if tx.Error != nil {
	return tx.Error
}
defer func() {
	if r := recover(); r != nil {
		tx.Rollback()
		panic(r)
	}
}()

if err := tx.Create(&order).Error; err != nil {
	tx.Rollback()
	return err
}
return tx.Commit().Error
```

**Nested transactions.** Calling `Transaction` on a `tx` creates a savepoint. If the inner function fails, only the inner work is undone:

```go
err := db.Transaction(func(tx *gorm.DB) error {
	if err := tx.Create(&order).Error; err != nil {
		return err
	}

	err := tx.Transaction(func(inner *gorm.DB) error {
		return inner.Create(&auditEntry).Error
	})
	if err != nil {
		log.Println("audit entry failed, order kept:", err)
	}
	return nil
})
```

### Rules for transaction code

1. **Use `tx` inside the function, never the outer `db`.** Using `db` runs that query on a different connection, outside the transaction. It will not roll back, and it cannot see uncommitted rows. This is a very common bug.
2. **Keep transactions short.** An open transaction holds locks and a connection from the pool.
3. **Do not call other services inside one.** No HTTP requests, no sending email. They are slow, and they cannot be rolled back. Commit first, then send.
4. **After a database error, the transaction is dead.** PostgreSQL rejects every further statement with "current transaction is aborted" until you roll back. Return the error, or use a nested `Transaction` (a savepoint) around the risky step.

GORM also wraps each single create, update, or delete in its own small transaction by default, so that hooks and association writes are atomic. Setting `SkipDefaultTransaction: true` in `gorm.Config` removes that overhead and can speed up write heavy code. Only do it once you understand the trade.

### The overselling problem

Picture the last seat on a flight. Two people try to book it at the same moment. Each one's booking app checks, sees one seat free, and confirms. Now two people hold the same seat.

In code, this is the same mistake:

```
Request A reads stock = 1
Request B reads stock = 1
Request A writes stock = 0
Request B writes stock = 0   (two sales, one item)
```

PostgreSQL's default isolation level (Read Committed) does not prevent this. The fix is to make the second request wait until the first finishes, by locking the row.

### Checkout with row locking

`SELECT ... FOR UPDATE` locks the rows it reads. Any other transaction that wants to lock or change them waits until this one commits or rolls back. GORM expresses it with `clause.Locking`.

Create `internal/store/orders.go`:

```go
package store

import (
	"cmp"
	"context"
	"errors"
	"slices"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"shop/internal/models"
)

type CheckoutItem struct {
	ProductID uint
	Quantity  int
}

type OrderStore struct {
	db *gorm.DB
}

func NewOrderStore(db *gorm.DB) *OrderStore {
	return &OrderStore{db: db}
}

func (s *OrderStore) Checkout(ctx context.Context, userID uint, items []CheckoutItem) (*models.Order, error) {
	if len(items) == 0 {
		return nil, ErrInvalidInput
	}
	for _, it := range items {
		if it.Quantity <= 0 {
			return nil, ErrInvalidInput
		}
	}

	/* Lock products in a fixed order so two checkouts cannot deadlock. */
	sorted := slices.Clone(items)
	slices.SortFunc(sorted, func(a, b CheckoutItem) int {
		return cmp.Compare(a.ProductID, b.ProductID)
	})

	order := &models.Order{UserID: userID, Status: models.OrderPending}

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, it := range sorted {
			var p models.Product
			err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&p, it.ProductID).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrNotFound
			}
			if err != nil {
				return err
			}
			if p.Stock < it.Quantity {
				return ErrInsufficientStock
			}

			err = tx.Model(&p).Update("stock", gorm.Expr("stock - ?", it.Quantity)).Error
			if err != nil {
				return err
			}

			order.Items = append(order.Items, models.OrderItem{
				ProductID:      p.ID,
				Quantity:       it.Quantity,
				UnitPriceCents: p.PriceCents,
			})
			order.TotalCents += p.PriceCents * int64(it.Quantity)
		}

		/* Inserts the order and all of its items together. */
		return tx.Create(order).Error
	})
	if err != nil {
		return nil, err
	}
	return order, nil
}
```

What to notice:

- The price comes from the locked product row, never from the client. Never trust a price sent by the browser.
- `UnitPriceCents` is copied into the order item, so later price changes do not rewrite history.
- If the same product appears on two lines, each line is processed in turn and sees the earlier decrement, so the stock check stays correct.
- Every sentinel error (`ErrInsufficientStock`) returned from the function rolls back the whole transaction, including stock already reduced for earlier items.
- If a client retries after a network failure, this code creates a second order. Real checkout endpoints accept an idempotency key (a unique value per attempt, stored with a unique index) so a repeat request returns the original order instead of making a new one.

### A lighter alternative: a conditional update

You can avoid explicit locking by letting the UPDATE check the condition itself. The row is locked for the duration of the statement, and the `WHERE` clause is checked against the latest committed version:

```go
res := tx.Model(&models.Product{}).
	Where("id = ? AND stock >= ?", it.ProductID, it.Quantity).
	Update("stock", gorm.Expr("stock - ?", it.Quantity))
if res.Error != nil {
	return res.Error
}
if res.RowsAffected == 0 {
	return ErrInsufficientStock
}
```

This is shorter and fast. The downside is that zero rows could mean "not enough stock" or "product does not exist", and you still need another query to read the price. The `CHECK (stock >= 0)` constraint in the schema is the last line of defense if any code path forgets the check.

### Moving money between wallets

Create `internal/store/wallets.go`:

```go
package store

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"shop/internal/models"
)

type WalletStore struct {
	db *gorm.DB
}

func NewWalletStore(db *gorm.DB) *WalletStore {
	return &WalletStore{db: db}
}

func (s *WalletStore) Transfer(ctx context.Context, fromID, toID uint, amountCents int64) error {
	if amountCents <= 0 || fromID == toID {
		return ErrInvalidInput
	}

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		/* Lock both wallets, lowest id first. */
		var wallets []models.Wallet
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id IN ?", []uint{fromID, toID}).
			Order("id ASC").
			Find(&wallets).Error
		if err != nil {
			return err
		}
		if len(wallets) != 2 {
			return ErrNotFound
		}

		for _, w := range wallets {
			if w.ID == fromID && w.BalanceCents < amountCents {
				return ErrInsufficientFunds
			}
		}

		err = tx.Model(&models.Wallet{}).Where("id = ?", fromID).
			Update("balance_cents", gorm.Expr("balance_cents - ?", amountCents)).Error
		if err != nil {
			return err
		}
		err = tx.Model(&models.Wallet{}).Where("id = ?", toID).
			Update("balance_cents", gorm.Expr("balance_cents + ?", amountCents)).Error
		if err != nil {
			return err
		}

		/* An immutable record of the movement. */
		return tx.Create(&models.Transfer{
			FromWalletID: fromID,
			ToWalletID:   toID,
			AmountCents:  amountCents,
		}).Error
	})
}
```

Why the lock order matters: imagine two people at a table with two pens, and each needs both pens to sign a form. If Ann grabs pen 1 while Bob grabs pen 2, each waits forever for the other. That is a deadlock. If everyone must pick up pen 1 first, one person finishes before the other starts. When transfer A to B and transfer B to A run at the same time, locking wallets in ascending id order gives every transaction the same "pen 1 first" rule. PostgreSQL does detect deadlocks and kills one transaction, but avoiding them is better than recovering from them.

Also notice the `transfers` row. Every movement of money gets its own record that is never edited or deleted. The balance column is a convenience. If the two ever disagree, the transfer history is how you find out what happened.

### Isolation levels and retries

For the strongest guarantee, you can ask for a stricter isolation level:

```go
err := db.Transaction(fn, &sql.TxOptions{Isolation: sql.LevelSerializable})
```

This needs `database/sql` imported. Under Serializable, PostgreSQL may abort a transaction with a serialization failure (SQLSTATE 40001) even though nothing is wrong, and your code must retry it. Start with Read Committed plus explicit row locks, as above. Move to Serializable only when you have a reason, and add a retry loop when you do.

[Back to contents](#contents)

## Lesson 11: The query toolbox

### Raw SQL

GORM does not need to write every query. For reports and complex joins, write SQL and map the result into a struct:

```go
type ProductSales struct {
	ProductID    uint
	Name         string
	Units        int64
	RevenueCents int64
}

var top []ProductSales
err := db.WithContext(ctx).Raw(`
	SELECT p.id AS product_id,
	       p.name,
	       SUM(oi.quantity)::bigint AS units,
	       SUM(oi.quantity * oi.unit_price_cents)::bigint AS revenue_cents
	FROM order_items oi
	JOIN products p ON p.id = oi.product_id
	GROUP BY p.id, p.name
	ORDER BY revenue_cents DESC
	LIMIT ?`, 10).Scan(&top).Error
```

`Scan` maps column names to struct fields the same way GORM does everywhere else (`revenue_cents` fills `RevenueCents`). Use `db.Exec("...")` for statements that return no rows.

There is no shame in raw SQL. A good rule: use GORM's builder for ordinary CRUD, and raw SQL for anything where the SQL itself is the clearest description.

### Upsert

An upsert inserts a row, or updates it if it already exists. In GORM it is a clause:

```go
// Insert categories, ignore the ones that already exist.
db.Clauses(clause.OnConflict{
	Columns:   []clause.Column{{Name: "name"}},
	DoNothing: true,
}).Create(&categories)

// Insert, or update chosen columns on conflict.
db.Clauses(clause.OnConflict{
	Columns:   []clause.Column{{Name: "name"}},
	DoUpdates: clause.AssignmentColumns([]string{"name"}),
}).Create(&categories)
```

The conflict target needs a matching unique index or constraint, which `categories.name UNIQUE` provides.

GORM also has `FirstOrCreate`, which looks for a row and inserts one if it is missing. It does two statements, so two requests at the same time can both find nothing and both insert. A unique constraint plus an upsert is the reliable version.

### Locking options

`clause.Locking` takes extra options for queue like workloads:

```go
// Fail at once instead of waiting if the row is locked.
tx.Clauses(clause.Locking{Strength: "UPDATE", Options: "NOWAIT"}).First(&job)

// Skip rows other workers have locked.
tx.Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
	Where("status = ?", "queued").Order("id").First(&job)
```

`SKIP LOCKED` lets several workers pull jobs from one table without fighting over the same row.

### Hooks

Hooks are methods GORM calls around database operations: `BeforeCreate`, `AfterCreate`, `BeforeUpdate`, `AfterUpdate`, `BeforeDelete`, `AfterFind`, and a few more:

```go
func (u *User) BeforeCreate(tx *gorm.DB) error {
	u.Email = strings.ToLower(strings.TrimSpace(u.Email))
	return nil
}
```

A hook runs inside the same transaction as the operation. Returning an error cancels it. Hooks are good for small, universal normalization. They are a poor home for business rules, because they run invisibly and are skipped by `UpdateColumn`, `Exec`, and raw SQL. Put business logic in the store or service layer where a reader can see it.

### Processing large tables

Do not `Find` a million rows into one slice. Process them in batches:

```go
var batch []models.Order
err := db.Where("status = ?", models.OrderPaid).
	FindInBatches(&batch, 500, func(tx *gorm.DB, n int) error {
		for _, o := range batch {
			/* handle one order */
			_ = o
		}
		return nil
	}).Error
```

### Seeing the SQL

- `db.Debug().Find(&users)` logs that one statement at Info level.
- `db.ToSQL(func(tx *gorm.DB) *gorm.DB { return tx.Where("id = ?", 1).Find(&users) })` returns the SQL string without running it.
- Running with `LogLevel: logger.Info` in development shows everything, which is the fastest way to learn what GORM does.

### Errors worth handling

- `gorm.ErrRecordNotFound`: `First`, `Take`, `Last` found nothing.
- `gorm.ErrDuplicatedKey`: a unique constraint failed (needs `TranslateError`).
- `gorm.ErrForeignKeyViolated`: a foreign key failed (needs `TranslateError`).
- `gorm.ErrMissingWhereClause`: you tried an unconditional update or delete.
- `context.DeadlineExceeded` and `context.Canceled`: the context ended the query.

Always compare with `errors.Is`, never `==`, because errors may be wrapped.

### Timeouts

Give every request a deadline, so a slow query cannot hold a connection forever:

```go
ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
defer cancel()
err := db.WithContext(ctx).First(&u, id).Error
```

[Back to contents](#contents)

## Lesson 12: The generics API

Since GORM v1.30, there is a second way to write queries that uses Go generics. The older chain style you have been using is still fully supported, and it is what most existing code and tutorials use, which is why this tutorial teaches it first. You can mix the two in one project. The generics API is newer and still being extended, so check the official docs for the exact method signatures in your GORM version.

```go
ctx := context.Background()

// Create
err := gorm.G[models.User](db).Create(ctx, &user)

// Read: typed results, no pointer to fill in
u, err := gorm.G[models.User](db).Where("email = ?", email).First(ctx)
users, err := gorm.G[models.User](db).Where("name ILIKE ?", "a%").Find(ctx)

// Update and delete report how many rows changed
rows, err := gorm.G[models.Product](db).Where("id = ?", id).Update(ctx, "stock", 10)
rows, err = gorm.G[models.Product](db).Where("id = ?", id).Updates(ctx, models.Product{Name: "Mug"})
rows, err = gorm.G[models.Product](db).Where("id = ?", id).Delete(ctx)

// Preload with options
orders, err := gorm.G[models.Order](db).
	Where("user_id = ?", userID).
	Preload("Items", func(db gorm.PreloadBuilder) error {
		db.Order("id ASC")
		return nil
	}).
	Find(ctx)
```

What changes:

- **A context is required** in the final call, so you cannot forget it.
- **Results are typed.** You get back a `models.User`, not a filled pointer.
- **No leaking state.** Each `gorm.G[T](db)` starts fresh, which removes the shared-query bug from Lesson 3.
- **A few methods were left out on purpose**, including `Save` and `FirstOrCreate`, which are easy to misuse.

A reasonable approach: learn the classic API well, since you will read it constantly. Use the generics API in new code if your team likes it. Transactions work the same way in both, through `db.Transaction`.

[Back to contents](#contents)

## Lesson 13: Testing the database layer

Mocking a database is tempting and usually unhelpful. A mock tells you that your code calls the method you expected, not that the SQL is correct, the constraints hold, or the migration matches the model. For store code, test against a real PostgreSQL. It is cheap, because Docker is already running.

### A test database helper

Create `internal/testdb/testdb.go`:

```go
package testdb

import (
	"errors"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"shop/internal/database"
)

var (
	once     sync.Once
	shared   *gorm.DB
	setupErr error
)

func setup(url string) (*gorm.DB, error) {
	m, err := database.NewMigrator(url)
	if err != nil {
		return nil, err
	}
	defer m.Close()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return nil, err
	}

	return database.Connect(database.Config{
		URL:          url,
		MaxOpenConns: 10,
		MaxIdleConns: 5,
		ConnMaxLife:  time.Minute,
		LogLevel:     logger.Silent,
	})
}

/* New returns a transaction that is rolled back when the test ends,
   so every test starts with an empty database. */
func New(t *testing.T) *gorm.DB {
	t.Helper()

	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}

	once.Do(func() { shared, setupErr = setup(url) })
	if setupErr != nil {
		t.Fatalf("set up test database: %v", setupErr)
	}

	tx := shared.Begin()
	if tx.Error != nil {
		t.Fatalf("begin transaction: %v", tx.Error)
	}
	t.Cleanup(func() { tx.Rollback() })
	return tx
}
```

Two good things happen here. The test database runs your real migrations, so tests exercise the same schema as production. And each test works inside a transaction that is thrown away at the end, so tests never see each other's data and need no cleanup code.

Run tests with:

```bash
TEST_DATABASE_URL='postgres://app:secret@localhost:5432/shop_test?sslmode=disable' go test ./...
```

Go runs different packages in parallel, and several may try to migrate at once. The advisory lock makes that safe.

### A test that catches schema drift

The biggest weakness of keeping structs and SQL separate is that they can disagree: someone adds a field to a struct and forgets the migration. This test fails when a struct field has no matching column. Create `internal/models/models_test.go`:

```go
package models_test

import (
	"testing"

	"gorm.io/gorm"

	"shop/internal/models"
	"shop/internal/testdb"
)

func TestModelsMatchSchema(t *testing.T) {
	db := testdb.New(t)

	tests := []struct {
		name  string
		model any
	}{
		{"User", &models.User{}},
		{"Profile", &models.Profile{}},
		{"Category", &models.Category{}},
		{"Product", &models.Product{}},
		{"Order", &models.Order{}},
		{"OrderItem", &models.OrderItem{}},
		{"Wallet", &models.Wallet{}},
		{"Transfer", &models.Transfer{}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			stmt := &gorm.Statement{DB: db}
			if err := stmt.Parse(tc.model); err != nil {
				t.Fatalf("parse model: %v", err)
			}
			for _, col := range stmt.Schema.DBNames {
				if !db.Migrator().HasColumn(tc.model, col) {
					t.Errorf("table %s has no column %s", stmt.Schema.Table, col)
				}
			}
		})
	}
}
```

Add a row to the table whenever you add a model.

### A table driven store test

Create `internal/store/orders_test.go`:

```go
package store_test

import (
	"context"
	"errors"
	"testing"

	"shop/internal/models"
	"shop/internal/store"
	"shop/internal/testdb"
)

func TestOrderStore_Checkout(t *testing.T) {
	tests := []struct {
		name      string
		stock     int
		quantity  int
		wantErr   error
		wantStock int
	}{
		{name: "reduces stock", stock: 5, quantity: 2, wantStock: 3},
		{name: "sells the last unit", stock: 1, quantity: 1, wantStock: 0},
		{name: "rejects more than in stock", stock: 1, quantity: 2, wantErr: store.ErrInsufficientStock, wantStock: 1},
		{name: "rejects zero quantity", stock: 5, quantity: 0, wantErr: store.ErrInvalidInput, wantStock: 5},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db := testdb.New(t)

			user := models.User{Email: "buyer@example.com", Name: "Buyer", PasswordHash: "x"}
			product := models.Product{SKU: "MUG-1", Name: "Mug", PriceCents: 1200, Stock: tc.stock}
			if err := db.Create(&user).Error; err != nil {
				t.Fatal(err)
			}
			if err := db.Create(&product).Error; err != nil {
				t.Fatal(err)
			}

			orders := store.NewOrderStore(db)
			_, err := orders.Checkout(context.Background(), user.ID, []store.CheckoutItem{
				{ProductID: product.ID, Quantity: tc.quantity},
			})
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("got error %v, want %v", err, tc.wantErr)
			}

			var got models.Product
			if err := db.First(&got, product.ID).Error; err != nil {
				t.Fatal(err)
			}
			if got.Stock != tc.wantStock {
				t.Errorf("stock is %d, want %d", got.Stock, tc.wantStock)
			}
		})
	}
}
```

`errors.Is(nil, nil)` is true, so the success cases pass with `wantErr` left empty.

### Limits of the rollback approach

- **A database error kills the surrounding test transaction.** If your test deliberately triggers a unique violation through a plain `Create`, make that the last database call in the test. Code that wraps its work in `db.Transaction` is safe, because the nested call uses a savepoint.
- **Concurrency tests need real commits.** To prove that two simultaneous checkouts cannot oversell, you need two connections that can see each other. Use the real `shared` database for those tests, run the goroutines, and clean up afterwards with `TRUNCATE users, products, orders, order_items RESTART IDENTITY CASCADE`.
- **Avoid SQLite in tests** when production runs PostgreSQL. Partial indexes, row locks, and error behavior differ, and a test that passes on SQLite proves very little.

[Back to contents](#contents)

## Lesson 14: Changing a live schema safely

Changing a table on a live system is like renovating a restaurant while it is serving dinner. You can do it, but you have to avoid blocking the kitchen.

Two facts drive everything in this lesson:

1. **Some schema changes lock the table.** While the lock is held, queries on that table wait.
2. **During a deploy, old and new versions of your app run at the same time.** The database must work for both.

### Quick guide to risk

Usually safe and fast:

- Adding a new table
- Adding a nullable column
- Adding a column with a constant default (PostgreSQL 11 and later does this without rewriting the table)
- Creating an index with `CONCURRENTLY`

Risky on a busy or large table:

- Adding a `NOT NULL` column with no default to a table that has rows (it fails)
- Changing a column's type (can rewrite the whole table)
- `CREATE INDEX` without `CONCURRENTLY` (blocks writes while it builds)
- Adding a foreign key or CHECK constraint that must validate every existing row
- Renaming or dropping a column that running code still uses

### Set a lock timeout

Schema changes need strong locks. If a long running query holds the table, your `ALTER TABLE` waits, and then every new query queues up behind your waiting statement. A small migration can freeze the whole application.

A lock timeout makes the change give up quickly instead:

```sql
BEGIN;
SET LOCAL lock_timeout = '5s';
ALTER TABLE users ADD COLUMN phone TEXT;
COMMIT;
```

If it times out, the migration fails cleanly. Run it again when traffic is lower.

### Example: add a column

```bash
migrate create -ext sql -dir db/migrations -seq add_phone_to_users
```

`000005_add_phone_to_users.up.sql`:

```sql
BEGIN;
SET LOCAL lock_timeout = '5s';

ALTER TABLE users ADD COLUMN phone TEXT;

COMMIT;
```

`000005_add_phone_to_users.down.sql`:

```sql
BEGIN;
SET LOCAL lock_timeout = '5s';

ALTER TABLE users DROP COLUMN phone;

COMMIT;
```

Then add `Phone *string` to the `User` struct. Order matters: deploy the migration first, then the code that uses it. Old code ignores the new column, so both versions work.

### Example: an index that does not block writes

`CREATE INDEX CONCURRENTLY` builds the index without blocking inserts and updates. It has one strict rule: it cannot run inside a transaction. So this migration is the exception to the `BEGIN`/`COMMIT` rule, and it must hold exactly one statement in its own file:

`000006_index_orders_status.up.sql`:

```sql
CREATE INDEX CONCURRENTLY IF NOT EXISTS orders_status_created_at_idx
    ON orders (status, created_at);
```

`000006_index_orders_status.down.sql`:

```sql
DROP INDEX CONCURRENTLY IF EXISTS orders_status_created_at_idx;
```

If a concurrent build fails, PostgreSQL leaves an invalid index behind, and `IF NOT EXISTS` will see the name and skip it. Drop the broken index by hand, then run the migration again.

### Example: a foreign key without a long lock

Adding a foreign key to a large table normally scans every row while holding a lock. Split it into two steps. Add the constraint without checking old rows, then validate separately, which takes a lighter lock:

```sql
BEGIN;
SET LOCAL lock_timeout = '5s';
ALTER TABLE order_items
    ADD CONSTRAINT order_items_product_fk
    FOREIGN KEY (product_id) REFERENCES products (id) NOT VALID;
COMMIT;
```

```sql
ALTER TABLE order_items VALIDATE CONSTRAINT order_items_product_fk;
```

Put these in two separate migrations.

### Renaming a column: expand and contract

Say you want to rename `users.name` to `users.full_name`. A single `ALTER TABLE ... RENAME COLUMN` looks tempting, but the moment it runs, every old app instance that still queries `name` breaks.

Think of replacing a bridge. You build the new bridge beside the old one, move traffic across, and only then take the old one down. In database terms, the pattern is called expand and contract:

1. **Expand.** Add `full_name` as a nullable column in a migration. Deploy app code that writes to both columns and reads `full_name`, falling back to `name` when it is empty.
2. **Backfill.** Copy existing data across in small batches. Repeat until it updates zero rows:

   ```sql
   UPDATE users SET full_name = name
   WHERE id IN (
       SELECT id FROM users WHERE full_name IS NULL ORDER BY id LIMIT 1000
   );
   ```

3. **Enforce.** Once every row has a value, add the constraint in a migration: `ALTER TABLE users ALTER COLUMN full_name SET NOT NULL;`. On very large tables this scans the table under a lock, so look up the `CHECK ... NOT VALID` technique used to avoid that.
4. **Contract.** Deploy app code that uses only `full_name`. After that version is fully rolled out, drop `name` in a final migration.

Each step is safe for the version of the app that is running at that moment. It takes more steps than a rename, and that is the price of zero downtime. For a small app with a maintenance window, a plain rename is fine. Choose based on what an outage would cost you.

### Data migrations and seed data

- **Small, permanent reference data** (a fixed list of categories or roles) can live in a migration. Write it so it can run twice: `INSERT ... ON CONFLICT DO NOTHING`.
- **Large backfills** should run as a separate job in batches, outside any migration transaction, so they do not hold a lock for minutes.
- **Fake development data** belongs in a separate command such as `cmd/seed`, never in migrations. Migrations run in production, and you do not want fake users there.

### Rolling back in practice

Down migrations are valuable on a developer's machine and in CI. In production, rolling back by running a down file is often not what you want:

- A down that drops a column or table destroys the data in it, and applying the matching up again does not bring it back.
- Newer app versions may already depend on the new schema.

In production the usual answer is to **fix forward** with a new migration, and to keep real backups (including a practiced restore) for the cases that truly need an undo. Take a backup before any risky change.

### Team workflow

- **Review migrations like code.** Ask: does it lock? does it rewrite a big table? will the previous app version still work after it?
- **Test the round trip in CI.** Applying every migration, undoing them all, and applying them again proves that every down file works. Run it against a throwaway database, because it wipes the schema:

```go
func TestMigrationsRoundTrip(t *testing.T) {
	url := os.Getenv("ROUNDTRIP_DATABASE_URL")
	if url == "" {
		t.Skip("ROUNDTRIP_DATABASE_URL is not set")
	}

	m, err := database.NewMigrator(url)
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()

	for _, step := range []func() error{m.Up, m.Down, m.Up} {
		if err := step(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			t.Fatal(err)
		}
	}
}
```

- **Resolve number clashes at merge time.** If two branches both added `000007`, the second one to merge renames its files to `000008` before merging. This is a reason some teams prefer timestamps.

[Back to contents](#contents)

## Lesson 15: Production checklist

Use this list before shipping a service built with this stack.

### Migrations

- Migrations run as a controlled step in your deploy pipeline, not by hand.
- The migration step uses a database role that can change the schema. The application uses a different role that can only SELECT, INSERT, UPDATE, and DELETE. If the app is compromised, it cannot drop tables.
- A recent backup exists, and someone has practiced restoring it.
- New migrations were tried against a copy of production sized data.
- DDL statements set a `lock_timeout`. Large indexes use `CONCURRENTLY`.
- No applied migration has been edited.
- Every migration has a tested down file, and the round trip test passes in CI.
- `AutoMigrate` is not called anywhere in production code.

### Connections

- The URL uses `sslmode=require` or stricter in production, and the password comes from a secret store or environment variable, never from source control.
- Pool settings are chosen on purpose. Open connections times app instances stays below PostgreSQL's `max_connections`, with room for the migration tool and admin sessions.
- `ConnMaxLifetime` is shorter than any idle timeout in your network path.
- If you use a connection pooler such as PgBouncer in transaction mode, check your driver's prepared statement settings, because they can conflict.

### Queries

- Every query has a context, and request level contexts carry timeouts.
- No N+1 queries: related rows are loaded with `Preload` or `Joins`.
- Columns used in `WHERE`, `ORDER BY`, and joins are indexed, including foreign keys. Slow queries are checked with `EXPLAIN ANALYZE`.
- List endpoints cap their page size. Very large lists use keyset pagination.
- Values always travel through `?` placeholders. Sort fields that come from users are checked against a fixed list.
- Wide tables are queried with `Select` for only the columns needed.

### Data integrity

- Money is stored as integer minor units, or as an exact decimal type.
- Anything that changes several rows, or depends on a value just read, runs in a transaction. Counters and balances use locks or atomic `gorm.Expr` updates.
- Important rules also exist as constraints in the database: foreign keys, CHECKs, unique indexes.
- Operations a client might retry (checkout, payments) use idempotency keys.
- Financial records are never deleted. Soft delete is used only where restoring makes sense, with partial unique indexes.

### Operations

- The logger runs at `Warn` in production, with a slow query threshold and `ParameterizedQueries` on.
- A health endpoint checks the database with `sqlDB.PingContext`.
- Shutdown is graceful: stop accepting requests, wait for in flight work, then close the pool with `sqlDB.Close()`.
- Pool statistics from `sqlDB.Stats()` (in use, idle, wait count, wait duration) are exported to your monitoring. A rising wait count means requests are queuing for connections.

[Back to contents](#contents)

## Lesson 16: Common mistakes

1. **Reusing a half built query.** Conditions pile up silently. Start from `db` each time, or add `.Session(&gorm.Session{})` to the reusable part.
2. **Updating with a struct and expecting zero values to be written.** Use a map, or `Select` the fields.
3. **Skipping `RowsAffected`.** An update or delete that matches nothing is not an error. Check the count.
4. **Using `db` instead of `tx` inside a transaction.** That query runs outside the transaction.
5. **Expecting `Find` to return `ErrRecordNotFound`.** Only `First`, `Take`, and `Last` do. `Find` returns an empty slice.
6. **Using `Save` on a stale struct.** It overwrites every column, including ones that someone else just changed.
7. **Forgetting `Preload`.** The association stays empty with no error, or you trigger one query per row.
8. **Losing soft deleted rows in history views.** Preloads and joins skip them unless you use `Unscoped`.
9. **A plain unique index with soft delete.** Deleted rows keep blocking new ones. Use a partial index with `WHERE deleted_at IS NULL`.
10. **Editing a migration that already ran.** Other databases have the old version. Add a new migration.
11. **Putting `CREATE INDEX CONCURRENTLY` inside `BEGIN` and `COMMIT`.** It errors. Give it its own file with no transaction.
12. **Running AutoMigrate and golang-migrate on the same tables.** The two will fight. Pick one owner.
13. **Expecting struct tags like `not null` or `size` to do anything.** With SQL migrations, they have no effect on the schema. The SQL is the only source.
14. **Forgetting the blank driver import** (`_ ".../database/postgres"`). The migrator then fails with "unknown driver".
15. **Storing money in floats.** Use integer cents or an exact decimal type.
16. **Passing user input to `Order()` or building SQL with `+`.** Both are injection risks.
17. **Opening too many connections.** Each app instance multiplies your pool size. Count them.
18. **Doing slow work inside a transaction.** Keep network calls and email outside, after the commit.

[Back to contents](#contents)

## Appendix A: AutoMigrate

AutoMigrate is GORM's built in schema tool. You list your models, and GORM brings the database in line:

```go
err := db.AutoMigrate(&models.User{}, &models.Product{})
```

According to the GORM documentation, it creates missing tables, foreign keys, constraints, columns, and indexes, and it can change an existing column's type, size, precision, or nullability. It does not delete unused columns, which protects your data but also means it cannot clean up after a rename.

When using it, the schema is described through struct tags, so the models carry more detail:

```go
type User struct {
	gorm.Model
	Email string `gorm:"size:255;not null;uniqueIndex"`
	Name  string `gorm:"size:100;not null"`
}
```

Where it fits:

- Quick prototypes and demos
- Throwaway databases in some test setups
- Learning GORM before learning migrations

Where it falls short:

- There is no history, no review step, and no down path.
- Renames look like "add a new column, leave the old one behind".
- Advanced features such as partial indexes, triggers, and data backfills need raw SQL anyway.
- Every app instance that starts up may try to alter the schema, which is risky in production.

If you like writing structs but want reviewed SQL files, tools like Atlas can read GORM models and generate versioned migration files for you. That gives you structs as the input and plain SQL as the reviewed output. Check the Atlas documentation for its GORM integration.

[Back to contents](#contents)

## Appendix B: Practice tasks

Do these in order. Each one builds a skill from the lessons.

1. **Add a column.** Write migration 5 for `users.phone`, add the `Phone *string` field, and run the schema drift test. Practice the up, down, up cycle.
2. **Add reviews.** Create a `reviews` table with `user_id`, `product_id`, a `rating` with a CHECK between 1 and 5, a `body`, and a unique pair on `(user_id, product_id)`. Add the model and a `ReviewStore.Create` that returns a domain error when someone reviews the same product twice.
3. **List reviews with authors.** Return a product's reviews with each user's name, using one `Joins` or one `Preload`. Turn on the Info logger and confirm there is no N+1.
4. **Average rating.** Add a method that returns the average rating and review count for a product, using `Select` and `Scan` or raw SQL.
5. **Cancel an order.** Write `CancelOrder` in a transaction. It should lock the order, allow cancellation only from `pending`, put the stock back, and set the status to `cancelled`.
6. **Keyset pagination.** Replace offset pagination on products with an `id < ?` cursor. Compare the SQL for page 1 and page 1000.
7. **Race test.** Write a test where 20 goroutines try to buy the last unit of a product. Exactly one must succeed, and stock must end at zero. Use real commits, not the rollback helper.
8. **Break a migration on purpose.** Add a migration with a typo halfway through, watch it fail, see the dirty flag with `version`, then recover with `force`.
9. **Expand and contract.** Rename `users.name` to `users.full_name` in the four steps from Lesson 14, including the batched backfill.
10. **Try the generics API.** Convert two store methods to `gorm.G[T]` and compare how the code reads.
11. **Wire up CI.** Run all migrations on a fresh PostgreSQL service, run the full test suite, and add the round trip test against a throwaway database.

[Back to contents](#contents)

## Appendix C: Command cheat sheet

### Docker and database

```bash
docker compose up -d
docker compose exec db psql -U app -d shop
docker compose exec db psql -U app -d shop -c '\dt'      # list tables
docker compose exec db psql -U app -d shop -c '\d users' # describe a table
```

### golang-migrate

```bash
migrate create -ext sql -dir db/migrations -seq <name>
migrate -path db/migrations -database "$DATABASE_URL" up [N]
migrate -path db/migrations -database "$DATABASE_URL" down N
migrate -path db/migrations -database "$DATABASE_URL" goto V
migrate -path db/migrations -database "$DATABASE_URL" version
migrate -path db/migrations -database "$DATABASE_URL" force V
migrate -path db/migrations -database "$DATABASE_URL" drop
```

### Tests

```bash
TEST_DATABASE_URL='postgres://app:secret@localhost:5432/shop_test?sslmode=disable' go test ./...
```

### GORM quick reference

```go
// Connect
db, err := gorm.Open(postgres.Open(url), &gorm.Config{TranslateError: true})

// Create
db.WithContext(ctx).Create(&x)
db.WithContext(ctx).CreateInBatches(items, 100)

// Read
db.First(&x, id)                    // by primary key, ErrRecordNotFound if missing
db.Where("col = ?", v).Find(&xs)    // zero or more rows
db.Preload("Items").Find(&orders)   // load a has-many
db.Joins("Profile").Find(&users)    // load a has-one in one query
db.Select("a", "b").Order("a DESC").Limit(10).Offset(20).Find(&xs)
db.Model(&X{}).Where("...").Count(&n)

// Update
db.Model(&x).Update("col", v)
db.Model(&x).Updates(map[string]any{"a": 1, "b": 0})
db.Model(&x).Select("A", "B").Updates(X{A: 1, B: 0})
db.Model(&x).Update("n", gorm.Expr("n - ?", 1))

// Delete
db.Delete(&X{}, id)                 // soft delete if DeletedAt exists
db.Unscoped().Delete(&X{}, id)      // hard delete

// Transactions and locking
db.Transaction(func(tx *gorm.DB) error { return nil })
tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&x, id)

// Raw SQL
db.Raw("SELECT ...", args).Scan(&rows)
db.Exec("UPDATE ...", args)

// Upsert
db.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "name"}}, DoNothing: true}).Create(&xs)
```

