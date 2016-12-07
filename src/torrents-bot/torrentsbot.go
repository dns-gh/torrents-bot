package main

import (
	"log"
	"os"

	"github.com/dns-gh/t411-client/t411client"
)

func main() {
	_, err := t411client.NewT411Client("", os.Getenv("T411_USERNAME"), os.Getenv("T411_PASSWORD"))
	if err != nil {
		log.Fatalln(err.Error())
	}
	log.Println("success")
}
