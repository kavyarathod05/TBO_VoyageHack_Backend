package main

import (
	"log"
	"math/rand"
	"time"

	"github.com/akashtripathi12/TBO_Backend/internal/models"
	"gorm.io/gorm"
)

// SeedFlights seeds global flights (not tied to any event)
func SeedFlights(db *gorm.DB) {
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

		log.Printf("   ✓ Created flight: %s (%s → DEL) - %s - ₹%.2f - %d/%d seats available",
			flightNumber, route.from, route.airline, basePrice, availableSeats, totalSeats)

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
				log.Printf("   ✓ Created flight: %s (%s → DEL) - %s - ₹%.2f - %d/%d seats available",
					eveningFlight.FlightNumber, route.from, route.airline, eveningFlight.BasePrice, eveningFlight.AvailableSeats, eveningFlight.TotalSeats)
			}
		}
	}

	log.Printf("✅ Seeded %d global flights successfully!\n", len(flights))
}
