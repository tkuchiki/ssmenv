package main

import (
	"log"

	"github.com/tkuchiki/ssmenv"
)

func main() {
	err := ssmenv.Run()

	if err != nil {
		log.Fatal(err)
	}

}
