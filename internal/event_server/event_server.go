package eventsrv

import (
	"context"
	"event/api/eventapi"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type server struct {
	eventapi.UnimplementedEventManagerServer
	eventsMap      map[int64]*eventapi.Event
	eventIDCounter int64
}

func MakeNewEventServer() eventapi.EventManagerServer {
	return &server{
		eventsMap:      make(map[int64]*eventapi.Event),
		eventIDCounter: 1,
	}
}

func (s *server) MakeEvent(ctx context.Context, in *eventapi.MakeEventRequest) (*eventapi.MakeEventResponse, error) {
	eventID := s.eventIDCounter
	s.eventIDCounter++
	event := &eventapi.Event{
		EventID:  eventID,
		SenderID: in.SenderID,
		Time:     in.Time,
		Name:     in.Name,
	}
	if s.eventsMap == nil {
		s.eventsMap = make(map[int64]*eventapi.Event)
	}
	s.eventsMap[eventID] = event
	log.Printf("%d made event: %d", event.SenderID, eventID)
	return &eventapi.MakeEventResponse{EventID: eventID}, status.New(codes.OK, "").Err()
}
func (s *server) GetEvent(ctx context.Context, in *eventapi.GetEventRequest) (*eventapi.GetEventResponse, error) {
	eventID := in.EventID
	senderID := in.SenderID
	event, exists := s.eventsMap[eventID]
	if !exists || (exists && (event.SenderID != senderID)) {
		return nil, status.Errorf(codes.NotFound, "Event with ID %d by user %d not found", eventID, senderID)
	}
	log.Printf("%d got event: %d", senderID, eventID)
	return &eventapi.GetEventResponse{
		Event: &eventapi.Event{
			SenderID: event.SenderID,
			EventID:  event.EventID,
			Time:     event.Time,
			Name:     event.Name,
		},
	}, status.New(codes.OK, "").Err()
}
func (s *server) DeleteEvent(ctx context.Context, in *eventapi.GetEventRequest) (*eventapi.DeleteEventResponse, error) {
	eventID := in.EventID
	senderID := in.SenderID
	event, exists := s.eventsMap[eventID]
	if !exists || event.SenderID != senderID {
		return nil, status.Errorf(codes.NotFound, "Event with ID %d by user %d not found", eventID, senderID)
	}
	delete(s.eventsMap, eventID)
	deleteresponse := fmt.Sprintf("%d deleted event: %d", senderID, eventID)
	log.Print(deleteresponse)
	return &eventapi.DeleteEventResponse{Deleteresponse: deleteresponse}, status.New(codes.OK, "").Err()
}
func (s *server) GetEvents(in *eventapi.GetEventsRequest, stream eventapi.EventManager_GetEventsServer) error {
	fromTime := time.Unix(0, in.Fromtime)
	toTime := time.Unix(0, in.Totime)
	var eventsToSend []*eventapi.Event

	for _, event := range s.eventsMap {
		eventTime := time.Unix(0, event.Time)
		if event.SenderID == in.SenderID && eventTime.After(fromTime) && eventTime.Before(toTime) {
			eventsToSend = append(eventsToSend, event)
		}
	}
	response := &eventapi.GetEventsResponse{
		SenderID: in.SenderID,
		Events:   eventsToSend,
	}
	log.Printf("%d requested events", in.SenderID)
	return stream.Send(response)
}
