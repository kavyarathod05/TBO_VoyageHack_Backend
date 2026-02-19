package main

import (
	"log"
	"math/rand"

	"github.com/akashtripathi12/TBO_Backend/internal/models"
	"gorm.io/gorm"
)

// SeedTransfers seeds global transfer options (not tied to any event)
func SeedTransfers(db *gorm.DB) {
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

			log.Printf("   ✓ Created transfer: %s (%s) - %d/%d available - ₹%.2f/cab",
				option.cabType, carModel, availableCount, totalCount, price)
		}
	}

	log.Printf("✅ Seeded %d global transfer options successfully!\n", len(transfers))
}
