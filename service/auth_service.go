package service

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/hotelpms/backend/config"
	"github.com/hotelpms/backend/dto"
	"github.com/hotelpms/backend/models"
	"github.com/hotelpms/backend/repository"
	"github.com/hotelpms/backend/utils"
)

type AuthService struct {
	userRepo  *repository.UserRepository
	roomRepo  *repository.RoomRepository
	hotelRepo *repository.HotelRepository
	jwtCfg    config.JWTConfig
}

type JWTClaims struct {
	UserID     string          `json:"user_id,omitempty"`
	Email      string          `json:"email,omitempty"`
	Role       models.UserRole `json:"role"`
	HotelID    string          `json:"hotel_id,omitempty"`
	RoomID     string          `json:"room_id,omitempty"`
	RoomNumber string          `json:"room_number,omitempty"`
	IsGuest    bool            `json:"is_guest"`
	jwt.RegisteredClaims
}

func NewAuthService(userRepo *repository.UserRepository, roomRepo *repository.RoomRepository, hotelRepo *repository.HotelRepository, jwtCfg config.JWTConfig) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		roomRepo:  roomRepo,
		hotelRepo: hotelRepo,
		jwtCfg:    jwtCfg,
	}
}

func (s *AuthService) Login(req dto.LoginRequest) (string, *models.User, error) {
	user, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		return "", nil, errors.New("invalid email or password")
	}

	if !user.IsActive {
		return "", nil, errors.New("account is deactivated")
	}

	if !utils.CheckPassword(req.Password, user.PasswordHash) {
		return "", nil, errors.New("invalid email or password")
	}

	// Block login if hotel is disabled (skip for super_admin — no hotel)
	if user.HotelID != nil && user.Role != models.RoleSuperAdmin {
		hotel, err := s.hotelRepo.FindByID(*user.HotelID)
		if err != nil {
			return "", nil, errors.New("hotel not found")
		}
		if !hotel.IsActive {
			return "", nil, errors.New("hotel is currently disabled, please contact support")
		}
	}

	token, err := s.generateToken(user)
	if err != nil {
		return "", nil, errors.New("failed to generate token")
	}

	return token, user, nil
}

func (s *AuthService) GuestLogin(req dto.GuestLoginRequest) (string, *models.Room, error) {
	// Block guest login if hotel is disabled
	hotel, err := s.hotelRepo.FindByID(req.HotelID)
	if err != nil {
		return "", nil, errors.New("hotel not found")
	}
	if !hotel.IsActive {
		return "", nil, errors.New("hotel is currently disabled, please contact support")
	}

	room, err := s.roomRepo.FindByRoomNumberAndHotel(req.RoomNumber, req.HotelID)
	if err != nil {
		return "", nil, errors.New("invalid room number or pin")
	}

	if room.AccessPin == "" || room.AccessPin != req.Pin {
		return "", nil, errors.New("invalid room number or pin")
	}

	if room.Status != models.RoomStatusOccupied {
		return "", nil, errors.New("room is not currently checked in")
	}

	token, err := s.generateGuestToken(room)
	if err != nil {
		return "", nil, errors.New("failed to generate token")
	}

	return token, room, nil
}

func (s *AuthService) CreateStaff(hotelID string, req dto.CreateStaffRequest) (*models.User, error) {
	role := models.UserRole(req.Role)
	if !models.IsValidRole(role) {
		return nil, errors.New("invalid role: " + req.Role)
	}
	if role == models.RoleSuperAdmin || role == models.RoleHotelAdmin {
		return nil, errors.New("cannot create super admin or hotel admin through this endpoint")
	}

	exists, err := s.userRepo.ExistsByEmail(req.Email)
	if err != nil {
		return nil, errors.New("failed to check existing user")
	}
	if exists {
		return nil, errors.New("email already registered")
	}

	hash, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	user := &models.User{
		Email:        req.Email,
		PasswordHash: hash,
		Name:         req.Name,
		Phone:        req.Phone,
		Role:         role,
		HotelID:      &hotelID,
		IsActive:     true,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, errors.New("failed to create staff")
	}

	return user, nil
}

func (s *AuthService) generateToken(user *models.User) (string, error) {
	hotelID := ""
	if user.HotelID != nil {
		hotelID = *user.HotelID
	}

	claims := JWTClaims{
		UserID:  user.ID.String(),
		Email:   user.Email,
		Role:    user.Role,
		HotelID: hotelID,
		IsGuest: false,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(s.jwtCfg.ExpiryHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtCfg.Secret))
}

func (s *AuthService) generateGuestToken(room *models.Room) (string, error) {
	claims := JWTClaims{
		Role:       "guest",
		HotelID:    room.HotelID,
		RoomID:     room.ID.String(),
		RoomNumber: room.RoomNumber,
		IsGuest:    true,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(s.jwtCfg.GuestExpiryHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtCfg.Secret))
}

func (s *AuthService) ValidateToken(tokenStr string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(s.jwtCfg.Secret), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}
