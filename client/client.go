package main

import (
	"fmt"
	"math/rand"
	"net"
	"os"
	"strconv"
	"time"
)

const maxAllowedInput = 999999999
const numbersPerPackage = 5

func main() {
	servAddr := "localhost:4000"
	tcpAddr, err := net.ResolveTCPAddr("tcp", servAddr)
	if err != nil {
		fmt.Println("ResolveTCPAddr failed:", err.Error())
		os.Exit(1)
	}

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		fmt.Println("Dial failed:", err.Error())
		os.Exit(1)
	}

	a := 0
	go func() {
		for {
			time.Sleep(1000 * time.Millisecond)
			fmt.Printf("Total packages sent: %v\n", a)
		}
	}()

	for {
		message := createMessage()
		a++
		_, err = conn.Write([]byte(message))
		if err != nil {
			fmt.Println("Write to server failed:", err.Error())
			break
		}
	}

	reply := make([]byte, 1024)

	_, err = conn.Read(reply)
	if err != nil {
		fmt.Println("Write to server failed:", err.Error())
		os.Exit(1)
	}

	fmt.Println("reply from server=", string(reply))

	conn.Close()
}

func createMessage() string {
	message := ""
	for i := 0; i < numbersPerPackage; i++ {
		num := rand.Intn(maxAllowedInput-1) + 1
		message = message + lpad(strconv.Itoa(num), "0", 9) + "\n"
	}
	return message
}

func lpad(s string, pad string, plength int) string {
	for i := len(s); i < plength; i++ {
		s = pad + s
	}
	return s
}
