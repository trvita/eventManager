package eventsrv

import (
	"context"
	"event/api/eventapi"
	"log"
)

type server struct {
	eventapi.UnimplementedGreeterServer
}

func MakeNewEventServer() eventapi.GreeterServer {
	return &server{}
}

func (s *server) SayHello(ctx context.Context, in *eventapi.HelloRequest) (*eventapi.HelloReply, error) {
	log.Printf("Received: %v", in.GetName())
	return &eventapi.HelloReply{Message: "Hello " + in.GetName()}, nil
}

// func (s *server) MakeEvent(ctx context.Context, in *eventapi.MakeEventRequest) (*eventapi.MakeEventResponse, error) {
// 	log.Printf("Sender ID: %d\nTime: %d\nEvent Name: %d", in.GetSenderID(), in.GetTime(), in.GetName())
// 	return &eventapi.HelloReply{Message: "Hello " + in.GetName()}, nil
// }