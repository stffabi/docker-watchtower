package actions

import (
	"os"

	"github.com/mozillazg/request"
)

const (
	SLACK_MESSAGE_STARTUP = "Watchtower startup"
	SLACK_MESSAGE_ERROR   = "An Error occurred while checking and redeployment. Please check logs."
	SLACK_MESSAGE_SUCCESS = "Successfully redeployed containers "
)

func Slack(slackUrl, message string, errorOccurred bool) error {
	msg := "[" + os.Hostname() + "]: " + message

	req = request.NewRequest(c)
	req.Json = map[string]string{
		"text": msg,
	}
	resp, err := req.Post(slackUrl)

	return err
}
