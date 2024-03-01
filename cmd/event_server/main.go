package main

import (
	"event/api/eventapi"
	eventsrv "event/internal/event_server"

	"flag"
	"fmt"
	"log"
	"net"

	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/grpc"
)

var (
	addr    = flag.String("h", "localhost", "The server IP address")
	port    = flag.Int("p", 50051, "The server port")
	brkaddr = flag.Int("b", 5672, "The broker port")
)

func main() {
	flag.Parse()

	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", *addr, *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()

	brk := fmt.Sprintf("amqp://guest:guest@localhost:%d/", *brkaddr)
	conn, err := amqp.Dial(brk)
	eventsrv.FailOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()
	ch, err := conn.Channel()
	eventsrv.FailOnError(err, "Failed to open a channel")
	defer ch.Close()
	// отдельная gouroutine которая понимает какое событие срабатывает чтобы отправить его
	server := eventsrv.MakeNewEventServer(ch)

	eventapi.RegisterEventManagerServer(s, server)

	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
