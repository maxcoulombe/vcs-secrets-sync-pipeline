package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/mq"
	"github.com/hashicorp/go-secure-stdlib/awsutil"
	"github.com/thanhpk/randstr"
	"log"
	"pack.ag/amqp"
)

// Get these from the Terraform run
const (
	brokerId        = "b-5612de2f-d7b4-4479-9612-e1627518c444"
	currentUsername = "potato"
)

type Step int

const (
	Create Step = iota
	Set
	Test
	Finalise
)

type Event struct {
	PreviousUsername string `json:"previousUsername"`
	Username         string `json:"username"`
	Password         string `json:"password"`
	BrokerId         string `json:"brokerId"`
	Step             Step   `json:"step"`
}

type Response struct {
	Success     bool   `json:"success"`
	NewUsername string `json:"newUsername"`
	NewPassword string `json:"newPassword"`
}

func createMQClient() (*mq.MQ, string, error) {
	creds, err := awsutil.RetrieveCreds("", "", "", nil)
	if err != nil {
		return nil, "", fmt.Errorf("unable to rerieve AWS credentials from provider chain: %w", err)
	}

	region, err := awsutil.GetRegion("")
	if err != nil {
		return nil, "", fmt.Errorf("unable to determine AWS region from config nor context: %w", err)
	}

	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: creds,
	})
	if err != nil {
		return nil, "", fmt.Errorf("unable to initialize AWS session: %w", err)
	}

	return mq.New(sess), region, nil
}

func handleCreateStep(ctx context.Context, event Event, mqClient *mq.MQ) (Response, error) {
	generatedUsername := randstr.Hex(16)
	generatedPassword := randstr.Hex(16)

	_, err := mqClient.CreateUserWithContext(ctx, &mq.CreateUserRequest{
		BrokerId: aws.String(event.BrokerId),
		Username: aws.String(generatedUsername),
		Password: aws.String(generatedPassword)})
	if err != nil {
		return Response{Success: false}, fmt.Errorf("unexpected error while trying to create user %s on broker %s", event.Username, event.BrokerId)
	} else {
		return Response{
			Success:     true,
			NewUsername: generatedUsername,
			NewPassword: generatedPassword,
		}, nil
	}
}

func handleSetStep(ctx context.Context, event Event, mqClient *mq.MQ) (Response, error) {
	_, err := mqClient.UpdateUserWithContext(ctx, &mq.UpdateUserRequest{
		BrokerId: aws.String(event.BrokerId),
		Username: aws.String(event.Username),
		Password: aws.String(event.Password),
	})
	if err != nil {
		return Response{Success: false}, fmt.Errorf("unexpected error while trying to update user %s on broker %s", event.Username, event.BrokerId)
	} else {
		return Response{
			Success:     true,
			NewUsername: event.Username,
			NewPassword: event.Password,
		}, nil
	}
}

func handleTestStep(event Event, region string) (Response, error) {
	client, err := amqp.Dial(fmt.Sprintf("amqps://%s-1.mq.%s.amazonaws.com:5671", event.BrokerId, region),
		amqp.ConnSASLPlain(event.Username, event.Password),
	)
	defer client.Close()

	if err != nil {
		return Response{Success: false}, fmt.Errorf("unable to connect to broker %s with the provided credentials", event.BrokerId)
	} else {
		return Response{Success: true}, nil
	}
}

func handleFinalizeStep(ctx context.Context, event Event, mqClient *mq.MQ) (Response, error) {
	_, err := mqClient.DeleteUserWithContext(ctx, &mq.DeleteUserInput{
		BrokerId: aws.String(event.BrokerId),
		Username: aws.String(event.PreviousUsername),
	})
	if err != nil {
		return Response{Success: false}, fmt.Errorf("unexpected error while trying to delete user %s on broker %s", event.Username, event.BrokerId)
	} else {
		return Response{Success: true}, nil
	}
}

func handleRequest(ctx context.Context, event Event) (Response, error) {
	mqClient, region, err := createMQClient()
	if err != nil {
		return Response{Success: false}, fmt.Errorf("unable to create MQ client: %w", err)
	}

	switch event.Step {
	case Create:
		return handleCreateStep(ctx, event, mqClient)
	case Set:
		return handleSetStep(ctx, event, mqClient)
	case Test:
		return handleTestStep(event, region)
	case Finalise:
		return handleFinalizeStep(ctx, event, mqClient)
	default:
		return Response{Success: true}, fmt.Errorf("step %v is not supported", event.Step)
	}
}

func main() {
	// Uncomment to run locally
	ctx := context.Background()

	event := Event{
		BrokerId: brokerId,
		Step:     Create,
	}
	r, err := handleRequest(ctx, event)
	if err != nil {
		log.Fatal(err)
	}

	event = Event{
		Username: r.NewUsername,
		Password: r.NewPassword,
		BrokerId: brokerId,
		Step:     Set,
	}
	r, err = handleRequest(ctx, event)
	if err != nil {
		log.Fatal(err)
	}

	event = Event{
		Username: r.NewUsername,
		Password: r.NewPassword,
		BrokerId: brokerId,
		Step:     Test,
	}
	r, err = handleRequest(ctx, event)
	if err != nil {
		log.Fatal(err)
	}

	event = Event{
		PreviousUsername: currentUsername,
		BrokerId:         brokerId,
		Step:             Finalise,
	}
	r, err = handleRequest(ctx, event)
	if err != nil {
		log.Fatal(err)
	}

	// Uncomment when running on AWS
	//lambda.Start(handleRequest)
}
