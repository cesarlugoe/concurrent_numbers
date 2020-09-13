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

// MaxClients allowed to connect to the server
const MaxClients = 5

const maxAllowedInput = 999999999
const requiredInputLength = 9
const reportInvervalMs = 10000
const terminateSignal = "terminate"

// Handler contains the state of the application
type Handler struct {
	c           chan string
	terminate   *bool
	count       *counter
	inputBuffer *bytes.Buffer
}

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

// StartServer returns a TCP listener
func (h *Handler) StartServer(port string) net.Listener {
	li, err := net.Listen("tcp", ":"+port)
	handleFaltalError(err)

	return li
}

// ServeListener handles the incoming connections
func (h *Handler) ServeListener(li net.Listener) {
	sema := make(chan struct{}, MaxClients)

	for {
		// semaphore to stop serving more than max number of clients
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

		if *h.terminate {
			_, err := conn.Write([]byte("Process terminated"))
			if err != nil {
				log.Println(err)
			}
			err = conn.Close()
			handleFaltalError(err)
			return
		}

		if !validateInput(ln) {
			<-sema
			err := conn.Close()
			handleFaltalError(err)
			return
		}

		if ln == terminateSignal {
			*h.terminate = true
			err := conn.Close()
			handleFaltalError(err)
			return
		}

		h.c <- ln
	}
}

// TrackInputs handles filtering and counting of numbers
func (h *Handler) TrackInputs() {

	// Stable Bloom Filter Algorithm can work with an unbounded data source
	sbf := boom.NewDefaultStableBloomFilter(10000, 0.01)

	for {
		ln := <-h.c

		if sbf.Test([]byte(ln)) {
			h.count.duplicateNumbers++
			continue
		}

		sbf.Add([]byte(ln))
		h.count.uniqueNumbers++
		h.inputBuffer.WriteString(ln + "\n")
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
		_, err := io.Copy(pipew, h.inputBuffer)
		handleFaltalError(err)
	}()

	_, err := io.Copy(file, piper)
	handleFaltalError(err)

	err = piper.Close()
	handleFaltalError(err)

	log.Println("Finished storing Numbers")
}

func validateInput(b string) bool {

	if b == terminateSignal {
		return true
	}

	num, err := strconv.Atoi(b)
	if err != nil || len(b) != requiredInputLength || num <= 0 || num >= maxAllowedInput {
		return false
	}

	return true
}

func handleFaltalError(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}
