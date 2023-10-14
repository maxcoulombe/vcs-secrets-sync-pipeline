package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type SecretsWriteEvent struct {
	GroupID string `json:"group_id"`
}

func handleRequest(_ context.Context, sqsEvent events.SQSEvent) error {
	for _, record := range sqsEvent.Records {
		var event *SecretsWriteEvent

		err := json.Unmarshal([]byte(record.Body), &event)
		if err != nil {
			log.Printf(fmt.Errorf("ERROR: %w", err).Error())
			return err
		}

		if event.GroupID == "error" {
			time.Sleep(time.Second * 3)
			return errors.New("oh no")
		} else {
			time.Sleep(time.Second * 3)
			log.Printf("SECRET EVENT: %+v\n", event)
		}
	}

	return nil
}

func main() {
	lambda.Start(handleRequest)
}
