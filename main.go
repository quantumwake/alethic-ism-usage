package main

import (
	"context"
	"encoding/json"
	"github.com/quantumwake/alethic-ism-core-go/pkg/repository/usage"
	"github.com/quantumwake/alethic-ism-core-go/pkg/routing"
	"github.com/quantumwake/alethic-ism-core-go/pkg/routing/nats"
	"log"
	"os"
)

var (
	natsRoute *routing.Route
	backend   *usage.BackendStorage
)

func onMessageReceived(ctx context.Context, envelop routing.MessageEnvelop) {
	defer func(envelop routing.MessageEnvelop, ctx context.Context) {
		err := envelop.Ack(ctx)
		if err != nil {

		}
	}(envelop, ctx)

	data, err := envelop.MessageRaw()
	if err != nil {
		log.Printf("Error getting raw message: %v", err)
		return
	}

	var model usage.Usage

	// unmarshal the message into a usage struct
	if err = json.Unmarshal(data, &model); err != nil {
		log.Printf("Error unmarshalling usage: %v", err)
		return
	}

	err = backend.InsertUsage(&model)
	if err != nil {
		log.Printf("Error inserting usage: %v", err)
	}

}

func main() {
	ctx := context.Background()

	// connect to usage database
	dsn, ok := os.LookupEnv("DSN")
	if !ok {
		dsn = "host=localhost port=5432 user=postgres password=postgres1 dbname=postgres sslmode=disable"
		log.Println("DSN environment variable not set")
	}
	backend = usage.NewBackend(dsn)

	//routingFile, ok := os.LookupEnv("ROUTING_FILE")
	//if !ok {
	//	routingFile = "../routing-nats.yaml"
	//}

	//// load the nats routing table
	//config, err := nats.LoadConfig(routingFile)
	//if err != nil {
	//	log.Fatal(err)
	//}

	// find the usage route we want to listen on
	//route, err := routes.FindRouteBySelector("processor/usage")
	//if err != nil {
	//	log.Fatal(err)
	//}

	// create a new route and subscribe to inbound messages
	route, err := nats.NewRouteSubscriberUsingSelector(ctx, "processor/usage", onMessageReceived)
	if err != nil {
		return
	}
	err = route.Subscribe(ctx)
	if err != nil {
		log.Fatalf("unable to subscribe to NATS: %v, route: %v", route, err)
	}

	// TODO need to have graceful shutdown here
	select {}
}
