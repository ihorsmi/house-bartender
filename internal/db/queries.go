package db

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type Queries struct {
	db *sql.DB
}

func unixNow() int64 { return time.Now().Unix() }
func b2i(b bool) int { if b { return 1 }; return 0 }
func i2b(i int) bool { return i != 0 }

func tFromUnix(u int64) time.Time {
	if u <= 0 {
		return time.Time{}
	}
	return time.Unix(u, 0)
}

func computedAvailExpr() string {
	// Force numeric 0/1 to make scanning stable.
	// if stock_count is set -> derived from stock_count>0 else is_available
	return `(CASE
		WHEN p.stock_count IS NOT NULL THEN (CASE WHEN p.stock_count > 0 THEN 1 ELSE 0 END)
		ELSE (CASE WHEN p.is_available = 1 THEN 1 ELSE 0 END)
	END)`
}

/* ---------------- Users ---------------- */

func (q *Queries) HasAnyAdmin() (bool, error) {
	row := q.db.QueryRow(`SELECT COUNT(1) FROM users WHERE role='ADMIN'`)
	var n int
	if err := row.Scan(&n); err != nil {
		return false, err
	}
	return n > 0, nil
}

func (q *Queries) GetUserByID(id int64) (*User, error) {
	row := q.db.QueryRow(`
		SELECT id,email,password_hash,role,display_name,is_active,on_duty,created_at,updated_at
		FROM users WHERE id=?`, id)
	var u User
	var isActive, onDuty int
	var ca, ua int64
	if err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.DisplayName, &isActive, &onDuty, &ca, &ua); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	u.IsActive = i2b(isActive)
	u.OnDuty = i2b(onDuty)
	u.CreatedAt = tFromUnix(ca)
	u.UpdatedAt = tFromUnix(ua)
	return &u, nil
}

func (q *Queries) GetUserByEmail(email string) (*User, error) {
	row := q.db.QueryRow(`
		SELECT id,email,password_hash,role,display_name,is_active,on_duty,created_at,updated_at
		FROM users WHERE email=?`, email)
	var u User
	var isActive, onDuty int
	var ca, ua int64
	if err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.DisplayName, &isActive, &onDuty, &ca, &ua); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	u.IsActive = i2b(isActive)
	u.OnDuty = i2b(onDuty)
	u.CreatedAt = tFromUnix(ca)
	u.UpdatedAt = tFromUnix(ua)
	return &u, nil
}

