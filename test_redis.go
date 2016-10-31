package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"

	"gopkg.in/redis.v5"
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

func imageContentToByteArray(url string) *bytes.Reader {
	content := getImage(url)
	r := bytes.NewReader(content)
	return r
}

func main() {
	// Test redis
	client := initRedis("localhost:6379", "", 0, 10)
	client.Close()
	// End test redis

	// Test getImage to file
	err := ioutil.WriteFile("Somefile.jpg", getImage("http://redis.io/images/redis-white.png"), 0644)

	if err != nil {
		log.Fatal("Could not save image", err)
	}
	// End test getImage to file

	// Test getImage to byte.Reader
	contentReader := imageContentToByteArray("http://redis.io/images/redis-white.png")

	toWrite, err := ioutil.ReadAll(contentReader)

	if err != nil {
		log.Fatal("Could not read bytes from contentReader", err)
	}

	err = ioutil.WriteFile("file_using_bytes.Reader.png", toWrite, 0644)

	if err != nil {
		log.Fatal("Could not save image using bytes reader", err)
	}
	// End test getImage to byte.Reader

}
