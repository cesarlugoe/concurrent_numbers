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
	}

	h := handler.New()
	li := h.StartServer("4000")

	defer li.Close()

	go h.ServeListener(li)
	go h.TrackInputs()
	h.StartReportLoop()
	h.SaveResultToFile(file)
}