func (q *Queries) ListUsers() ([]User, error) {
	rows, err := q.db.Query(`
		SELECT id,email,password_hash,role,display_name,is_active,on_duty,created_at,updated_at
		FROM users ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []User
	for rows.Next() {
		var u User
		var isActive, onDuty int
		var ca, ua int64
		if err := rows.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.DisplayName, &isActive, &onDuty, &ca, &ua); err != nil {
			return nil, err
		}
		u.IsActive = i2b(isActive)
		u.OnDuty = i2b(onDuty)
		u.CreatedAt = tFromUnix(ca)
		u.UpdatedAt = tFromUnix(ua)
		out = append(out, u)
	}
	return out, nil
}

func (q *Queries) CreateUser(p CreateUserParams) (int64, error) {
	res, err := q.db.Exec(`
		INSERT INTO users(email,password_hash,role,display_name,is_active,on_duty,created_at,updated_at)
		VALUES(?,?,?,?,?,?,?,?)`,
		p.Email, p.PasswordHash, p.Role, p.DisplayName, b2i(p.IsActive), b2i(p.OnDuty), unixNow(), unixNow())
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (q *Queries) UpdateUser(p UpdateUserParams) error {
	_, err := q.db.Exec(`
		UPDATE users SET email=?, role=?, display_name=?, updated_at=? WHERE id=?`,
		p.Email, p.Role, p.DisplayName, unixNow(), p.ID)
	return err
}

func (q *Queries) SetUserPassword(id int64, hash string) error {
	_, err := q.db.Exec(`UPDATE users SET password_hash=?, updated_at=? WHERE id=?`, hash, unixNow(), id)
	return err
}

func (q *Queries) SetUserActive(id int64, active bool) error {
	_, err := q.db.Exec(`UPDATE users SET is_active=?, updated_at=? WHERE id=?`, b2i(active), unixNow(), id)
	return err
}

func (q *Queries) SetUserDuty(id int64, onDuty bool) error {
	_, err := q.db.Exec(`UPDATE users SET on_duty=?, updated_at=? WHERE id=?`, b2i(onDuty), unixNow(), id)
	return err
}

/* ---------------- Products ---------------- */

func (q *Queries) ListProducts(search string) ([]Product, error) {
	search = strings.TrimSpace(strings.ToLower(search))
	var rows *sql.Rows
	var err error
	if search == "" {
		rows, err = q.db.Query(fmt.Sprintf(`
			SELECT
				p.id,
				COALESCE(p.name,'') AS name,
				COALESCE(p.category,'') AS category,
				COALESCE(p.abv_percent, 0) AS abv_percent,
				COALESCE(p.allergen_flags,'') AS allergen_flags,
				COALESCE(p.notes,'') AS notes,
				COALESCE(p.is_available, 0) AS is_available,
				p.stock_count,
				%s AS computed_avail,
				p.created_at,p.updated_at
			FROM products p
			ORDER BY p.category, p.name`, computedAvailExpr()))
	} else {
		like := "%" + search + "%"
		rows, err = q.db.Query(fmt.Sprintf(`
			SELECT
				p.id,
				COALESCE(p.name,'') AS name,
				COALESCE(p.category,'') AS category,
				COALESCE(p.abv_percent, 0) AS abv_percent,
				COALESCE(p.allergen_flags,'') AS allergen_flags,
				COALESCE(p.notes,'') AS notes,
				COALESCE(p.is_available, 0) AS is_available,
				p.stock_count,
				%s AS computed_avail,
				p.created_at,p.updated_at
			FROM products p
			WHERE lower(p.name) LIKE ? OR lower(p.category) LIKE ?
			ORDER BY p.category, p.name`, computedAvailExpr()), like, like)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Product
	for rows.Next() {
		var p Product
		var isAvail, comp int
		var ca, ua int64
		if err := rows.Scan(&p.ID, &p.Name, &p.Category, &p.ABVPercent, &p.AllergenFlags, &p.Notes, &isAvail, &p.StockCount, &comp, &ca, &ua); err != nil {
			return nil, err
		}
		p.IsAvailable = i2b(isAvail)
		p.ComputedAvail = i2b(comp)
		p.CreatedAt = tFromUnix(ca)
		p.UpdatedAt = tFromUnix(ua)
		out = append(out, p)
	}
	return out, nil
}

func (q *Queries) CreateProduct(p CreateProductParams) (int64, error) {
	res, err := q.db.Exec(`
		INSERT INTO products(name,category,abv_percent,allergen_flags,notes,is_available,stock_count,created_at,updated_at)
		VALUES(?,?,?,?,?,?,?,?,?)`,
		p.Name, p.Category, p.ABVPercent, p.AllergenFlags, p.Notes, b2i(p.IsAvailable), p.StockCount, unixNow(), unixNow())
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (q *Queries) UpdateProduct(p UpdateProductParams) error {
	_, err := q.db.Exec(`
		UPDATE products
		SET name=?, category=?, abv_percent=?, allergen_flags=?, notes=?, is_available=?, stock_count=?, updated_at=?
		WHERE id=?`,
		p.Name, p.Category, p.ABVPercent, p.AllergenFlags, p.Notes, b2i(p.IsAvailable), p.StockCount, unixNow(), p.ID)
	return err
}

func (q *Queries) DeleteProduct(id int64) error {
	_, err := q.db.Exec(`DELETE FROM products WHERE id=?`, id)
	return err
}

func (q *Queries) ToggleProductAvailability(id int64, avail bool) error {
	_, err := q.db.Exec(`UPDATE products SET is_available=?, updated_at=? WHERE id=?`, b2i(avail), unixNow(), id)
	return err
}

func (q *Queries) SetProductStock(id int64, stock *int64) error {
	_, err := q.db.Exec(`UPDATE products SET stock_count=?, updated_at=? WHERE id=?`, stock, unixNow(), id)
	return err
}

/* ---------------- Cocktails ---------------- */

func (q *Queries) GetCocktailByID(id int64) (*Cocktail, error) {
	row := q.db.QueryRow(`
		SELECT
			id,
			COALESCE(name,''),
			COALESCE(description,''),
			COALESCE(image_path,''),
			COALESCE(tags,''),
			COALESCE(difficulty,'easy'),
			COALESCE(prep_time_minutes, 0),
			COALESCE(instructions,''),
			COALESCE(is_enabled,0),
			created_at,updated_at
		FROM cocktails WHERE id=?`, id)
	var c Cocktail
	var enabled int
	var ca, ua int64
	if err := row.Scan(&c.ID, &c.Name, &c.Description, &c.ImagePath, &c.Tags, &c.Difficulty, &c.PrepTimeMinutes, &c.Instructions, &enabled, &ca, &ua); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	c.IsEnabled = i2b(enabled)
	c.CreatedAt = tFromUnix(ca)
	c.UpdatedAt = tFromUnix(ua)
	return &c, nil
}

func (q *Queries) ListCocktailsComputed(onlyAvailable bool) ([]Cocktail, error) {
	// computed availability: enabled AND all required ingredients are available
	sqlq := fmt.Sprintf(`
		SELECT * FROM (
			SELECT
				c.id,
				COALESCE(c.name,'') AS name,
				COALESCE(c.description,'') AS description,
				COALESCE(c.image_path,'') AS image_path,
				COALESCE(c.tags,'') AS tags,
				COALESCE(c.difficulty,'easy') AS difficulty,
				COALESCE(c.prep_time_minutes, 0) AS prep_time_minutes,
				COALESCE(c.instructions,'') AS instructions,
				COALESCE(c.is_enabled,0) AS is_enabled,
				CASE
					WHEN COALESCE(c.is_enabled,0) = 0 THEN 0
					WHEN EXISTS (
						SELECT 1
						FROM cocktail_ingredients ci
						JOIN products p ON p.id = ci.product_id
						WHERE ci.cocktail_id = c.id
						  AND ci.required = 1
						  AND (%s) = 0
					) THEN 0
					ELSE 1
				END AS computed_avail,
				c.created_at,c.updated_at
			FROM cocktails c
		) WHERE (? = 0 OR computed_avail = 1)
		ORDER BY name`, computedAvailExpr())

	rows, err := q.db.Query(sqlq, b2i(onlyAvailable))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Cocktail
	for rows.Next() {
		var c Cocktail
		var enabled, comp int
		var ca, ua int64
		if err := rows.Scan(&c.ID, &c.Name, &c.Description, &c.ImagePath, &c.Tags, &c.Difficulty, &c.PrepTimeMinutes, &c.Instructions, &enabled, &comp, &ca, &ua); err != nil {
			return nil, err
		}
		c.IsEnabled = i2b(enabled)
		c.ComputedAvail = i2b(comp)
		c.CreatedAt = tFromUnix(ca)
		c.UpdatedAt = tFromUnix(ua)
		out = append(out, c)
	}
	return out, nil
}

func (q *Queries) CreateCocktail(p CreateCocktailParams) (int64, error) {
	res, err := q.db.Exec(`
		INSERT INTO cocktails(name,description,image_path,tags,difficulty,prep_time_minutes,instructions,is_enabled,created_at,updated_at)
		VALUES(?,?,?,?,?,?,?,?,?,?)`,
		p.Name, p.Description, p.ImagePath, p.Tags, p.Difficulty, p.PrepTimeMinutes, p.Instructions, b2i(p.IsEnabled), unixNow(), unixNow())
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (q *Queries) UpdateCocktail(p UpdateCocktailParams) error {
	_, err := q.db.Exec(`
		UPDATE cocktails
		SET name=?, description=?, image_path=?, tags=?, difficulty=?, prep_time_minutes=?, instructions=?, is_enabled=?, updated_at=?
		WHERE id=?`,
		p.Name, p.Description, p.ImagePath, p.Tags, p.Difficulty, p.PrepTimeMinutes, p.Instructions, b2i(p.IsEnabled), unixNow(), p.ID)
	return err
}

func (q *Queries) ToggleCocktailEnabled(id int64, enabled bool) error {
	_, err := q.db.Exec(`UPDATE cocktails SET is_enabled=?, updated_at=? WHERE id=?`, b2i(enabled), unixNow(), id)
	return err
}

func (q *Queries) DeleteCocktail(id int64) error {
	_, err := q.db.Exec(`DELETE FROM cocktails WHERE id=?`, id)
	return err
}

func (q *Queries) GetCocktailIngredients(cocktailID int64) ([]CocktailIngredient, error) {
	sqlq := fmt.Sprintf(`
		SELECT
			ci.id,ci.cocktail_id,ci.product_id,ci.quantity,COALESCE(ci.unit,''),ci.required,
			COALESCE(p.name,''),COALESCE(p.category,''),
			%s AS product_avail
		FROM cocktail_ingredients ci
		JOIN products p ON p.id = ci.product_id
		WHERE ci.cocktail_id=?
		ORDER BY ci.required DESC, p.category, p.name`, computedAvailExpr())

	rows, err := q.db.Query(sqlq, cocktailID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []CocktailIngredient
	for rows.Next() {
		var ci CocktailIngredient
		var req, pav int
		if err := rows.Scan(&ci.ID, &ci.CocktailID, &ci.ProductID, &ci.Quantity, &ci.Unit, &req, &ci.ProductName, &ci.ProductCategory, &pav); err != nil {
			return nil, err
		}
		ci.Required = i2b(req)
		ci.ProductAvail = i2b(pav)
		out = append(out, ci)
	}
	return out, nil
}

func (q *Queries) ReplaceCocktailIngredients(cocktailID int64, items []IngredientUpsertItem) error {
	tx, err := q.db.Begin()
	if err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM cocktail_ingredients WHERE cocktail_id=?`, cocktailID); err != nil {
		_ = tx.Rollback()
		return err
	}
	for _, it := range items {
		if it.ProductID <= 0 {
			continue
		}
		if _, err := tx.Exec(`
			INSERT INTO cocktail_ingredients(cocktail_id,product_id,quantity,unit,required)
			VALUES(?,?,?,?,?)`, cocktailID, it.ProductID, it.Quantity, it.Unit, b2i(it.Required)); err != nil {
			_ = tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

/* ---------------- Orders ---------------- */

func (q *Queries) CreateOrder(p CreateOrderParams) (int64, error) {
	tx, err := q.db.Begin()
	if err != nil {
		return 0, err
	}
	res, err := tx.Exec(`
		INSERT INTO orders(user_id,cocktail_id,quantity,notes,location,status,assigned_bartender_id,created_at,updated_at)
		VALUES(?,?,?,?,?,'PLACED',NULL,?,?)`,
		p.UserID, p.CocktailID, p.Quantity, p.Notes, p.Location, unixNow(), unixNow())
	if err != nil {
		_ = tx.Rollback()
		return 0, err
	}
	id, _ := res.LastInsertId()

	if _, err := tx.Exec(`
		INSERT INTO order_events(order_id,from_status,to_status,changed_by_user_id,created_at)
		VALUES(?, '', 'PLACED', NULL, ?)`, id, unixNow()); err != nil {
		_ = tx.Rollback()
		return 0, err
	}

	return id, tx.Commit()
}

func (q *Queries) GetOrderByID(id int64) (*Order, error) {
	row := q.db.QueryRow(`
		SELECT
			o.id,o.user_id,o.cocktail_id,o.quantity,COALESCE(o.notes,''),COALESCE(o.location,''),COALESCE(o.status,''),o.assigned_bartender_id,o.created_at,o.updated_at,
			COALESCE(u.display_name,''),
			COALESCE(c.name,''),COALESCE(c.image_path,''),
			COALESCE(ub.display_name,'')
		FROM orders o
		JOIN users u ON u.id=o.user_id
		JOIN cocktails c ON c.id=o.cocktail_id
		LEFT JOIN users ub ON ub.id=o.assigned_bartender_id
		WHERE o.id=?`, id)

	var o Order
	var bid sql.NullInt64
	var ca, ua int64
	if err := row.Scan(&o.ID, &o.UserID, &o.CocktailID, &o.Quantity, &o.Notes, &o.Location, &o.Status, &bid, &ca, &ua,
		&o.UserDisplayName, &o.CocktailName, &o.CocktailImagePath, &o.AssignedBartenderName); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if bid.Valid {
		o.AssignedBartenderID = &bid.Int64
	}
	o.CreatedAt = tFromUnix(ca)
	o.UpdatedAt = tFromUnix(ua)
	return &o, nil
}

func (q *Queries) ListOrdersForUser(userID int64) ([]Order, error) {
	rows, err := q.db.Query(`
		SELECT
			o.id,o.user_id,o.cocktail_id,o.quantity,COALESCE(o.notes,''),COALESCE(o.location,''),COALESCE(o.status,''),o.assigned_bartender_id,o.created_at,o.updated_at,
			COALESCE(u.display_name,''),
			COALESCE(c.name,''),COALESCE(c.image_path,''),
			COALESCE(ub.display_name,'')
		FROM orders o
		JOIN users u ON u.id=o.user_id
		JOIN cocktails c ON c.id=o.cocktail_id
		LEFT JOIN users ub ON ub.id=o.assigned_bartender_id
		WHERE o.user_id=?
		ORDER BY o.created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Order
	for rows.Next() {
		var o Order
		var bid sql.NullInt64
		var ca, ua int64
		if err := rows.Scan(&o.ID, &o.UserID, &o.CocktailID, &o.Quantity, &o.Notes, &o.Location, &o.Status, &bid, &ca, &ua,
			&o.UserDisplayName, &o.CocktailName, &o.CocktailImagePath, &o.AssignedBartenderName); err != nil {
			return nil, err
		}
		if bid.Valid {
			o.AssignedBartenderID = &bid.Int64
		}
		o.CreatedAt = tFromUnix(ca)
		o.UpdatedAt = tFromUnix(ua)
		out = append(out, o)
	}
	return out, nil
}

func (q *Queries) ListOrderQueue() ([]Order, error) {
	rows, err := q.db.Query(`
		SELECT
			o.id,o.user_id,o.cocktail_id,o.quantity,COALESCE(o.notes,''),COALESCE(o.location,''),COALESCE(o.status,''),o.assigned_bartender_id,o.created_at,o.updated_at,
			COALESCE(u.display_name,''),
			COALESCE(c.name,''),COALESCE(c.image_path,''),
			COALESCE(ub.display_name,'')
		FROM orders o
		JOIN users u ON u.id=o.user_id
		JOIN cocktails c ON c.id=o.cocktail_id
		LEFT JOIN users ub ON ub.id=o.assigned_bartender_id
		WHERE o.status NOT IN ('DELIVERED','CANCELLED')
		ORDER BY o.created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Order
	for rows.Next() {
		var o Order
		var bid sql.NullInt64
		var ca, ua int64
		if err := rows.Scan(&o.ID, &o.UserID, &o.CocktailID, &o.Quantity, &o.Notes, &o.Location, &o.Status, &bid, &ca, &ua,
			&o.UserDisplayName, &o.CocktailName, &o.CocktailImagePath, &o.AssignedBartenderName); err != nil {
			return nil, err
		}
		if bid.Valid {
			o.AssignedBartenderID = &bid.Int64
		}
		o.CreatedAt = tFromUnix(ca)
		o.UpdatedAt = tFromUnix(ua)
		out = append(out, o)
	}
	return out, nil
}

func (q *Queries) AssignOrder(orderID int64, bartenderID *int64) error {
	_, err := q.db.Exec(`UPDATE orders SET assigned_bartender_id=?, updated_at=? WHERE id=?`, bartenderID, unixNow(), orderID)
	return err
}

func (q *Queries) UpdateOrderStatus(orderID int64, from, to string, changedBy *int64) error {
	tx, err := q.db.Begin()
	if err != nil {
		return err
	}
	if _, err := tx.Exec(`UPDATE orders SET status=?, updated_at=? WHERE id=?`, to, unixNow(), orderID); err != nil {
		_ = tx.Rollback()
		return err
	}
	_, err = tx.Exec(`
		INSERT INTO order_events(order_id,from_status,to_status,changed_by_user_id,created_at)
		VALUES(?,?,?,?,?)`, orderID, from, to, changedBy, unixNow())
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

func (q *Queries) ListOrderEvents(orderID int64) ([]OrderEvent, error) {
	rows, err := q.db.Query(`
		SELECT
			e.id,e.order_id,COALESCE(e.from_status,''),COALESCE(e.to_status,''),e.changed_by_user_id,e.created_at,
			COALESCE(u.display_name,'')
		FROM order_events e
		LEFT JOIN users u ON u.id=e.changed_by_user_id
		WHERE e.order_id=?
		ORDER BY e.created_at ASC, e.id ASC`, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []OrderEvent
	for rows.Next() {
		var e OrderEvent
		var cb sql.NullInt64
		var ca int64
		if err := rows.Scan(&e.ID, &e.OrderID, &e.FromStatus, &e.ToStatus, &cb, &ca, &e.ChangedByName); err != nil {
			return nil, err
		}
		if cb.Valid {
			e.ChangedByUserID = &cb.Int64
		}
		e.CreatedAt = tFromUnix(ca)
		out = append(out, e)
	}
	return out, nil
}

/* ---------------- Debug ---------------- */

func (q *Queries) DebugCounts() (string, error) {
	type c struct {
		name string
		qry  string
	}
	checks := []c{
		{"users", "SELECT COUNT(1) FROM users"},
		{"products", "SELECT COUNT(1) FROM products"},
		{"cocktails", "SELECT COUNT(1) FROM cocktails"},
		{"orders", "SELECT COUNT(1) FROM orders"},
		{"events", "SELECT COUNT(1) FROM order_events"},
	}
	var parts []string
	for _, it := range checks {
		row := q.db.QueryRow(it.qry)
		var n int
		if err := row.Scan(&n); err != nil {
			return "", err
		}
		parts = append(parts, fmt.Sprintf("%s=%d", it.name, n))
	}
	return strings.Join(parts, " | "), nil
}
