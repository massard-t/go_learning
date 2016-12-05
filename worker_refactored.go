package main

import (
	"bytes"
	_ "expvar"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"strings"

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

type job struct {
	azure_manager storage.BlobStorageClient
	payload       string
}

type worker struct {
	id int
}

func (w worker) process(j job) {
	fmt.Printf("worker%d: started", w.id)
	// Do request with job.payload
	makeMagicHappen(j)
	fmt.Printf("worker%d: completed", w.id)
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
	log.Println("[CONFIG] Redis channel: ", channel)

	pubsub, err := client.Subscribe(channel)

	if err != nil {
		log.Fatal("[ERROR] Could not listen to channel: ", err)
	}

	return pubsub
}

func initAzure() storage.BlobStorageClient {
	log.Println("[CONFIG] Azure account name: ", acc_name)

	client, err := storage.NewBasicClient(acc_name, acc_key)

	if err != nil {
		log.Fatal("[ERROR]Could not reach Azure ", err)
	}

	blob_service := client.GetBlobService()
	return blob_service
}

func getImage(url string) *bytes.Reader {
	resp, err := http.Get(url)

	if err != nil {
		log.Println("[ERROR]Could not download url: ", err)
	}

	defer resp.Body.Close()

	contents, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Println("[ERROR]Could not read content", err)
	}

	r := bytes.NewReader(contents)
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

func getUrlAndDest(msg string) (string, string) {
	splitted_msg := strings.Split(msg, "@@")
	url, dest := splitted_msg[0], splitted_msg[1]
	dest = strings.TrimPrefix(dest, "blob://")
	return url, dest
}

func makeMagicHappen(j job) {
	url, dest := getUrlAndDest(j.payload)
	if url != "" && dest != "" {
		content := getImage(url)
		bytesToAzure(j.azure_manager, content, dest)
	}
}
func requestHandler(jobCh chan job, blob_service storage.BlobStorageClient, p string) {

	// Create Job and push the work onto the jobCh.
	job := job{blob_service, p}
	go func() {
		fmt.Printf("added: %s\n", job.payload)
		jobCh <- job
	}()

	return
}

func getOrder(client *redis.Client) string {
	msg := client.RPop(channel)
	return msg.String()
}

func main() {
	var (
		maxQueueSize = flag.Int("max_queue_size", 100, "The size of job queue")
		maxWorkers   = flag.Int("max_workers", 5, "The number of workers to start")
	)
	flag.Parse()

	redis_client := initRedis(redis_host, redis_password)
	//pubsub := initSubscriber(redis_client)

	blob_service := initAzure()
	// create job channel
	jobCh := make(chan job, *maxQueueSize)

	// create workers
	for i := 0; i < *maxWorkers; i++ {
		w := worker{i}
		go func(w worker) {
			for j := range jobCh {
				w.process(j)
			}
		}(w)
	}

	// handler for adding jobs
	// HANDLE REQUESTS WITH REDIS
	for {
		msg := getOrder(redis_client)

		//	if err != nil {
		//		log.Println("[DEBUG] No message")
		//	} else if msg == "kill" {
		if msg == "kill" {
			log.Println("[INFO] Killing the Subscriber.")
			break
		} else {
			log.Println(msg)
			requestHandler(jobCh, blob_service, msg)
		}
	}
	log.Println("[INFO] Done listening, exiting Program.")
}
