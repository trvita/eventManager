/*
 *
 * Copyright 2015 gRPC authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

// Package main implements a client for Greeter service.
package main

import (
	"bufio"
	"context"
	"event/api/eventapi"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

var (
	addr = flag.String("dst", "localhost", "The server IP address")
	port = flag.Int("p", 50051, "The server port")
	sdid = flag.Int64("sender-id", 0, "Sender ID")
)

func main() {
	flag.Parse()
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", *addr, *port), grpc.WithTransportCredentials(insecure.NewCredentials()))
	failOnError(err, "Failed to connect to server")
	defer conn.Close()
	c := eventapi.NewEventManagerClient(conn)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
	r, err := c.GreetSender(context.Background(), &eventapi.GreetSenderRequest{
		SenderID: *sdid,
	})
	failOnError(err, "Failed to get sender ID")
	*sdid = r.SenderID
	log.Printf("Your ID: %d}", *sdid)

	for {
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		command := strings.Fields(scanner.Text())

		switch strings.ToLower(command[0]) {
		case "makeevent":
			if len(command) != 4 {
				log.Printf("Usage: %s <sender-id> <event-time> <event-name>", command[0])
				continue
			}
			senderID, err := strconv.ParseInt(command[1], 10, 64)
			if err != nil {
				log.Printf("Error parsing senderID: %v", err)
				continue
			}
			time, err := strconv.ParseInt(command[2], 10, 64)
			if err != nil {
				log.Printf("Error parsing event time: %v", err)
				continue
			}
			r, err := c.MakeEvent(context.Background(), &eventapi.MakeEventRequest{
				SenderID: senderID,
				Time:     time,
				Name:     command[3],
			})
			if err != nil {
				log.Printf("Could not make event: %v", err)
				continue
			}
			log.Printf("Created{Event ID: %d}", r.GetEventID())

		case "getevent":
			if len(command) != 4 {
				log.Printf("Usage: %s <sender-id> <event-id>", command[0])
				continue
			}
			senderID, err := strconv.ParseInt(command[1], 10, 64)
			if err != nil {
				log.Printf("Error parsing senderID: %v", err)
				continue
			}
			eventID, err := strconv.ParseInt(command[2], 10, 64)
			if err != nil {
				log.Printf("Error parsing eventID: %v", err)
				continue
			}
			r, err := c.GetEvent(ctx, &eventapi.GetEventRequest{
				SenderID: senderID,
				EventID:  eventID,
			})
			if err != nil {
				log.Printf("Error getting event: %v", err)
				continue
			} else if r == nil {
				fmt.Println("Event not found")
				continue
			} else {
				fmt.Printf("Event{sender_id:%d, eventId:%d, time:%d, name:%s}\n",
					r.Event.SenderID, r.Event.EventID, r.Event.Time, r.Event.Name)
			}

		case "deleteevent":
			if len(command) != 4 {
				log.Printf("Usage: %s <sender-id> <event-id>", command[0])
				continue
			}
			senderID, err := strconv.ParseInt(command[1], 10, 64)
			if err != nil {
				log.Printf("Error parsing senderID: %v", err)
				continue
			}
			eventID, err := strconv.ParseInt(command[2], 10, 64)
			if err != nil {
				log.Printf("Error parsing eventID: %v", err)
				continue
			}
			r, err := c.DeleteEvent(ctx, &eventapi.GetEventRequest{
				SenderID: senderID,
				EventID:  eventID,
			})
			if err != nil {
				log.Printf("Error getting event: %v", err)
				continue
			} else if r == nil {
				fmt.Println("Event not found")
				continue
			} else {
				fmt.Printf("%s\n", r.Deleteresponse)
			}

		case "getevents":
			if len(command) != 4 {
				log.Printf("Usage: %s <sender-id> <from-time> <to-time>", command[0])
				continue
			}
			senderID, err := strconv.ParseInt(command[1], 10, 64)
			if err != nil {
				log.Printf("Error parsing senderID: %v", err)
				continue
			}
			fromtime, err := strconv.ParseInt(command[2], 10, 64)
			if err != nil {
				log.Printf("Error parsing from time: %v", err)
				continue
			}
			totime, err := strconv.ParseInt(command[3], 10, 64)
			if err != nil {
				log.Printf("Error parsing to time: %v", err)
				continue
			}
			stream, err := c.GetEvents(ctx, &eventapi.GetEventsRequest{
				SenderID: senderID,
				Fromtime: fromtime,
				Totime:   totime,
			})
			if err != nil {
				log.Printf("Error getting events: %v", err)
				continue
			}
			fmt.Printf("Events by %d from %d to %d:\n", senderID, fromtime, totime)
			for {
				r, err := stream.Recv()
				if err == io.EOF {
					break
				}
				if err != nil {
					log.Printf("Error receiving events: %v", err)
					continue
				}
				for _, event := range r.Events {
					fmt.Printf("Event{eventId:%d, time:%d, name:%s}\n", event.EventID, event.Time, event.Name)
				}
			}
		case "exit":
			r, err := c.Exit(context.Background(), &eventapi.ExitRequest{
				SenderID: *sdid,
			})
			failOnError(err, "Failed to get sender ID")
			log.Printf(r.Goodbye)
			os.Exit(0)
		default:
			fmt.Println("Unknown command. Available commands: MakeEvent, GetEvent, DeleteEvent, GetEvents, Exit")
		}
	}
}
