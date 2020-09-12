package main

import (
	"log"
	"os"

	"github.com/cesarlugoe/concurrent_numbers/handler"
)

type counter struct {
	uniqueNumbers    int
	duplicateNumbers int
}

func main() {

	file, err := os.Create("./numbers.log")
	if err != nil {
		log.Fatalln(err)
		os.Exit(1)
	}

	h := handler.New()
	li, err := h.StartServer("4000")
	if err != nil {
		log.Fatalln(err)
	}

	defer li.Close()

	go h.ServeListener(li)
	go h.TrackInputs()
	h.StartReportLoop()
	h.SaveResultToFile(file)
}
