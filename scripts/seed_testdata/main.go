package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/hotelpms/backend/config"
	"github.com/hotelpms/backend/database"
	"github.com/hotelpms/backend/models"
	"github.com/hotelpms/backend/utils"
	"gorm.io/gorm"
)

// ──────────────────────────────────────────────────────────────────────────────
// Test-data seeder — run with:   go run ./scripts/seed_testdata
// Delete the data by dropping/recreating the DB or reverting your migration.
// ──────────────────────────────────────────────────────────────────────────────

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	db, err := database.Connect(&cfg.Database)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	log.Println("=== Seeding massive test data ===")

	// Load all permissions (already seeded by the main seeder)
	allPerms := loadPermissions(db)
	log.Printf("Loaded %d permissions", len(allPerms))

	hotels := seedHotels(db)
	for i, hotel := range hotels {
		log.Printf("── Hotel %d/%d: %s (%s)", i+1, len(hotels), hotel.Name, hotel.ID)

		roles := seedRoles(db, hotel.ID.String(), allPerms)
		roomTypes := seedRoomTypes(db, hotel.ID.String())
		amenities := seedAmenities(db, hotel.ID.String())
		rooms := seedRooms(db, hotel.ID.String(), roomTypes, amenities)
		staff := seedStaff(db, hotel.ID.String(), i, roles)
		guests := seedGuests(db, hotel.ID.String())
		menuItems := seedMenuItems(db, hotel.ID.String())
		activities := seedActivities(db, hotel.ID.String())
		// Last hotel (Sunrise Bay) gets extra today check-ins and check-outs
		if i == len(hotels)-1 {
			seedReservationsAndRelated(db, hotel.ID.String(), rooms, guests, staff, menuItems, activities)
			seedTodayCheckIns(db, hotel.ID.String(), rooms[30:50], guests, 20)
			seedTodayCheckOuts(db, hotel.ID.String(), rooms[20:30], guests[10:20], staff, menuItems, activities)
		} else {
			seedReservationsAndRelated(db, hotel.ID.String(), rooms, guests, staff, menuItems, activities)
		}
		seedGuestSettings(db, hotel.ID.String())

		log.Printf("   Rooms: %d | Staff: %d | Guests: %d | Menu: %d | Activities: %d | Roles: %d",
			len(rooms), len(staff), len(guests), len(menuItems), len(activities), len(roles))
	}

	log.Println("=== Done! ===")
}

// ─── Permissions ─────────────────────────────────────────────────────────────

func loadPermissions(db *gorm.DB) []models.Permission {
	var perms []models.Permission
	if err := db.Find(&perms).Error; err != nil {
		log.Fatalf("Failed to load permissions: %v", err)
	}
	return perms
}

func permsByModule(perms []models.Permission) map[string][]models.Permission {
	m := make(map[string][]models.Permission)
	for _, p := range perms {
		m[p.Module] = append(m[p.Module], p)
	}
	return m
}

func pickPerms(byModule map[string][]models.Permission, modules ...string) []models.Permission {
	var out []models.Permission
	for _, mod := range modules {
		out = append(out, byModule[mod]...)
	}
	return out
}

// ─── Roles ───────────────────────────────────────────────────────────────────

func seedRoles(db *gorm.DB, hotelID string, allPerms []models.Permission) map[string]models.Role {
	byModule := permsByModule(allPerms)

	// Define role → modules mapping
	roleDefs := []struct {
		Name, Slug, Desc string
		Modules          []string
	}{
		{
			"Hotel Admin", "hotel-admin", "Full access to all hotel operations",
			[]string{"dashboard", "rooms", "room_types", "reservations", "amenities",
				"housekeeping", "menu", "orders", "activities", "activity_bookings",
				"billing", "staff", "guest_settings", "roles"},
		},
		{
			"Manager", "manager", "Manages daily operations, staff, and reservations",
			[]string{"dashboard", "rooms", "room_types", "reservations", "amenities",
				"housekeeping", "menu", "orders", "activities", "activity_bookings",
				"billing", "staff", "guest_settings"},
		},
		{
			"Front Desk", "front-desk", "Handles check-in/out and reservations",
			[]string{"dashboard", "rooms", "room_types", "reservations", "amenities",
				"billing", "orders", "activities", "activity_bookings", "guest_settings"},
		},
		{
			"Housekeeping", "housekeeping", "Manages room cleaning tasks",
			[]string{"housekeeping", "rooms"},
		},
		{
			"Room Service", "room-service", "Handles food orders and delivery",
			[]string{"orders", "menu"},
		},
		{
			"Staff", "staff", "Basic staff with limited access",
			[]string{"rooms", "dashboard"},
		},
	}

	roles := make(map[string]models.Role)
	for _, rd := range roleDefs {
		hid := hotelID
		role := models.Role{
			HotelID:     &hid,
			Name:        rd.Name,
			Slug:        rd.Slug,
			Description: rd.Desc,
			IsSystem:    false,
		}
		if err := db.Create(&role).Error; err != nil {
			log.Fatalf("role %s: %v", rd.Name, err)
		}

		// Attach permissions
		perms := pickPerms(byModule, rd.Modules...)
		if len(perms) > 0 {
			if err := db.Model(&role).Association("Permissions").Replace(perms); err != nil {
				log.Fatalf("role perms %s: %v", rd.Name, err)
			}
		}

		roles[rd.Slug] = role
		log.Printf("   Role: %s (%d permissions)", rd.Name, len(perms))
	}
	return roles
}

