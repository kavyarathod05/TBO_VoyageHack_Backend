package handlers

import (
	"strconv"

	"github.com/akashtripathi12/TBO_Backend/internal/models"
	"github.com/gofiber/fiber/v2"
)

// GetBanquetsByHotel retrieves banquet halls for a specific hotel, optionally filtered by capacity
func (r *Repository) GetBanquetsByHotel(c *fiber.Ctx) error {
	hotelCode := c.Params("hotelCode")
	capacityStr := c.Query("capacity")

	var banquets []models.BanquetHall
	query := r.DB.Where("hotel_id = ?", hotelCode)

	if capacityStr != "" {
		capacity, err := strconv.Atoi(capacityStr)
		if err == nil {
			query = query.Where("capacity >= ?", capacity)
		}
	}

	if err := query.Find(&banquets).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch banquets",
		})
	}

	return c.JSON(banquets)
}

// GetCateringByHotel retrieves catering menus for a specific hotel
func (r *Repository) GetCateringByHotel(c *fiber.Ctx) error {
	hotelCode := c.Params("hotelCode")

	var menus []models.CateringMenu
	if err := r.DB.Where("hotel_id = ?", hotelCode).Find(&menus).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch catering menus",
		})
	}

	return c.JSON(menus)
}
