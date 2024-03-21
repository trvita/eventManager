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
		log.Panicf("\u001b[31m%s: %s", msg, err)
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
				fmt.Printf("\u001b[93mCLIENT:\u001b[0m Unknown command. Available commands: MakeEvent, GetEvent, DeleteEvent, GetEvents, Exit\n")
			}
		case <-ctx.Done():
			return
		}
	}
}

func GetMessages(sdid *int64, c eventapi.EventManagerClient, brkconn *amqp.Connection) {
	fmt.Printf("\u001b[93mCLIENT:\u001b[0m Waiting for messages\n")
	ch, err := brkconn.Channel()
	FailOnError(err, "Failed to open a channel")
	defer ch.Close()

	queueName := fmt.Sprintf("%d", *sdid)

	msgs, err := ch.Consume(
		queueName, // queue
		"",        // consumer
		true,      // auto-ack
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)
	FailOnError(err, "Failed to register a consumer")
	for d := range msgs {
		fmt.Printf("\u001b[93mCLIENT:\u001b[0mReceived a message: %s\n", d.Body)
	}
}

func MakeEvent(command []string, c eventapi.EventManagerClient) {
	if len(command) != 4 {
		fmt.Printf("\u001b[93mCLIENT:\u001b[0m Usage: %s <sender-id> <event-time> <event-name>\n", command[0])
		return
	}
	senderID, err := strconv.ParseInt(command[1], 10, 64)
	if err != nil {
		fmt.Printf("\u001b[93mCLIENT:\u001b[0m Error parsing senderID: %v\n", err)
		return
	}
	time, err := strconv.ParseInt(command[2], 10, 64)
	if err != nil {
		fmt.Printf("\u001b[93mCLIENT:\u001b[0m Error parsing event time: %v\n", err)
		return
	}
	r, err := c.MakeEvent(context.Background(), &eventapi.MakeEventRequest{
		SenderID: senderID,
		Time:     time,
		Name:     command[3],
	})
	if err != nil {
		fmt.Printf("\u001b[93mCLIENT:\u001b[0m Could not make event: %v\n", err)
		return
	}
	fmt.Printf("\u001b[93mCLIENT:\u001b[0m Created{Event ID: %d}\n", r.GetEventID())
}

func GetEvent(command []string, c eventapi.EventManagerClient, ctx context.Context) {
	if len(command) != 3 {
		fmt.Printf("\u001b[93mCLIENT:\u001b[0m Usage: %s <sender-id> <event-id>", command[0])
		return
	}
	senderID, err := strconv.ParseInt(command[1], 10, 64)
	if err != nil {
		fmt.Printf("\u001b[93mCLIENT:\u001b[0m Error parsing senderID: %v\n", err)
		return
	}
	eventID, err := strconv.ParseInt(command[2], 10, 64)
	if err != nil {
		fmt.Printf("\u001b[93mCLIENT:\u001b[0m Error parsing eventID: %v\n", err)
		return
	}
	r, err := c.GetEvent(ctx, &eventapi.GetEventRequest{
		SenderID: senderID,
		EventID:  eventID,
	})
	if err != nil {
		fmt.Printf("\u001b[93mCLIENT:\u001b[0m Error getting event: %v\n", err)
		return
	} else if r == nil {
		fmt.Printf("\u001b[93mCLIENT:\u001b[0m Event not found\n")
		return
	} else {
		fmt.Printf("\u001b[93mCLIENT:\u001b[0m Event{sender_id:%d, eventId:%d, time:%d, name:%s}\n",
			r.Event.SenderID, r.Event.EventID, r.Event.Time, r.Event.Name)
	}
}

func DeleteEvent(command []string, c eventapi.EventManagerClient, ctx context.Context) {
	if len(command) != 3 {
		fmt.Printf("\u001b[93mCLIENT:\u001b[0m Usage: %s <sender-id> <event-id>\n", command[0])
		return
	}
	senderID, err := strconv.ParseInt(command[1], 10, 64)
	if err != nil {
		fmt.Printf("\u001b[93mCLIENT:\u001b[0m Error parsing senderID: %v\n", err)
		return
	}
	eventID, err := strconv.ParseInt(command[2], 10, 64)
	if err != nil {
		fmt.Printf("\u001b[93mCLIENT:\u001b[0m Error parsing eventID: %v\n", err)
		return
	}
	r, err := c.DeleteEvent(ctx, &eventapi.GetEventRequest{
		SenderID: senderID,
		EventID:  eventID,
	})
	if err != nil {
		fmt.Printf("\u001b[93mCLIENT:\u001b[0m Error getting event: %v\n", err)
		return
	} else if r == nil {
		fmt.Printf("\u001b[93mCLIENT:\u001b[0m Event not found\n")
		return
	} else {
		fmt.Printf("\u001b[93mCLIENT:\u001b[0m %s\n", r.Deleteresponse)
	}

}

func GetEvents(command []string, c eventapi.EventManagerClient, ctx context.Context) {
	if len(command) != 4 {
		fmt.Printf("\u001b[93mCLIENT:\u001b[0m Usage: %s <sender-id> <from-time> <to-time>\n", command[0])
		return
	}
	senderID, err := strconv.ParseInt(command[1], 10, 64)
	if err != nil {
		fmt.Printf("\u001b[93mCLIENT:\u001b[0m Error parsing senderID: %v\n", err)
		return
	}
	fromtime, err := strconv.ParseInt(command[2], 10, 64)
	if err != nil {
		fmt.Printf("\u001b[93mCLIENT:\u001b[0m Error parsing from time: %v\n", err)
		return
	}
	totime, err := strconv.ParseInt(command[3], 10, 64)
	if err != nil {
		fmt.Printf("\u001b[93mCLIENT:\u001b[0m Error parsing to time: %v\n", err)
		return
	}
	stream, err := c.GetEvents(ctx, &eventapi.GetEventsRequest{
		SenderID: senderID,
		Fromtime: fromtime,
		Totime:   totime,
	})
	if err != nil {
		fmt.Printf("\u001b[93mCLIENT:\u001b[0m Error getting events: %v\n", err)
		return
	}
	for {
		r, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Printf("\u001b[93mCLIENT:\u001b[0m Error receiving events: %v\n", err)
			return
		}
		if r.Events != nil {
			fmt.Printf("\u001b[93mCLIENT:\u001b[0m Events by %d from %d to %d:\n", senderID, fromtime, totime)
			for _, event := range r.Events {
				fmt.Printf("\u001b[93m\tEvent{eventId:%d, time:%d, name:%s}\n", event.EventID, event.Time, event.Name)
			}
		} else {
			fmt.Println("\u001b[93m > No events found.")
		}
	}
}
func Exit(command []string, c eventapi.EventManagerClient, ctx context.Context, sdid *int64) {
	r, err := c.Exit(context.Background(), &eventapi.ExitRequest{
		SenderID: *sdid,
	})
	FailOnError(err, "Failed to get sender ID")
	fmt.Println("\u001b[93m >", r.Goodbye, "")
	os.Exit(0)
}