// ─── Hotels ──────────────────────────────────────────────────────────────────

var hotelDefs = []struct {
	Name, Slug, Address, City, State, Country, Zip, Phone, Email, Desc string
}{
	{"The Grand Palace", "grand-palace", "123 Royal Avenue", "Mumbai", "Maharashtra", "India", "400001", "+91-22-12345678", "info@grandpalace.com", "A luxury 5-star hotel in the heart of Mumbai"},
	{"Ocean Breeze Resort", "ocean-breeze", "456 Coastal Road", "Goa", "Goa", "India", "403001", "+91-832-2345678", "stay@oceanbreeze.com", "Beachside resort with panoramic ocean views"},
	{"Mountain View Lodge", "mountain-view", "789 Hilltop Lane", "Shimla", "Himachal Pradesh", "India", "171001", "+91-177-3456789", "hello@mountainview.com", "Cozy mountain retreat with stunning valley views"},
	{"City Central Hotel", "city-central", "321 Business Park Road", "Bangalore", "Karnataka", "India", "560001", "+91-80-4567890", "book@citycentral.com", "Modern business hotel in the tech capital"},
	{"Sunrise Bay Hotel", "sunrise-bay", "55 Marina Drive", "Chennai", "Tamil Nadu", "India", "600001", "+91-44-5678901", "stay@sunrisebay.com", "Premium waterfront hotel with stunning sunrise views"},
}

func seedHotels(db *gorm.DB) []models.Hotel {
	var hotels []models.Hotel
	for _, h := range hotelDefs {
		hotel := models.Hotel{
			Name: h.Name, Slug: h.Slug, Address: h.Address,
			City: h.City, State: h.State, Country: h.Country,
			ZipCode: h.Zip, Phone: h.Phone, Email: h.Email,
			Description: h.Desc, IsActive: true,
		}
		if err := db.Create(&hotel).Error; err != nil {
			log.Fatalf("Failed to create hotel %s: %v", h.Name, err)
		}
		hotels = append(hotels, hotel)
	}
	return hotels
}

// ─── Room Types ──────────────────────────────────────────────────────────────

func seedRoomTypes(db *gorm.DB, hotelID string) []models.RoomType {
	defs := []struct{ Name, Desc string }{
		{"Standard", "Comfortable standard room"},
		{"Deluxe", "Spacious deluxe room with city view"},
		{"Suite", "Luxury suite with living area"},
		{"Presidential Suite", "Top-floor presidential suite"},
		{"Family Room", "Large room ideal for families"},
		{"Twin Room", "Room with two single beds"},
	}
	var out []models.RoomType
	for _, d := range defs {
		rt := models.RoomType{HotelID: hotelID, Name: d.Name, Description: d.Desc, IsActive: true}
		if err := db.Create(&rt).Error; err != nil {
			log.Fatalf("room type: %v", err)
		}
		out = append(out, rt)
	}
	return out
}

// ─── Amenities ───────────────────────────────────────────────────────────────

