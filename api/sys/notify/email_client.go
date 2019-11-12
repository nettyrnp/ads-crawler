package notify

import (
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/nettyrnp/ads-crawler/api/common"
	"github.com/pkg/errors"
)

type SESConfig struct {
	AWSKey       string `json:"awsKey"`
	AWSSecretKey string `json:"awsSecretKey"`
	AWSRegion    string `json:"awsRegion"`
	Sender       string `json:"sender"`
}

func (c SESConfig) Validate() []error {
	var errs []error
	if len(c.AWSKey) == 0 {
		errs = append(errs, errors.New("SESConfig requires a non-empty AWSKey value"))
	}
	if len(c.AWSSecretKey) == 0 {
		errs = append(errs, errors.New("SESConfig requires a non-empty AWSSecretKey value"))
	}
	if len(c.AWSRegion) == 0 {
		errs = append(errs, errors.New("SESConfig requires a non-empty AWSRegion value"))
	}
	if len(c.Sender) == 0 {
		errs = append(errs, errors.New("SESConfig requires a non-empty Sender email"))
	}
	return errs
}

type EmailNotifier interface {
	Send(ctx context.Context, toAddr, msg string) error
}

type EmailClient struct {
	Kind string
	SESConfig
}

func (c *EmailClient) Send(ctx context.Context, toAddr, msg string) error {
	awsSession := session.New(&aws.Config{
		Region:      aws.String(c.AWSRegion),
		Credentials: credentials.NewStaticCredentials(c.AWSKey, c.AWSSecretKey, ""),
	})

	sesSession := ses.New(awsSession)

	sesEmailInput := &ses.SendEmailInput{
		Destination: &ses.Destination{
			ToAddresses: []*string{aws.String(toAddr)},
		},
		Message: &ses.Message{
			Body: &ses.Body{
				Html: &ses.Content{
					Data: aws.String(msg)},
			},
			Subject: &ses.Content{
				Data: aws.String("Email verification"),
			},
		},
		Source: aws.String(c.Sender),
		ReplyToAddresses: []*string{
			aws.String(c.Sender),
		},
	}

	_, err := sesSession.SendEmail(sesEmailInput)
	if err != nil {
		common.LogError(err.Error())
		return err
	}
	common.LogInfof(">> sent email to %v: %v\n", c.Kind, msg)
	return nil
}
