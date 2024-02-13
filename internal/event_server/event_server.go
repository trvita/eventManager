package eventsrv

import (
	"context"
	"event/api/eventapi"
	"log"
)

type server struct {
	eventapi.UnimplementedEventServer
}

func MakeNewEventServer() eventapi.EventServer {
	return &server{}
}

func (s *server) MakeEvent(ctx context.Context, in *eventapi.MakeEventRequest) (*eventapi.MakeEventResponse, error) {
	log.Printf("Received: %v", in)
	return &eventapi.MakeEventResponse{}, nil
}
func (s *server) GetEvent(ctx context.Context, in *eventapi.GetEventRequest) (*eventapi.GetEventResponse, error) {
	log.Printf("Received: %v", in)
	return &eventapi.GetEventResponse{}, nil
}
func (s *server) DeleteEvent(ctx context.Context, in *eventapi.GetEventRequest) (*eventapi.DeleteEventResponse, error) {
	log.Printf("Received: %v", in)
	return &eventapi.DeleteEventResponse{}, nil
}
func (s *server) GetEvents(*eventapi.GetEventsRequest, eventapi.Event_GetEventsServer) error {
	log.Printf("Received: %v", "GetEvents")
	return nil
}
