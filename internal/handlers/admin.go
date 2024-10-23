// internal/handlers/admin.go
package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/joshuscurtis/go-booking/models"
	"github.com/joshuscurtis/go-booking/templates/pages"
	"github.com/joshuscurtis/go-booking/templates/partials"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// AdminMiddleware checks if the user is authenticated as admin
func AdminMiddleware(username, password string) echo.MiddlewareFunc {
	return middleware.BasicAuth(func(u, p string, c echo.Context) (bool, error) {
		if u == username && p == password {
			return true, nil
		}
		return false, nil
	})
}

func AdminDashboardHandler(c echo.Context) error {
	return pages.AdminDashboard().Render(c.Request().Context(), c.Response().Writer)
}

func AdminSlotsHandler(c echo.Context, db *sql.DB) error {
	date := c.QueryParam("date")
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}
	slots, err := getTimeSlotsFromDB(db, date)
	if err != nil {
		return err
	}
	return partials.AdminSlots(slots, date).Render(c.Request().Context(), c.Response().Writer)
}

func AdminCreateSlotHandler(c echo.Context, db *sql.DB) error {
	date := c.FormValue("date")
	timeSlot := c.FormValue("time")
	capacity := c.FormValue("capacity")

	if date == "" || timeSlot == "" || capacity == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "All fields are required"})
	}

	capacityInt, err := strconv.Atoi(capacity)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid capacity value"})
	}

	_, err = db.Exec(`
		INSERT INTO time_slots (date, time, available, capacity, booked_count) 
		VALUES (?, ?, true, ?, 0)
	`, date, timeSlot, capacityInt)
	if err != nil {
		log.Printf("Error creating time slot: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create time slot"})
	}

	slots, err := getTimeSlotsFromDB(db, date)
	if err != nil {
		return err
	}
	return partials.AdminSlots(slots, date).Render(c.Request().Context(), c.Response().Writer)
}

func AdminDeleteSlotHandler(c echo.Context, db *sql.DB) error {
	date := c.FormValue("date")
	timeSlot := c.FormValue("time")

	_, err := db.Exec("DELETE FROM time_slots WHERE date = ? AND time = ?", date, timeSlot)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete time slot"})
	}

	slots, err := getTimeSlotsFromDB(db, date)
	if err != nil {
		return err
	}
	return partials.AdminSlots(slots, date).Render(c.Request().Context(), c.Response().Writer)
}

func AdminBookingsHandler(c echo.Context, db *sql.DB) error {
	bookings, err := getAllBookings(db)
	if err != nil {
		return err
	}
	return partials.AdminBookings(bookings).Render(c.Request().Context(), c.Response().Writer)
}

func AdminCancelBookingHandler(c echo.Context, db *sql.DB) error {
	bookingIDStr := c.FormValue("booking_id")
	if bookingIDStr == "" {
		return c.HTML(http.StatusBadRequest, "<div id='calendar-view-container'>Booking ID is required</div>")
	}

	// Convert string ID to integer
	bookingID, err := strconv.Atoi(bookingIDStr)
	if err != nil {
		log.Printf("Error converting booking ID: %v", err)
		return c.HTML(http.StatusBadRequest, "<div id='calendar-view-container'>Invalid booking ID</div>")
	}

	tx, err := db.Begin()
	if err != nil {
		log.Printf("Error beginning transaction: %v", err)
		return c.HTML(http.StatusInternalServerError, "<div id='calendar-view-container'>Database error</div>")
	}
	defer tx.Rollback()

	// Get booking details before deletion
	var date string
	var timeSlot string
	err = tx.QueryRow("SELECT date, time_slot FROM bookings WHERE id = ?", bookingID).Scan(&date, &timeSlot)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.HTML(http.StatusNotFound, "<div id='calendar-view-container'>Booking not found</div>")
		}
		log.Printf("Error fetching booking details: %v", err)
		return c.HTML(http.StatusInternalServerError, "<div id='calendar-view-container'>Database error</div>")
	}

	// Delete the booking
	_, err = tx.Exec("DELETE FROM bookings WHERE id = ?", bookingID)
	if err != nil {
		log.Printf("Error deleting booking: %v", err)
		return c.HTML(http.StatusInternalServerError, "<div id='calendar-view-container'>Failed to delete booking</div>")
	}

	// Update time slot availability
	_, err = tx.Exec(`
        UPDATE time_slots 
        SET booked_count = booked_count - 1,
            available = CASE 
                WHEN booked_count - 1 < capacity THEN true 
                ELSE false 
            END
        WHERE date = ? AND time = ?
    `, date, timeSlot)
	if err != nil {
		log.Printf("Error updating time slot: %v", err)
		return c.HTML(http.StatusInternalServerError, "<div id='calendar-view-container'>Failed to update time slot</div>")
	}

	if err = tx.Commit(); err != nil {
		log.Printf("Error committing transaction: %v", err)
		return c.HTML(http.StatusInternalServerError, "<div id='calendar-view-container'>Database error</div>")
	}

	// Parse the date to get the current month for the calendar view
	parsedDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		parsedDate = time.Now()
	}

	// Get all bookings for the current month
	firstDay := time.Date(parsedDate.Year(), parsedDate.Month(), 1, 0, 0, 0, 0, parsedDate.Location())
	lastDay := firstDay.AddDate(0, 1, -1)

	bookings, err := getBookingsForDateRange(db, firstDay, lastDay)
	if err != nil {
		log.Printf("Error fetching updated bookings: %v", err)
		return c.HTML(http.StatusInternalServerError, "<div id='calendar-view-container'>Failed to fetch updated bookings</div>")
	}

	// Set Content-Type header
	c.Response().Header().Set("Content-Type", "text/html")

	// Render the updated calendar view with the container div
	return partials.CalendarView(bookings, parsedDate).Render(c.Request().Context(), c.Response().Writer)
}

func getAllBookings(db *sql.DB) ([]models.Booking, error) {
	rows, err := db.Query(`
		SELECT id, date, time_slot, name, email 
		FROM bookings 
		ORDER BY date DESC, time_slot ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bookings []models.Booking
	for rows.Next() {
		var b models.Booking
		if err := rows.Scan(&b.ID, &b.Date, &b.TimeSlot, &b.Name, &b.Email); err != nil {
			return nil, err
		}
		bookings = append(bookings, b)
	}
	return bookings, nil
}
