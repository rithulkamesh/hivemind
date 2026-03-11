package email

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
)

// SES sends transactional email via AWS SES.
type SES struct {
	client  *ses.Client
	from    string
	replyTo string
}

// NewSES creates an SES client from environment.
func NewSES(ctx context.Context, region, from, replyTo string) (*SES, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}
	return &SES{
		client:  ses.NewFromConfig(cfg),
		from:    from,
		replyTo: replyTo,
	}, nil
}

// Send sends a simple email (subject + body).
func (s *SES) Send(ctx context.Context, to, subject, bodyText, bodyHTML string) error {
	body := &types.Body{}
	if bodyHTML != "" {
		body.Html = &types.Content{Data: aws.String(bodyHTML)}
	}
	if bodyText != "" {
		body.Text = &types.Content{Data: aws.String(bodyText)}
	}
	_, err := s.client.SendEmail(ctx, &ses.SendEmailInput{
		Source: aws.String(s.from),
		Destination: &types.Destination{
			ToAddresses: []string{to},
		},
		Message: &types.Message{
			Subject: &types.Content{Data: aws.String(subject)},
			Body:    body,
		},
		ReplyToAddresses: []string{s.replyTo},
	})
	return err
}

// SendPublishSuccess sends "Your package X Y is now live".
func (s *SES) SendPublishSuccess(ctx context.Context, to, name, version string) error {
	subject := "Package " + name + " " + version + " is now live"
	body := "Your package " + name + " " + version + " has passed verification and is now published on the Hivemind Registry."
	return s.Send(ctx, to, subject, body, "")
}

// SendPublishFailed sends verification failure notification.
func (s *SES) SendPublishFailed(ctx context.Context, to, name, version, report string) error {
	subject := "Verification failed for " + name + " " + version
	body := "Verification failed for " + name + " " + version + ".\n\nReport:\n" + report
	return s.Send(ctx, to, subject, body, "")
}
