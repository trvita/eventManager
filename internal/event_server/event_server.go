package event_server

import (
	"context"

	"log"

	pb "api/eventapi"

	"google.golang.org/grpc/examples/helloworld/helloworld"
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

// func (s *server) MakeEvent(ctx context.Context, in *pb.MakeEventRequest) (*pb.MakeEventResponse, error) {
// 	log.Printf("Sender ID: %d\nTime: %d\nEvent Name: %d", in.GetSenderID(), in.GetTime(), in.GetName())
// 	return &pb.HelloReply{Message: "Hello " + in.GetName()}, nil
// }