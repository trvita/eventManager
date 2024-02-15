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
	"context"
	"event/api/eventapi"
	"flag"
	"fmt"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	addr = flag.String("dst", "localhost", "The server IP address")
	port = flag.Int("p", 50051, "The server port")
)

func main() {
	flag.Parse()
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", *addr, *port), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()
	c := eventapi.NewEventClient(conn)
	r, err := c.MakeEvent(context.Background(), &eventapi.MakeEventRequest{
		SenderID: 123,
		Time:     20,
		Name:     "some event",
	})
	if err != nil {
		log.Fatalf("could not make event: %v", err)
	}
	log.Printf("Event ID: %d", r.GetEventID())
}
