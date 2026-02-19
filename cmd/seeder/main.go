package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/akashtripathi12/TBO_Backend/internal/models"
	"github.com/akashtripathi12/TBO_Backend/internal/store"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// RoomsInventoryItem represents a single room type in inventory
type RoomsInventoryItem struct {
	RoomOfferID  string `json:"room_offer_id"`
	RoomName     string `json:"room_name"`
	Available    int    `json:"available"`
	MaxCapacity  int    `json:"max_capacity"`
	PricePerRoom int    `json:"price_per_room"`
}

func main() {
	rand.Seed(time.Now().UnixNano())

	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  No .env file found, relying on environment variables")
	}

	store.InitDB()
	db := store.DB

	// 1. Full Database Reset
	log.Println("⚠️  STARTING DATABASE RESET...")
	// Order matters: drop child tables first
	tables := []string{
		"cart_items",
		"flight_bookings",
		"transfer_bookings",
		"guest_allocations",
		"room_offers",
		"banquet_halls",
		"catering_menus",
		"hotels",
		"cities",
		"countries",
		"events",
		"agent_profiles",
		"guests",
		"users",
		"flights",
		"transfers",
	}

	for _, table := range tables {
		if err := db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", table)).Error; err != nil {
			log.Fatalf("❌ Failed to drop table %s: %v", table, err)
		}
	}
	log.Println("✅ Database Cleared.")

	// 2. Automigration
	log.Println("🛠️  Running Automigrate...")
	err := db.AutoMigrate(
		&models.User{},
		&models.AgentProfile{},
		&models.Country{},
		&models.City{},
		&models.Hotel{},
		&models.RoomOffer{},
		&models.BanquetHall{},
		&models.CateringMenu{},
		&models.Event{},
		&models.Guest{},
		&models.GuestAllocation{},
		&models.CartItem{},
		&models.Flight{},
		&models.FlightBooking{},
		&models.Transfer{},
		&models.TransferBooking{},
	)
	if err != nil {
		log.Fatal("❌ Migration Failed:", err)
	}
	log.Println("✅ All tables created successfully!")

	// 3. Seed Users
	log.Println("👤 Seeding Users...")
	agentUser := models.User{
		ID:           uuid.New(),
		Name:         "Demo Agent",
		Email:        "agent@demo.com",
		PasswordHash: "$2a$10$5XQ/P7S1q1.H.5W/5.6.7uGvW8R.y6YEqg/kL5jD0j6qP6hD.fA1S", // Admin@123
		Phone:        "+91-9876543210",
		Role:         "agent",
	}
	db.Create(&agentUser)

	agentProfile := models.AgentProfile{
		UserID:        agentUser.ID,
		AgencyName:    "Demo Travel Agency",
		AgencyCode:    "DTA001",
		Location:      "New Delhi, India",
		BusinessPhone: "+91-11-99999999",
	}
	db.Create(&agentProfile)

	headGuestUser := models.User{
		ID:           uuid.New(),
		Name:         "Global Event Manager",
		Email:        "headguest@demo.com",
		PasswordHash: "$2a$10$5XQ/P7S1q1.H.5W/5.6.7uGvW8R.y6YEqg/kL5jD0j6qP6hD.fA1S", // Admin@123
		Phone:        "+91-9123456789",
		Role:         "head_guest",
	}
	db.Create(&headGuestUser)

	// 4. Seed Geography
	log.Println("🌍 Seeding Geography...")
	country := models.Country{
		Code:      "IN",
		Name:      "India",
		PhoneCode: "91",
	}
	if err := db.Create(&country).Error; err != nil {
		log.Printf("❌ Failed to create country: %v", err)
	} else {
		log.Println("   ✓ Created Country: India")
	}

	city := models.City{
		ID:          "DEL",
		CountryCode: "IN",
		Name:        "New Delhi",
		IsPopular:   true,
	}
	if err := db.Create(&city).Error; err != nil {
		log.Printf("❌ Failed to create city: %v", err)
	} else {
		log.Println("   ✓ Created City: New Delhi")
	}

	// 5. Seed Hotels (50 hotels)
	log.Println("🏨 Seeding 50 Hotels...")
	hotelNames := []string{"Grand", "Royal", "Palace", "Residency", "Heritage", "Luxury", "Budget", "Plaza", "Continental", "Imperial"}
	hotelTypes := []string{"Plaza", "Suites", "Inn", "Resort", "Towers", "Gardens", "Boutique"}

	// Realistic Unsplash Hotel Images
	hotelImages := []string{
		"https://images.unsplash.com/photo-1566073771259-6a8506099945",
		"https://images.unsplash.com/photo-1582719478250-c89cae4dc85b",
		"https://images.unsplash.com/photo-1542314831-068cd1dbfeeb",
		"https://images.unsplash.com/photo-1520250497591-112f2f40a3f4",
		"https://images.unsplash.com/photo-1571896349842-33c89424de2d",
		"https://images.unsplash.com/photo-1445019980597-93fa8acb246c",
		"https://images.unsplash.com/photo-1564501049412-61c2a3083791",
		"https://images.unsplash.com/photo-1551882547-ff40c63fe5fa",
		"https://images.unsplash.com/photo-1535827841776-24afc1e255ac",
		"https://images.unsplash.com/photo-1618773928121-c32242e63f39",
	}

	var firstHotel models.Hotel
	for i := 1; i <= 50; i++ {
		name := fmt.Sprintf("%s %s %d", hotelNames[rand.Intn(len(hotelNames))], hotelTypes[rand.Intn(len(hotelTypes))], i)
		starRating := rand.Intn(3) + 3    // 3 to 5 stars
		occupancy := rand.Intn(501) + 500 // 500 to 1000

		facilities, _ := json.Marshal([]string{"WiFi", "Pool", "Restaurant", "Gym", "Spa", "Valet Parking"})

		// Use a specific image and a few random ones for each hotel
		imgURL := hotelImages[rand.Intn(len(hotelImages))] + "?auto=format&fit=crop&w=800&q=80"
		imageUrls, _ := json.Marshal([]string{imgURL, "https://images.unsplash.com/photo-1566665797739-1674de7a421a?auto=format&fit=crop&w=800&q=80"})

		h := models.Hotel{
			ID:         fmt.Sprintf("HOTEL%03d", i),
			CityID:     "DEL",
			Name:       name,
			StarRating: starRating,
			Address:    fmt.Sprintf("Address Line %d, New Delhi - 110001", i),
			Facilities: datatypes.JSON(facilities),
			ImageUrls:  datatypes.JSON(imageUrls),
			Occupancy:  occupancy,
		}
		db.Create(&h)
		if i == 1 {
			firstHotel = h
		}

		// 6. Seed Room Offers for each hotel
		roomTypes := []struct {
			Name     string
			Capacity int
			BaseFare float64
		}{
			{"Standard Room", 2, 3000.0},
			{"Deluxe Room", 3, 5000.0},
			{"Executive Suite", 4, 8000.0},
			{"Presidential Suite", 5, 15000.0},
		}

		for _, rt := range roomTypes {
			count := rand.Intn(51) + 50 // 50 to 100
			fare := rt.BaseFare + float64(rand.Intn(1000))

			db.Create(&models.RoomOffer{
				ID:          uuid.New().String(),
				HotelID:     h.ID,
				Name:        rt.Name,
				BookingCode: fmt.Sprintf("BOOK-%s-%s", h.ID, rt.Name[:3]),
				MaxCapacity: rt.Capacity,
				TotalFare:   fare,
				Currency:    "INR",
				Count:       count,
			})
		}

		// 7. Seed Banquets and Catering for every 4th hotel (25%)
		if i%4 == 0 {
			banquetImages, _ := json.Marshal([]string{
				"https://images.unsplash.com/photo-1519167758481-83f550bb49b3?auto=format&fit=crop&w=800&q=80",
				"https://images.unsplash.com/photo-1511795409834-ef04bbd61622?auto=format&fit=crop&w=800&q=80",
			})

			db.Create(&models.BanquetHall{
				HotelID:     h.ID,
				Name:        "Crystal Ballroom",
				Capacity:    300 + rand.Intn(201), // 300-500
				PricePerDay: 50000.0 + float64(rand.Intn(20000)),
				ImageUrls:   datatypes.JSON(banquetImages),
			})

			loungeImages, _ := json.Marshal([]string{
				"https://images.unsplash.com/photo-1541336032412-2048a678540d?auto=format&fit=crop&w=800&q=80",
			})

			db.Create(&models.BanquetHall{
				HotelID:     h.ID,
				Name:        "Executive Lounge",
				Capacity:    50 + rand.Intn(51), // 50-100
				PricePerDay: 20000.0 + float64(rand.Intn(5000)),
				ImageUrls:   datatypes.JSON(loungeImages),
			})
			cateringImages, _ := json.Marshal([]string{
				"https://images.unsplash.com/photo-1555244162-803834f70033?auto=format&fit=crop&w=800&q=80",
				"https://images.unsplash.com/photo-1544333346-64e4fe1f8ff2?auto=format&fit=crop&w=800&q=80",
			})

			db.Create(&models.CateringMenu{
				HotelID:       h.ID,
				Name:          "Premium Wedding Menu",
				Type:          "mixed",
				PricePerPlate: 1500.0,
				ImageUrls:     datatypes.JSON(cateringImages),
			})

			buffetImages, _ := json.Marshal([]string{
				"https://images.unsplash.com/photo-1504674900247-0877df9cc836?auto=format&fit=crop&w=800&q=80",
			})

			db.Create(&models.CateringMenu{
				HotelID:       h.ID,
				Name:          "Corporate Buffet",
				Type:          "veg",
				PricePerPlate: 800.0,
				ImageUrls:     datatypes.JSON(buffetImages),
			})
		}
	}

	// 8. Seed Main Event
	log.Println("📅 Seeding Main Event...")
	var roomOffers []models.RoomOffer
	db.Where("hotel_id = ?", firstHotel.ID).Find(&roomOffers)

	var invItems []RoomsInventoryItem
	for _, ro := range roomOffers {
		invItems = append(invItems, RoomsInventoryItem{
			RoomOfferID:  ro.ID,
			RoomName:     ro.Name,
			Available:    ro.Count,
			MaxCapacity:  ro.MaxCapacity,
			PricePerRoom: int(ro.TotalFare),
		})
	}
	invJSON, _ := json.Marshal(invItems)

	event := models.Event{
		ID:             uuid.New(),
		AgentID:        agentUser.ID,
		HeadGuestID:    headGuestUser.ID,
		HotelID:        firstHotel.ID,
		Name:           "Global Tech Summit 2026",
		Location:       "New Delhi, India",
		RoomsInventory: datatypes.JSON(invJSON),
		Status:         "active",
		StartDate:      time.Now().AddDate(0, 1, 0),
		EndDate:        time.Now().AddDate(0, 1, 4),
	}
	db.Create(&event)

	// 9. Seed 500 Guests for the event
	log.Println("👥 Seeding 500 Guests organized in families...")
	guestCount := 0
	for guestCount < 500 {
		familyID := uuid.New()
		familySize := rand.Intn(5) + 1 // 1 to 5 members
		if guestCount+familySize > 500 {
			familySize = 500 - guestCount
		}

		for j := 0; j < familySize; j++ {
			age := 5 + rand.Intn(60)
			gType := "adult"
			if age < 12 {
				gType = "child"
			}

			guest := models.Guest{
				ID:            uuid.New(),
				Name:          fmt.Sprintf("Guest %d", guestCount+1),
				Age:           age,
				Type:          gType,
				Phone:         fmt.Sprintf("+91-90000%05d", guestCount),
				Email:         fmt.Sprintf("guest%d@demo.com", guestCount),
				EventID:       event.ID,
				FamilyID:      familyID,
				ArrivalDate:   event.StartDate,
				DepartureDate: event.EndDate,
			}
			db.Create(&guest)
			guestCount++

			// 10. Randomly allocate ~100 guests to show status
			if guestCount <= 100 {
				ro := roomOffers[rand.Intn(len(roomOffers))]
				db.Create(&models.GuestAllocation{
					ID:           uuid.New(),
					EventID:      event.ID,
					GuestID:      guest.ID,
					RoomOfferID:  &ro.ID,
					LockedPrice:  ro.TotalFare,
					Status:       "allocated",
					AssignedMode: "agent_manual",
				})
			}
		}
	}

	// 11. Seed Global Flights
	flights := seedGlobalFlights(db)

	// 12. Seed Global Transfers
	transfers := seedGlobalTransfers(db)

	// 13. Seed Cart Data (linked to bookings)
	log.Println("🛒 Seeding Cart with Flight and Transfer Bookings...")

	// Find one flight to book
	if len(flights) > 0 {
		flight := flights[0]
		fb := models.FlightBooking{
			FlightID:    flight.ID,
			EventID:     event.ID,
			SeatsBooked: 2,
			PriceLocked: flight.BasePrice,
			Status:      "pending",
			BookedBy:    headGuestUser.ID,
		}
		db.Create(&fb)

		cartItemF := models.CartItem{
			EventID:         event.ID,
			Type:            "flight",
			RefID:           flight.ID.String(),
			FlightBookingID: &fb.ID,
			Status:          "wishlist",
			Quantity:        2,
			LockedPrice:     flight.BasePrice,
			AddedBy:         headGuestUser.ID,
		}
		db.Create(&cartItemF)
		log.Printf("   ✓ Added Flight Booking to Cart: %s (%s)", fb.ID, flight.FlightNumber)
	}

	// Find one transfer to book
	if len(transfers) > 0 {
		transfer := transfers[0]
		tb := models.TransferBooking{
			TransferID:     transfer.ID,
			EventID:        event.ID,
			CabsBooked:     1,
			PriceLocked:    transfer.BasePricePerCab,
			PickupLocation: "Airport",
			DropLocation:   "Hotel",
			Status:         "pending",
			BookedBy:       headGuestUser.ID,
		}
		db.Create(&tb)

		cartItemT := models.CartItem{
			EventID:           event.ID,
			Type:              "transfer",
			RefID:             transfer.ID.String(),
			TransferBookingID: &tb.ID,
			Status:            "wishlist",
			Quantity:          1,
			LockedPrice:       transfer.BasePricePerCab,
			AddedBy:           headGuestUser.ID,
		}
		db.Create(&cartItemT)
		log.Printf("   ✓ Added Transfer Booking to Cart: %s (%s)", tb.ID, transfer.CarModel)
	}

	log.Println("🎉 DEMO SEEDING COMPLETED!")
	log.Printf("📊 Summary:")
	log.Printf("   - Users: 2 (1 Agent, 1 Head Guest)")
	log.Printf("   - Hotels: 50")
	log.Printf("   - Room Offers: 200")
	log.Printf("   - Events: 1 (%s)", event.Name)
	log.Printf("   - Guests: 500")
	log.Printf("   - Allocations: 100")
	log.Printf("   - Flights: %d", len(flights))
	log.Printf("   - Transfers: %d", len(transfers))
}

