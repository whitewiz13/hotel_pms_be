package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/hotelpms/backend/cache"
	"github.com/hotelpms/backend/dto"
	"github.com/hotelpms/backend/models"
	"github.com/hotelpms/backend/repository"
	"gorm.io/gorm"
)

type PlanService struct {
	db        *gorm.DB
	planRepo  *repository.PlanRepository
	hotelRepo *repository.HotelRepository
	cache     *cache.Cache
}

func NewPlanService(db *gorm.DB, planRepo *repository.PlanRepository, hotelRepo *repository.HotelRepository, cache *cache.Cache) *PlanService {
	return &PlanService{db: db, planRepo: planRepo, hotelRepo: hotelRepo, cache: cache}
}

func (s *PlanService) GetAllPlans() ([]models.Plan, error) {
	return s.planRepo.FindAll()
}

func (s *PlanService) GetHotelSubscription(hotelID string) (*models.Subscription, error) {
	sub, err := s.planRepo.FindSubscriptionByHotelID(hotelID)
	if err != nil {
		return nil, errors.New("subscription not found")
	}
	return sub, nil
}

func (s *PlanService) GetHotelSubscriptionDetails(hotelID string) (*dto.SubscriptionResponse, error) {
	hotel, err := s.hotelRepo.FindByID(hotelID)
	if err != nil {
		return nil, errors.New("hotel not found")
	}

	sub, err := s.planRepo.FindSubscriptionByHotelID(hotelID)
	if err != nil {
		return nil, errors.New("subscription not found")
	}

	return s.buildSubscriptionResponse(sub, hotel), nil
}

func (s *PlanService) GetHotelPlan(hotelID string) (*models.Plan, error) {
	cacheKey := "plan:" + hotelID
	if cached, ok := s.cache.Get(cacheKey); ok {
		plan := cached.(models.Plan)
		return &plan, nil
	}

	sub, err := s.planRepo.FindSubscriptionByHotelID(hotelID)
	if err != nil {
		// Default to free plan if no subscription found
		plan, err := s.planRepo.FindByID(models.PlanFree)
		if err == nil {
			s.cache.Set(cacheKey, *plan)
		}
		return plan, err
	}

	s.cache.Set(cacheKey, sub.Plan)
	return &sub.Plan, nil
}

func (s *PlanService) ChangeHotelPlan(hotelID, planID string) (*models.Subscription, error) {
	s.cache.Delete("plan:" + hotelID)

	status := models.SubscriptionStatusActive
	_, err := s.UpdateHotelSubscription(hotelID, dto.UpdateHotelSubscriptionRequest{
		PlanID: &planID,
		Status: &status,
	})
	if err != nil {
		return nil, err
	}

	return s.planRepo.FindSubscriptionByHotelID(hotelID)
}