func seedAmenities(db *gorm.DB, hotelID string) []models.Amenity {
	defs := []struct {
		Name, Desc, Icon string
		Cat              models.AmenityCategory
	}{
		{"Wi-Fi", "High-speed wireless internet", "wifi", models.AmenityCategoryRoom},
		{"TV", "55-inch Smart TV", "tv", models.AmenityCategoryRoom},
		{"Mini Bar", "Fully stocked mini bar", "wine", models.AmenityCategoryRoom},
		{"Safe", "In-room digital safe", "lock", models.AmenityCategoryRoom},
		{"Air Conditioning", "Central air conditioning", "snowflake", models.AmenityCategoryRoom},
		{"Coffee Maker", "Nespresso coffee machine", "coffee", models.AmenityCategoryRoom},
		{"Iron", "Iron and ironing board", "shirt", models.AmenityCategoryRoom},
		{"Bathrobe", "Luxury cotton bathrobe", "user", models.AmenityCategoryBathroom},
		{"Rainfall Shower", "Oversized rainfall shower head", "droplet", models.AmenityCategoryBathroom},
		{"Jacuzzi", "Private jacuzzi tub", "waves", models.AmenityCategoryBathroom},
		{"Swimming Pool", "Outdoor infinity pool", "pool", models.AmenityCategoryRecreation},
		{"Gym", "24-hour fitness center", "dumbbell", models.AmenityCategoryRecreation},
		{"Spa Access", "Complimentary spa access", "heart", models.AmenityCategoryRecreation},
		{"Restaurant", "Fine dining restaurant", "utensils", models.AmenityCategoryDining},
		{"Room Service", "24-hour room service", "bell", models.AmenityCategoryDining},
		{"Parking", "Complimentary valet parking", "car", models.AmenityCategoryGeneral},
		{"Concierge", "24-hour concierge service", "info", models.AmenityCategoryGeneral},
		{"Laundry", "Same-day laundry service", "shirt", models.AmenityCategoryGeneral},
	}
	var out []models.Amenity
	for _, d := range defs {
		a := models.Amenity{
			HotelID: hotelID, Name: d.Name, Description: d.Desc,
			Icon: d.Icon, Category: d.Cat, IsActive: true,
		}
		if err := db.Create(&a).Error; err != nil {
			log.Fatalf("amenity: %v", err)
		}
		out = append(out, a)
	}
	return out
}

// ─── Rooms (50 per hotel) ────────────────────────────────────────────────────

func seedRooms(db *gorm.DB, hotelID string, roomTypes []models.RoomType, amenities []models.Amenity) []models.Room {
	typeNames := []string{"Standard", "Deluxe", "Suite", "Presidential Suite", "Family Room", "Twin Room"}
	prices := map[string]float64{
		"Standard": 3500, "Deluxe": 6000, "Suite": 12000,
		"Presidential Suite": 25000, "Family Room": 8000, "Twin Room": 4500,
	}
	maxOcc := map[string]int{
		"Standard": 2, "Deluxe": 2, "Suite": 3,
		"Presidential Suite": 4, "Family Room": 5, "Twin Room": 2,
	}

	var rooms []models.Room
	roomNum := 100
	for floor := 1; floor <= 10; floor++ {
		roomsPerFloor := 5
		for r := 0; r < roomsPerFloor; r++ {
			roomNum++
			typeName := typeNames[roomNum%len(typeNames)]
			pin, _ := utils.GeneratePin(6)
			room := models.Room{
				HotelID:       hotelID,
				RoomNumber:    fmt.Sprintf("%d", roomNum),
				RoomType:      typeName,
				Floor:         floor,
				Status:        models.RoomStatusAvailable,
				PricePerNight: prices[typeName] + float64(floor*200),
				Description:   fmt.Sprintf("Floor %d - %s room %d", floor, typeName, roomNum),
				MaxOccupancy:  maxOcc[typeName],
				AccessPin:     pin,
				IsActive:      true,
			}
			if err := db.Create(&room).Error; err != nil {
				log.Fatalf("room: %v", err)
			}
			// Attach 3-6 random amenities
			if len(amenities) > 0 {
				count := 3 + rand.Intn(4)
				if count > len(amenities) {
					count = len(amenities)
				}
				picked := pickRandomAmenities(amenities, count)
				db.Model(&room).Association("Amenities").Replace(picked)
			}
			rooms = append(rooms, room)
		}
	}
	return rooms
}

func pickRandomAmenities(amenities []models.Amenity, n int) []models.Amenity {
	perm := rand.Perm(len(amenities))
	var out []models.Amenity
	for i := 0; i < n && i < len(perm); i++ {
		out = append(out, amenities[perm[i]])
	}
	return out
}

// ─── Staff (15-20 per hotel) ─────────────────────────────────────────────────

