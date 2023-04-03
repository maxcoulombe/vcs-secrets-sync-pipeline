package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/hashicorp/eventlogger"
	"github.com/hashicorp/secrets-store-syncer/sinkers/store"
	"github.com/hashicorp/secrets-store-syncer/syncer"
)

const (
	FormatterNodeName = "formatter"
	StoreNodeName     = "store"
	PipelineID        = "vcs-secrets-sync-pipeline"
	EventType         = "vcs-secrets-sync-event"
	GlobalTagKey      = "hashicorp:vcs:secret"
	SecretKeyPattern  = "hashicorp/vcs/%s/%s"
)

type SecretsWriteEvent struct {
	PublicTenantId   string           `json:"public_tenant_id"`
	PrivateTenantId  string           `json:"private_tenant_id"`
	AppName          string           `json:"app_name"`
	SecretName       string           `json:"secret_name"`
	SecretToken      string           `json:"secret_token"`
	IntegrationType  string           `json:"integration_type"`
	IntegrationToken string           `json:"integration_token"`
	Operation        syncer.Operation `json:"operation"`
}

func parseEvent(sqsEvent events.SQSEvent) ([]*SecretsWriteEvent, error) {
	var secretsWriteEvents []*SecretsWriteEvent

	for _, record := range sqsEvent.Records {
		var secretsWriteEvent *SecretsWriteEvent

		err := json.Unmarshal([]byte(record.Body), &secretsWriteEvent)
		if err != nil {
			log.Printf(fmt.Errorf("ERROR: %w", err).Error())
			return nil, err
		}
		log.Printf("SECRET EVENT: %+v\n", secretsWriteEvent)

		secretsWriteEvents = append(secretsWriteEvents, secretsWriteEvent)
	}

	return secretsWriteEvents, nil
}

// TODO Get this from the Tokenization service
func getSecretValue(event *SecretsWriteEvent) (string, error) {
	return "test", nil
}

// TODO Add a cache to reuse the syncer between lambda invocations
func getOrInitStore(event *SecretsWriteEvent) (*syncer.SecretsSyncer, error) {
	connDetails, err := getConnectionDetails(event)
	if err != nil {

	}

	return initStoreSinker(event, connDetails)
}

// TODO Get this from the Tokenization service
func getConnectionDetails(event *SecretsWriteEvent) (map[string]any, error) {
	return map[string]any{}, nil
}

func initStoreSinker(event *SecretsWriteEvent, connDetails map[string]any) (*syncer.SecretsSyncer, error) {
	sync := syncer.NewSyncer()

	err := sync.RegisterNode(FormatterNodeName, syncer.NewMetadataFormatter(
		map[string]string{GlobalTagKey: ""},
		"Source: <HCP URL>",
		"",
	))
	if err != nil {
		log.Printf(fmt.Errorf("ERROR: %w", err).Error())
		return nil, err
	}

	err = store.CreateAndRegisterStore(sync, StoreNodeName, store.Type(event.IntegrationType), connDetails)
	if err != nil {
		log.Printf(fmt.Errorf("ERROR: %w", err).Error())
		return nil, err
	}

	err = sync.RegisterPipeline(eventlogger.Pipeline{
		PipelineID: PipelineID,
		EventType:  EventType,
		NodeIDs: []eventlogger.NodeID{
			FormatterNodeName,
			StoreNodeName,
		},
	})
	if err != nil {
		log.Printf(fmt.Errorf("ERROR: %w", err).Error())
		return nil, err
	}

	return sync, nil
}

// TODO Reach back to the Secrets-Service with the sync result
func reportSecretSyncStatus(event *SecretsWriteEvent, err error) error {
	if err != nil {
		log.Printf(fmt.Errorf("ERROR: %w", err).Error())
	}

	return err
}

func handleRequest(ctx context.Context, sqsEvent events.SQSEvent) error {
	secretWriteEvents, err := parseEvent(sqsEvent)
	if err != nil {
		log.Printf(fmt.Errorf("ERROR: %w", err).Error())
		return err
	}

	for _, event := range secretWriteEvents {
		secretValue, err := getSecretValue(event)
		if err != nil {
			return reportSecretSyncStatus(event, err)
		}

		syncEvent := syncer.SecretSyncEvent{
			Key:       fmt.Sprintf(SecretKeyPattern, event.AppName, event.SecretName),
			Value:     secretValue,
			Operation: event.Operation,
		}
		log.Printf("SYNC EVENT: %+v\n", syncEvent)

		sinker, err := getOrInitStore(event)
		if err != nil {
			return reportSecretSyncStatus(event, err)
		}

		status, err := sinker.Send(ctx, EventType, syncEvent)
		if err != nil || len(status.Warnings) != 0 {
			return reportSecretSyncStatus(event, err)
		}

		err = reportSecretSyncStatus(event, nil)
		if err != nil {
			log.Printf(fmt.Errorf("ERROR: %w", err).Error())
			return err
		}
	}

	return nil
}

func main() {
	lambda.Start(handleRequest)
}
