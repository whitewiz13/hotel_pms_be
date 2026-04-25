package models

type MenuCategory string

const (
	MenuCategoryAppetizer MenuCategory = "appetizer"
	MenuCategoryMainCourse MenuCategory = "main_course"
	MenuCategoryDessert   MenuCategory = "dessert"
	MenuCategoryBeverage  MenuCategory = "beverage"
	MenuCategorySnack     MenuCategory = "snack"
)

type MenuItem struct {
	BaseModel
	HotelID     string       `gorm:"type:uuid;not null;index" json:"hotel_id"`
	Name        string       `gorm:"not null;size:255" json:"name"`
	Description string       `gorm:"type:text" json:"description"`
	Price       float64      `gorm:"not null;default:0" json:"price"`
	Category    MenuCategory `gorm:"type:varchar(30);not null;default:'main_course'" json:"category"`
	IsAvailable bool         `gorm:"default:true" json:"is_available"`

	Hotel Hotel `gorm:"foreignKey:HotelID" json:"-"`
}

func (MenuItem) TableName() string {
	return "menu_items"
}
