package main

import (
	"bytes"
	"github.com/loldesign/azure"
	"gopkg.in/redis.v5"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)
var (
	container      = os.Getenv("AZURE_IMAGE_CONTAINER")
	acc_name       = os.Getenv("AZURE_ACCOUNT_NAME")
	acc_key        = os.Getenv("AZURE_ACCOUNT_KEY")
	channel        = os.Getenv("REDIS_CHANNEL")
	redis_host     = os.Getenv("REDIS_HOST")
	redis_password = os.Getenv("REDIS_PASSWORD")
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

func initAzure(acc_name string, acc_key string) azure.Azure {
	client := azure.New(acc_name, acc_key)
	return client
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

	// Test Azure
	clientAzure := initAzure(os.Getenv("AZURE_ACCOUNT"), os.Getenv("AZURE_KEY"))
	// End test Azure

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

	// Test imageContentToByteArray to azure upload
	contentReaderAzure := imageContentToByteArray("http://redis.io/images/redis-white.png")
	res, err := clientAzure.FileUpload("images", "test_golang.png", contentReaderAzure)

	if err != nil {
		log.Fatal("Could not upload to blob", err)
	}

	log.Println(res)
}
