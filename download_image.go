package main

import (
	//	"bytes"
	//	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	//	"os"
	//	"path"
	//	"strings"
)

func getImage(url string) []byte {
	fmt.Println("Trying to get url: %s", url)

	resp, err := http.Get(url)
	defer resp.Body.Close()

	if err != nil {
		log.Fatal("Could not download url: %s", url)
	}

	contents, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Fatal("Could not read content !", err)
	}

	return contents
}

func main() {
	fmt.Println("Test")
	content := getImage("https://avatars2.githubusercontent.com/u/16706490?v=3&s=88")
	err := ioutil.WriteFile("picture.jpg", content, 0644)
	if err != nil {
		log.Fatal("Something went wrong")
	}
}