func (s *PlanService) UpdateHotelSubscription(hotelID string, req dto.UpdateHotelSubscriptionRequest) (*dto.SubscriptionResponse, error) {
	s.cache.Delete("plan:" + hotelID)

	if !req.HasChanges() {
		return nil, errors.New("at least one subscription field must be provided")
	}

	hotel, err := s.hotelRepo.FindByID(hotelID)
	if err != nil {
		return nil, errors.New("hotel not found")
	}

	var response *dto.SubscriptionResponse
	err = s.db.Transaction(func(tx *gorm.DB) error {
		planRepo := repository.NewPlanRepository(tx)
		hotelRepo := repository.NewHotelRepository(tx)

		sub, err := planRepo.FindSubscriptionByHotelID(hotelID)
		isNew := false
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("failed to load subscription")
			}
			isNew = true
			sub = &models.Subscription{
				HotelID: hotelID,
				Status:  models.SubscriptionStatusActive,
			}
		}

		if isNew && req.PlanID == nil {
			return errors.New("plan_id is required when creating a subscription")
		}

		if sub.AssignedAt == nil && !sub.CreatedAt.IsZero() {
			assignedAt := sub.CreatedAt.UTC()
			sub.AssignedAt = &assignedAt
		}

		planChanged := false
		if req.PlanID != nil {
			planID := strings.TrimSpace(*req.PlanID)
			if planID == "" {
				return errors.New("plan_id cannot be empty")
			}
			if _, err := planRepo.FindByID(planID); err != nil {
				return errors.New("invalid plan")
			}
			planChanged = sub.PlanID != planID
			sub.PlanID = planID
		}

		if sub.PlanID == "" {
			return errors.New("plan_id is required")
		}

		if req.AssignedAt != nil {
			sub.AssignedAt = req.AssignedAt
		} else if isNew || planChanged {
			now := time.Now().UTC()
			sub.AssignedAt = &now
		}

		if req.Status != nil {
			status := strings.TrimSpace(*req.Status)
			if status == "" {
				return errors.New("status cannot be empty")
			}
			sub.Status = status
		} else if sub.Status == "" {
			sub.Status = models.SubscriptionStatusActive
		}

		if req.RenewedAt != nil {
			sub.RenewedAt = req.RenewedAt
		}
		if req.AccessUntil != nil {
			sub.ExpiresAt = req.AccessUntil
			if req.RenewedAt == nil && sub.Status == models.SubscriptionStatusActive {
				renewedAt := time.Now().UTC()
				sub.RenewedAt = &renewedAt
			}
		}
		if req.SuspensionReason != nil {
			reason := strings.TrimSpace(*req.SuspensionReason)
			if reason == "" {
				sub.SuspensionReason = nil
			} else {
				sub.SuspensionReason = &reason
			}
		}
		if req.SuspendedAt != nil {
			sub.SuspendedAt = req.SuspendedAt
		}

		if req.Status != nil {
			switch sub.Status {
			case models.SubscriptionStatusSuspended:
				if req.SuspendedAt == nil {
					now := time.Now().UTC()
					sub.SuspendedAt = &now
				}
			case models.SubscriptionStatusActive, models.SubscriptionStatusPastDue:
				if req.SuspendedAt == nil {
					sub.SuspendedAt = nil
				}
				if req.SuspensionReason == nil {
					sub.SuspensionReason = nil
				}
			}
		}

		if isNew {
			if err := planRepo.CreateSubscription(sub); err != nil {
				return errors.New("failed to create subscription")
			}
		} else {
			if err := planRepo.UpdateSubscription(sub); err != nil {
				return errors.New("failed to update subscription")
			}
		}

		if req.HotelIsActive != nil {
			hotel.IsActive = *req.HotelIsActive
			if err := hotelRepo.Update(hotel); err != nil {
				return errors.New("failed to update hotel access")
			}
		}

		sub, err = planRepo.FindSubscriptionByHotelID(hotelID)
		if err != nil {
			return errors.New("failed to load updated subscription")
		}

		response = s.buildSubscriptionResponse(sub, hotel)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return response, nil
}

// --- Limit checking helpers ---

func (s *PlanService) CheckRoomLimit(hotelID string) error {
	plan, err := s.GetHotelPlan(hotelID)
	if err != nil {
		return err
	}
	if plan.MaxRooms == -1 {
		return nil
	}
	count, err := s.planRepo.CountRoomsByHotel(hotelID)
	if err != nil {
		return err
	}
	if count >= int64(plan.MaxRooms) {
		return fmt.Errorf("room limit reached (%d). Upgrade your plan to add more rooms", plan.MaxRooms)
	}
	return nil
}

func (s *PlanService) CheckStaffLimit(hotelID string) error {
	plan, err := s.GetHotelPlan(hotelID)
	if err != nil {
		return err
	}
	if plan.MaxStaff == -1 {
		return nil
	}
	count, err := s.planRepo.CountStaffByHotel(hotelID)
	if err != nil {
		return err
	}
	if count >= int64(plan.MaxStaff) {
		return fmt.Errorf("staff limit reached (%d). Upgrade your plan to add more staff", plan.MaxStaff)
	}
	return nil
}

func (s *PlanService) CheckReservationLimit(hotelID string) error {
	plan, err := s.GetHotelPlan(hotelID)
	if err != nil {
		return err
	}
	if plan.MaxReservationsMonth == -1 {
		return nil
	}
	count, err := s.planRepo.CountReservationsThisMonth(hotelID)
	if err != nil {
		return err
	}
	if count >= int64(plan.MaxReservationsMonth) {
		return fmt.Errorf("monthly reservation limit reached (%d). Upgrade your plan for more reservations", plan.MaxReservationsMonth)
	}
	return nil
}

