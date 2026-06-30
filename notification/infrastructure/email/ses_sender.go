package email

import (
	"context"
	"errors"
	"fmt"

	"notification/domain/entity"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
)

type SesSender struct {
	client    *ses.Client
	fromEmail string
}

func NewSesSender(ctx context.Context, fromEmail string) (*SesSender, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("AWS設定の読み込みに失敗: %w", err)
	}

	return &SesSender{
		client:    ses.NewFromConfig(cfg),
		fromEmail: fromEmail,
	}, nil
}

func (s *SesSender) Send(ctx context.Context, to, subject, body string) error {
	input := &ses.SendEmailInput{
		Source: aws.String(s.fromEmail),
		Destination: &types.Destination{
			ToAddresses: []string{to},
		},
		Message: &types.Message{
			Subject: &types.Content{
				Data:    aws.String(subject),
				Charset: aws.String("UTF-8"),
			},
			Body: &types.Body{
				Text: &types.Content{
					Data:    aws.String(body),
					Charset: aws.String("UTF-8"),
				},
			},
		},
	}

	if _, err := s.client.SendEmail(ctx, input); err != nil {
		var msgRejected *types.MessageRejected
		if errors.As(err, &msgRejected) {
			return fmt.Errorf("%w: %v", entity.ErrInvalidRecipient, err)
		}
		return fmt.Errorf("SESメール送信に失敗: %w", err)
	}

	return nil
}
