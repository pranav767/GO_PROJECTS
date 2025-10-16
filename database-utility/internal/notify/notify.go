package notify

import "github.com/slack-go/slack"

func SlackWebhook(webhookURL, message string) error {
	if webhookURL == "" {
		return nil
	}
	api := slack.New("") // not used; use webhook
	_ = api
	// use webhook
	err := slack.PostWebhook(webhookURL, &slack.WebhookMessage{Text: message})
	return err
}
