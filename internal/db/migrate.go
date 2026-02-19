package db

import "database/sql"

func Migrate(db *sql.DB) error {
	stmts := []string{
		`PRAGMA foreign_keys = ON;`,

		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			email TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
			role TEXT NOT NULL CHECK(role IN ('ADMIN','BARTENDER','USER')),
			display_name TEXT NOT NULL,
			is_active INTEGER NOT NULL DEFAULT 1,
			on_duty INTEGER NOT NULL DEFAULT 0,
			created_at INTEGER NOT NULL DEFAULT (strftime('%s','now')),
			updated_at INTEGER NOT NULL DEFAULT (strftime('%s','now'))
		);`,

		`CREATE TABLE IF NOT EXISTS products (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			category TEXT NOT NULL,
			abv_percent REAL NULL,
			allergen_flags TEXT NOT NULL DEFAULT '',
			notes TEXT NOT NULL DEFAULT '',
			is_available INTEGER NOT NULL DEFAULT 1,
			stock_count INTEGER NULL,
			created_at INTEGER NOT NULL DEFAULT (strftime('%s','now')),
			updated_at INTEGER NOT NULL DEFAULT (strftime('%s','now'))
		);`,

		`CREATE TABLE IF NOT EXISTS cocktails (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			description TEXT NOT NULL DEFAULT '',
			image_path TEXT NOT NULL DEFAULT '',
			tags TEXT NOT NULL DEFAULT '',
			difficulty TEXT NOT NULL DEFAULT 'easy',
			prep_time_minutes INTEGER NOT NULL DEFAULT 5,
			instructions TEXT NOT NULL DEFAULT '',
			is_enabled INTEGER NOT NULL DEFAULT 1,
			created_at INTEGER NOT NULL DEFAULT (strftime('%s','now')),
			updated_at INTEGER NOT NULL DEFAULT (strftime('%s','now'))
		);`,

		`CREATE TABLE IF NOT EXISTS cocktail_ingredients (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			cocktail_id INTEGER NOT NULL,
			product_id INTEGER NOT NULL,
			quantity REAL NULL,
			unit TEXT NOT NULL DEFAULT '',
			required INTEGER NOT NULL DEFAULT 1,
			FOREIGN KEY(cocktail_id) REFERENCES cocktails(id) ON DELETE CASCADE,
			FOREIGN KEY(product_id) REFERENCES products(id) ON DELETE RESTRICT
		);`,

		`CREATE TABLE IF NOT EXISTS orders (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			cocktail_id INTEGER NOT NULL,
			quantity INTEGER NOT NULL DEFAULT 1,
			notes TEXT NOT NULL DEFAULT '',
			location TEXT NOT NULL DEFAULT '',
			status TEXT NOT NULL CHECK(status IN ('PLACED','ACCEPTED','IN_PROGRESS','READY','DELIVERED','CANCELLED')),
			assigned_bartender_id INTEGER NULL,
			created_at INTEGER NOT NULL DEFAULT (strftime('%s','now')),
			updated_at INTEGER NOT NULL DEFAULT (strftime('%s','now')),
			FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE,
			FOREIGN KEY(cocktail_id) REFERENCES cocktails(id) ON DELETE RESTRICT,
			FOREIGN KEY(assigned_bartender_id) REFERENCES users(id) ON DELETE SET NULL
		);`,

		`CREATE TABLE IF NOT EXISTS order_events (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			order_id INTEGER NOT NULL,
			from_status TEXT NOT NULL DEFAULT '',
			to_status TEXT NOT NULL,
			changed_by_user_id INTEGER NULL,
			created_at INTEGER NOT NULL DEFAULT (strftime('%s','now')),
			FOREIGN KEY(order_id) REFERENCES orders(id) ON DELETE CASCADE,
			FOREIGN KEY(changed_by_user_id) REFERENCES users(id) ON DELETE SET NULL
		);`,

		`CREATE INDEX IF NOT EXISTS idx_orders_status_created ON orders(status, created_at);`,
		`CREATE INDEX IF NOT EXISTS idx_order_events_order_created ON order_events(order_id, created_at);`,
		`CREATE INDEX IF NOT EXISTS idx_cocktail_ingredients_cocktail ON cocktail_ingredients(cocktail_id);`,
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	for _, s := range stmts {
		if _, err := tx.Exec(s); err != nil {
			_ = tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}
