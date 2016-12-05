package main

import (
	"log"

	"gopkg.in/redis.v5"
)

var (
	client *redis.Client
)

func main() {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // To change based on env
		Password: "",
		DB:       0,
	})

	msg := client.RPop("test")
	if msg != nil {
		log.Println(msg)
	}

}
