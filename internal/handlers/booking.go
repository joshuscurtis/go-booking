package handlers

import (
	"database/sql"
	"htmx-templ-echo-project/models"
	"htmx-templ-echo-project/templates/pages"
	"htmx-templ-echo-project/templates/partials"
	"log"
	"time"

	"github.com/labstack/echo/v4"
)

func BookingHandler(c echo.Context) error {
	return pages.Booking().Render(c.Request().Context(), c.Response().Writer)
}

func BookingFormHandler(c echo.Context, db *sql.DB) error {
	date := c.QueryParam("date")
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}
	timeSlots, err := getTimeSlotsFromDB(db, date)
	if err != nil {
		log.Printf("Error fetching time slots: %v", err)
		return c.String(500, "An error occurred while fetching time slots")
	}
	return partials.BookingForm(timeSlots).Render(c.Request().Context(), c.Response().Writer)
}

func TimeSlotHandler(c echo.Context, db *sql.DB) error {
	date := c.QueryParam("date")
	if date == "" {
		return c.String(400, "Date is required")
	}
	timeSlots, err := getTimeSlotsFromDB(db, date)
	if err != nil {
		log.Printf("Error fetching time slots: %v", err)
		return c.String(500, "An error occurred while fetching time slots")
	}
	return partials.TimeSlotOptions(timeSlots).Render(c.Request().Context(), c.Response().Writer)
}

func BookingSubmitHandler(c echo.Context, db *sql.DB) error {
	date := c.FormValue("date")
	timeSlot := c.FormValue("time-slot")
	name := c.FormValue("name")
	email := c.FormValue("email")

	log.Printf("Received booking request - Date: %s, Time: %s, Name: %s, Email: %s", date, timeSlot, name, email)

	if date == "" || timeSlot == "" || name == "" || email == "" {
		return partials.BookingResponse(false, "Please fill in all fields.").Render(c.Request().Context(), c.Response().Writer)
	}

	err := createBooking(db, date, timeSlot, name, email)
	if err != nil {
		log.Printf("Error creating booking: %v", err)
		return partials.BookingResponse(false, "An error occurred while booking. Please try again.").Render(c.Request().Context(), c.Response().Writer)
	}

	return partials.BookingResponse(true, "Your booking has been confirmed. We look forward to seeing you!").Render(c.Request().Context(), c.Response().Writer)
}

func BookingsDataHandler(c echo.Context, db *sql.DB) error {
	currentDate := time.Now()
	startOfMonth := time.Date(currentDate.Year(), currentDate.Month(), 1, 0, 0, 0, 0, currentDate.Location())
	endOfMonth := startOfMonth.AddDate(0, 1, -1)

	bookings, err := getBookingsForDateRange(db, startOfMonth, endOfMonth)
	if err != nil {
		log.Printf("Error fetching bookings: %v", err)
		return c.JSON(500, map[string]string{"error": "An error occurred while fetching bookings"})
	}
	return c.JSON(200, bookings)
}

func CalendarViewHandler(c echo.Context, db *sql.DB) error {
	dateStr := c.QueryParam("date")
	var currentDate time.Time
	var err error

	if dateStr == "" {
		currentDate = time.Now()
	} else {
		currentDate, err = time.Parse("2006-01", dateStr)
		if err != nil {
			currentDate = time.Now()
		}
	}

	firstDay := time.Date(currentDate.Year(), currentDate.Month(), 1, 0, 0, 0, 0, currentDate.Location())
	lastDay := firstDay.AddDate(0, 1, -1)

	bookings, err := getBookingsForDateRange(db, firstDay, lastDay)
	if err != nil {
		return err
	}

	return partials.CalendarView(bookings, currentDate).Render(c.Request().Context(), c.Response().Writer)
}

func createBooking(db *sql.DB, date, timeSlot, name, email string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	available, err := isTimeSlotAvailable(db, date, timeSlot)
	if err != nil {
		return err
	}
	if !available {
		return echo.NewHTTPError(400, "This time slot is no longer available. Please choose another.")
	}

	_, err = tx.Exec("INSERT INTO bookings (date, time_slot, name, email) VALUES (?, ?, ?, ?)", date, timeSlot, name, email)
	if err != nil {
		return err
	}

	result, err := tx.Exec("UPDATE time_slots SET available = false, booked_count = booked_count + 1 WHERE date = ? AND time = ?", date, timeSlot)
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return echo.NewHTTPError(400, "The selected time slot is not available. Please choose another.")
	}

	return tx.Commit()
}

func getTimeSlotsFromDB(db *sql.DB, date string) ([]models.TimeSlot, error) {
	rows, err := db.Query("SELECT time, available, booked_count FROM time_slots WHERE date = ?", date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var timeSlots []models.TimeSlot
	for rows.Next() {
		var ts models.TimeSlot
		err := rows.Scan(&ts.Time, &ts.Available, &ts.BookedCount)
		if err != nil {
			return nil, err
		}
		timeSlots = append(timeSlots, ts)
	}

	if len(timeSlots) == 0 {
		log.Printf("No time slots found for date: %s", date)
	}

	return timeSlots, nil
}

func isTimeSlotAvailable(db *sql.DB, date string, timeSlot string) (bool, error) {
	var available bool
	err := db.QueryRow("SELECT available FROM time_slots WHERE date = ? AND time = ?", date, timeSlot).Scan(&available)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("No time slot found for date: %s and time: %s", date, timeSlot)
			return false, echo.NewHTTPError(404, "Time slot not found")
		}
		return false, err
	}
	return available, nil
}

func getBookingsForDateRange(db *sql.DB, start, end time.Time) ([]models.Booking, error) {
	rows, err := db.Query("SELECT id, date, time_slot, name, email FROM bookings WHERE date BETWEEN ? AND ? ORDER BY date, time_slot", start.Format("2006-01-02"), end.Format("2006-01-02"))
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
