<<<<<<< HEAD
# house-bartender
=======
House Bartender 🍸🏠

A tiny, self-hosted cocktail ordering app for home use.

House Bartender has two portals:

User portal: browse available cocktails and place an order

Bartender portal: manage products (ingredients), enable/disable cocktails, and handle the live order queue

Built as a lightweight, containerized app with a single Go backend binary, server-rendered HTML + HTMX, SQLite persistence, and SSE notifications.

Table of contents

Features (v1)

Tech stack

Quick start (Docker Compose)

First-time setup

Seeding products & cocktails

Roles & access

How availability works

Data persistence & backups

Development (local)

Troubleshooting

Screenshots

Roadmap ideas

Contributing

Security notes

Features (v1)
User portal

View cocktails that are available right now

Availability = cocktail enabled AND all required ingredients available

Filter by alcohol / tags / include/exclude ingredient

View cocktail recipe details (ingredients + instructions)

Place an order (quantity, location, notes)

View your order history + timeline events

Bartender portal

Products (ingredients)

Toggle availability

Optional stock count (if set, availability comes from stock > 0)

Cocktails

Create / edit / delete cocktails

Enable / disable cocktails

Manage recipe ingredients (required/optional)

Queue

Live updates via SSE

Assign orders to bartenders

Update status flow (placed → in progress → delivered / cancelled)

Admin

Create users, assign roles (USER / BARTENDER / ADMIN)

Enable/disable accounts

Toggle bartender “on duty”

Set/reset user passwords

View DB counts + run idempotent seed

Tech stack

Backend: Go (single binary)

UI: server-rendered HTML + HTMX (no React)

Database: SQLite (persisted in Docker volume)

Realtime: Server-Sent Events (SSE)

Auth: cookie session + RBAC (USER/BARTENDER/ADMIN)

Deploy: Docker / Docker Compose

Quick start (Docker Compose)
1) Clone
git clone <your-repo-url>
cd house-bartender
2) Configure environment (recommended)

Create a .env file (or use your existing Compose env wiring) with at least stable session keys:

# required for stable logins across restarts
SESSION_HASH_KEY_HEX=$(openssl rand -hex 32)

# optional (recommended if your session implementation uses encryption)
SESSION_BLOCK_KEY_HEX=$(openssl rand -hex 32)

# optional: bootstrap first admin without onboarding page
BOOTSTRAP_ADMIN_EMAIL=admin@local
BOOTSTRAP_ADMIN_PASSWORD=change-me-strong
BOOTSTRAP_ADMIN_NAME=admin

# optional app config
ADDR=:8080
BASE_URL=http://localhost:8080
DATA_DIR=/data
DB_PATH=/data/housebartender.db
UPLOAD_DIR=/data/uploads

If SESSION_HASH_KEY_HEX is missing/too short, the app will generate an ephemeral key and you’ll be logged out on restart.

3) Run
docker compose up -d --build
docker compose logs -f app

Open:

http://localhost:8080

First-time setup
Option A: Bootstrap admin via env (recommended)

If BOOTSTRAP_ADMIN_EMAIL / BOOTSTRAP_ADMIN_PASSWORD / BOOTSTRAP_ADMIN_NAME are set and no admin exists yet, the app will create the first admin automatically.

Option B: Onboarding page

If no admin exists and you didn’t provide bootstrap env vars, you’ll be redirected to:

/onboarding

Create the first admin there.

Seeding products & cocktails

The app seeds the catalog only if the products and cocktails tables are empty, so your changes persist.

You can also run an idempotent seed from:

/admin/settings → Seed

To verify your DB quickly inside the Docker volume:

docker compose ps --services
# (your service is usually "app")

# If you know the volume name (example):
vol="house-bartender_housebartender_data"
echo "select count(*) from products; select count(*) from cocktails;" \
| docker run --rm -i -v ${vol}:/data alpine:3.20 sh -lc \
  "apk add --no-cache sqlite >/dev/null; sqlite3 /data/housebartender.db"
Roles & access
USER

Browse available cocktails

Place orders

View own orders

BARTENDER

Manage products (availability/stock)

Manage cocktails (enable/disable, edit recipes)

Work order queue (assign, status changes)

ADMIN

Everything a bartender can do

Manage users & roles + passwords

Settings & seed tools

How availability works
Product availability

If stock_count is set → available when stock_count > 0

Else → available when is_available = 1

Cocktail availability

A cocktail is available when:

is_enabled = 1

AND every required ingredient’s product is available (rule above)

Data persistence & backups

By default in Docker:

SQLite DB: /data/housebartender.db

Uploads: /data/uploads

/data is a Docker volume (survives container recreation).

Backup example:

# Stop the app first to avoid partial writes
docker compose stop

# Copy DB out of the volume using a helper container
vol="house-bartender_housebartender_data"
docker run --rm -v ${vol}:/data -v "$(pwd)":/backup alpine:3.20 \
  sh -lc "cp /data/housebartender.db /backup/housebartender.db.backup"

docker compose start
Development (local)
Requirements

Go toolchain

SQLite build support (CGO enabled)

Run locally
go run ./cmd/housebartender

Or build:

go build -o housebartender ./cmd/housebartender
./housebartender
Troubleshooting
Sessions reset on restart

Set a stable SESSION_HASH_KEY_HEX (and optionally SESSION_BLOCK_KEY_HEX) in env.

No cocktails appear for user

User portal only shows cocktails that are computed available:

Ensure cocktails are enabled

Ensure required ingredients are available (stock > 0 or is_available=1)

Template parse errors

If you edited templates recently:

Make sure all {{if}} / {{else}} / {{end}} blocks are properly nested

Avoid using template functions unless they exist in Go template.FuncMap

Screenshots

Roadmap ideas

Per-bartender notifications (only assigned bartender)

Inventory consumption on order completion

Printer-friendly “ticket” view for orders

Mobile-first queue UX improvements

Export/import catalog JSON

Contributing

PRs welcome. Suggested workflow:

Open an issue describing the change

Create a feature branch

Keep changes focused and small

Ensure templates render correctly for USER / BARTENDER / ADMIN paths

Security notes

This app is designed for home/self-hosted use. If exposing beyond your LAN:

Put behind a reverse proxy with TLS

Consider IP allowlists / auth hardening

Use strong admin passwords

Keep session keys secret
>>>>>>> b36a881 (Initial release (v1.0.0))
