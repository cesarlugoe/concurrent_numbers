package main

import (
	"fmt"
	"bufio"
	"log"
	"net"
	"time"
	"strconv"
	"github.com/tylertreat/BoomFilters"
)

const maxClients = 5

func main() {

	li, err := net.Listen("tcp", ":4000")
	if err != nil {
		log.Fatalln(err)
	}

	defer li.Close()

	sema := make(chan struct{}, maxClients)
	c := make(chan string)
	terminate := false

	go func() {
		for {
			sema <- struct{}{}
			defer func() { <-sema }()
			conn, err := li.Accept()
			if err != nil {
				log.Println(err)
				// check
				continue
			}

			go handleConn(conn, c, &terminate)
		}
	}()

	var uniqueNumbers []string
	numberOfDuplicates := 0

	go report(&uniqueNumbers, &numberOfDuplicates)

	sbf := boom.NewDefaultStableBloomFilter(10000, 0.01)

	for {
		newValue := <-c

		if sbf.Test([]byte(newValue)) {
			numberOfDuplicates++
			continue
		} 

		sbf.Add([]byte(newValue))
		uniqueNumbers = append(uniqueNumbers, newValue)
	}
}

func handleConn(conn net.Conn, c chan string, terminate *bool) {

	scanner := bufio.NewScanner(conn)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		ln := scanner.Text()

		if  *terminate || !validateInput(ln) {	
			conn.Close()
			return
		}

		if (ln == "terminate") {
			*terminate = true
			conn.Close()
			return
		}

		c <- ln
	}

	defer conn.Close()
}

func report(uniqueNumbers *[]string, numberOfDuplicates *int){
	intervalCountUnique := 0
	lastTotal := 0
	intervalCountDuplicate := 0
	lastDuplicateTotal := 0

	for {
		time.Sleep(10000 * time.Millisecond)
		quantity := len(*uniqueNumbers)
		intervalCountUnique = quantity - lastTotal
		lastTotal = quantity

		duplicateTotal := *numberOfDuplicates
		intervalCountDuplicate = duplicateTotal - lastDuplicateTotal
		lastDuplicateTotal = duplicateTotal

		fmt.Printf("Received %v unique numbers, %v duplicates. Unique total: %v \n", intervalCountUnique, intervalCountDuplicate, quantity)
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