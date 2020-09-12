## How to Run the aplication (locally)

1. You need to have `go` installed
2. After installed, run `go install`
3. If you have any issue getting the "github.com/cesarlugoe/concurrent_numbers/handler" pkg as is in a private repository, make sure you have
   your terminal linked to your github account (ssh-key)
4. If you still have an issue, please follow this solution https://stackoverflow.com/questions/27500861/whats-the-proper-way-to-go-get-a-private-repository
5. To start the server: inside the server folder, run `go run main.go`
6. You can connect via `telnet localhost 4000` to manually test inputs
7. I've created a Client so you can run in loop random numbers to send and test performance, in client folder run `go run client.go`, you can run    this several times to check concurrent connections (only with several connections will full server performace be achieved)


## Notes

- The `numbers.log` file is created at the end of the sequence, so, only when one of the connections sends `terminate` the file will be created according to specifications.
- The Maximum amount of connections to be served at a given time are 5, however, additional clients can connect, but will not be served until old clients are disconnected.
- The termination secuence waits for the last report to be printed to the console, to be able to check the result of the log file vs this last print.
- the `StartReportLoop` function serves as the main "clock" that keeps the application running, as it is not a go rutine, only when `terminate` is triggered it will exit. I thought about implementing a wait group, or something of the sort, but in the end, this seemed sufficient for the task at hand.
- The Client prints the number of packages it sents to the server, there are 5 numbers on each pkg sent

