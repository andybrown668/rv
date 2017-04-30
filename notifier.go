package van

import (
	"github.com/nlopes/slack"
	"fmt"
	"os"
	"time"
)

//https://hooks.slack.com/services/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
var slackApi *slack.Client
var token = os.Getenv("SLACK_API_TOKEN")
var enabled = true
var sent time.Time
const Channel = "camera"

func notify(text string) {
	if slackApi == nil {
		if token == "" {
			enabled = false
		} else {
			slackApi = slack.New(token)
			cleanup()
		}
	}

	//ignore?
	if !enabled || time.Since(sent) < 5 * time.Minute{
		fmt.Printf("NO_SLACK: %s\n", text)
		return
	}

	_, _, err := slackApi.PostMessage(Channel, text, slack.PostMessageParameters{})
	if err != nil {
		fmt.Printf("failed to send: %s\n", err)
	}

	//don't flood slack - throttle messages to one every five minutes
	sent = time.Now()
}

func cleanup() {
	//remove all old messages
	channelId := ""
	if channels, err := slackApi.GetChannels(true); err != nil {
		fmt.Printf("failed to get channel: %s\n", err)
		return
	} else {
		for _, channel := range channels {
			if channel.Name == Channel {
				channelId = channel.ID
			}
		}
	}

	if history, err := slackApi.GetChannelHistory(channelId, slack.NewHistoryParameters()); err != nil {
		fmt.Printf("failed to get history: %s\n", err)
	} else {
		for _, message := range history.Messages {
			if _, _, err := slackApi.DeleteMessage(channelId, message.Timestamp); err != nil {
				fmt.Printf("failed to delete message: %s\n", err)
				break
			} else {
				fmt.Printf("Deleted msg from %s", message.Timestamp)
			}
		}
	}

}
