package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
	"github.com/go-redis/redis/v8"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

const (
	// QueueTicket is the Ticket queue key in redis
	queueTicket = "queue:ticket"
	// hashUser is the user hash key in redis
	hashUser          = "hash:user"
	updateTicketTopic = "ticket.update"
	httpUserHeader    = "X-User-Id"
	dummyUser         = 0
	concurrency       = 100
)

var (
	logger = watermill.NewStdLogger(false, false)

	db            *gorm.DB
	redisClient   *redis.Client
	mysqlHost     = os.Getenv("MYSQL_HOST")
	mysqlUser     = os.Getenv("MYSQL_USER")
	mysqlPassword = os.Getenv("MYSQL_PASSWORD")
	mysqlDatabase = os.Getenv("MYSQL_DATABASE")
	redisHost     = os.Getenv("REDIS_HOST")
	httpPort      = os.Getenv("HTTP_PORT")
	pubsub        *gochannel.GoChannel
	pubsubRouter  *message.Router
)

func initMySQL() {
	var err error
	db, err = gorm.Open(mysql.Open(fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", mysqlUser, mysqlPassword, mysqlHost, mysqlDatabase)), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal(err)
	}
	sqlDB.SetMaxOpenConns(concurrency)
	fmt.Println("database connected")

	db.AutoMigrate(&Ticket{})
}

func initRedis() {
	redisClient = redis.NewClient(&redis.Options{
		Addr:     redisHost,
		DB:       0,
		PoolSize: concurrency,
	})
	if _, err := redisClient.Ping(context.Background()).Result(); err != nil {
		log.Fatal(err)
	}
	fmt.Println("redis connected. pool size 100")
}

func initPubSub() {
	pubsub = gochannel.NewGoChannel(gochannel.Config{Persistent: true}, logger)
	var err error
	pubsubRouter, err = message.NewRouter(message.RouterConfig{}, logger)
	if err != nil {
		log.Fatal(err)
	}
	pubsubRouter.AddNoPublisherHandler(
		updateTicketTopic+"_handler",
		updateTicketTopic,
		pubsub,
		updateTicket,
	)
}

func prepareData() {
	tickets := make([]Ticket, 0, concurrency)
	if err := db.Debug().Where("user_id = ?", dummyUser).Limit(concurrency).Find(&tickets).Error; err != nil {
		log.Fatal(err)
	}
	if len(tickets) < concurrency {
		neededTickets := concurrency - len(tickets)
		for i := 0; i < neededTickets; i++ {
			err := db.Create(&Ticket{
				UserID: dummyUser,
			}).Error
			if err != nil {
				log.Fatal(err)
			}
		}
		if err := db.Debug().Where("user_id = ?", dummyUser).Limit(concurrency).Find(&tickets).Error; err != nil {
			log.Fatal(err)
		}
	}
	redisTiketLen, err := redisClient.LLen(context.Background(), queueTicket).Result()
	if err != nil {
		log.Fatal(err)
	}
	if redisTiketLen != concurrency {
		if err := redisClient.Del(context.Background(), queueTicket).Err(); err != nil {
			log.Fatal(err)
		}
		for i := range tickets {
			if err := redisClient.LPush(context.Background(), queueTicket, tickets[i].ID).Err(); err != nil {
				log.Fatal(err)
			}
		}
		fmt.Printf("loaded %d tickets to redis\n", concurrency)
	}
}
