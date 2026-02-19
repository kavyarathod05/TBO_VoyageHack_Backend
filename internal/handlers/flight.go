package handlers

import (
	"github.com/akashtripathi12/TBO_Backend/internal/models"
	"github.com/akashtripathi12/TBO_Backend/internal/utils"
	"github.com/gofiber/fiber/v2"
)

// GetAllFlights returns all global flights (not event-specific)
// GET /api/v1/flights
func (m *Repository) GetAllFlights(c *fiber.Ctx) error {
	var flights []models.Flight

	// Optional query parameters for filtering
	departureCode := c.Query("departure_code")
	arrivalCode := c.Query("arrival_code")
	status := c.Query("status", "active")

	query := m.DB.Model(&models.Flight{}).Where("status = ?", status)

	if departureCode != "" {
		query = query.Where("departure_code = ?", departureCode)
	}
	if arrivalCode != "" {
		query = query.Where("arrival_code = ?", arrivalCode)
	}

	// Order by departure time
	if err := query.Order("departure_time ASC").Find(&flights).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch flights")
	}

	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{
		"count": len(flights),
		"data":  flights,
	})
}

// GetFlight returns a single global flight by ID
// GET /api/v1/flights/:id
func (m *Repository) GetFlight(c *fiber.Ctx) error {
	flightID := c.Params("id")

	var flight models.Flight
	if err := m.DB.First(&flight, "id = ?", flightID).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Flight not found")
	}

	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{
		"data": flight,
	})
}

// GetEventFlightBookings returns all flight bookings for a specific event
// GET /api/v1/events/:id/flights
func (m *Repository) GetEventFlightBookings(c *fiber.Ctx) error {
	eventID := c.Params("id")

	var bookings []models.FlightBooking
	if err := m.DB.Preload("Flight").Where("event_id = ?", eventID).Find(&bookings).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch flight bookings")
	}

	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{
		"count": len(bookings),
		"data":  bookings,
	})
}

// BookFlightForEvent creates a flight booking for an event
// POST /api/v1/events/:id/flights/book
func (m *Repository) BookFlightForEvent(c *fiber.Ctx) error {
	eventID := c.Params("id")

	type BookFlightRequest struct {
		FlightID    string `json:"flight_id" validate:"required"`
		SeatsBooked int    `json:"seats_booked" validate:"required,min=1"`
	}

	var req BookFlightRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	// Validate event exists
	var event models.Event
	if err := m.DB.First(&event, "id = ?", eventID).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Event not found")
	}

	// Validate flight exists and has enough seats
	var flight models.Flight
	if err := m.DB.First(&flight, "id = ?", req.FlightID).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Flight not found")
	}

	// Check availability
	if flight.AvailableSeats < req.SeatsBooked {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Not enough seats available")
	}

	// Start transaction
	tx := m.DB.Begin()

	// Reduce available seats
	flight.AvailableSeats -= req.SeatsBooked
	if err := tx.Save(&flight).Error; err != nil {
		tx.Rollback()
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to update flight availability")
	}

	// Create booking
	booking := models.FlightBooking{
		FlightID:    flight.ID,
		EventID:     event.ID,
		SeatsBooked: req.SeatsBooked,
		PriceLocked: flight.BasePrice, // Lock current price
		Status:      "confirmed",
	}

	if err := tx.Create(&booking).Error; err != nil {
		tx.Rollback()
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to create booking")
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to complete booking")
	}

	// Load flight details
	m.DB.Preload("Flight").First(&booking, booking.ID)

	return utils.SuccessResponse(c, fiber.StatusCreated, fiber.Map{
		"message": "Flight booked successfully",
		"data":    booking,
	})
}

// CreateFlight creates a new global flight (admin only)
// POST /api/v1/flights
func (m *Repository) CreateFlight(c *fiber.Ctx) error {
	var flight models.Flight

	if err := c.BodyParser(&flight); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	// Set default status if not provided
	if flight.Status == "" {
		flight.Status = "active"
	}

	// Set available seats equal to total seats if not provided
	if flight.AvailableSeats == 0 {
		flight.AvailableSeats = flight.TotalSeats
	}

	if err := m.DB.Create(&flight).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to create flight")
	}

	return utils.SuccessResponse(c, fiber.StatusCreated, fiber.Map{
		"message": "Flight created successfully",
		"data":    flight,
	})
}

// UpdateFlight updates a global flight
// PUT /api/v1/flights/:id
func (m *Repository) UpdateFlight(c *fiber.Ctx) error {
	flightID := c.Params("id")

	var flight models.Flight
	if err := m.DB.First(&flight, "id = ?", flightID).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Flight not found")
	}

	type UpdateFlightRequest struct {
		TotalSeats     *int     `json:"total_seats"`
		AvailableSeats *int     `json:"available_seats"`
		BasePrice      *float64 `json:"base_price"`
		Status         *string  `json:"status"`
	}

	var req UpdateFlightRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	// Update fields if provided
	if req.TotalSeats != nil {
		flight.TotalSeats = *req.TotalSeats
	}
	if req.AvailableSeats != nil {
		flight.AvailableSeats = *req.AvailableSeats
	}
	if req.BasePrice != nil {
		flight.BasePrice = *req.BasePrice
	}
	if req.Status != nil {
		flight.Status = *req.Status
	}

	if err := m.DB.Save(&flight).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to update flight")
	}

	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{
		"message": "Flight updated successfully",
		"data":    flight,
	})
}

// DeleteFlight deletes a global flight
// DELETE /api/v1/flights/:id
func (m *Repository) DeleteFlight(c *fiber.Ctx) error {
	flightID := c.Params("id")

	var flight models.Flight
	if err := m.DB.First(&flight, "id = ?", flightID).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Flight not found")
	}

	if err := m.DB.Delete(&flight).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to delete flight")
	}

	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{
		"message": "Flight deleted successfully",
	})
}

// CancelFlightBooking cancels a flight booking and restores seats
// DELETE /api/v1/events/:event_id/flights/:booking_id
func (m *Repository) CancelFlightBooking(c *fiber.Ctx) error {
	eventID := c.Params("event_id")
	bookingID := c.Params("booking_id")

	var booking models.FlightBooking
	if err := m.DB.Preload("Flight").First(&booking, "id = ? AND event_id = ?", bookingID, eventID).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Booking not found")
	}

	// Start transaction
	tx := m.DB.Begin()

	// Restore available seats
	var flight models.Flight
	if err := tx.First(&flight, booking.FlightID).Error; err != nil {
		tx.Rollback()
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to find flight")
	}

	flight.AvailableSeats += booking.SeatsBooked
	if err := tx.Save(&flight).Error; err != nil {
		tx.Rollback()
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to restore seats")
	}

	// Update booking status
	booking.Status = "cancelled"
	if err := tx.Save(&booking).Error; err != nil {
		tx.Rollback()
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to cancel booking")
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to complete cancellation")
	}

	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{
		"message": "Booking cancelled successfully",
		"data":    booking,
	})
}
