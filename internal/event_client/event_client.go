package eventcl

import (
	"context"
	"event/api/eventapi"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	amqp "github.com/rabbitmq/amqp091-go"
)

func FailOnError(err error, msg string) {
	if err != nil {
		log.Panicf("\u001b[31m%s: %s\u001b[0m", msg, err)
	}
}
func ProcessCommands(c eventapi.EventManagerClient, ctx context.Context, sdid *int64, ssid int64, commandChan chan []string) {
	for {
		select {
		case command := <-commandChan:
			switch strings.ToLower(command[0]) {
			case "makeevent":
				MakeEvent(command, c)
			case "getevent":
				GetEvent(command, c, ctx)
			case "deleteevent":
				DeleteEvent(command, c, ctx)
			case "getevents":
				GetEvents(command, c, ctx)
			case "exit":
				Exit(command, c, ctx, sdid)
			default:
				fmt.Printf("\u001b[93m > Unknown command. Available commands: MakeEvent, GetEvent, DeleteEvent, GetEvents, Exit\n\u001b[0m")
			}
		case <-ctx.Done():
			return
		}
	}
}
func GetMessages(sdid *int64, c eventapi.EventManagerClient) {
	brkconn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	FailOnError(err, "Failed to connect to RabbitMQ")
	defer brkconn.Close()

	ch, err := brkconn.Channel()
	FailOnError(err, "Failed to open a channel")
	defer ch.Close()

	queueName := fmt.Sprintf("%d", sdid)
	q, err := ch.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	FailOnError(err, "Failed to declare a queue")

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	FailOnError(err, "Failed to register a consumer")
	for d := range msgs {
		log.Printf("Received a message: %s", d.Body)
		log.Printf("Done")
	}
}

func MakeEvent(command []string, c eventapi.EventManagerClient) {
	if len(command) != 4 {
		fmt.Printf("\u001b[93m > Usage: %s <sender-id> <event-time> <event-name>\n\u001b[0m", command[0])
		return
	}
	senderID, err := strconv.ParseInt(command[1], 10, 64)
	if err != nil {
		fmt.Printf("\u001b[93m > Error parsing senderID: %v\n\u001b[0m", err)
		return
	}
	time, err := strconv.ParseInt(command[2], 10, 64)
	if err != nil {
		fmt.Printf("\u001b[93m > Error parsing event time: %v\n\u001b[0m", err)
		return
	}
	r, err := c.MakeEvent(context.Background(), &eventapi.MakeEventRequest{
		SenderID: senderID,
		Time:     time,
		Name:     command[3],
	})
	if err != nil {
		fmt.Printf("\u001b[93m > Could not make event: %v\n\u001b[0m", err)
		return
	}
	fmt.Printf("\u001b[93m > Created{Event ID: %d}\n\u001b[0m", r.GetEventID())
}

func GetEvent(command []string, c eventapi.EventManagerClient, ctx context.Context) {
	if len(command) != 3 {
		fmt.Printf("\u001b[93m > Usage: %s <sender-id> <event-id>\u001b[0m", command[0])
		return
	}
	senderID, err := strconv.ParseInt(command[1], 10, 64)
	if err != nil {
		fmt.Printf("\u001b[93m > Error parsing senderID: %v\n\u001b[0m", err)
		return
	}
	eventID, err := strconv.ParseInt(command[2], 10, 64)
	if err != nil {
		fmt.Printf("\u001b[93m > Error parsing eventID: %v\n\u001b[0m", err)
		return
	}
	r, err := c.GetEvent(ctx, &eventapi.GetEventRequest{
		SenderID: senderID,
		EventID:  eventID,
	})
	if err != nil {
		fmt.Printf("\u001b[93m > Error getting event: %v\n\u001b[0m", err)
		return
	} else if r == nil {
		fmt.Printf("\u001b[93m > Event not found\n\u001b[0m")
		return
	} else {
		fmt.Printf("\u001b[93m > Event{sender_id:%d, eventId:%d, time:%d, name:%s}\n\u001b[0m",
			r.Event.SenderID, r.Event.EventID, r.Event.Time, r.Event.Name)
	}
}

func DeleteEvent(command []string, c eventapi.EventManagerClient, ctx context.Context) {
	if len(command) != 3 {
		fmt.Printf("\u001b[93m > Usage: %s <sender-id> <event-id>\n\u001b[0m", command[0])
		return
	}
	senderID, err := strconv.ParseInt(command[1], 10, 64)
	if err != nil {
		fmt.Printf("\u001b[93m > Error parsing senderID: %v\n\u001b[0m", err)
		return
	}
	eventID, err := strconv.ParseInt(command[2], 10, 64)
	if err != nil {
		fmt.Printf("\u001b[93m > Error parsing eventID: %v\n\u001b[0m", err)
		return
	}
	r, err := c.DeleteEvent(ctx, &eventapi.GetEventRequest{
		SenderID: senderID,
		EventID:  eventID,
	})
	if err != nil {
		fmt.Printf("\u001b[93m > Error getting event: %v\n\u001b[0m", err)
		return
	} else if r == nil {
		fmt.Printf("\u001b[93m > Event not found\n\u001b[0m")
		return
	} else {
		fmt.Printf("\u001b[93m > %s\n\u001b[0m", r.Deleteresponse)
	}

}

func GetEvents(command []string, c eventapi.EventManagerClient, ctx context.Context) {
	if len(command) != 4 {
		fmt.Printf("\u001b[93m > Usage: %s <sender-id> <from-time> <to-time>\n\u001b[0m", command[0])
		return
	}
	senderID, err := strconv.ParseInt(command[1], 10, 64)
	if err != nil {
		fmt.Printf("\u001b[93m > Error parsing senderID: %v\n\u001b[0m", err)
		return
	}
	fromtime, err := strconv.ParseInt(command[2], 10, 64)
	if err != nil {
		fmt.Printf("\u001b[93m > Error parsing from time: %v\n\u001b[0m", err)
		return
	}
	totime, err := strconv.ParseInt(command[3], 10, 64)
	if err != nil {
		fmt.Printf("\u001b[93m > Error parsing to time: %v\n\u001b[0m", err)
		return
	}
	stream, err := c.GetEvents(ctx, &eventapi.GetEventsRequest{
		SenderID: senderID,
		Fromtime: fromtime,
		Totime:   totime,
	})
	if err != nil {
		fmt.Printf("\u001b[93m > Error getting events: %v\n\u001b[0m", err)
		return
	}
	for {
		r, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Printf("\u001b[93m > Error receiving events: %v\n\u001b[0m", err)
			return
		}
		if r.Events != nil {
			fmt.Printf("\u001b[93m > Events by %d from %d to %d:\n", senderID, fromtime, totime)
			for _, event := range r.Events {
				fmt.Printf("\u001b[93m\tEvent{eventId:%d, time:%d, name:%s}\n\u001b[0m", event.EventID, event.Time, event.Name)
			}
		} else {
			fmt.Println("\u001b[93m > No events found.\u001b[0m")
		}
	}
}
func Exit(command []string, c eventapi.EventManagerClient, ctx context.Context, sdid *int64) {
	r, err := c.Exit(context.Background(), &eventapi.ExitRequest{
		SenderID: *sdid,
	})
	FailOnError(err, "Failed to get sender ID")
	fmt.Println("\u001b[93m >", r.Goodbye, "\u001b[0m")
	os.Exit(0)
}