func seedStaff(db *gorm.DB, hotelID string, hotelIndex int, roles map[string]models.Role) []models.User {
	hash, _ := utils.HashPassword("staff@123")

	// Map user role → role slug for RoleID assignment
	roleSlugMap := map[string]string{
		string(models.RoleHotelAdmin):  "hotel-admin",
		string(models.RoleManager):     "manager",
		string(models.RoleFrontDesk):   "front-desk",
		string(models.RoleHousekeeping): "housekeeping",
		string(models.RoleRoomService):  "room-service",
		string(models.RoleStaff):        "staff",
	}

	staffDefs := []struct {
		Name, Role string
	}{
		{"Hotel Admin", string(models.RoleHotelAdmin)},
		{"Front Desk Manager", string(models.RoleManager)},
		{"Night Manager", string(models.RoleManager)},
		{"Reception Agent 1", string(models.RoleFrontDesk)},
		{"Reception Agent 2", string(models.RoleFrontDesk)},
		{"Reception Agent 3", string(models.RoleFrontDesk)},
		{"Head Housekeeper", string(models.RoleHousekeeping)},
		{"Housekeeper A", string(models.RoleHousekeeping)},
		{"Housekeeper B", string(models.RoleHousekeeping)},
		{"Housekeeper C", string(models.RoleHousekeeping)},
		{"Housekeeper D", string(models.RoleHousekeeping)},
		{"Room Service Lead", string(models.RoleRoomService)},
		{"Room Service Waiter 1", string(models.RoleRoomService)},
		{"Room Service Waiter 2", string(models.RoleRoomService)},
		{"Room Service Waiter 3", string(models.RoleRoomService)},
		{"Bell Boy", string(models.RoleStaff)},
		{"Concierge", string(models.RoleStaff)},
		{"Maintenance Staff", string(models.RoleStaff)},
	}

	var users []models.User
	for j, s := range staffDefs {
		hid := hotelID
		email := fmt.Sprintf("h%d.%s.%d@hotelpms.com", hotelIndex+1, slugify(s.Role), j+1)

		// Resolve RoleID from the roles map
		var roleID *string
		if slug, ok := roleSlugMap[s.Role]; ok {
			if role, found := roles[slug]; found {
				id := role.ID.String()
				roleID = &id
			}
		}

		u := models.User{
			Email:        email,
			PasswordHash: hash,
			Name:         s.Name,
			Phone:        fmt.Sprintf("+91-99%04d%04d", hotelIndex+1, j+1),
			Role:         models.UserRole(s.Role),
			RoleID:       roleID,
			HotelID:      &hid,
			IsActive:     true,
		}
		if err := db.Create(&u).Error; err != nil {
			log.Fatalf("staff: %v", err)
		}
		users = append(users, u)
	}
	return users
}

func slugify(s string) string {
	out := ""
	for _, c := range s {
		if c == ' ' || c == '_' {
			out += "-"
		} else if c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '-' {
			out += string(c)
		} else if c >= 'A' && c <= 'Z' {
			out += string(c + 32)
		}
	}
	return out
}

// ─── Guests (40 per hotel) ───────────────────────────────────────────────────

var guestNames = []string{
	"Arjun Patel", "Priya Sharma", "Rahul Verma", "Ananya Gupta", "Vikram Singh",
	"Neha Reddy", "Amit Kumar", "Kavya Nair", "Sanjay Joshi", "Divya Menon",
	"Rajesh Iyer", "Meera Kapoor", "Aditya Rao", "Pooja Deshmukh", "Rohan Malhotra",
	"Sunita Das", "Karthik Pillai", "Deepa Choudhury", "Nikhil Bose", "Shreya Mukherjee",
	"John Smith", "Emily Johnson", "Michael Brown", "Sarah Williams", "David Jones",
	"Jessica Davis", "Chris Miller", "Amanda Wilson", "James Moore", "Laura Taylor",
	"Robert Anderson", "Lisa Thomas", "Daniel Jackson", "Karen White", "Matthew Harris",
	"Jennifer Martin", "Andrew Thompson", "Stephanie Garcia", "Ryan Martinez", "Nicole Robinson",
}

func seedGuests(db *gorm.DB, hotelID string) []models.Guest {
	var guests []models.Guest
	for i, name := range guestNames {
		g := models.Guest{
			HotelID: hotelID,
			Name:    name,
			Email:   fmt.Sprintf("guest.%d.%s@example.com", i+1, uuid.New().String()[:8]),
			Phone:   fmt.Sprintf("+91-98%04d%04d", rand.Intn(10000), i+1),
		}
		if err := db.Create(&g).Error; err != nil {
			log.Fatalf("guest: %v", err)
		}
		guests = append(guests, g)
	}
	return guests
}

// ─── Menu Items (30 per hotel) ───────────────────────────────────────────────

