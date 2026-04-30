package service

import (
	"errors"
	"fmt"

	"github.com/hotelpms/backend/dto"
	"github.com/hotelpms/backend/models"
	"github.com/hotelpms/backend/repository"
	"gorm.io/gorm"
)

type OrderService struct {
	db                  *gorm.DB
	orderRepo           *repository.OrderRepository
	menuRepo            *repository.MenuRepository
	roomRepo            *repository.RoomRepository
	reservationRepo     *repository.ReservationRepository
	userRepo            *repository.UserRepository
	notificationService *NotificationService
}

func NewOrderService(
	db *gorm.DB,
	orderRepo *repository.OrderRepository,
	menuRepo *repository.MenuRepository,
	roomRepo *repository.RoomRepository,
	reservationRepo *repository.ReservationRepository,
	userRepo *repository.UserRepository,
	notificationService *NotificationService,
) *OrderService {
	return &OrderService{
		db:                  db,
		orderRepo:           orderRepo,
		menuRepo:            menuRepo,
		roomRepo:            roomRepo,
		reservationRepo:     reservationRepo,
		userRepo:            userRepo,
		notificationService: notificationService,
	}
}

func (s *OrderService) Create(hotelID string, req dto.CreateOrderRequest) (*models.Order, error) {
	// Verify room belongs to hotel
	room, err := s.roomRepo.FindByID(req.RoomID)
	if err != nil || room.HotelID != hotelID {
		return nil, errors.New("room not found")
	}

	// Verify reservation exists and is checked in
	reservation, err := s.reservationRepo.FindByIDAndHotel(req.ReservationID, hotelID)
	if err != nil {
		return nil, errors.New("reservation not found")
	}
	if reservation.Status != models.ReservationStatusCheckedIn {
		return nil, errors.New("orders can only be placed for checked-in reservations")
	}
	if reservation.RoomID != req.RoomID {
		return nil, errors.New("reservation does not match the specified room")
	}

	// Build order items and calculate total
	var totalAmount float64
	var orderItems []models.OrderItem

	for _, itemReq := range req.Items {
		menuItem, err := s.menuRepo.FindByIDAndHotel(itemReq.MenuItemID, hotelID)
		if err != nil {
			return nil, errors.New("menu item not found: " + itemReq.MenuItemID)
		}
		if !menuItem.IsAvailable {
			return nil, errors.New("menu item not available: " + menuItem.Name)
		}

		subtotal := menuItem.Price * float64(itemReq.Quantity)
		totalAmount += subtotal

		orderItems = append(orderItems, models.OrderItem{
			MenuItemID: itemReq.MenuItemID,
			Quantity:   itemReq.Quantity,
			UnitPrice:  menuItem.Price,
			Subtotal:   subtotal,
			Notes:      itemReq.Notes,
		})
	}

	order := &models.Order{
		HotelID:       hotelID,
		RoomID:        req.RoomID,
		ReservationID: req.ReservationID,
		GuestID:       reservation.GuestID,
		GuestName:     req.GuestName,
		Status:        models.OrderStatusPending,
		TotalAmount:   totalAmount,
		Notes:         req.Notes,
		Items:         orderItems,
	}

	if req.AssignedToID != "" {
		// Verify the assigned user belongs to the hotel
		assignedUser, err := s.userRepo.FindByIDAndHotel(req.AssignedToID, hotelID)
		if err != nil || assignedUser == nil {
			return nil, errors.New("assigned staff member not found")
		}
		order.AssignedToID = &req.AssignedToID
	}

	if err := s.orderRepo.Create(order); err != nil {
		return nil, errors.New("failed to create order")
	}

	// Notify hotel staff about the new order
	if s.notificationService != nil {
		go s.notificationService.SendToHotelStaff(
			hotelID,
			"New Room Service Order",
			fmt.Sprintf("Order from %s — Room %s", req.GuestName, room.RoomNumber),
			map[string]string{
				"type":     "new_order",
				"order_id": order.ID.String(),
				"hotel_id": hotelID,
			},
			"orders:read", "orders:update_status",
		)

		// Notify the assigned staff member
		if req.AssignedToID != "" {
			go s.notificationService.SendToUser(
				req.AssignedToID,
				"Order Assigned to You",
				fmt.Sprintf("Order for %s — Room %s", req.GuestName, room.RoomNumber),
				map[string]string{
					"type":     "order_assigned",
					"order_id": order.ID.String(),
					"hotel_id": hotelID,
				},
			)
		}
	}

	return s.orderRepo.FindByID(order.ID.String())
}

