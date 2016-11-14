package main

import (
	"bytes"
	"gopkg.in/redis.v5"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"github.com/Azure/azure-sdk-for-go/storage"
)

var (
	container = os.Getenv("AZURE_IMAGE_CONTAINER")
)

func getImage(url string) *bytes.Reader {
	resp, err := http.Get(url)

	if err != nil {
		log.Fatal("[ERROR]Could not download url: ", err)
	}

	defer resp.Body.Close()

	contents, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Fatal("[ERROR]Could not read content", err)
	}

	r := bytes.NewReader(contents)
	return r
}

func initAzure() storage.Client{
	acc_name := os.Getenv("AZURE_ACCOUNT")
	acc_key := os.Getenv("AZURE_KEY")

	log.Println("[CONFIG] Azure account name: ", acc_name)

	client := storage.NewBasicClient(acc_name, acc_key)

	return client
}

func initRedis(host string, password string) *redis.Client {
	log.Println("[CONFIG] Redis host: ", host)
	log.Println("[CONFIG] Redis password: ", password)

	client := redis.NewClient(&redis.Options{
		Addr:     host,
		Password: password,
		DB:       0,
		PoolSize: 10,
	})

	_, err := client.Ping().Result()

	if err != nil {
		log.Fatal("[ERROR] Could not reach Redis host: ", err)
	}

	return client
}

func initSubscriber(client *redis.Client) *redis.PubSub {
	channel := os.Getenv("REDIS_CHANNEL")

	log.Println("[CONFIG] Redis channel: ", channel)

	pubsub, err := client.Subscribe(channel)

	if err != nil {
		log.Fatal("[ERROR] Could not listen to channel: ", err)
	}

	return pubsub
}

func bytesToAzure(client azure.Azure, content *bytes.Reader, dest string) {
	log.Println(dest)
	resp, err := client.FileUpload(container, dest, content)

	if err != nil {
		log.Fatal("[ERROR] Could not upload image: ", err)
	} else {
		log.Println("[SUCCESS] Destination: ", dest)
		log.Println(resp)
	}
}

func getUrlAndDest(msg string) (string, string) {
	splitted_msg := strings.Split(msg, "@@")
	url, dest := splitted_msg[0], splitted_msg[1]
	dest = strings.TrimPrefix(dest, "blob://")
	return url, dest
}

func makeMagicHappen(client azure.Azure, msg string) {
	url, dest := getUrlAndDest(msg)
	content := getImage(url)
	bytesToAzure(client, content, dest)
}

func runDownloader(pubsub *redis.PubSub, blob_service azure.Azure) {
	for {
		msg, err := pubsub.ReceiveMessage()

		if err != nil {
			log.Println("[DEBUG] No message")
		} else if msg.Payload == "kill" {
			log.Println("[INFO] Killing the subcriber.")
			break
		} else {
			log.Println(msg.Payload)
			makeMagicHappen(blob_service, msg.Payload)
		}
	}
	log.Println("[INFO] Done listening, exiting program.")
}

func main() {
	redis_host := os.Getenv("REDIS_HOST")
	redis_password := os.Getenv("REDIS_PASSWORD")
	log.Println("#####################################")
	log.Println("           Configuration             ")
	log.Println("#####################################")
	redis_client := initRedis(redis_host, redis_password)

	pubsub := initSubscriber(redis_client)

	blob_service := initAzure()
	runDownloader(pubsub, blob_service)
}