func seedMenuItems(db *gorm.DB, hotelID string) []models.MenuItem {
	defs := []struct {
		Name, Desc string
		Price      float64
		Cat        models.MenuCategory
	}{
		// Appetizers
		{"Paneer Tikka", "Marinated cottage cheese grilled in tandoor", 350, models.MenuCategoryAppetizer},
		{"Chicken Wings", "Crispy buffalo-style chicken wings", 450, models.MenuCategoryAppetizer},
		{"Spring Rolls", "Vegetable spring rolls with sweet chili sauce", 280, models.MenuCategoryAppetizer},
		{"Soup of the Day", "Chef's special daily soup", 220, models.MenuCategoryAppetizer},
		{"Bruschetta", "Toasted bread with tomato basil topping", 300, models.MenuCategoryAppetizer},
		{"Caesar Salad", "Classic caesar salad with croutons", 380, models.MenuCategoryAppetizer},
		// Main Course
		{"Butter Chicken", "Creamy tomato-based chicken curry", 550, models.MenuCategoryMainCourse},
		{"Dal Makhani", "Slow-cooked black lentils in cream", 400, models.MenuCategoryMainCourse},
		{"Grilled Salmon", "Atlantic salmon with lemon herb butter", 950, models.MenuCategoryMainCourse},
		{"Lamb Biryani", "Fragrant basmati rice with tender lamb", 650, models.MenuCategoryMainCourse},
		{"Pasta Carbonara", "Classic Italian carbonara with bacon", 520, models.MenuCategoryMainCourse},
		{"Margherita Pizza", "Fresh mozzarella, basil, tomato sauce", 480, models.MenuCategoryMainCourse},
		{"Grilled Chicken Steak", "Herb-marinated chicken with mashed potatoes", 600, models.MenuCategoryMainCourse},
		{"Vegetable Thai Curry", "Mixed vegetables in coconut curry", 450, models.MenuCategoryMainCourse},
		{"Fish and Chips", "Beer-battered fish with crispy fries", 550, models.MenuCategoryMainCourse},
		{"Mushroom Risotto", "Creamy arborio rice with wild mushrooms", 520, models.MenuCategoryMainCourse},
		// Desserts
		{"Gulab Jamun", "Traditional Indian milk dessert", 200, models.MenuCategoryDessert},
		{"Chocolate Lava Cake", "Warm chocolate cake with molten center", 350, models.MenuCategoryDessert},
		{"Tiramisu", "Classic Italian coffee-flavored dessert", 380, models.MenuCategoryDessert},
		{"Ice Cream Sundae", "Three scoops with toppings", 280, models.MenuCategoryDessert},
		{"Fruit Platter", "Seasonal fresh fruits", 250, models.MenuCategoryDessert},
		// Beverages
		{"Fresh Lime Soda", "Refreshing lime with soda", 150, models.MenuCategoryBeverage},
		{"Mango Lassi", "Sweet mango yogurt drink", 180, models.MenuCategoryBeverage},
		{"Cappuccino", "Italian espresso with steamed milk", 200, models.MenuCategoryBeverage},
		{"Green Tea", "Organic green tea", 150, models.MenuCategoryBeverage},
		{"Fresh Orange Juice", "Freshly squeezed orange juice", 220, models.MenuCategoryBeverage},
		{"Mineral Water", "500ml premium mineral water", 80, models.MenuCategoryBeverage},
		// Snacks
		{"French Fries", "Crispy golden fries with dip", 200, models.MenuCategorySnack},
		{"Club Sandwich", "Triple-decker chicken club sandwich", 380, models.MenuCategorySnack},
		{"Cheese Nachos", "Loaded nachos with cheese and salsa", 320, models.MenuCategorySnack},
	}

	var items []models.MenuItem
	for _, d := range defs {
		m := models.MenuItem{
			HotelID: hotelID, Name: d.Name, Description: d.Desc,
			Price: d.Price, Category: d.Cat, IsAvailable: true,
		}
		if err := db.Create(&m).Error; err != nil {
			log.Fatalf("menu: %v", err)
		}
		items = append(items, m)
	}
	return items
}

// ─── Activities (15 per hotel) ───────────────────────────────────────────────

func seedActivities(db *gorm.DB, hotelID string) []models.Activity {
	defs := []struct {
		Name, Desc string
		Price      float64
		Cat        models.ActivityCategory
	}{
		// Cab
		{"Airport Transfer", "Luxury sedan to/from airport", 2500, models.ActivityCategoryCab},
		{"City Tour Cab", "Full-day chauffeured city tour", 4000, models.ActivityCategoryCab},
		{"Station Pickup", "Railway station pickup service", 1500, models.ActivityCategoryCab},
		// Spa
		{"Swedish Massage", "60-minute full body relaxation massage", 3000, models.ActivityCategorySpa},
		{"Aromatherapy", "90-minute aromatherapy session", 4500, models.ActivityCategorySpa},
		{"Facial Treatment", "Deep cleansing facial treatment", 2500, models.ActivityCategorySpa},
		{"Couples Spa Package", "120-minute couples spa retreat", 8000, models.ActivityCategorySpa},
		// Tour
		{"Heritage Walk", "Guided 3-hour heritage walking tour", 1500, models.ActivityCategoryTour},
		{"Food Trail", "Local street food tasting tour", 2000, models.ActivityCategoryTour},
		{"Sunset Cruise", "Evening sunset cruise experience", 3500, models.ActivityCategoryTour},
		// Laundry
		{"Express Laundry", "Same-day express laundry service", 500, models.ActivityCategoryLaundry},
		{"Dry Cleaning", "Professional dry cleaning per piece", 300, models.ActivityCategoryLaundry},
		// Other
		{"Yoga Session", "Morning yoga with certified instructor", 800, models.ActivityCategoryOther},
		{"Cooking Class", "Learn local cuisine with our chef", 2500, models.ActivityCategoryOther},
		{"Bicycle Rental", "Full-day bicycle rental", 600, models.ActivityCategoryOther},
	}

	var out []models.Activity
	for _, d := range defs {
		a := models.Activity{
			HotelID: hotelID, Name: d.Name, Description: d.Desc,
			Price: d.Price, Category: d.Cat, IsAvailable: true,
		}
		if err := db.Create(&a).Error; err != nil {
			log.Fatalf("activity: %v", err)
		}
		out = append(out, a)
	}
	return out
}

