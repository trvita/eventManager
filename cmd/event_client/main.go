package main

import (
	"bufio"
	"context"
	"event/api/eventapi"
	eventcl "event/internal/event_client"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Color string

const (
	Black      Color = "\u001b[30m"
	Red        Color = "\u001b[31m"
	Green      Color = "\u001b[32m"
	Yellow     Color = "\u001b[93m"
	Blue       Color = "\u001b[34m"
	ColorReset Color = "\u001b[0m"
)

var (
	addr = flag.String("dst", "localhost", "The server IP address")
	port = flag.Int("p", 50051, "The server port")
	sdid = flag.Int64("sender-id", 0, "Sender ID")
)

func main() {
	flag.Parse()
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", *addr, *port), grpc.WithTransportCredentials(insecure.NewCredentials()))
	eventcl.FailOnError(err, "Failed to connect to server")
	defer conn.Close()

	c := eventapi.NewEventManagerClient(conn)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	ssid := time.Now().Unix()
	r, err := c.GreetSender(context.Background(), &eventapi.GreetSenderRequest{
		SenderID:  *sdid,
		SessionID: ssid,
	})
	eventcl.FailOnError(err, "Failed to get sender ID")

	*sdid = r.SenderID
	fmt.Printf("\u001b[93m > Your ID: %d\n\u001b[0m", *sdid)

	commandChan := make(chan []string)
	go eventcl.ProcessCommands(c, ctx, sdid, ssid, commandChan)
	scanner := bufio.NewScanner(os.Stdin)
	for {
		if scanner.Scan() {
			command := strings.Fields(scanner.Text())
			commandChan <- command
		} else {
			fmt.Println("Error reading input:", scanner.Err())
			break
		}
	}

	go eventcl.GetMessages(sdid, c)
}
