package main

import (
	"flag"
	"log"
	"math/rand"
	"time"

	"github.com/noppawitt/pammy"
)

func main() {
	var token string

	flag.StringVar(&token, "t", "", "Bot Token")
	flag.Parse()

	rand.Seed(time.Now().UnixNano())

	pammy, err := pammy.NewClient(token)
	if err != nil {
		log.Fatal(err)
	}

	if err := pammy.Start(); err != nil {
		log.Fatal(err)
	}
}
