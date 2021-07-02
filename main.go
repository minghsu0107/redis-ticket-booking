package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
)

func init() {
	initMySQL()
	initRedis()
	initPubSub()
	prepareData()
}

func main() {
	go func() {
		if err := pubsubRouter.Run(context.Background()); err != nil {
			log.Fatal(err)
		}
	}()
	if httpPort == "" {
		httpPort = "8080"
	}
	http.HandleFunc("/ticket", ticketHandler)
	fmt.Printf("listen on %s\n", httpPort)
	fmt.Println(http.ListenAndServe(fmt.Sprintf(":%s", httpPort), nil))
}
