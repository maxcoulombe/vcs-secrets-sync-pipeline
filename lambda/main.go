package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/mq"
	"github.com/hashicorp/go-secure-stdlib/awsutil"
	"github.com/thanhpk/randstr"

	"pack.ag/amqp"
)

type Step string

const (
	Create   Step = "CREATE"
	Set      Step = "SET"
	Test     Step = "TEST"
	Finalise Step = "FINALISE"
)

type LambdaEvent struct {
	Body string `json:"body"`
}

type Event struct {
	PreviousUsername string `json:"previousUsername,omitempty"`
	Username         string `json:"username,omitempty"`
	Password         string `json:"password,omitempty"`
	BrokerId         string `json:"brokerId"`
	Step             Step   `json:"step"`
}

type LambdaResponse struct {
	IsBase64Encoded bool   `json:"isBase64Encoded"`
	StatusCode      int    `json:"statusCode"`
	Body            string `json:"body"`
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
		return Response{Success: false}, fmt.Errorf("unexpected error while trying to create user on broker %s: %w", event.BrokerId, err)
	} else {
		return Response{
			Success:     true,
			NewUsername: generatedUsername,
			NewPassword: generatedPassword,
		}, nil
	}
}

func handleSetStep(ctx context.Context, event Event, mqClient *mq.MQ) (Response, error) {
	if event.Username != "" && event.Password != "" {
		_, err := mqClient.CreateUserWithContext(ctx, &mq.CreateUserRequest{
			BrokerId: aws.String(event.BrokerId),
			Username: aws.String(event.Username),
			Password: aws.String(event.Password),
		})
		if err != nil {
			return Response{Success: false}, fmt.Errorf("unexpected error while trying to update user %s on broker %s: %w", event.Username, event.BrokerId, err)
		}
	}

	_, err := mqClient.RebootBrokerWithContext(ctx, &mq.RebootBrokerInput{
		BrokerId: aws.String(event.BrokerId),
	})
	if err != nil {
		return Response{Success: false}, fmt.Errorf("unexpected error while trying to update user %s on broker %s: %w", event.Username, event.BrokerId, err)
	} else {
		return Response{Success: true}, nil
	}
}

func handleTestStep(event Event, region string) (Response, error) {
	_, err := amqp.Dial(fmt.Sprintf("amqps://%s-1.mq.%s.amazonaws.com:5671", event.BrokerId, region),
		amqp.ConnSASLPlain(event.Username, event.Password),
		amqp.ConnConnectTimeout(5*time.Second),
	)

	if err != nil {
		return Response{Success: false}, fmt.Errorf(
			"unable to connect to broker %s with the provided credentials: %w", event.BrokerId, err)
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
		return Response{Success: false}, fmt.Errorf("unexpected error while trying to delete previous user %s on broker %s: %w", event.PreviousUsername, event.BrokerId, err)
	} else {
		return Response{Success: true}, nil
	}
}

func handleRequest(ctx context.Context, event LambdaEvent) (LambdaResponse, error) {
	e := Event{}
	err := json.Unmarshal([]byte(event.Body), &e)
	if err != nil {
		return LambdaResponse{
			StatusCode: 500,
		}, fmt.Errorf("unable to parse event: %v", event.Body)
	}

	mqClient, region, err := createMQClient()
	if err != nil {
		return LambdaResponse{
			StatusCode: 500,
		}, fmt.Errorf("unable to create MQ client: %s", event.Body)
	}

	r := Response{}
	switch e.Step {
	case Create:
		r, err = handleCreateStep(ctx, e, mqClient)
		break
	case Set:
		r, err = handleSetStep(ctx, e, mqClient)
		break
	case Test:
		r, err = handleTestStep(e, region)
		break
	case Finalise:
		r, err = handleFinalizeStep(ctx, e, mqClient)
		break
	default:
		return LambdaResponse{
			StatusCode: 500,
		}, fmt.Errorf("step %v is not supported", e.Step)
	}
	if err != nil {
		return LambdaResponse{
			StatusCode: 500,
		}, fmt.Errorf("error handling request: %w", err)
	}

	body, err := json.Marshal(r)
	if err != nil {
		return LambdaResponse{
			StatusCode: 500,
		}, fmt.Errorf("unable to marshal response body: %w", err)
	}

	return LambdaResponse{StatusCode: 200, Body: string(body)}, err
}

func main() {
	lambda.Start(handleRequest)
}
