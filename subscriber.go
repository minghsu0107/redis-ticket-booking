package main

import (
	"encoding/json"
	"log"

	"github.com/ThreeDotsLabs/watermill/message"
)

type CheckTicket struct {
	UserID int
}

func updateTicket(msg *message.Message) error {
	var ticket Ticket
	if err := json.Unmarshal(msg.Payload, &ticket); err != nil {
		return err
	}
	var checkTicket CheckTicket
	if err := db.Model(&Ticket{}).Select("user_id").Where("id = ?", ticket.ID).First(&checkTicket).Error; err != nil {
		return err
	}
	if checkTicket.UserID != dummyUser {
		log.Fatalf("Race condition happened. Ticket %d has been sold to %d\n", ticket.ID, checkTicket.UserID)
		return nil
	}
	// mark the ticket taken
	if err := db.Model(&Ticket{}).Where("id = ?", ticket.ID).
		Updates(
			Ticket{
				UserID: ticket.UserID,
				SoldAt: ticket.SoldAt,
			}).Error; err != nil {
		return err
	}
	log.Printf("Ticket %d is sold to user %d", ticket.ID, ticket.UserID)
	return nil
}
