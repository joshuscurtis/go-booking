package handlers

import (
	"database/sql"
	"fmt"
	"htmx-templ-echo-project/templates/pages"
	"htmx-templ-echo-project/templates/partials"

	"github.com/labstack/echo/v4"
)

func RegisterHandlers(e *echo.Echo, db *sql.DB) {
	e.GET("/", HomeHandler)
	e.GET("/about", AboutHandler)
	e.GET("/contact", ContactHandler)
	e.GET("/partials/greeting", GreetingHandler)
	e.POST("/partials/contact", ContactFormHandler)
	e.POST("/partials/bookings", func(c echo.Context) error {
		return BookingSubmitHandler(c, db)
	})
	e.GET("/booking", func(c echo.Context) error { return pages.Booking().Render(c.Request().Context(), c.Response().Writer) })
	e.GET("/partials/booking-form", func(c echo.Context) error { return BookingFormHandler(c, db) })
	e.GET("/partials/time-slots", func(c echo.Context) error { return TimeSlotHandler(c, db) })
	e.POST("/partials/submit-booking", func(c echo.Context) error { return BookingSubmitHandler(c, db) })
	e.GET("/partials/bookings-data", func(c echo.Context) error { return BookingsDataHandler(c, db) })

	e.GET("/partials/calendar-view", func(c echo.Context) error {
		return CalendarViewHandler(c, db)
	})
}

func HomeHandler(c echo.Context) error {
	return pages.Home().Render(c.Request().Context(), c.Response().Writer)
}

func AboutHandler(c echo.Context) error {
	return pages.About().Render(c.Request().Context(), c.Response().Writer)
}

func ContactHandler(c echo.Context) error {
	return pages.Contact().Render(c.Request().Context(), c.Response().Writer)
}

func GreetingHandler(c echo.Context) error {
	return partials.Greeting().Render(c.Request().Context(), c.Response().Writer)
}

func ContactFormHandler(c echo.Context) error {
	name := c.FormValue("name")
	email := c.FormValue("email")
	message := c.FormValue("message")

	if name == "" || email == "" || message == "" {
		return partials.FormResponse(false, "Please fill in all fields.").Render(c.Request().Context(), c.Response().Writer)
	}
	fmt.Println("name: ", name)
	fmt.Println("email: ", email)
	fmt.Println("message: ", message)

	return partials.FormResponse(true, "Thank you for your message. We'll get back to you soon!").Render(c.Request().Context(), c.Response().Writer)
}