// ─── Reservations + Orders + Activity Bookings + Bills + Housekeeping ────────

func seedReservationsAndRelated(
	db *gorm.DB, hotelID string,
	rooms []models.Room, guests []models.Guest, staff []models.User,
	menuItems []models.MenuItem, activities []models.Activity,
) {
	now := time.Now()
	statuses := []models.ReservationStatus{
		models.ReservationStatusReserved,
		models.ReservationStatusCheckedIn,
		models.ReservationStatusCheckedOut,
	}

	// Find housekeeping + room service staff
	var housekeepers, roomServiceStaff []models.User
	for _, s := range staff {
		switch s.Role {
		case models.RoleHousekeeping:
			housekeepers = append(housekeepers, s)
		case models.RoleRoomService:
			roomServiceStaff = append(roomServiceStaff, s)
		}
	}
	managerID := ""
	for _, s := range staff {
		if s.Role == models.RoleManager {
			managerID = s.ID.String()
			break
		}
	}

	// Create 30 reservations per hotel
	numReservations := 30
	if numReservations > len(rooms) {
		numReservations = len(rooms)
	}

	for i := 0; i < numReservations; i++ {
		room := rooms[i]
		guest := guests[i%len(guests)]
		status := statuses[i%len(statuses)]

		// Spread reservations across past, present, and future
		daysOffset := (i - numReservations/2) * 2
		checkIn := now.AddDate(0, 0, daysOffset)
		checkOut := checkIn.AddDate(0, 0, 2+rand.Intn(5))

		res := models.Reservation{
			HotelID:      hotelID,
			RoomID:       room.ID.String(),
			GuestID:      guest.ID.String(),
			GuestName:    guest.Name,
			GuestPhone:   guest.Phone,
			CheckInDate:  checkIn,
			CheckOutDate: checkOut,
			Status:       status,
			Notes:        fmt.Sprintf("Test reservation #%d", i+1),
		}
		if err := db.Create(&res).Error; err != nil {
			log.Fatalf("reservation: %v", err)
		}

		// Mark some rooms as occupied
		if status == models.ReservationStatusCheckedIn {
			db.Model(&room).Update("status", models.RoomStatusOccupied)
		}

		// Room inventory entries
		for d := checkIn; d.Before(checkOut); d = d.AddDate(0, 0, 1) {
			inv := models.RoomInventory{
				HotelID:       hotelID,
				RoomID:        room.ID.String(),
				Date:          d,
				IsAvailable:   false,
				ReservationID: strPtr(res.ID.String()),
			}
			db.Create(&inv) // ignore dups
		}

		// Orders (2 per checked-in reservation)
		if status == models.ReservationStatusCheckedIn && len(menuItems) > 0 {
			for o := 0; o < 2; o++ {
				orderItems := pickRandomMenuItems(menuItems, 2+rand.Intn(3))
				total := 0.0
				var ois []models.OrderItem
				for _, mi := range orderItems {
					qty := 1 + rand.Intn(2)
					sub := mi.Price * float64(qty)
					total += sub
					ois = append(ois, models.OrderItem{
						MenuItemID: mi.ID.String(),
						Quantity:   qty,
						UnitPrice:  mi.Price,
						Subtotal:   sub,
					})
				}

				var assignedTo *string
				if len(roomServiceStaff) > 0 {
					id := roomServiceStaff[rand.Intn(len(roomServiceStaff))].ID.String()
					assignedTo = &id
				}

				order := models.Order{
					HotelID:       hotelID,
					RoomID:        room.ID.String(),
					ReservationID: res.ID.String(),
					GuestID:       guest.ID.String(),
					GuestName:     guest.Name,
					Status:        models.OrderStatusDelivered,
					TotalAmount:   total,
					AssignedToID:  assignedTo,
					Notes:         fmt.Sprintf("Test order %d for reservation %d", o+1, i+1),
				}
				if err := db.Create(&order).Error; err != nil {
					log.Fatalf("order: %v", err)
				}
				for k := range ois {
					ois[k].OrderID = order.ID.String()
				}
				if err := db.Create(&ois).Error; err != nil {
					log.Fatalf("order items: %v", err)
				}
			}
		}

		// Activity Bookings (1 per checked-in reservation)
		if status == models.ReservationStatusCheckedIn && len(activities) > 0 {
			act := activities[rand.Intn(len(activities))]
			scheduled := now.Add(time.Duration(rand.Intn(48)) * time.Hour)
			ab := models.ActivityBooking{
				HotelID:       hotelID,
				RoomID:        room.ID.String(),
				ReservationID: res.ID.String(),
				ActivityID:    act.ID.String(),
				GuestID:       guest.ID.String(),
				GuestName:     guest.Name,
				ScheduledAt:   &scheduled,
				Status:        models.ActivityBookingConfirmed,
				Amount:        act.Price,
				Notes:         "Auto-seeded booking",
			}
			if err := db.Create(&ab).Error; err != nil {
				log.Fatalf("activity booking: %v", err)
			}
		}

		// Housekeeping tasks (for checked-out rooms)
		if status == models.ReservationStatusCheckedOut && len(housekeepers) > 0 && managerID != "" {
			hk := housekeepers[rand.Intn(len(housekeepers))]
			hkID := hk.ID.String()
			task := models.HousekeepingTask{
				HotelID:      hotelID,
				RoomID:       room.ID.String(),
				AssignedToID: &hkID,
				AssignedByID: managerID,
				Status:       models.HousekeepingStatusPending,
				Priority:     models.HousekeepingPriorityNormal,
				Notes:        fmt.Sprintf("Post-checkout cleaning for room %s", room.RoomNumber),
			}
			if err := db.Create(&task).Error; err != nil {
				log.Fatalf("housekeeping: %v", err)
			}
			db.Model(&room).Update("status", models.RoomStatusDirty)
		}

		// Bills (for checked-out reservations)
		if status == models.ReservationStatusCheckedOut {
			nights := int(checkOut.Sub(checkIn).Hours() / 24)
			if nights < 1 {
				nights = 1
			}
			roomCharges := room.PricePerNight * float64(nights)
			roomServiceTotal := float64(rand.Intn(3000))
			activityTotal := float64(rand.Intn(5000))
			subtotal := roomCharges + roomServiceTotal + activityTotal
			taxRate := 18.0
			taxAmt := subtotal * taxRate / 100
			totalAmt := subtotal + taxAmt

			bill := models.Bill{
				HotelID:          hotelID,
				ReservationID:    res.ID.String(),
				RoomID:           room.ID.String(),
				GuestID:          guest.ID.String(),
				GuestName:        guest.Name,
				RoomCharges:      roomCharges,
				UpfrontPaid:      roomCharges * 0.3,
				RoomServiceTotal: roomServiceTotal,
				ActivityTotal:    activityTotal,
				Subtotal:         subtotal,
				TaxRate:          taxRate,
				TaxAmount:        taxAmt,
				TotalAmount:      totalAmt,
				Status:           models.BillStatusPaid,
				Notes:            "Auto-generated bill",
			}
			if err := db.Create(&bill).Error; err != nil {
				log.Fatalf("bill: %v", err)
			}

			// Line items
			lines := []models.BillLineItem{
				{BillID: bill.ID.String(), Type: models.BillLineRoom, Description: fmt.Sprintf("Room %s - %d nights", room.RoomNumber, nights), Amount: roomCharges},
				{BillID: bill.ID.String(), Type: models.BillLineUpfront, Description: "Upfront payment at check-in", Amount: -(roomCharges * 0.3)},
			}
			if roomServiceTotal > 0 {
				lines = append(lines, models.BillLineItem{BillID: bill.ID.String(), Type: models.BillLineRoomService, Description: "Room service charges", Amount: roomServiceTotal})
			}
			if activityTotal > 0 {
				lines = append(lines, models.BillLineItem{BillID: bill.ID.String(), Type: models.BillLineActivity, Description: "Activity charges", Amount: activityTotal})
			}
			if err := db.Create(&lines).Error; err != nil {
				log.Fatalf("bill lines: %v", err)
			}
		}
	}
}

