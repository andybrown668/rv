package van

import (
	"github.com/nlopes/slack"
	"fmt"
	"os"
)

//https://hooks.slack.com/services/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
var slackApi *slack.Client
var token = os.Getenv("SLACK_API_TOKEN")
var enabled = true
func notify(text string) {
	if slackApi == nil {
		if token == "" {
			enabled = false
		} else {
			slackApi = slack.New(token)
		}
	}

	if !enabled {
		fmt.Printf("NO_SLACK: %s\n", text)
		return
	}

	channel, time, err := slackApi.PostMessage("camera", text, slack.PostMessageParameters{})
	if err == nil {
		fmt.Printf("sent to %s at %s\n", channel, time)
	} else {
		fmt.Printf("failed to send: %s\n", err)
	}
}
