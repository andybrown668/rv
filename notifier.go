package van

import (
	"github.com/nlopes/slack"
	"fmt"
)

//https://hooks.slack.com/services/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
var slackApi *slack.Client
func notify(text string) {
	if slackApi == nil {
		slackApi = slack.New("xoxp-176548880658-176548880674-176643424389-51d48d4395082845782eebc5483a1e27")
	}

	channel, time, err := slackApi.PostMessage("camera", text, slack.PostMessageParameters{})
	if err == nil {
		fmt.Printf("sent to %s at %s\n", channel, time)
	} else {
		fmt.Printf("failed to send: %s\n", err)
	}
}
