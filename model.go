package main

import "time"

// Ticket is the ticket type
type Ticket struct {
	ID        int        `gorm:"primaryKey"` // auto-increment
	UserID    int        // bigint
	CreatedAt int64      `gorm:"autoUpdateTime:milli"`
	SoldAt    *time.Time // use pointer to avoid zero-value field
}
