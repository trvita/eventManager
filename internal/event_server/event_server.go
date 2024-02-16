package eventsrv

import (
	"context"
	"event/api/eventapi"
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	eventapi.UnimplementedEventManagerServer
	eventsMap       map[int64]*eventapi.Event
	eventIDCounter  int64
	senderIDCounter int64
	sessionID       int64
	channel         *amqp.Channel
}

func MakeNewEventServer(ch *amqp.Channel) eventapi.EventManagerServer {
	return &Server{
		eventsMap:       make(map[int64]*eventapi.Event),
		eventIDCounter:  1,
		senderIDCounter: 0,
		sessionID:       0,
		channel:         ch,
	}
}

func (s *Server) GreetSender(ctx context.Context, in *eventapi.GreetSenderRequest) (*eventapi.GreetSenderResponse, error) {
	if in.SenderID == 0 {
		s.senderIDCounter++
		senderID := s.senderIDCounter
		log.Printf("%d connected", senderID)

		exchangeName := fmt.Sprintf("%d", senderID)
		err := s.channel.ExchangeDeclare(
			exchangeName,
			"direct", // или другой тип обмена, который вам нужен
			true,     // durable
			false,    // autoDelete
			false,    // internal
			false,    // noWait
			nil,      // arguments
		)
		if err != nil {
			return nil, err
		}
		queueName := fmt.Sprintf("%d:%d", senderID, s.sessionID)
		_, err = s.channel.QueueDeclare(
			queueName,
			true,  // durable
			false, // autoDelete
			false, // not exclusive, others can subscribe too
			false, // noWait
			nil,   // arguments
		)
		if err != nil {
			return nil, err
		}
		err = s.channel.QueueBind(
			queueName,
			queueName,
			exchangeName,
			false, // noWait
			nil,   // arguments
		)
		if err != nil {
			return nil, err
		}
		return &eventapi.GreetSenderResponse{SenderID: senderID}, nil
	} else {
		log.Printf("%d connected", in.SenderID)
		return &eventapi.GreetSenderResponse{SenderID: in.SenderID}, nil
	}
}

func (s *Server) MakeEvent(ctx context.Context, in *eventapi.MakeEventRequest) (*eventapi.MakeEventResponse, error) {
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
	exchangeName := fmt.Sprintf("%d", in.SenderID)
	err := s.channel.PublishWithContext(
		ctx,
		exchangeName, // exchange
		exchangeName, // routing key
		false,        // mandatory
		false,        // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        []byte(fmt.Sprintf(`{"eventID": %d, "senderID": %d, "time": %d, "name": "%s"}`, eventID, in.SenderID, in.Time, in.Name)),
		},
	)
	if err != nil {
		return nil, err
	}
	return &eventapi.MakeEventResponse{EventID: eventID}, status.New(codes.OK, "").Err()
}

func (s *Server) GetEvent(ctx context.Context, in *eventapi.GetEventRequest) (*eventapi.GetEventResponse, error) {
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

func (s *Server) DeleteEvent(ctx context.Context, in *eventapi.GetEventRequest) (*eventapi.DeleteEventResponse, error) {
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

func (s *Server) GetEvents(in *eventapi.GetEventsRequest, stream eventapi.EventManager_GetEventsServer) error {
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

func (s *Server) Exit(ctx context.Context, in *eventapi.ExitRequest) (*eventapi.ExitResponse, error) {
	log.Printf("%d exited", in.SenderID)
	goodbye := fmt.Sprintf("Goodbye, %d", in.SenderID)
	return &eventapi.ExitResponse{Goodbye: goodbye}, nil
}
