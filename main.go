package main

import (
	"alethic-ism-usage/pkg/data"
	"context"
	"encoding/json"
	"github.com/nats-io/nats.go"
	"github.com/quantumwake/alethic-ism-core-go/pkg/routing"
	"log"
	"os"
)

var (
	natsRoute  *routing.NATSRoute
	dataAccess *data.Access
)

func onMessageReceived(route *routing.NATSRoute, msg *nats.Msg) {
	var usage data.Usage
	err := json.Unmarshal(msg.Data, &usage)
	if err != nil {
		log.Printf("Error unmarshalling usage: %v", err)
		return
	}

	err = dataAccess.InsertUsage(&usage)
	if err != nil {
		log.Printf("Error inserting usage: %v", err)
	}

	err = msg.Ack()
	if err != nil {
		return
	}
}

func main() {
	// connect to usage database
	dsn, ok := os.LookupEnv("DSN")
	if !ok {
		dsn = "host=localhost port=5432 user=postgres password=postgres1 dbname=postgres sslmode=disable"
		log.Println("DSN environment variable not set")
	}
	dataAccess = data.NewDataAccess(dsn)

	routingFile, ok := os.LookupEnv("ROUTING_FILE")
	if !ok {
		routingFile = "../routing-nats.yaml"
	}

	// load the nats routing table
	routes, err := routing.LoadConfig(routingFile)
	if err != nil {
		log.Fatal(err)
	}

	// find the usage route we want to listen on
	route, err := routes.FindRouteBySelector("processor/usage")
	if err != nil {
		log.Fatal(err)
	}

	// create a new route and subscribe to inbound messages
	natsRoute = routing.NewNATSRoute(route, onMessageReceived)
	err = natsRoute.Subscribe(context.TODO())
	if err != nil {
		log.Fatalf("unable to subscribe to NATS: %v, route: %v", route, err)
	}

	// TODO need to have graceful shutdown here
	select {}
}
