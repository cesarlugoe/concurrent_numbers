package main

import (
	"log"
	"os"

	"github.com/cesarlugoe/concurrent_numbers/handler"
)

const port = "4000"

func main() {

	file, err := os.Create("./numbers.log")
	if err != nil {
		log.Fatalln(err)
	}

	h := handler.New()
	li := h.StartServer(port)

	defer li.Close()

	go h.ServeListener(li)
	go h.TrackInputs()

	// when h.terminate = true, report loop breaks and saveResultToFile() is called
	h.StartReportLoop()
	h.SaveResultToFile(file)
}
