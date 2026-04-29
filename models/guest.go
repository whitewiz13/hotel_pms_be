package models

type Guest struct {
	BaseModel
	HotelID       string `gorm:"type:uuid;not null;index" json:"hotel_id"`
	Name          string `gorm:"not null;size:255" json:"name"`
	Email         string `gorm:"size:255" json:"email"`
	Phone         string `gorm:"size:20" json:"phone"`
	IDType        string `gorm:"size:30" json:"id_type"`         // aadhaar, pan, passport, driving_license, voter_id
	IDNumber      string `gorm:"size:50" json:"id_number"`       // masked/stored document number
	IDDocumentURL string `gorm:"size:500" json:"id_document_url"` // Firebase Storage URL

	Hotel Hotel `gorm:"foreignKey:HotelID" json:"-"`
}

func (Guest) TableName() string {
	return "guests"
}
