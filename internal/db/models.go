package db

import "time"

type User struct {
	ID           int64
	Email        string
	PasswordHash string
	Role         string
	DisplayName  string
	IsActive     bool
	OnDuty       bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type Product struct {
	ID            int64
	Name          string
	Category      string
	ABVPercent    *float64
	AllergenFlags string
	Notes         string
	IsAvailable   bool
	StockCount    *int64
	ComputedAvail bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type Cocktail struct {
	ID              int64
	Name            string
	Description     string
	ImagePath       string
	Tags            string
	Difficulty      string
	PrepTimeMinutes int64
	Instructions    string
	IsEnabled       bool
	ComputedAvail   bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type CocktailIngredient struct {
	ID         int64
	CocktailID int64
	ProductID  int64
	Quantity   *float64
	Unit       string
	Required   bool

	ProductName     string
	ProductCategory string
	ProductAvail    bool
}

type IngredientUpsertItem struct {
	ProductID int64
	Quantity  *float64
	Unit      string
	Required  bool
}

type Order struct {
	ID                 int64
	UserID             int64
	CocktailID         int64
	Quantity           int64
	Notes              string
	Location           string
	Status             string
	AssignedBartenderID *int64
	CreatedAt          time.Time
	UpdatedAt          time.Time

	UserDisplayName       string
	CocktailName          string
	CocktailImagePath     string
	AssignedBartenderName string
}

type OrderEvent struct {
	ID              int64
	OrderID         int64
	FromStatus      string
	ToStatus        string
	ChangedByUserID *int64
	ChangedByName   string
	CreatedAt       time.Time
}

/* ---------- parameter structs ---------- */

type CreateUserParams struct {
	Email        string
	PasswordHash string
	Role         string
	DisplayName  string
	IsActive     bool
	OnDuty       bool
}

type UpdateUserParams struct {
	ID          int64
	Email       string
	Role        string
	DisplayName string
}

type CreateProductParams struct {
	Name          string
	Category      string
	ABVPercent    *float64
	AllergenFlags string
	Notes         string
	IsAvailable   bool
	StockCount    *int64
}

type UpdateProductParams struct {
	ID            int64
	Name          string
	Category      string
	ABVPercent    *float64
	AllergenFlags string
	Notes         string
	IsAvailable   bool
	StockCount    *int64
}

type CreateCocktailParams struct {
	Name            string
	Description     string
	ImagePath       string
	Tags            string
	Difficulty      string
	PrepTimeMinutes int64
	Instructions    string
	IsEnabled       bool
}

type UpdateCocktailParams struct {
	ID              int64
	Name            string
	Description     string
	ImagePath       string
	Tags            string
	Difficulty      string
	PrepTimeMinutes int64
	Instructions    string
	IsEnabled       bool
}

type CreateOrderParams struct {
	UserID     int64
	CocktailID int64
	Quantity   int64
	Notes      string
	Location   string
}