func (s *PlanService) HasFeature(hotelID, feature string) (bool, error) {
	plan, err := s.GetHotelPlan(hotelID)
	if err != nil {
		return false, err
	}
	switch feature {
	case "room_service":
		return plan.FeatureRoomService, nil
	case "activities":
		return plan.FeatureActivities, nil
	case "guest_portal":
		return plan.FeatureGuestPortal, nil
	case "notifications":
		return plan.FeatureNotifications, nil
	case "analytics":
		return plan.FeatureAnalytics, nil
	case "adv_analytics":
		return plan.FeatureAdvAnalytics, nil
	case "custom_roles":
		return plan.FeatureCustomRoles, nil
	case "guest_uploads":
		return plan.FeatureGuestUploads, nil
	default:
		return false, fmt.Errorf("unknown feature: %s", feature)
	}
}

// PlanUsage holds current usage counts and plan limits for a hotel.
type PlanUsage struct {
	Plan   models.Plan `json:"plan"`
	Usage  UsageCounts `json:"usage"`
	Limits UsageLimits `json:"limits"`
}

type UsageCounts struct {
	Rooms             int64 `json:"rooms"`
	Staff             int64 `json:"staff"`
	ReservationsMonth int64 `json:"reservations_month"`
}

type UsageLimits struct {
	Rooms             int `json:"rooms"`
	Staff             int `json:"staff"`
	ReservationsMonth int `json:"reservations_month"`
	StorageMB         int `json:"storage_mb"`
}

func (s *PlanService) GetUsage(hotelID string) (*PlanUsage, error) {
	sub, err := s.planRepo.FindSubscriptionByHotelID(hotelID)
	if err != nil {
		return nil, errors.New("subscription not found")
	}

	rooms, _ := s.planRepo.CountRoomsByHotel(hotelID)
	staff, _ := s.planRepo.CountStaffByHotel(hotelID)
	reservations, _ := s.planRepo.CountReservationsThisMonth(hotelID)

	return &PlanUsage{
		Plan: sub.Plan,
		Usage: UsageCounts{
			Rooms:             rooms,
			Staff:             staff,
			ReservationsMonth: reservations,
		},
		Limits: UsageLimits{
			Rooms:             sub.Plan.MaxRooms,
			Staff:             sub.Plan.MaxStaff,
			ReservationsMonth: sub.Plan.MaxReservationsMonth,
			StorageMB:         sub.Plan.MaxStorageMB,
		},
	}, nil
}

func (s *PlanService) buildSubscriptionResponse(sub *models.Subscription, hotel *models.Hotel) *dto.SubscriptionResponse {
	daysLeft, isOverdue, overdueDays := calculateAccessTiming(sub.ExpiresAt)
	canAccess := hotel.IsActive && sub.Status != models.SubscriptionStatusSuspended && sub.Status != models.SubscriptionStatusCancelled

	return &dto.SubscriptionResponse{
		ID:               sub.ID.String(),
		CreatedAt:        sub.CreatedAt,
		UpdatedAt:        sub.UpdatedAt,
		HotelID:          sub.HotelID,
		PlanID:           sub.PlanID,
		Plan:             sub.Plan,
		Status:           sub.Status,
		AssignedAt:       sub.AssignedAt,
		RenewedAt:        sub.RenewedAt,
		AccessUntil:      sub.ExpiresAt,
		ExpiresAt:        sub.ExpiresAt,
		SuspendedAt:      sub.SuspendedAt,
		SuspensionReason: sub.SuspensionReason,
		DaysLeft:         daysLeft,
		IsOverdue:        isOverdue,
		OverdueDays:      overdueDays,
		HotelIsActive:    hotel.IsActive,
		CanAccess:        canAccess,
	}
}

func calculateAccessTiming(accessUntil *time.Time) (*int, bool, int) {
	if accessUntil == nil {
		return nil, false, 0
	}

	today := time.Now().UTC()
	todayDate := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.UTC)
	accessDate := time.Date(accessUntil.UTC().Year(), accessUntil.UTC().Month(), accessUntil.UTC().Day(), 0, 0, 0, 0, time.UTC)
	diffDays := int(accessDate.Sub(todayDate).Hours() / 24)
	if diffDays >= 0 {
		return &diffDays, false, 0
	}

	zero := 0
	return &zero, true, -diffDays
}
