package handlers

import (
	"github.com/akashtripathi12/TBO_Backend/internal/models"
	"github.com/akashtripathi12/TBO_Backend/internal/utils"
	"github.com/gofiber/fiber/v2"
)

// GetFlightLocations returns unique departure and arrival codes
// GET /api/v1/flights/locations
func (m *Repository) GetFlightLocations(c *fiber.Ctx) error {
	var departureCodes []string
	var arrivalCodes []string

	// Get distinct departure codes
	if err := m.DB.Model(&models.Flight{}).Distinct("departure_code").Pluck("departure_code", &departureCodes).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch departure codes")
	}

	// Get distinct arrival codes
	if err := m.DB.Model(&models.Flight{}).Distinct("arrival_code").Pluck("arrival_code", &arrivalCodes).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch arrival codes")
	}

	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{
		"departure_codes": departureCodes,
		"arrival_codes":   arrivalCodes,
	})
}
