package main

import (
    "net"
    "os"
    "math/rand"
    "strconv"
)

func main() {
    servAddr := "localhost:4000"
    tcpAddr, err := net.ResolveTCPAddr("tcp", servAddr)
    if err != nil {
        println("ResolveTCPAddr failed:", err.Error())
        os.Exit(1)
    }

    conn, err := net.DialTCP("tcp", nil, tcpAddr)
    if err != nil {
        println("Dial failed:", err.Error())
        os.Exit(1)
    }
    a := 0
    for {
        message := createMessage()
        a++
        println(a)
         _, err = conn.Write([]byte(message))
        if err != nil {
            println("Write to server failed:", err.Error())
            os.Exit(1)
        }
    }
   

    reply := make([]byte, 1024)

    _, err = conn.Read(reply)
    if err != nil {
        println("Write to server failed:", err.Error())
        os.Exit(1)
    }

    println("reply from server=", string(reply))

    conn.Close()
}

func createMessage() string {
    message := ""
    for i := 0; i < 5; i++ {
        num := rand.Intn(999999999 - 000000001) + 000000001
        message = message + lpad(strconv.Itoa(num), "0", 9) + "\n"
    }
    return message
}

func lpad(s string, pad string, plength int) string {
    for i := len(s); i<plength; i++ {
        s = pad + s
    }
    return s
}