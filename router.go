package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
)

func ticketHandler(w http.ResponseWriter, r *http.Request) {
	uidStr := r.Header.Get(httpUserHeader)
	if uidStr == "" {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = fmt.Fprint(w, "Unauthorized")
		return
	}
	uid, err := strconv.Atoi(uidStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = fmt.Fprint(w, "Bad Request")
		return
	}

	// Check wether the current user has bought a ticket
	if result, err := redisClient.HIncrBy(context.Background(), hashUser, uidStr, 1).Result(); err != nil || result != 1 {
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = fmt.Fprint(w, "Too Many Request")
		return
	}

	// pop out a ticket from queue
	ticketIdStr, err := redisClient.LPop(context.Background(), queueTicket).Result()
	if err != nil {
		log.Println(err.Error())
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, "No Ticket")
		return
	}
	ticketID, err := strconv.Atoi(ticketIdStr)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = fmt.Fprint(w, "Internal Server Error")
		return
	}

	// publish update ticket
	soldAt := time.Now()
	payload, err := json.Marshal(&Ticket{
		ID:     ticketID,
		UserID: uid,
		SoldAt: &soldAt,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = fmt.Fprint(w, "Internal Server Error")
		return
	}
	msg := message.NewMessage(watermill.NewUUID(), payload)
	if err := pubsub.Publish(updateTicketTopic, msg); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = fmt.Fprint(w, "Internal Server Error")
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprint(w, ticketIdStr)
}
