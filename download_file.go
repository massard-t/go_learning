package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/Azure/azure-sdk-for-go/storage"
	"gopkg.in/redis.v5"
)

var (
	container      = os.Getenv("AZURE_IMAGE_CONTAINER")
	acc_name       = os.Getenv("AZURE_ACCOUNT_NAME")
	acc_key        = os.Getenv("AZURE_ACCOUNT_KEY")
	channel        = os.Getenv("REDIS_CHANNEL")
	redis_host     = os.Getenv("REDIS_HOST")
	redis_password = os.Getenv("REDIS_PASSWORD")
)

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

func initAzure(acc_name string, acc_key string) storage.BlobStorageClient {
	client, err := storage.NewBasicClient(acc_name, acc_key)

	if err != nil {
		log.Fatal("Could not reach Azure", err)
	}

	return client.GetBlobService()
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

func bytesToAzure(client storage.BlobStorageClient, content *bytes.Reader, dest string) {
	log.Println(dest)
	m := make(map[string]string)
	readerSize := uint64(content.Size())
	if readerSize != 0 {
		err := client.CreateBlockBlobFromReader(container, dest, readerSize, content, m)

		if err != nil {
			log.Fatal("[ERROR] Could not upload image: ", err)
		} else {
			log.Println("[SUCCESS] Destination: ", dest)
		}
	} else {
		log.Println("[ERROR]Empty content.")
	}
}

func main() {
	log.Println(acc_name, acc_key, container)
	// Test redis
	client := initRedis("localhost:6379", "", 0, 10)
	client.Close()
	// End test redis

	// Test Azure
	clientAzure := initAzure(acc_name, acc_key)
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

	bytesToAzure(clientAzure, contentReaderAzure, "test.png")

	if err != nil {
		log.Fatal("Could not upload to blob", err)
	}

	log.Println("Done")
}
