package main

import (
	"database/sql"
	"github.com/joshuscurtis/go-booking/internal/handlers"
	"log"
	"time"

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

func migrate(db *sql.DB) {
	sql := `
	CREATE TABLE IF NOT EXISTS bookings(
		id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		date TEXT NOT NULL,
		time_slot TEXT NOT NULL,
		name TEXT NOT NULL,
		email TEXT NOT NULL
	);

	DROP TABLE IF EXISTS time_slots;

	CREATE TABLE time_slots(
		id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		date TEXT NOT NULL,
		time TEXT NOT NULL,
		available BOOLEAN NOT NULL,
		booked_count INTEGER NOT NULL
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
