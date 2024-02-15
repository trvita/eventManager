package eventsrv

import (
	"context"
	"event/api/eventapi"
	"fmt"
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
	return &eventapi.MakeEventResponse{EventID: eventID}, status.New(codes.OK, "").Err()
}
func (s *server) GetEvent(ctx context.Context, in *eventapi.GetEventRequest) (*eventapi.GetEventResponse, error) {
	eventID := in.EventID
	event, exists := s.eventsMap[eventID]
	if !exists && event.SenderID != in.SenderID {
		return nil, status.Errorf(codes.NotFound, "Event with ID %d by user %d not found", eventID, event.SenderID)
	}
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
	event, exists := s.eventsMap[eventID]
	if !exists && event.SenderID != in.SenderID {
		return nil, status.Errorf(codes.NotFound, "Event with ID %d by user %d not found", eventID, event.SenderID)
	}
	delete(s.eventsMap, eventID)
	deleteresponse := fmt.Sprintf("Event with ID %d deleted", eventID)
	return &eventapi.DeleteEventResponse{Deleteresponse: deleteresponse}, status.New(codes.OK, "").Err()
}
func (s *server) GetEvents(req *eventapi.GetEventsRequest, stream eventapi.EventManager_GetEventsServer) error {
	fromTime := time.Unix(0, req.Fromtime)
	toTime := time.Unix(0, req.Totime)
	var eventsToSend []*eventapi.Event

	for _, event := range s.eventsMap {
		eventTime := time.Unix(0, event.Time)
		if event.SenderID == req.SenderID && eventTime.After(fromTime) && eventTime.Before(toTime) {
			eventsToSend = append(eventsToSend, event)
		}
	}
	response := &eventapi.GetEventsResponse{
		SenderID: req.SenderID,
		Fromtime: req.Fromtime,
		Totime:   req.Totime,
		Events:   eventsToSend,
	}
	return stream.Send(response)
}