func (s *OrderService) GetByID(id, hotelID string) (*models.Order, error) {
	order, err := s.orderRepo.FindByIDAndHotel(id, hotelID)
	if err != nil {
		return nil, errors.New("order not found")
	}
	return order, nil
}

func (s *OrderService) List(hotelID string, query dto.ListOrdersQuery) ([]models.Order, int64, error) {
	page := query.Page
	perPage := query.PerPage
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	return s.orderRepo.FindByHotelID(hotelID, query.Status, query.RoomID, query.ReservationID, query.AssignedToID, page, perPage)
}

func (s *OrderService) UpdateStatus(id, hotelID string, req dto.UpdateOrderStatusRequest) (*models.Order, error) {
	order, err := s.orderRepo.FindByIDAndHotel(id, hotelID)
	if err != nil {
		return nil, errors.New("order not found")
	}

	newStatus := models.OrderStatus(req.Status)

	// Validate status transitions
	switch order.Status {
	case models.OrderStatusPending:
		if newStatus != models.OrderStatusPreparing && newStatus != models.OrderStatusCancelled {
			return nil, errors.New("pending orders can only move to preparing or cancelled")
		}
	case models.OrderStatusPreparing:
		if newStatus != models.OrderStatusReady && newStatus != models.OrderStatusCancelled {
			return nil, errors.New("preparing orders can only move to ready or cancelled")
		}
	case models.OrderStatusReady:
		if newStatus != models.OrderStatusDelivered {
			return nil, errors.New("ready orders can only move to delivered")
		}
	default:
		return nil, errors.New("order cannot be updated from current status")
	}

	order.Status = newStatus
	if err := s.orderRepo.Update(order); err != nil {
		return nil, errors.New("failed to update order status")
	}

	// Notify relevant staff about status change
	if s.notificationService != nil {
		title := fmt.Sprintf("Order %s", string(newStatus))
		body := fmt.Sprintf("Order for %s — Room %s is now %s", order.GuestName, order.Room.RoomNumber, string(newStatus))

		if newStatus == models.OrderStatusReady {
			// Notify front desk/admin so they can reassign for delivery
			go s.notificationService.SendToHotelStaff(
				hotelID, title, body,
				map[string]string{
					"type":     "order_ready_for_assignment",
					"order_id": id,
					"status":   string(newStatus),
					"hotel_id": hotelID,
				},
				"orders:assign",
			)
		} else {
			go s.notificationService.SendToHotelStaff(
				hotelID, title, body,
				map[string]string{
					"type":     "order_status_update",
					"order_id": id,
					"status":   string(newStatus),
					"hotel_id": hotelID,
				},
				"orders:read", "orders:update_status",
			)
		}
	}

	return order, nil
}

func (s *OrderService) Assign(id, hotelID string, req dto.AssignOrderRequest) (*models.Order, error) {
	order, err := s.orderRepo.FindByIDAndHotel(id, hotelID)
	if err != nil {
		return nil, errors.New("order not found")
	}

	if order.Status == models.OrderStatusDelivered || order.Status == models.OrderStatusCancelled {
		return nil, errors.New("cannot assign completed or cancelled orders")
	}

	// Verify staff belongs to hotel
	user, err := s.userRepo.FindByID(req.AssignedToID)
	if err != nil {
		return nil, errors.New("staff member not found")
	}
	if user.HotelID == nil || *user.HotelID != hotelID {
		return nil, errors.New("staff member does not belong to this hotel")
	}

	order.AssignedToID = &req.AssignedToID
	if err := s.orderRepo.Update(order); err != nil {
		return nil, errors.New("failed to assign order")
	}

	// Notify the assigned staff member
	if s.notificationService != nil {
		go s.notificationService.SendToUser(
			req.AssignedToID,
			"Order Assigned to You",
			fmt.Sprintf("Order for %s — Room %s", order.GuestName, order.Room.RoomNumber),
			map[string]string{
				"type":     "order_assigned",
				"order_id": id,
				"hotel_id": hotelID,
			},
		)
	}

	return s.orderRepo.FindByIDAndHotel(id, hotelID)
}
