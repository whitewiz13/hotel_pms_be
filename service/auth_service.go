package service

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/hotelpms/backend/cache"
	"github.com/hotelpms/backend/config"
	"github.com/hotelpms/backend/dto"
	"github.com/hotelpms/backend/models"
	"github.com/hotelpms/backend/repository"
	"github.com/hotelpms/backend/utils"
)

type AuthService struct {
	userRepo    *repository.UserRepository
	roomRepo    *repository.RoomRepository
	hotelRepo   *repository.HotelRepository
	roleRepo    *repository.RoleRepository
	planService *PlanService
	jwtCfg      config.JWTConfig
	cache       *cache.Cache
}

type JWTClaims struct {
	UserID      string          `json:"user_id,omitempty"`
	Email       string          `json:"email,omitempty"`
	Role        models.UserRole `json:"role"`
	RoleID      string          `json:"role_id,omitempty"`
	HotelID     string          `json:"hotel_id,omitempty"`
	RoomID      string          `json:"room_id,omitempty"`
	RoomNumber  string          `json:"room_number,omitempty"`
	IsGuest     bool            `json:"is_guest"`
	Permissions []string        `json:"permissions,omitempty"`
	jwt.RegisteredClaims
}

func NewAuthService(userRepo *repository.UserRepository, roomRepo *repository.RoomRepository, hotelRepo *repository.HotelRepository, roleRepo *repository.RoleRepository, planService *PlanService, jwtCfg config.JWTConfig, cache *cache.Cache) *AuthService {
	return &AuthService{
		userRepo:    userRepo,
		roomRepo:    roomRepo,
		hotelRepo:   hotelRepo,
		roleRepo:    roleRepo,
		planService: planService,
		jwtCfg:      jwtCfg,
		cache:       cache,
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
		if err := s.CheckHotelAccess(*user.HotelID); err != nil {
			return "", nil, err
		}
	}

	token, err := s.generateToken(user)
	if err != nil {
		return "", nil, errors.New("failed to generate token")
	}

	return token, user, nil
}

func (s *AuthService) GuestLogin(req dto.GuestLoginRequest) (string, *models.Room, error) {
	if err := s.CheckHotelAccess(req.HotelID); err != nil {
		return "", nil, err
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

func (s *AuthService) CheckHotelAccess(hotelID string) error {
	cacheKey := "hotel_access:" + hotelID
	if cached, ok := s.cache.Get(cacheKey); ok {
		if cached == nil {
			return nil
		}
		return cached.(error)
	}

	hotel, err := s.hotelRepo.FindByID(hotelID)
	if err != nil {
		accessErr := errors.New("hotel not found")
		s.cache.Set(cacheKey, accessErr)
		return accessErr
	}
	if !hotel.IsActive {
		accessErr := errors.New("hotel is currently disabled, please contact support")
		s.cache.Set(cacheKey, accessErr)
		return accessErr
	}

	sub, err := s.planService.GetHotelSubscription(hotelID)
	if err != nil {
		// No subscription — allow access
		s.cache.Set(cacheKey, nil)
		return nil
	}

	if sub.Status == models.SubscriptionStatusSuspended || sub.Status == models.SubscriptionStatusCancelled {
		accessErr := errors.New("hotel access is currently suspended, please contact support")
		s.cache.Set(cacheKey, accessErr)
		return accessErr
	}

	s.cache.Set(cacheKey, nil)
	return nil
}

func (s *AuthService) CreateStaff(hotelID string, req dto.CreateStaffRequest) (*models.User, error) {
	// Check plan staff limit
	if err := s.planService.CheckStaffLimit(hotelID); err != nil {
		return nil, err
	}

	// Validate role exists and belongs to this hotel
	role, err := s.roleRepo.FindByID(req.RoleID)
	if err != nil {
		return nil, errors.New("invalid role")
	}
	if role.HotelID == nil || *role.HotelID != hotelID {
		return nil, errors.New("role does not belong to this hotel")
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
		Role:         models.UserRole(role.Slug),
		RoleID:       &req.RoleID,
		HotelID:      &hotelID,
		IsActive:     true,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, errors.New("failed to create staff")
	}

	return user, nil
}

func (s *AuthService) GetMe(userID string, roleID string) (*dto.MeResponse, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	hotelID := ""
	if user.HotelID != nil {
		hotelID = *user.HotelID
	}

	var permissions []string
	if roleID != "" {
		codes, err := s.roleRepo.GetPermissionCodes(roleID)
		if err == nil {
			permissions = codes
		}
	}

	if permissions == nil {
		permissions = []string{}
	}

	return &dto.MeResponse{
		UserID:      user.ID.String(),
		Email:       user.Email,
		Name:        user.Name,
		Role:        string(user.Role),
		RoleID:      roleID,
		HotelID:     hotelID,
		Permissions: permissions,
	}, nil
}

func (s *AuthService) generateToken(user *models.User) (string, error) {
	hotelID := ""
	if user.HotelID != nil {
		hotelID = *user.HotelID
	}

	var permissions []string
	roleID := ""
	if user.RoleID != nil {
		roleID = *user.RoleID
	}

	claims := JWTClaims{
		UserID:      user.ID.String(),
		Email:       user.Email,
		Role:        user.Role,
		RoleID:      roleID,
		HotelID:     hotelID,
		IsGuest:     false,
		Permissions: permissions,
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
