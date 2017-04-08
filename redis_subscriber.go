package main

import (
	"fmt"
	"gopkg.in/redis.v5"
)

func initRedis(host string) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     host,
		Password: "",
		DB:       0,
		PoolSize: 10,
	})

	pong, err := client.Ping().Result()

	if err != nil {
		fmt.Println("Could not ping Redis server")
	}

	fmt.Println(pong)
	fmt.Println("Succesfully pinged Redis server")
	return client
}

func runSubscriber(client *redis.Client) {
	pubsub, err := client.Subscribe("channel")

	if err != nil {
		fmt.Println(err)
	}

	for {
		msg, err := pubsub.ReceiveMessage()

		if err != nil {
			fmt.Println(err)
		} else if msg.String() == "kill" {
			fmt.Println("Killing the subscriber.")
			break
		} else {
			fmt.Println(msg)
		}
	}

	fmt.Println("Done listening")
}

func main() {
	client := initRedis("localhost:6379")

	runSubscriber(client)
}
