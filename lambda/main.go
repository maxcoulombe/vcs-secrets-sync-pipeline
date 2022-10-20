package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/mq"
	"github.com/hashicorp/go-secure-stdlib/awsutil"
	"github.com/thanhpk/randstr"
)

type Step string

const (
	Create   Step = "create"
	Set      Step = "set"
	Test     Step = "test"
	Finalize Step = "finalize"
)

type Event struct {
	Stage          Step        `json:"stage"`
	StaticInput    StaticInput `json:"static_input,omitempty"`
	CurrentVersion Secret      `json:"current_version,omitempty"`
	NextVersion    Secret      `json:"next_version,omitempty"`
	SecretPath     string      `json:"secret_path,omitempty"`
}

type Response struct {
	Error       string `json:"error,omitempty"`
	NextVersion Secret `json:"next_version,omitempty"`
}

type StaticInput struct {
	BrokerId string `json:"broker_id,omitempty"`
}

type Secret struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func createMQClient() (*mq.MQ, error) {
	creds, err := awsutil.RetrieveCreds("", "", "", nil)
	if err != nil {
		return nil, fmt.Errorf("unable to rerieve AWS credentials from provider chain: %w", err)
	}

	region, err := awsutil.GetRegion("")
	if err != nil {
		return nil, fmt.Errorf("unable to determine AWS region from config nor context: %w", err)
	}

	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: creds,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to initialize AWS session: %w", err)
	}

	return mq.New(sess), nil
}

func handleCreateStep(ctx context.Context, event Event, mqClient *mq.MQ) (Response, error) {
	generatedUsername := randstr.Hex(16)
	generatedPassword := randstr.Hex(16)

	_, err := mqClient.CreateUserWithContext(ctx, &mq.CreateUserRequest{
		BrokerId: aws.String(event.StaticInput.BrokerId),
		Username: aws.String(generatedUsername),
		Password: aws.String(generatedPassword)})
	if err != nil {
		err = fmt.Errorf("unexpected error while trying to create user on broker %s: %w", event.StaticInput.BrokerId, err)
		return Response{Error: err.Error()}, err
	} else {
		return Response{
			NextVersion: Secret{
				Username: generatedUsername,
				Password: generatedPassword,
			},
		}, nil
	}
}

func handleSetStep(ctx context.Context, event Event, mqClient *mq.MQ) (Response, error) {
	_, err := mqClient.CreateUserWithContext(ctx, &mq.CreateUserRequest{
		BrokerId: aws.String(event.StaticInput.BrokerId),
		Username: aws.String(event.NextVersion.Username),
		Password: aws.String(event.NextVersion.Password),
	})
	if err != nil && strings.Contains(err.Error(), "already exists") {
		// It's ok
	} else if err != nil {
		err = fmt.Errorf("unexpected error while trying to update user on broker %s: %w", event.StaticInput.BrokerId, err)
		return Response{Error: err.Error()}, err
	}

	_, err = mqClient.RebootBrokerWithContext(ctx, &mq.RebootBrokerInput{
		BrokerId: aws.String(event.StaticInput.BrokerId),
	})
	if err != nil && strings.Contains(err.Error(), "REBOOT_IN_PROGRESS") {
		// It's ok
	} else if err != nil {
		err = fmt.Errorf("unexpected error while trying to update user on broker %s: %w", event.StaticInput.BrokerId, err)
		return Response{Error: err.Error()}, err
	}

	return Response{NextVersion: Secret{
		Username: event.NextVersion.Username,
		Password: event.NextVersion.Password,
	}}, nil

}

func handleTestStep(_ context.Context, event Event, _ *mq.MQ) (Response, error) {
	// If you want to use the real Test
	//_, err := amqp.Dial(fmt.Sprintf("amqps://%s-1.mq.%s.amazonaws.com:5671", event.BrokerId, region),
	//	amqp.ConnSASLPlain(event.Username, event.Password),
	//	amqp.ConnConnectTimeout(5*time.Second),
	//)

	// If you want to succeed 10% of the time to test a retry mechanism
	//success := rand.Intn(100)%10 == 0
	//if !success {
	//	err := fmt.Errorf("unable to connect to broker %s with the provided credentials", event.StaticInput.BrokerId)
	//	return Response{Error: err.Error()}, err
	//} else {
	//	return Response{NextVersion: Secret{
	//		Username: event.NextVersion.Username,
	//		Password: event.NextVersion.Password,
	//	}}, nil
	//}

	// Short-circuiting to success
	return Response{NextVersion: Secret{
		Username: event.NextVersion.Username,
		Password: event.NextVersion.Password,
	}}, nil
}

func handleFinalizeStep(ctx context.Context, event Event, mqClient *mq.MQ) (Response, error) {
	if event.CurrentVersion.Username == "" {
		return Response{NextVersion: Secret{
			Username: event.NextVersion.Username,
			Password: event.NextVersion.Password,
		}}, nil
	}

	_, err := mqClient.DeleteUserWithContext(ctx, &mq.DeleteUserInput{
		BrokerId: aws.String(event.StaticInput.BrokerId),
		Username: aws.String(event.CurrentVersion.Username),
	})
	if err != nil {
		err = fmt.Errorf("unexpected error while trying to delete previous user on broker %s: %w", event.StaticInput.BrokerId, err)
		return Response{Error: err.Error()}, err
	} else {
		return Response{NextVersion: Secret{
			Username: event.NextVersion.Username,
			Password: event.NextVersion.Password,
		}}, nil
	}
}

func handleRequest(ctx context.Context, event Event) (Response, error) {
	mqClient, err := createMQClient()
	if err != nil {
		err = fmt.Errorf("unable to create MQ client: %w", err)
		return Response{
			Error: err.Error(),
		}, err
	}

	r := Response{}
	switch event.Stage {
	case Create:
		r, err = handleCreateStep(ctx, event, mqClient)
		break
	case Set:
		r, err = handleSetStep(ctx, event, mqClient)
		break
	case Test:
		r, err = handleTestStep(ctx, event, mqClient)
		break
	case Finalize:
		r, err = handleFinalizeStep(ctx, event, mqClient)
		break
	default:
		err = fmt.Errorf("stage %v is not supported", event.Stage)
		return Response{
			Error: err.Error(),
		}, err
	}
	if err != nil {
		err = fmt.Errorf("error handling request: %w", err)
		return Response{
			Error: err.Error(),
		}, err
	} else {
		return r, nil
	}
}

func main() {
	lambda.Start(handleRequest)
}
