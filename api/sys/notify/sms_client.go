package notify

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/nettyrnp/ads-crawler/api/common"
	"github.com/pkg/errors"
	"net/http"
	"net/url"
	"strings"
)

type SNSConfig struct {
	SmsSenderAccountSid string `json:"smsSenderAccountSid"`
	SmsSenderAuthToken  string `json:"smsSenderAuthToken"`
	SmsSenderPhone      string `json:"smsSenderPhone"`
}

func (c SNSConfig) Validate() []error {
	var errs []error
	if len(c.SmsSenderAccountSid) == 0 {
		errs = append(errs, errors.New("SNSConfig requires a non-empty SmsSenderAccountSid value"))
	}
	if len(c.SmsSenderAuthToken) == 0 {
		errs = append(errs, errors.New("SNSConfig requires a non-empty SmsSenderAuthToken value"))
	}
	if len(c.SmsSenderPhone) == 0 {
		errs = append(errs, errors.New("SNSConfig requires a non-empty SmsSenderPhone value"))
	}
	return errs
}

type SmsNotifier interface {
	Send(ctx context.Context, toAddr, msg string) error
}

type SmsClient struct {
	Kind string
	SNSConfig
}

func (c *SmsClient) Send(ctx context.Context, toPhone, msg string) error {
	urlStr := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%v/Messages.json", c.SmsSenderAccountSid)
	msgData := url.Values{}
	msgData.Set("To", toPhone)
	msgData.Set("From", c.SmsSenderPhone)
	msgData.Set("Body", msg)
	msgDataReader := *strings.NewReader(msgData.Encode())

	client := &http.Client{}
	req, err := http.NewRequest("POST", urlStr, &msgDataReader)
	if err != nil {
		return err
	}
	req.SetBasicAuth(c.SmsSenderAccountSid, c.SmsSenderAuthToken)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		var data map[string]interface{}
		decoder := json.NewDecoder(resp.Body)
		err := decoder.Decode(&data)
		if err != nil {
			return err
		}
		common.LogInfof(">> sent sms to %v: %v. SID: %v\n", c.Kind, msg, data["sid"])
		return nil
	}
	return errors.Errorf("Sending sms to customer phone '%v'", c.SmsSenderPhone)
}
