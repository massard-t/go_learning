package main

import (
	"gopkg.in/redis.v5"
	"io/ioutil"
	"log"
	"net/http"
)

type processFunc func(string, string)

func getImage(url string) []byte {
	log.Println("Trying to get url: ", url)

	resp, err := http.Get(url)
	defer resp.Body.Close()

	if err != nil {
		log.Fatal("Could not download url: ", url)
	}

	contents, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Fatal("Could not read content", err)
	}

	return contents
}

func initRedis(host string, password string, db int, poolsize int) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     host,
		Password: password,
		DB:       db,
		PoolSize: poolsize,
	})

	pong, err := client.Ping().Result()

	if err != nil {
		log.Fatal("Could not reach Redis server: ", err)
	}

	log.Println("Pinged redis: ", pong)
	return client
}

func main() {
	client := initRedis("localhost:6379", "", 0, 10)
	client.Close()
	err := ioutil.WriteFile("Somefile.jpg", getImage("http://redis.io/images/redis-white.png"), 0644)

	if err != nil {
		log.Fatal("Could not save image", err)
	}

}
