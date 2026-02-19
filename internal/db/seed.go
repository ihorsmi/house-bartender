package db

import (
	"database/sql"
	"fmt"
)

type seedProduct struct {
	Name          string
	Category      string
	ABVPercent    float64
	AllergenFlags string
	Notes         string
	IsAvailable   bool
	StockCount    *int64
}

type seedCocktail struct {
	Name            string
	Description     string
	ImagePath       string
	Tags            string
	Difficulty      string
	PrepTimeMinutes int64
	Instructions    string
	IsEnabled       bool
}

type seedIng struct {
	CocktailName string
	ProductName  string
	Quantity     *float64
	Unit         string
	Required     bool
}

func Seed(db *sql.DB) error {
	// Keep existing callers working: /admin/settings "Seed" calls db.Seed(...)
	return SeedCatalog(db)
}

func SeedCatalog(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	products := []seedProduct{
		{Name: "Water", Category: "Basics", ABVPercent: 0, IsAvailable: true},
		{Name: "Ice", Category: "Basics", ABVPercent: 0, IsAvailable: true, Notes: "Cubes"},
		{Name: "Lime", Category: "Fruit", ABVPercent: 0, IsAvailable: true},
		{Name: "Mint", Category: "Herbs", ABVPercent: 0, IsAvailable: true},
		{Name: "Sugar Syrup", Category: "Sweeteners", ABVPercent: 0, IsAvailable: true, Notes: "Simple syrup"},
		{Name: "Soda Water", Category: "Mixers", ABVPercent: 0, IsAvailable: true},
		{Name: "Tonic Water", Category: "Mixers", ABVPercent: 0, IsAvailable: true},
		{Name: "Ginger Beer", Category: "Mixers", ABVPercent: 0, IsAvailable: true},
		{Name: "Cola", Category: "Soft Drinks", ABVPercent: 0, IsAvailable: true},
		{Name: "Orange Juice", Category: "Juice", ABVPercent: 0, IsAvailable: true},
		{Name: "Cranberry Juice", Category: "Juice", ABVPercent: 0, IsAvailable: true},

		{Name: "Non-Alcoholic Beer", Category: "Beer", ABVPercent: 0, IsAvailable: true},
		{Name: "Beer (Lager)", Category: "Beer", ABVPercent: 5, IsAvailable: true},
		{Name: "Red Wine", Category: "Wine", ABVPercent: 13, IsAvailable: true},

		{Name: "Gin", Category: "Spirits", ABVPercent: 40, IsAvailable: true},
		{Name: "Vodka", Category: "Spirits", ABVPercent: 40, IsAvailable: true},
		{Name: "White Rum", Category: "Spirits", ABVPercent: 40, IsAvailable: true},
		{Name: "Tequila", Category: "Spirits", ABVPercent: 40, IsAvailable: true},
		{Name: "Triple Sec", Category: "Liqueurs", ABVPercent: 30, IsAvailable: true},
	}

	cocktails := []seedCocktail{
		{
			Name: "Gin & Tonic",
			Description: "Crisp and bitter-sweet — a true staple.",
			Tags: "alcoholic,classic,refreshing",
			Difficulty: "easy",
			PrepTimeMinutes: 2,
			Instructions: "Fill a glass with ice.\nAdd gin.\nTop with tonic water.\nGarnish with lime.",
			IsEnabled: true,
		},
		{
			Name: "Mojito",
			Description: "Mint, lime, rum — bright and refreshing.",
			Tags: "alcoholic,classic,mint,citrus,refreshing",
			Difficulty: "medium",
			PrepTimeMinutes: 5,
			Instructions: "Muddle mint with sugar syrup and lime.\nAdd rum and ice.\nTop with soda water.\nStir gently.",
			IsEnabled: true,
		},
		{
			Name: "Margarita",
			Description: "Tequila + triple sec + lime. Straight to the point.",
			Tags: "alcoholic,sour,classic",
			Difficulty: "medium",
			PrepTimeMinutes: 4,
			Instructions: "Add tequila, triple sec and lime to a glass with ice.\nStir (or shake if you prefer).\nServe cold.",
			IsEnabled: true,
		},
		{
			Name: "Moscow Mule",
			Description: "Vodka + ginger beer + lime — spicy and cold.",
			Tags: "alcoholic,ginger,refreshing",
			Difficulty: "easy",
			PrepTimeMinutes: 3,
			Instructions: "Fill a glass with ice.\nAdd vodka.\nTop with ginger beer.\nSqueeze lime and stir.",
			IsEnabled: true,
		},
		{
			Name: "Cuba Libre",
			Description: "Rum and cola with lime.",
			Tags: "alcoholic,rum,cola,easy",
			Difficulty: "easy",
			PrepTimeMinutes: 2,
			Instructions: "Fill glass with ice.\nAdd rum.\nTop with cola.\nAdd lime.",
			IsEnabled: true,
		},
		{
			Name: "Vodka Soda",
			Description: "Clean and simple.",
			Tags: "alcoholic,simple,refreshing",
			Difficulty: "easy",
			PrepTimeMinutes: 2,
			Instructions: "Ice in glass.\nAdd vodka.\nTop with soda water.\nOptional lime.",
			IsEnabled: true,
		},
		{
			Name: "Rum & Ginger",
			Description: "Rum with ginger beer and lime.",
			Tags: "alcoholic,ginger,rum",
			Difficulty: "easy",
			PrepTimeMinutes: 3,
			Instructions: "Ice.\nAdd rum.\nTop with ginger beer.\nAdd lime.",
			IsEnabled: true,
		},
		{
			Name: "Cola",
			Description: "Classic cola, served cold.",
			Tags: "non-alcoholic,soda,easy",
			Difficulty: "easy",
			PrepTimeMinutes: 1,
			Instructions: "Pour cola into a glass.\nAdd ice if desired.",
			IsEnabled: true,
		},
		{
			Name: "Ginger Lime Fizz",
			Description: "Ginger beer with lime — spicy and bright.",
			Tags: "non-alcoholic,ginger,citrus,refreshing",
			Difficulty: "easy",
			PrepTimeMinutes: 2,
			Instructions: "Ice.\nAdd ginger beer.\nSqueeze lime.\nStir.",
			IsEnabled: true,
		},
		{
			Name: "Virgin Mojito",
			Description: "Mint + lime + soda, no alcohol.",
			Tags: "non-alcoholic,mint,citrus,refreshing",
			Difficulty: "easy",
			PrepTimeMinutes: 4,
			Instructions: "Muddle mint with sugar syrup and lime.\nAdd ice.\nTop with soda water.\nStir gently.",
			IsEnabled: true,
		},
		{
			Name: "Lime Soda",
			Description: "Soda water with lime.",
			Tags: "non-alcoholic,citrus,refreshing",
			Difficulty: "easy",
			PrepTimeMinutes: 2,
			Instructions: "Ice.\nTop with soda water.\nSqueeze lime.\nStir.",
			IsEnabled: true,
		},
		{
			Name: "Orange Spritzer",
			Description: "Orange juice topped with soda.",
			Tags: "non-alcoholic,citrus,refreshing",
			Difficulty: "easy",
			PrepTimeMinutes: 2,
			Instructions: "Ice.\nAdd orange juice.\nTop with soda water.\nStir.",
			IsEnabled: true,
		},
		{
			Name: "Cranberry Fizz",
			Description: "Cranberry juice topped with soda.",
			Tags: "non-alcoholic,fruity,refreshing",
			Difficulty: "easy",
			PrepTimeMinutes: 2,
			Instructions: "Ice.\nAdd cranberry juice.\nTop with soda water.\nStir.",
			IsEnabled: true,
		},
		{
			Name: "Non-Alcoholic Beer",
			Description: "Zero/low alcohol beer served cold.",
			Tags: "non-alcoholic,beer,easy",
			Difficulty: "easy",
			PrepTimeMinutes: 1,
			Instructions: "Serve chilled.\nOptional: pour into a glass.",
			IsEnabled: true,
		},
	}

	// quantities
	q30 := f64(30)
	q50 := f64(50)
	q60 := f64(60)
	q90 := f64(90)
	q120 := f64(120)
	q150 := f64(150)
	q1 := f64(1)
	q8 := f64(8)

	ings := []seedIng{
		// Gin & Tonic
		{CocktailName: "Gin & Tonic", ProductName: "Gin", Quantity: &q50, Unit: "ml", Required: true},
		{CocktailName: "Gin & Tonic", ProductName: "Tonic Water", Quantity: &q150, Unit: "ml", Required: true},
		{CocktailName: "Gin & Tonic", ProductName: "Lime", Quantity: &q1, Unit: "pc", Required: false},
		{CocktailName: "Gin & Tonic", ProductName: "Ice", Quantity: &q8, Unit: "pc", Required: false},

		// Mojito
		{CocktailName: "Mojito", ProductName: "White Rum", Quantity: &q50, Unit: "ml", Required: true},
		{CocktailName: "Mojito", ProductName: "Lime", Quantity: &q1, Unit: "pc", Required: true},
		{CocktailName: "Mojito", ProductName: "Mint", Quantity: &q8, Unit: "leaves", Required: true},
		{CocktailName: "Mojito", ProductName: "Sugar Syrup", Quantity: &q30, Unit: "ml", Required: true},
		{CocktailName: "Mojito", ProductName: "Soda Water", Quantity: &q90, Unit: "ml", Required: true},
		{CocktailName: "Mojito", ProductName: "Ice", Quantity: &q8, Unit: "pc", Required: false},

		// Margarita
		{CocktailName: "Margarita", ProductName: "Tequila", Quantity: &q50, Unit: "ml", Required: true},
		{CocktailName: "Margarita", ProductName: "Triple Sec", Quantity: &q30, Unit: "ml", Required: true},
		{CocktailName: "Margarita", ProductName: "Lime", Quantity: &q1, Unit: "pc", Required: true},
		{CocktailName: "Margarita", ProductName: "Ice", Quantity: &q8, Unit: "pc", Required: false},

		// Moscow Mule
		{CocktailName: "Moscow Mule", ProductName: "Vodka", Quantity: &q50, Unit: "ml", Required: true},
		{CocktailName: "Moscow Mule", ProductName: "Ginger Beer", Quantity: &q150, Unit: "ml", Required: true},
		{CocktailName: "Moscow Mule", ProductName: "Lime", Quantity: &q1, Unit: "pc", Required: true},
		{CocktailName: "Moscow Mule", ProductName: "Ice", Quantity: &q8, Unit: "pc", Required: false},

		// Cuba Libre
		{CocktailName: "Cuba Libre", ProductName: "White Rum", Quantity: &q50, Unit: "ml", Required: true},
		{CocktailName: "Cuba Libre", ProductName: "Cola", Quantity: &q150, Unit: "ml", Required: true},
		{CocktailName: "Cuba Libre", ProductName: "Lime", Quantity: &q1, Unit: "pc", Required: false},
		{CocktailName: "Cuba Libre", ProductName: "Ice", Quantity: &q8, Unit: "pc", Required: false},

		// Vodka Soda
		{CocktailName: "Vodka Soda", ProductName: "Vodka", Quantity: &q50, Unit: "ml", Required: true},
		{CocktailName: "Vodka Soda", ProductName: "Soda Water", Quantity: &q150, Unit: "ml", Required: true},
		{CocktailName: "Vodka Soda", ProductName: "Lime", Quantity: &q1, Unit: "pc", Required: false},

		// Rum & Ginger
		{CocktailName: "Rum & Ginger", ProductName: "White Rum", Quantity: &q50, Unit: "ml", Required: true},
		{CocktailName: "Rum & Ginger", ProductName: "Ginger Beer", Quantity: &q150, Unit: "ml", Required: true},
		{CocktailName: "Rum & Ginger", ProductName: "Lime", Quantity: &q1, Unit: "pc", Required: false},

		// Cola
		{CocktailName: "Cola", ProductName: "Cola", Quantity: &q150, Unit: "ml", Required: true},
		{CocktailName: "Cola", ProductName: "Ice", Quantity: &q8, Unit: "pc", Required: false},

		// Ginger Lime Fizz
		{CocktailName: "Ginger Lime Fizz", ProductName: "Ginger Beer", Quantity: &q150, Unit: "ml", Required: true},
		{CocktailName: "Ginger Lime Fizz", ProductName: "Lime", Quantity: &q1, Unit: "pc", Required: true},
		{CocktailName: "Ginger Lime Fizz", ProductName: "Ice", Quantity: &q8, Unit: "pc", Required: false},

		// Virgin Mojito
		{CocktailName: "Virgin Mojito", ProductName: "Lime", Quantity: &q1, Unit: "pc", Required: true},
		{CocktailName: "Virgin Mojito", ProductName: "Mint", Quantity: &q8, Unit: "leaves", Required: true},
		{CocktailName: "Virgin Mojito", ProductName: "Sugar Syrup", Quantity: &q30, Unit: "ml", Required: true},
		{CocktailName: "Virgin Mojito", ProductName: "Soda Water", Quantity: &q150, Unit: "ml", Required: true},
		{CocktailName: "Virgin Mojito", ProductName: "Ice", Quantity: &q8, Unit: "pc", Required: false},

		// Lime Soda
		{CocktailName: "Lime Soda", ProductName: "Soda Water", Quantity: &q150, Unit: "ml", Required: true},
		{CocktailName: "Lime Soda", ProductName: "Lime", Quantity: &q1, Unit: "pc", Required: true},

		// Orange Spritzer
		{CocktailName: "Orange Spritzer", ProductName: "Orange Juice", Quantity: &q120, Unit: "ml", Required: true},
		{CocktailName: "Orange Spritzer", ProductName: "Soda Water", Quantity: &q60, Unit: "ml", Required: true},
		{CocktailName: "Orange Spritzer", ProductName: "Ice", Quantity: &q8, Unit: "pc", Required: false},

		// Cranberry Fizz
		{CocktailName: "Cranberry Fizz", ProductName: "Cranberry Juice", Quantity: &q120, Unit: "ml", Required: true},
		{CocktailName: "Cranberry Fizz", ProductName: "Soda Water", Quantity: &q60, Unit: "ml", Required: true},
		{CocktailName: "Cranberry Fizz", ProductName: "Ice", Quantity: &q8, Unit: "pc", Required: false},

		// Non-Alcoholic Beer
		{CocktailName: "Non-Alcoholic Beer", ProductName: "Non-Alcoholic Beer", Quantity: nil, Unit: "bottle", Required: true},
	}

	// Upsert products
	prodIDs := map[string]int64{}
	for _, p := range products {
		id, err := upsertProduct(tx, p)
		if err != nil {
			return fmt.Errorf("seed product %q: %w", p.Name, err)
		}
		prodIDs[p.Name] = id
	}

	// Upsert cocktails
	cockIDs := map[string]int64{}
	for _, c := range cocktails {
		id, err := upsertCocktail(tx, c)
		if err != nil {
			return fmt.Errorf("seed cocktail %q: %w", c.Name, err)
		}
		cockIDs[c.Name] = id
	}

	// Insert ingredients if cocktail has none yet (preserve bartender edits)
	for _, c := range cocktails {
		cid := cockIDs[c.Name]
		has, err := cocktailHasIngredients(tx, cid)
		if err != nil {
			return fmt.Errorf("check ingredients %q: %w", c.Name, err)
		}
		if has {
			continue
		}
		for _, it := range ings {
			if it.CocktailName != c.Name {
				continue
			}
			pid := prodIDs[it.ProductName]
			if _, err := tx.Exec(`
				INSERT INTO cocktail_ingredients(cocktail_id,product_id,quantity,unit,required)
				VALUES(?,?,?,?,?)`,
				cid, pid, it.Quantity, it.Unit, b2i(it.Required),
			); err != nil {
				return fmt.Errorf("insert ingredient %q -> %q: %w", it.CocktailName, it.ProductName, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

func upsertProduct(tx *sql.Tx, p seedProduct) (int64, error) {
	// Conditional insert (valid in SQLite)
	_, err := tx.Exec(`
		INSERT INTO products(name,category,abv_percent,allergen_flags,notes,is_available,stock_count,created_at,updated_at)
		SELECT ?,?,?,?,?,?,?,?,?
		WHERE NOT EXISTS (SELECT 1 FROM products WHERE name=?);`,
		p.Name, p.Category, p.ABVPercent, p.AllergenFlags, p.Notes, b2i(p.IsAvailable), p.StockCount, unixNow(), unixNow(),
		p.Name,
	)
	if err != nil {
		return 0, err
	}

	// Always update metadata (idempotent, keeps row current)
	_, err = tx.Exec(`
		UPDATE products
		SET category=?, abv_percent=?, allergen_flags=?, notes=?, is_available=?, stock_count=?, updated_at=?
		WHERE name=?;`,
		p.Category, p.ABVPercent, p.AllergenFlags, p.Notes, b2i(p.IsAvailable), p.StockCount, unixNow(),
		p.Name,
	)
	if err != nil {
		return 0, err
	}

	var id int64
	if err := tx.QueryRow(`SELECT id FROM products WHERE name=?;`, p.Name).Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

func upsertCocktail(tx *sql.Tx, c seedCocktail) (int64, error) {
	_, err := tx.Exec(`
		INSERT INTO cocktails(name,description,image_path,tags,difficulty,prep_time_minutes,instructions,is_enabled,created_at,updated_at)
		SELECT ?,?,?,?,?,?,?,?, ?,?
		WHERE NOT EXISTS (SELECT 1 FROM cocktails WHERE name=?);`,
		c.Name, c.Description, c.ImagePath, c.Tags, c.Difficulty, c.PrepTimeMinutes, c.Instructions, b2i(c.IsEnabled),
		unixNow(), unixNow(),
		c.Name,
	)
	if err != nil {
		return 0, err
	}

	_, err = tx.Exec(`
		UPDATE cocktails
		SET description=?, image_path=?, tags=?, difficulty=?, prep_time_minutes=?, instructions=?, is_enabled=?, updated_at=?
		WHERE name=?;`,
		c.Description, c.ImagePath, c.Tags, c.Difficulty, c.PrepTimeMinutes, c.Instructions, b2i(c.IsEnabled), unixNow(),
		c.Name,
	)
	if err != nil {
		return 0, err
	}

	var id int64
	if err := tx.QueryRow(`SELECT id FROM cocktails WHERE name=?;`, c.Name).Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

func cocktailHasIngredients(tx *sql.Tx, cocktailID int64) (bool, error) {
	var n int
	if err := tx.QueryRow(`SELECT COUNT(1) FROM cocktail_ingredients WHERE cocktail_id=?;`, cocktailID).Scan(&n); err != nil {
		return false, err
	}
	return n > 0, nil
}

func f64(v float64) float64 { return v }
