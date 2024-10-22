package calendar

import (
	"time"

	"github.com/joshuscurtis/go-booking/models"
)

type CalendarDay struct {
	Number  int
	InMonth bool
	Date    time.Time
}

func GenerateCalendarDays(date time.Time) []CalendarDay {
	firstDay := time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location())
	lastDay := firstDay.AddDate(0, 1, -1)

	// Get the previous month's days that should appear
	var days []CalendarDay
	for i := int(firstDay.Weekday()); i > 0; i-- {
		prevDay := firstDay.AddDate(0, 0, -i)
		days = append(days, CalendarDay{
			Number:  prevDay.Day(),
			InMonth: false,
			Date:    prevDay,
		})
	}

	// Current month's days
	for i := 1; i <= lastDay.Day(); i++ {
		days = append(days, CalendarDay{
			Number:  i,
			InMonth: true,
			Date:    time.Date(date.Year(), date.Month(), i, 0, 0, 0, 0, date.Location()),
		})
	}

	// Next month's days to complete the grid
	daysNeeded := 42 - len(days) // 6 rows * 7 days = 42
	for i := 1; i <= daysNeeded; i++ {
		nextDay := lastDay.AddDate(0, 0, i)
		days = append(days, CalendarDay{
			Number:  nextDay.Day(),
			InMonth: false,
			Date:    nextDay,
		})
	}

	return days
}

func FilterBookingsForDate(bookings []models.Booking, date time.Time) []models.Booking {
	var filtered []models.Booking
	dateStr := date.Format("2006-01-02")
	for _, booking := range bookings {
		if booking.Date == dateStr {
			filtered = append(filtered, booking)
		}
	}
	return filtered
}
