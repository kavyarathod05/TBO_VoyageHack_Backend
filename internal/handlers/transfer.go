package handlers

import (
	"github.com/akashtripathi12/TBO_Backend/internal/models"
	"github.com/akashtripathi12/TBO_Backend/internal/utils"
	"github.com/gofiber/fiber/v2"
)

// GetAllTransfers returns all global transfer options (not event-specific)
// GET /api/v1/transfers
func (m *Repository) GetAllTransfers(c *fiber.Ctx) error {
	var transfers []models.Transfer

	// Optional query parameters for filtering
	cabType := c.Query("cab_type")
	status := c.Query("status", "active")

	query := m.DB.Model(&models.Transfer{}).Where("status = ?", status)

	if cabType != "" {
		query = query.Where("cab_type = ?", cabType)
	}

	// Order by cab type and price
	if err := query.Order("cab_type ASC, base_price_per_cab ASC").Find(&transfers).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch transfers")
	}

	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{
		"count": len(transfers),
		"data":  transfers,
	})
}

// GetTransfer returns a single global transfer by ID
// GET /api/v1/transfers/:id
func (m *Repository) GetTransfer(c *fiber.Ctx) error {
	transferID := c.Params("id")

	var transfer models.Transfer
	if err := m.DB.First(&transfer, "id = ?", transferID).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Transfer not found")
	}

	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{
		"data": transfer,
	})
}

// GetEventTransferBookings returns all transfer bookings for a specific event
// GET /api/v1/events/:id/transfers
func (m *Repository) GetEventTransferBookings(c *fiber.Ctx) error {
	eventID := c.Params("id")

	var bookings []models.TransferBooking
	if err := m.DB.Preload("Transfer").Where("event_id = ?", eventID).Find(&bookings).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch transfer bookings")
	}

	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{
		"count": len(bookings),
		"data":  bookings,
	})
}

// BookTransferForEvent creates a transfer booking for an event
// POST /api/v1/events/:id/transfers/book
func (m *Repository) BookTransferForEvent(c *fiber.Ctx) error {
	eventID := c.Params("id")

	type BookTransferRequest struct {
		TransferID     string `json:"transfer_id" validate:"required"`
		CabsBooked     int    `json:"cabs_booked" validate:"required,min=1"`
		PickupLocation string `json:"pickup_location"`
		DropLocation   string `json:"drop_location"`
		PickupTime     string `json:"pickup_time"` // ISO 8601 format
	}

	var req BookTransferRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	// Validate event exists
	var event models.Event
	if err := m.DB.First(&event, "id = ?", eventID).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Event not found")
	}

	// Validate transfer exists and has enough cabs
	var transfer models.Transfer
	if err := m.DB.First(&transfer, "id = ?", req.TransferID).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Transfer not found")
	}

	// Check availability
	if transfer.AvailableCount < req.CabsBooked {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Not enough cabs available")
	}

	// Start transaction
	tx := m.DB.Begin()

	// Reduce available count
	transfer.AvailableCount -= req.CabsBooked
	if err := tx.Save(&transfer).Error; err != nil {
		tx.Rollback()
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to update transfer availability")
	}

	// Create booking
	booking := models.TransferBooking{
		TransferID:     transfer.ID,
		EventID:        event.ID,
		CabsBooked:     req.CabsBooked,
		PriceLocked:    transfer.BasePricePerCab, // Lock current price
		PickupLocation: req.PickupLocation,
		DropLocation:   req.DropLocation,
		Status:         "confirmed",
	}

	if err := tx.Create(&booking).Error; err != nil {
		tx.Rollback()
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to create booking")
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to complete booking")
	}

	// Load transfer details
	m.DB.Preload("Transfer").First(&booking, booking.ID)

	return utils.SuccessResponse(c, fiber.StatusCreated, fiber.Map{
		"message": "Transfer booked successfully",
		"data":    booking,
	})
}

// CreateTransfer creates a new global transfer option (admin only)
// POST /api/v1/transfers
func (m *Repository) CreateTransfer(c *fiber.Ctx) error {
	var transfer models.Transfer

	if err := c.BodyParser(&transfer); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	// Set default status if not provided
	if transfer.Status == "" {
		transfer.Status = "active"
	}

	// Set available count equal to total count if not provided
	if transfer.AvailableCount == 0 {
		transfer.AvailableCount = transfer.TotalCount
	}

	if err := m.DB.Create(&transfer).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to create transfer")
	}

	return utils.SuccessResponse(c, fiber.StatusCreated, fiber.Map{
		"message": "Transfer created successfully",
		"data":    transfer,
	})
}

// UpdateTransfer updates a global transfer
// PUT /api/v1/transfers/:id
func (m *Repository) UpdateTransfer(c *fiber.Ctx) error {
	transferID := c.Params("id")

	var transfer models.Transfer
	if err := m.DB.First(&transfer, "id = ?", transferID).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Transfer not found")
	}

	type UpdateTransferRequest struct {
		TotalCount      *int     `json:"total_count"`
		AvailableCount  *int     `json:"available_count"`
		BasePricePerCab *float64 `json:"base_price_per_cab"`
		Status          *string  `json:"status"`
	}

	var req UpdateTransferRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	// Update fields if provided
	if req.TotalCount != nil {
		transfer.TotalCount = *req.TotalCount
	}
	if req.AvailableCount != nil {
		transfer.AvailableCount = *req.AvailableCount
	}
	if req.BasePricePerCab != nil {
		transfer.BasePricePerCab = *req.BasePricePerCab
	}
	if req.Status != nil {
		transfer.Status = *req.Status
	}

	if err := m.DB.Save(&transfer).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to update transfer")
	}

	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{
		"message": "Transfer updated successfully",
		"data":    transfer,
	})
}

// DeleteTransfer deletes a global transfer
// DELETE /api/v1/transfers/:id
func (m *Repository) DeleteTransfer(c *fiber.Ctx) error {
	transferID := c.Params("id")

	var transfer models.Transfer
	if err := m.DB.First(&transfer, "id = ?", transferID).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Transfer not found")
	}

	if err := m.DB.Delete(&transfer).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to delete transfer")
	}

	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{
		"message": "Transfer deleted successfully",
	})
}

// CancelTransferBooking cancels a transfer booking and restores cabs
// DELETE /api/v1/events/:event_id/transfers/:booking_id
func (m *Repository) CancelTransferBooking(c *fiber.Ctx) error {
	eventID := c.Params("event_id")
	bookingID := c.Params("booking_id")

	var booking models.TransferBooking
	if err := m.DB.Preload("Transfer").First(&booking, "id = ? AND event_id = ?", bookingID, eventID).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Booking not found")
	}

	// Start transaction
	tx := m.DB.Begin()

	// Restore available count
	var transfer models.Transfer
	if err := tx.First(&transfer, booking.TransferID).Error; err != nil {
		tx.Rollback()
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to find transfer")
	}

	transfer.AvailableCount += booking.CabsBooked
	if err := tx.Save(&transfer).Error; err != nil {
		tx.Rollback()
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to restore cabs")
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
