package main

import (
	"context"
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
	brkaddr = flag.Int("brk", 5672, "The broker port")
)

func main() {
	flag.Parse()

	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", *addr, *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()

	brkconn, err := amqp.Dial(fmt.Sprintf("amqp://guest:guest@localhost:%d/", *brkaddr))
	eventsrv.FailOnError(err, "Failed to brkconnect to RabbitMQ")
	defer brkconn.Close()
	ch, err := brkconn.Channel()
	eventsrv.FailOnError(err, "Failed to open a channel")
	defer ch.Close()

	server := eventsrv.MakeNewEventServer(ch).(*eventsrv.Server)
	eventapi.RegisterEventManagerServer(s, server)
	err = ch.ExchangeDelete("EventExchange", false, false) // для изменения типа точки обмена
	if err != nil {
		log.Printf("Failed to delete exchange: %v", err)
	}
	err = ch.ExchangeDeclare(
		eventsrv.ExchangeName,
		"direct",
		true,  // durable
		false, // autoDelete
		false, // internal
		false, // noWait
		nil,   // arguments
	)
	eventsrv.FailOnError(err, "Failed to declare an exchange")

	go eventsrv.ProcessEvents(context.Background(), server)

	log.Printf("Server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