// ─── Today Check-Ins (reserved, checking in today) ─────────────────────────

func seedTodayCheckIns(db *gorm.DB, hotelID string, rooms []models.Room, guests []models.Guest, count int) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 14, 0, 0, 0, now.Location())

	for i := 0; i < count && i < len(rooms); i++ {
		room := rooms[i]
		guest := guests[i%len(guests)]
		checkOut := today.AddDate(0, 0, 2+rand.Intn(4))

		res := models.Reservation{
			HotelID:      hotelID,
			RoomID:       room.ID.String(),
			GuestID:      guest.ID.String(),
			GuestName:    guest.Name,
			GuestPhone:   guest.Phone,
			CheckInDate:  today,
			CheckOutDate: checkOut,
			Status:       models.ReservationStatusReserved,
			Notes:        fmt.Sprintf("Today check-in #%d", i+1),
		}
		if err := db.Create(&res).Error; err != nil {
			log.Fatalf("today check-in reservation: %v", err)
		}

		// Room inventory
		for d := today; d.Before(checkOut); d = d.AddDate(0, 0, 1) {
			inv := models.RoomInventory{
				HotelID: hotelID, RoomID: room.ID.String(),
				Date: d, IsAvailable: false, ReservationID: strPtr(res.ID.String()),
			}
			db.Create(&inv)
		}
	}
	log.Printf("   Today check-ins: %d", count)
}

