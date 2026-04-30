package service

import (
	"errors"
	"fmt"

	"github.com/hotelpms/backend/models"
	"github.com/hotelpms/backend/repository"
)

type PlanService struct {
	planRepo *repository.PlanRepository
}

func NewPlanService(planRepo *repository.PlanRepository) *PlanService {
	return &PlanService{planRepo: planRepo}
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

func (s *PlanService) GetHotelPlan(hotelID string) (*models.Plan, error) {
	sub, err := s.planRepo.FindSubscriptionByHotelID(hotelID)
	if err != nil {
		// Default to free plan if no subscription found
		return s.planRepo.FindByID(models.PlanFree)
	}
	return &sub.Plan, nil
}

func (s *PlanService) ChangeHotelPlan(hotelID, planID string) (*models.Subscription, error) {
	// Validate plan exists
	_, err := s.planRepo.FindByID(planID)
	if err != nil {
		return nil, errors.New("invalid plan")
	}

	sub, err := s.planRepo.FindSubscriptionByHotelID(hotelID)
	if err != nil {
		// Create new subscription
		sub = &models.Subscription{
			HotelID: hotelID,
			PlanID:  planID,
			Status:  models.SubscriptionStatusActive,
		}
		if err := s.planRepo.CreateSubscription(sub); err != nil {
			return nil, errors.New("failed to create subscription")
		}
	} else {
		sub.PlanID = planID
		sub.Status = models.SubscriptionStatusActive
		if err := s.planRepo.UpdateSubscription(sub); err != nil {
			return nil, errors.New("failed to update subscription")
		}
	}

	// Re-fetch with preloaded plan
	return s.planRepo.FindSubscriptionByHotelID(hotelID)
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
