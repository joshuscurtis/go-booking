package models

type TimeSlot struct {
	Time        string
	Available   bool
	BookedCount int
}

type DateAvailability struct {
	Date      string
	Available bool
}

type Booking struct {
	ID       int
	Date     string
	TimeSlot string
	Name     string
	Email    string
}
