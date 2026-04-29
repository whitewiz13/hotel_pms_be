package repository

import (
	"github.com/hotelpms/backend/models"
	"gorm.io/gorm"
)

type FCMTokenRepository struct {
	db *gorm.DB
}

func NewFCMTokenRepository(db *gorm.DB) *FCMTokenRepository {
	return &FCMTokenRepository{db: db}
}

// Upsert saves a device token. If the same token already exists for the user, it updates the timestamp.
func (r *FCMTokenRepository) Upsert(token *models.FCMToken) error {
	var existing models.FCMToken
	err := r.db.Where("user_id = ? AND device_token = ?", token.UserID, token.DeviceToken).First(&existing).Error
	if err == nil {
		// Already exists — touch updated_at
		return r.db.Model(&existing).Update("updated_at", token.UpdatedAt).Error
	}
	return r.db.Create(token).Error
}

// FindByUserID returns all FCM tokens for a user.
func (r *FCMTokenRepository) FindByUserID(userID string) ([]models.FCMToken, error) {
	var tokens []models.FCMToken
	err := r.db.Where("user_id = ?", userID).Find(&tokens).Error
	return tokens, err
}

// FindTokensByHotelAndPermission returns device tokens for all hotel staff
// whose role includes at least one of the given permission codes.
// Hotel admins are always included (they bypass permission checks).
func (r *FCMTokenRepository) FindTokensByHotelAndPermission(hotelID string, permissions ...string) ([]string, error) {
	var tokens []string

	// Hotel admins always get notified
	adminQuery := r.db.Model(&models.FCMToken{}).
		Select("fcm_tokens.device_token").
		Joins("JOIN users ON users.id = fcm_tokens.user_id").
		Where("fcm_tokens.hotel_id = ? AND users.role = ? AND users.deleted_at IS NULL",
			hotelID, models.RoleHotelAdmin)

	var adminTokens []string
	if err := adminQuery.Pluck("fcm_tokens.device_token", &adminTokens).Error; err != nil {
		return nil, err
	}
	tokens = append(tokens, adminTokens...)

	// Staff with matching permissions via their assigned role
	if len(permissions) > 0 {
		staffQuery := r.db.Model(&models.FCMToken{}).
			Select("DISTINCT fcm_tokens.device_token").
			Joins("JOIN users ON users.id = fcm_tokens.user_id").
			Joins("JOIN roles ON roles.id = users.role_id").
			Joins("JOIN role_permissions ON role_permissions.role_id = roles.id").
			Joins("JOIN permissions ON permissions.id = role_permissions.permission_id").
			Where("fcm_tokens.hotel_id = ? AND permissions.code IN ? AND users.deleted_at IS NULL",
				hotelID, permissions)

		var staffTokens []string
		if err := staffQuery.Pluck("fcm_tokens.device_token", &staffTokens).Error; err != nil {
			return nil, err
		}
		tokens = append(tokens, staffTokens...)
	}

	// Deduplicate
	seen := make(map[string]bool, len(tokens))
	unique := make([]string, 0, len(tokens))
	for _, t := range tokens {
		if !seen[t] {
			seen[t] = true
			unique = append(unique, t)
		}
	}

	return unique, nil
}

// DeleteByToken removes a specific device token (e.g. on logout).
func (r *FCMTokenRepository) DeleteByToken(userID, deviceToken string) error {
	return r.db.Where("user_id = ? AND device_token = ?", userID, deviceToken).
		Delete(&models.FCMToken{}).Error
}
