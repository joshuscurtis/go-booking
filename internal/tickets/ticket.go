package tickets

import (
	"encoding/base64"
	"fmt"

	"github.com/google/uuid"
	"github.com/joshuscurtis/go-booking/models"
	"github.com/skip2/go-qrcode"
)

type TicketData struct {
	TicketID   string
	BookingID  int
	Date       string
	Time       string
	Name       string
	Email      string
	VerifyURL  string
	QRCodeData string
}

func GenerateTicket(booking models.Booking, baseURL string) (TicketData, error) {
	// Generate unique ticket ID
	ticketID := uuid.New().String()

	// Create verification URL
	verifyURL := fmt.Sprintf("%s/verify-ticket/%s", baseURL, ticketID)

	// Generate QR code
	qr, err := qrcode.New(verifyURL, qrcode.Medium)
	if err != nil {
		return TicketData{}, err
	}

	// Convert QR code to base64 string
	png, err := qr.PNG(256)
	if err != nil {
		return TicketData{}, err
	}
	qrBase64 := base64.StdEncoding.EncodeToString(png)

	return TicketData{
		TicketID:   ticketID,
		BookingID:  booking.ID,
		Date:       booking.Date,
		Time:       booking.TimeSlot,
		Name:       booking.Name,
		Email:      booking.Email,
		VerifyURL:  verifyURL,
		QRCodeData: qrBase64,
	}, nil
}
