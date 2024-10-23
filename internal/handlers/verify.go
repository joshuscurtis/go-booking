package handlers

import (
	"database/sql"
	"fmt"

	"github.com/joshuscurtis/go-booking/models"
	"github.com/joshuscurtis/go-booking/templates/partials"
	"github.com/labstack/echo/v4"
)

func VerifyTicketHandler(c echo.Context, db *sql.DB) error {
	ticketID := c.Param("ticketID")
	if ticketID == "" {
		return partials.VerifyTicketResponse(false, "Invalid ticket ID", nil).Render(c.Request().Context(), c.Response().Writer)
	}

	// Query the database for the ticket
	var booking models.Booking
	err := db.QueryRow(`
        SELECT b.* FROM bookings b
        JOIN tickets t ON b.id = t.booking_id
        WHERE t.ticket_id = ?
    `, ticketID).Scan(&booking.ID, &booking.Date, &booking.TimeSlot, &booking.Name, &booking.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			return partials.VerifyTicketResponse(false, "Ticket not found", nil).Render(c.Request().Context(), c.Response().Writer)
		}
		return partials.VerifyTicketResponse(false, "Error verifying ticket", nil).Render(c.Request().Context(), c.Response().Writer)
	}

	fmt.Println("booking", booking)
	return partials.VerifyTicketResponse(true, "Valid ticket", &booking).Render(c.Request().Context(), c.Response().Writer)
}