// ─── Today Check-Outs (checked-in guests leaving today) ─────────────────────

func seedTodayCheckOuts(
	db *gorm.DB, hotelID string,
	rooms []models.Room, guests []models.Guest,
	staff []models.User, menuItems []models.MenuItem, activities []models.Activity,
) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 11, 0, 0, 0, now.Location())
	count := len(rooms)
	if count > len(guests) {
		count = len(guests)
	}

	for i := 0; i < count; i++ {
		room := rooms[i]
		guest := guests[i%len(guests)]
		nights := 2 + rand.Intn(4)
		checkIn := today.AddDate(0, 0, -nights)

		res := models.Reservation{
			HotelID:      hotelID,
			RoomID:       room.ID.String(),
			GuestID:      guest.ID.String(),
			GuestName:    guest.Name,
			GuestPhone:   guest.Phone,
			CheckInDate:  checkIn,
			CheckOutDate: today,
			Status:       models.ReservationStatusCheckedIn,
			Notes:        fmt.Sprintf("Today check-out #%d", i+1),
		}
		if err := db.Create(&res).Error; err != nil {
			log.Fatalf("today check-out reservation: %v", err)
		}

		db.Model(&room).Update("status", models.RoomStatusOccupied)

		// Room inventory
		for d := checkIn; d.Before(today); d = d.AddDate(0, 0, 1) {
			inv := models.RoomInventory{
				HotelID: hotelID, RoomID: room.ID.String(),
				Date: d, IsAvailable: false, ReservationID: strPtr(res.ID.String()),
			}
			db.Create(&inv)
		}

		// Add an order for each
		if len(menuItems) > 0 {
			items := pickRandomMenuItems(menuItems, 2+rand.Intn(3))
			total := 0.0
			var ois []models.OrderItem
			for _, mi := range items {
				qty := 1 + rand.Intn(2)
				sub := mi.Price * float64(qty)
				total += sub
				ois = append(ois, models.OrderItem{
					MenuItemID: mi.ID.String(), Quantity: qty,
					UnitPrice: mi.Price, Subtotal: sub,
				})
			}
			order := models.Order{
				HotelID: hotelID, RoomID: room.ID.String(),
				ReservationID: res.ID.String(), GuestID: guest.ID.String(),
				GuestName: guest.Name, Status: models.OrderStatusDelivered,
				TotalAmount: total, Notes: "Pre-checkout order",
			}
			if err := db.Create(&order).Error; err != nil {
				log.Fatalf("today checkout order: %v", err)
			}
			for k := range ois {
				ois[k].OrderID = order.ID.String()
			}
			db.Create(&ois)
		}

		// Activity booking
		if len(activities) > 0 {
			act := activities[rand.Intn(len(activities))]
			scheduled := now.Add(-time.Duration(rand.Intn(24)) * time.Hour)
			ab := models.ActivityBooking{
				HotelID: hotelID, RoomID: room.ID.String(),
				ReservationID: res.ID.String(), ActivityID: act.ID.String(),
				GuestID: guest.ID.String(), GuestName: guest.Name,
				ScheduledAt: &scheduled, Status: models.ActivityBookingCompleted,
				Amount: act.Price, Notes: "Completed during stay",
			}
			db.Create(&ab)
		}
	}
	log.Printf("   Today check-outs: %d", count)
}

// ─── Guest Settings ──────────────────────────────────────────────────────────

func seedGuestSettings(db *gorm.DB, hotelID string) {
	gs := models.GuestSettings{
		HotelID:         hotelID,
		WifiPassword:    fmt.Sprintf("Welcome%s!", uuid.New().String()[:4]),
		AllowOrders:     true,
		AllowActivities: true,
	}
	if err := db.Create(&gs).Error; err != nil {
		log.Fatalf("guest settings: %v", err)
	}
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

func strPtr(s string) *string { return &s }

func pickRandomMenuItems(items []models.MenuItem, n int) []models.MenuItem {
	perm := rand.Perm(len(items))
	var out []models.MenuItem
	for i := 0; i < n && i < len(perm); i++ {
		out = append(out, items[perm[i]])
	}
	return out
}
