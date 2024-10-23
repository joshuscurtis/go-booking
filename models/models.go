package models

type TimeSlot struct {
	Time        string
	Available   bool
	Capacity    int
	BookedCount int
}

type Booking struct {
	ID       int
	Date     string
	TimeSlot string
	Name     string
	Email    string
}
