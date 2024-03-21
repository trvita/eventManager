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

var ExchangeName = "EventExchange"

type Server struct {
	eventapi.UnimplementedEventManagerServer
	eventsMap       map[int64]*eventapi.Event
	eventIDCounter  int64
	senderIDCounter int64
	sessionID       int64
	channel         *amqp.Channel
}

func FailOnError(err error, msg string) {
	if err != nil {
		log.Panicf("\u001b[31m%s: %s\u001b[0m", msg, err)
	}
}
func ProcessEvents(ctx context.Context, s *Server) {
	for {
		currentTime := time.Now().Unix()
		for eventID, event := range s.eventsMap {
			if currentTime >= event.Time {
				err := s.PublishMessages(ctx, event)
				if err != nil {
					log.Printf("\u001b[93mSERVER\u001b[0m: Error publishing message for eventID %d: %v", eventID, err)
					continue
				}
				delete(s.eventsMap, eventID)
			}
		}
		time.Sleep(1 * time.Second)
	}
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
	senderID := in.SenderID
	if senderID == 0 { // null senderID means new sender
		s.senderIDCounter++
		senderID = s.senderIDCounter
	}
	queueName := fmt.Sprintf("%d", senderID)
	_, err := s.channel.QueueDeclare(
		queueName,
		true,  // durable
		false, // autoDelete
		false, // not exclusive, others can subscribe too
		false, // noWait
		nil,   // arguments
	)
	FailOnError(err, "Failed to declare a queue")

	err = s.channel.QueueBind(
		queueName,
		queueName, // routing key
		ExchangeName,
		false, // noWait
		nil,   // arguments
	)
	if err != nil {
		return nil, err
	}
	log.Printf("\u001b[93mSERVER\u001b[0m: Connected senderID %d, sessionID: %d", in.SenderID, in.SessionID)
	return &eventapi.GreetSenderResponse{SenderID: senderID}, nil
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
	log.Printf("\u001b[93mSERVER\u001b[0m: Made event senderID %d, eventID %d", event.SenderID, eventID)
	return &eventapi.MakeEventResponse{EventID: eventID}, status.New(codes.OK, "").Err()
}

func (s *Server) GetEvent(ctx context.Context, in *eventapi.GetEventRequest) (*eventapi.GetEventResponse, error) {
	eventID := in.EventID
	senderID := in.SenderID
	event, exists := s.eventsMap[eventID]
	if !exists || (exists && (event.SenderID != senderID)) {
		return nil, status.Errorf(codes.NotFound, "Event with ID %d by user %d not found", eventID, senderID)
	}
	log.Printf("\u001b[93mSERVER\u001b[0m: Got event senderID %d, eventID %d", senderID, eventID)
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
		return nil, status.Errorf(codes.NotFound, "EventID %d by senderID %d not found", eventID, senderID)
	}
	delete(s.eventsMap, eventID)
	deleteresponse := fmt.Sprintf("EventID %d by senderID %d deleted", eventID, senderID)
	log.Printf("\u001b[93mSERVER\u001b[0m: Deleted event senderID %d, eventID %d", senderID, eventID)
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
	log.Printf("\u001b[93mSERVER\u001b[0m: Requested events senderID %d", in.SenderID)
	return stream.Send(response)
}

func (s *Server) Exit(ctx context.Context, in *eventapi.ExitRequest) (*eventapi.ExitResponse, error) {
	log.Printf("\u001b[93mSERVER\u001b[0m: Exited senderID %d", in.SenderID)
	goodbye := fmt.Sprintf("Goodbye, %d", in.SenderID)
	return &eventapi.ExitResponse{Goodbye: goodbye}, nil
}

func (s *Server) PublishMessages(ctx context.Context, event *eventapi.Event) error {
	ExchangeName := "EventExchange"
	queueName := fmt.Sprintf("%d", event.SenderID)
	expiration := fmt.Sprintf("%d", event.Time)
	err := s.channel.PublishWithContext(
		ctx,
		ExchangeName, // exchange
		queueName,    // routing key
		false,        // mandatory
		false,        // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Expiration:   expiration,
			Body:         []byte(fmt.Sprintf(`{"eventID": %d, "senderID": %d, "time": %d, "name": "%s"}`, event.EventID, event.SenderID, event.Time, event.Name)),
		},
	)
	if err != nil {
		return err
	}
	log.Printf("\u001b[93mSERVER\u001b[0m: Published message senderID %d: eventID %d", event.SenderID, event.EventID)
	return nil
}
