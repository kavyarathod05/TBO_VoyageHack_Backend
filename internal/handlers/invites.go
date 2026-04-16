package handlers

import (
	"fmt"
	"log"

	"github.com/akashtripathi12/TBO_Backend/internal/models"
	"github.com/akashtripathi12/TBO_Backend/internal/queue"
	"github.com/akashtripathi12/TBO_Backend/internal/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// SendInvites triggers the asynchronous email invitation process
func (m *Repository) SendInvites(c *fiber.Ctx) error {
	eventID := c.Params("id")
	if _, err := uuid.Parse(eventID); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid Event ID")
	}

	// 1. Fetch Event details for the email context
	var event models.Event
	if err := m.DB.Where("id = ?", eventID).First(&event).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Event not found")
	}

	// 2. Fetch all guests for the event
	var guests []models.Guest
	if err := m.DB.Where("event_id = ?", eventID).Find(&guests).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch guests")
	}

	// 3. Group guests by FamilyID
	// FamilyID -> []Guest
	families := make(map[uuid.UUID][]models.Guest)

	// Handle guests without FamilyID (treat as single-person family)
	// We'll use a random UUID or their own ID as key if FamilyID is nil/empty
	// Assuming FamilyID is a value type UUID, checking for empty
	emptyUUID := uuid.Nil

	for _, g := range guests {
		famID := g.FamilyID
		if famID == emptyUUID {
			// Create a temporary unique key for this guest
			// In reality, we probably want to skip or handle them individually.
			// Let's treat them as their own family.
			famID = g.ID
		}
		families[famID] = append(families[famID], g)
	}

	queuedCount := 0

	// 4. Iterate and enqueue tasks
	for _, familyMembers := range families {
		// Find a representative with an email
		var targetGuest models.Guest
		found := false
		for _, member := range familyMembers {
			if member.Email != "" {
				targetGuest = member
				found = true
				break
			}
		}

		if !found {
			log.Printf("⚠️ No email found for family group (Size: %d). Skipping.", len(familyMembers))
			continue
		}

		// Construct Email Body
		// In a real app, use html/template
		param := fmt.Sprintf("family_id=%s", targetGuest.FamilyID.String())
		portalLink := fmt.Sprintf("%s/events/%s/portal/%s?%s", m.App.FrontendURL, event.ID, targetGuest.ID, param)

		subject := fmt.Sprintf("You are invited to %s!", event.Name)
		body := fmt.Sprintf(`
			<h1>Hello %s and Family,</h1>
			<p>You have been invited to <strong>%s</strong> at %s.</p>
			<p>We have arranged flights, hotels, and cabs for you.</p>
			<p>Please click the link below to view your itinerary and details:</p>
			<a href="%s" style="display:inline-block;padding:10px 20px;background-color:#007bff;color:white;text-decoration:none;border-radius:5px;">View Invitation</a>
			<p>Looking forward to seeing you!</p>
		`, targetGuest.Name, event.Name, event.Location, portalLink)

		// Create Task
		task, err := queue.NewEmailTask(targetGuest.Email, subject, body)
		if err != nil {
			log.Printf("❌ Failed to create email task: %v", err)
			continue
		}

		// Enqueue Task
		if m.QueueClient != nil {
			_, err = m.QueueClient.Enqueue(task)
			if err != nil {
				log.Printf("❌ Failed to enqueue email task: %v", err)
				continue
			}
			queuedCount++
		} else {
			log.Printf("⚠️ Queue disabled - Email skipped for: %s", targetGuest.Email)
		}
	}

	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{
		"message":       "Invitation process started",
		"emails_queued": queuedCount,
	})
}
