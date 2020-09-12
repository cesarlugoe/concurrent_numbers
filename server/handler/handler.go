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

// MaxClients is the maximum number of clients allowed to connect to the server
const MaxClients = 5
const maxAllowedInput = 999999999
const requiredInputLength = 9
const reportInvervalMs = 10000

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
func (h *Handler) StartServer(port string) (net.Listener, error) {
	li, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalln(err)
		return nil, err
	}

	return li, nil
}

// ServeListener handles the incoming connections
func (h *Handler) ServeListener(li net.Listener) {
	sema := make(chan struct{}, MaxClients)

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

	for scanner.Scan() {
		ln := scanner.Text()

		if *h.terminate || !validateInput(ln) {
			<-sema
			if *h.terminate {
				conn.Write([]byte("Process terminated"))
			}
			conn.Close()
			return
		}

		if ln == "terminate" {
			*h.terminate = true
			conn.Close()
			return
		}

		h.c <- ln
	}

	defer conn.Close()
}

// TrackInputs handles filtering and counting of numbers
func (h *Handler) TrackInputs() {

	// Stable Bloom Filter Algorithm can work with an unbounded data source
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
	var uniqueTotal int
	var uniqueInterval int
	var prevUniqueTotal int

	var duplicateTotal int
	var prevDuplicateTotal int
	var duplicateInterval int

	for !*h.terminate {
		time.Sleep(reportInvervalMs * time.Millisecond)

		uniqueTotal = h.count.uniqueNumbers
		uniqueInterval = uniqueTotal - prevUniqueTotal
		prevUniqueTotal = uniqueTotal

		duplicateTotal = h.count.duplicateNumbers
		duplicateInterval = duplicateTotal - prevDuplicateTotal
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

	if err != nil || len(b) != requiredInputLength || num <= 0 || num >= maxAllowedInput {
		return false
	}

	return true
}