// seedGlobalFlights seeds global flights (not tied to any event) and returns them
func seedGlobalFlights(db *gorm.DB) []models.Flight {
	log.Println("✈️  Seeding Global Flights...")

	// Routes to Delhi from various cities
	routes := []struct {
		from     string
		fromCode string
		airline  string
		prefix   string
		distance int // for pricing
	}{
		{"Mumbai", "BOM", "Air India", "AI", 1400},
		{"Bangalore", "BLR", "IndiGo", "6E", 2100},
		{"Dubai", "DXB", "Emirates", "EK", 2200},
		{"Singapore", "SIN", "Singapore Airlines", "SQ", 4100},
		{"London", "LHR", "British Airways", "BA", 6700},
		{"New York", "JFK", "Air India", "AI", 12000},
		{"Kolkata", "CCU", "Vistara", "UK", 1500},
		{"Chennai", "MAA", "IndiGo", "6E", 2180},
		{"Hyderabad", "HYD", "Air India", "AI", 1580},
		{"Ahmedabad", "AMD", "Vistara", "UK", 1050},
	}

	var flights []models.Flight
	for i, route := range routes {
		// Generate flight number
		flightNum := rand.Intn(900) + 100
		flightNumber := route.prefix + " " + string(rune(flightNum/100+'0')) + string(rune((flightNum/10)%10+'0')) + string(rune(flightNum%10+'0'))

		// Random departure time (next 7 days)
		daysAhead := rand.Intn(7) + 1
		hour := rand.Intn(18) + 6 // 6 AM to 11 PM
		minute := []int{0, 15, 30, 45}[rand.Intn(4)]
		departureTime := time.Now().AddDate(0, 0, daysAhead).
			Truncate(24 * time.Hour).
			Add(time.Duration(hour)*time.Hour + time.Duration(minute)*time.Minute)

		// Calculate arrival time (based on distance)
		flightDuration := time.Duration(route.distance/800+1) * time.Hour
		arrivalTime := departureTime.Add(flightDuration)

		// Calculate price based on distance
		basePrice := float64(route.distance) * 4.5
		if route.distance > 5000 {
			basePrice = float64(route.distance) * 6.5 // International premium
		}

		// Random seat configuration
		totalSeats := []int{150, 180, 200, 250}[rand.Intn(4)]
		availableSeats := totalSeats - rand.Intn(50) // Some seats already booked

		flight := models.Flight{
			FlightNumber:   flightNumber,
			AirlineName:    route.airline,
			DepartureTime:  departureTime,
			ArrivalTime:    arrivalTime,
			DepartureCode:  route.fromCode,
			ArrivalCode:    "DEL",
			TotalSeats:     totalSeats,
			AvailableSeats: availableSeats,
			BasePrice:      basePrice,
			Status:         "active",
		}

		flights = append(flights, flight)

		// Create the flight
		if err := db.Create(&flight).Error; err != nil {
			log.Printf("   ❌ Failed to create flight %s: %v", flightNumber, err)
			continue
		}

		// Add some variety - create a second flight for popular routes
		if i < 5 {
			// Create evening flight
			eveningDeparture := departureTime.Add(8 * time.Hour)
			eveningArrival := eveningDeparture.Add(flightDuration)

			eveningFlight := models.Flight{
				FlightNumber:   route.prefix + " " + string(rune((flightNum+100)/100+'0')) + string(rune(((flightNum+100)/10)%10+'0')) + string(rune((flightNum+100)%10+'0')),
				AirlineName:    route.airline,
				DepartureTime:  eveningDeparture,
				ArrivalTime:    eveningArrival,
				DepartureCode:  route.fromCode,
				ArrivalCode:    "DEL",
				TotalSeats:     totalSeats,
				AvailableSeats: totalSeats - rand.Intn(30),
				BasePrice:      basePrice * 1.1, // Evening flights slightly more expensive
				Status:         "active",
			}

			if err := db.Create(&eveningFlight).Error; err == nil {
				flights = append(flights, eveningFlight)
			}
		}
	}

	log.Printf("✅ Seeded %d global flights successfully!\n", len(flights))
	return flights
}

