package main

import (
	"database/sql"
	"log"
	"time"

	"github.com/joshuscurtis/go-booking/internal/handlers"
	"github.com/joshuscurtis/go-booking/models"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/mattn/go-sqlite3"
)

func initDB(filepath string) *sql.DB {
	db, err := sql.Open("sqlite3", filepath)
	if err != nil {
		log.Fatal(err)
	}
	if db == nil {
		log.Fatal("db nil")
	}
	return db
}

// In main.go, update the migrate function

func migrate(db *sql.DB) {
	sql := `
    CREATE TABLE IF NOT EXISTS bookings(
        id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
        date TEXT NOT NULL,
        time_slot TEXT NOT NULL,
        name TEXT NOT NULL,
        email TEXT NOT NULL
    );

    CREATE TABLE IF NOT EXISTS tickets (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      booking_id INTEGER NOT NULL,
      ticket_id TEXT NOT NULL UNIQUE,
      created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
      FOREIGN KEY (booking_id) REFERENCES bookings(id)
  );


    DROP TABLE IF EXISTS time_slots;

    CREATE TABLE time_slots(
        id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
        date TEXT NOT NULL,
        time TEXT NOT NULL,
        available BOOLEAN NOT NULL,
        capacity INTEGER NOT NULL DEFAULT 5,
        booked_count INTEGER NOT NULL DEFAULT 0
    );
    `
	_, err := db.Exec(sql)
	if err != nil {
		log.Fatal(err)
	}
}

func seedTimeSlots(db *sql.DB) {
	timeSlots := []string{"09:00", "10:00", "11:00", "12:00", "13:00", "14:00", "15:00", "16:00", "17:00"}
	// Generate time slots for the next 7 days
	for i := 0; i < 7; i++ {
		date := time.Now().AddDate(0, 0, i).Format("2006-01-02")
		for _, slot := range timeSlots {
			_, err := db.Exec("INSERT INTO time_slots (date, time, available, booked_count) VALUES (?, ?, ?, ?)", date, slot, true, 0)
			if err != nil {
				log.Printf("Error seeding time slot %s on %s: %v", slot, date, err)
			}
		}
	}
	log.Println("Time slots seeded successfully")
}

const ticketSchema = `
CREATE TABLE IF NOT EXISTS tickets (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    booking_id INTEGER NOT NULL,
    ticket_id TEXT NOT NULL UNIQUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (booking_id) REFERENCES bookings(id)
);
`

func storeTicketID(db *sql.DB, bookingID int, ticketID string) error {
	_, err := db.Exec("INSERT INTO tickets (booking_id, ticket_id) VALUES (?, ?)", bookingID, ticketID)
	return err
}

func getBookingByTicketID(db *sql.DB, ticketID string) (models.Booking, error) {
	var booking models.Booking
	err := db.QueryRow(`
        SELECT b.* FROM bookings b
        JOIN tickets t ON b.id = t.booking_id
        WHERE t.ticket_id = ?
    `, ticketID).Scan(&booking.ID, &booking.Date, &booking.TimeSlot, &booking.Name, &booking.Email)
	return booking, err
}

func main() {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Initialize SQLite
	db := initDB("bookings.db")
	migrate(db)
	seedTimeSlots(db)
	defer db.Close()

	handlers.RegisterHandlers(e, db)

	log.Fatal(e.Start(":8080"))
}
