package main

import (
	"bufio"
	"log"
	"net"
	"time"
	"strconv"
	"github.com/tylertreat/BoomFilters"
	"bytes"
	"io"
	"os"
)

const maxClients = 5

type counter struct {
	uniqueNumbers 	 int
	duplicateNumbers int
}

func main() {

	file, err := os.Create("./numbers.log")
	if err != nil {
		log.Fatalln(err)
		os.Exit(1)
	}

	li, err := net.Listen("tcp", ":4000")
	if err != nil {
		log.Fatalln(err)
	}

	defer li.Close()

	c := make(chan string)
	terminate := false

	go serveListener(li, c, &terminate)

	count := counter{uniqueNumbers: 0, duplicateNumbers: 0}
	inputBuffer := new(bytes.Buffer)

	go trackInputs(c, &count, inputBuffer)
	startReportLoop(&count, &terminate)
	saveToFile(inputBuffer, file)
}

func serveListener(li net.Listener, c chan string, terminate *bool) {
	sema := make(chan struct{}, maxClients)	

	for {
		// stop serving more than max number of clients
		sema <- struct{}{}

		conn, err := li.Accept()
		if err != nil {
			log.Println(err)
			<-sema
			continue
		}

		go handleConn(conn, c, terminate, sema)
	}
}

func handleConn(conn net.Conn, c chan string, terminate *bool, sema chan struct{}) {

	scanner := bufio.NewScanner(conn)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		ln := scanner.Text()

		if  *terminate || !validateInput(ln) {	
			<-sema
			if *terminate {
				conn.Write([]byte("Process terminated"))
			}
			conn.Close()
			return
		}

		if ln == "terminate" {
			*terminate = true
			conn.Close()
			return
		}

		c <- ln
	}

	defer conn.Close()
}

func saveToFile(inputBuffer *bytes.Buffer, file *os.File) {
	piper, pipew := io.Pipe()
	go func() {
		defer pipew.Close()
		io.Copy(pipew, inputBuffer)
	}()
	
	io.Copy(file, piper)
	piper.Close()
	log.Println("Finished storing Numbers")
}

func trackInputs(c chan string, count *counter, inputBuffer *bytes.Buffer) {
	sbf := boom.NewDefaultStableBloomFilter(10000, 0.01)	
	
	for {
		newValue := <-c

		if sbf.Test([]byte(newValue)) {
			count.duplicateNumbers++
			continue
		} 

		sbf.Add([]byte(newValue))
		count.uniqueNumbers++
		inputBuffer.WriteString(newValue + "\n")
	}
}



func startReportLoop(count *counter, terminate *bool){
	prevUniqueTotal := 0
	prevDuplicateTotal := 0

	for !*terminate {
		time.Sleep(10000 * time.Millisecond)

		uniqueTotal := count.uniqueNumbers
		uniqueInterval := uniqueTotal - prevUniqueTotal
		prevUniqueTotal = uniqueTotal

		duplicateTotal := count.duplicateNumbers
		duplicateInterval := duplicateTotal - prevDuplicateTotal
		prevDuplicateTotal = duplicateTotal

		log.Printf("Received %v unique numbers, %v duplicates. Unique total: %v \n", uniqueInterval, duplicateInterval, uniqueTotal)
	}
}

func validateInput(b string) bool {
	num, err := strconv.Atoi(b)

	if b == "terminate" {
		return true
	}

	if err != nil || len(b) != 9 || num <= 0 || num >= 999999999 {
		return false
	}

	return true
}