package handler

import (
	"bufio"
	"bytes"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"time"

	boom "github.com/tylertreat/BoomFilters"
)

const maxClients = 5

// Handler handles
type Handler struct {
	c           chan string
	terminate   *bool
	count       *counter
	inputBuffer *bytes.Buffer
}

//Counter keepts track of the recieved numbers
type counter struct {
	uniqueNumbers    int
	duplicateNumbers int
}

// New creates a new handler
func New() *Handler {
	terminate := false
	return &Handler{
		c:           make(chan string),
		terminate:   &terminate,
		count:       &counter{uniqueNumbers: 0, duplicateNumbers: 0},
		inputBuffer: new(bytes.Buffer),
	}
}

// StartServer listens to TCP connections
func (h *Handler) StartServer(address string) net.Listener {
	li, err := net.Listen("tcp", ":"+address)
	if err != nil {
		log.Fatalln(err)
	}

	return li
}

// ServeListener handles the incoming connections
func (h *Handler) ServeListener(li net.Listener) {
	sema := make(chan struct{}, maxClients)

	for {
		// stop serving more than max number of clients
		sema <- struct{}{}

		conn, err := li.Accept()
		if err != nil {
			<-sema
			continue
		}

		go h.handleConn(conn, sema)
	}
}

func (h *Handler) handleConn(conn net.Conn, sema chan struct{}) {

	scanner := bufio.NewScanner(conn)
	scanner.Split(bufio.ScanLines)
	terminate := h.terminate

	for scanner.Scan() {
		ln := scanner.Text()
		if *terminate || !validateInput(ln) {
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

		h.c <- ln
	}

	defer conn.Close()
}

// TrackInputs handles filtering and counting of numbers
func (h *Handler) TrackInputs() {

	sbf := boom.NewDefaultStableBloomFilter(10000, 0.01)

	for {
		newValue := <-h.c

		if sbf.Test([]byte(newValue)) {
			h.count.duplicateNumbers++
			continue
		}

		sbf.Add([]byte(newValue))
		h.count.uniqueNumbers++
		h.inputBuffer.WriteString(newValue + "\n")
	}
}

// StartReportLoop prints a report on intervals of incoming numbers to the console
func (h *Handler) StartReportLoop() {
	prevUniqueTotal := 0
	prevDuplicateTotal := 0
	terminate := h.terminate

	for !*terminate {
		time.Sleep(10000 * time.Millisecond)

		uniqueTotal := h.count.uniqueNumbers
		uniqueInterval := uniqueTotal - prevUniqueTotal
		prevUniqueTotal = uniqueTotal

		duplicateTotal := h.count.duplicateNumbers
		duplicateInterval := duplicateTotal - prevDuplicateTotal
		prevDuplicateTotal = duplicateTotal

		log.Printf("Received %v unique numbers, %v duplicates. Unique total: %v \n", uniqueInterval, duplicateInterval, uniqueTotal)
	}
}

// SaveResultToFile stores unique numbers on the created file when the terminate switch turns to false
func (h *Handler) SaveResultToFile(file *os.File) {
	piper, pipew := io.Pipe()
	go func() {
		defer pipew.Close()
		io.Copy(pipew, h.inputBuffer)
	}()

	io.Copy(file, piper)
	piper.Close()
	log.Println("Finished storing Numbers")
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
