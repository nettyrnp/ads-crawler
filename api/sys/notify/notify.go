package notify

import (
	"context"
	"github.com/nettyrnp/ads-crawler/config"
)

func NewEmailNotifier(conf config.Config, kind string) EmailNotifier {
	conf2 := SESConfig{
		AWSKey:       conf.SESKey,
		AWSSecretKey: conf.SESSecretKey,
		AWSRegion:    conf.SESRegion,
		Sender:       conf.SESSender,
	}
	return &EmailClient{kind, conf2}
}

func NewSmsNotifier(conf config.Config, kind string) SmsNotifier {
	//TODO: enable before PROD, because it is for price model of country xxx
	//conf2 := SNSConfig{
	//	SmsSenderAccountSid: conf.SmsSenderAccountSid,
	//	SmsSenderAuthToken:  conf.SmsSenderAuthToken,
	//	SmsSenderPhone:      conf.SmsSenderPhone,
	//}
	//return &SmsClient{kind, conf2}

	return &noopSmsNotifier{}
}

type noopSmsNotifier struct {
}

func (n *noopSmsNotifier) Send(ctx context.Context, toAddr, msg string) error {
	return nil
}