// seedGlobalTransfers seeds global transfer options (not tied to any event) and returns them
func seedGlobalTransfers(db *gorm.DB) []models.Transfer {
	log.Println("🚗 Seeding Global Transfers...")

	transferOptions := []struct {
		cabType   string
		models    []string
		basePrice float64
		capacity  int
	}{
		{
			cabType:   "hatchback",
			models:    []string{"Maruti Swift", "Hyundai i20", "Tata Altroz"},
			basePrice: 1500,
			capacity:  15,
		},
		{
			cabType:   "sedan",
			models:    []string{"Honda City", "Hyundai Verna", "Maruti Ciaz", "Toyota Etios"},
			basePrice: 2500,
			capacity:  20,
		},
		{
			cabType:   "suv",
			models:    []string{"Toyota Innova", "Toyota Fortuner", "Mahindra XUV700", "MG Hector"},
			basePrice: 4000,
			capacity:  12,
		},
	}

	var transfers []models.Transfer
	for _, option := range transferOptions {
		for _, carModel := range option.models {
			// Random pricing variation
			priceVariation := rand.Float64()*500 - 250 // ±250
			price := option.basePrice + priceVariation

			// Random availability
			totalCount := option.capacity + rand.Intn(10) - 5 // ±5 from base capacity
			if totalCount < 5 {
				totalCount = 5
			}
			availableCount := totalCount - rand.Intn(totalCount/3) // Some already booked

			transfer := models.Transfer{
				CabType:         option.cabType,
				CarModel:        carModel,
				TotalCount:      totalCount,
				AvailableCount:  availableCount,
				BasePricePerCab: price,
				Status:          "active",
			}

			transfers = append(transfers, transfer)

			// Create the transfer
			if err := db.Create(&transfer).Error; err != nil {
				log.Printf("   ❌ Failed to create transfer %s (%s): %v", option.cabType, carModel, err)
				continue
			}
		}
	}

	log.Printf("✅ Seeded %d global transfer options successfully!\n", len(transfers))
	return transfers
}
