package eventsrv

import (
	"context"

	"log"

	pb "github.com/trvita/eventManager/api/eventapi/eventapi"

	"github.com/trvita/eventManager/api/eventapi/helloworld"
)

type server struct {
	pb.UnimplementedGreeterServer
}

func MakeNewEventServer() (helloworld.GreeterServer) {
	return &server{}
}

func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	log.Printf("Received: %v", in.GetName())
	return &pb.HelloReply{Message: "Hello " + in.GetName()}, nil
}

func (s *server) MakeEvent(ctx context.Context, in *pb.MakeEventRequest) (*pb.MakeEventResponse, error) {
	log.Printf("Sender ID: %d\nTime: %d\nEvent Name: %d", in.GetSenderID(), in.GetTime(), in.GetName())
	return &pb.HelloReply{Message: "Hello " + in.GetName()}, nil
}