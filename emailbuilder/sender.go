package emailbuilder

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/cenkalti/backoff/v4"
	"github.com/samber/lo"
)

type SESDriverConfig struct {
	FilterAddressesRegex                      string
	FromEmailAddress                          string
	FromName                                  string
	SubjectCharset                            string
	HTMLBodyCharset                           string
	TextBodyCharset                           string
	ConfigurationSetName                      string
	FeedbackForwardingEmailAddress            string
	FeedbackForwardingEmailAddressIdentityArn string
	FromEmailAddressIdentityArn               string
	ContactListName                           string
	TopicName                                 string
}

type DefaultSESDriver struct {
	Client *sesv2.Client
	Config SESDriverConfig
}

func NewDefaultSESDriver(config SESDriverConfig) *DefaultSESDriver {
	return &DefaultSESDriver{
		Client: sesv2.NewFromConfig(MustNewAWSConfig(context.Background())),
		Config: config,
	}
}

func (d *DefaultSESDriver) SendEmail(ctx context.Context, input *sesv2.SendEmailInput) (*sesv2.SendEmailOutput, error) {
	var output *sesv2.SendEmailOutput
	var err error
	err = backoff.Retry(func() error {
		output, err = d.Client.SendEmail(ctx, input)
		return err
	}, backoff.NewExponentialBackOff(
		backoff.WithInitialInterval(500*time.Millisecond),
		backoff.WithMaxInterval(2*time.Second),
		backoff.WithMaxElapsedTime(5*time.Second),
	))
	if err != nil {
		return nil, err
	}
	return output, nil
}

func MustNewAWSConfig(ctx context.Context) aws.Config {
	return lo.Must1(awsconfig.LoadDefaultConfig(ctx))
}
