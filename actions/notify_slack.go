package actions

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/mozillazg/request"
)

type SlackNotifier struct {
	slackUrl string
	identity string
}

const (
	SLACK_MESSAGE_STARTUP = "Watchtower startup"
	SLACK_MESSAGE_ERROR   = "Some errors while checking and redeployment (Please check logs):"
	SLACK_MESSAGE_SUCCESS = "Successfully redeployed images:"
)

func NewSlackNotifier(slackUrl, identity string) *SlackNotifier {
	identity = strings.Trim(identity, " ")

	if len(identity) != 0 {
		identity = fmt.Sprintf("[%s]: ", identity)
	}

	return &SlackNotifier{
		slackUrl: slackUrl,
		identity: identity,
	}
}

func (s SlackNotifier) sendNotification(json map[string]interface{}) {
	c := new(http.Client)
	req := request.NewRequest(c)
	req.Json = json
	_, err := req.Post(s.slackUrl)

	if err != nil {
		fmt.Println(err)
	}
}

func (s SlackNotifier) NotifyStartup() {
	s.sendNotification(map[string]interface{}{
		"text": fmt.Sprintf("%s%s", s.identity, SLACK_MESSAGE_STARTUP),
	})
}

func buildAttachment(items []string, title, color string) map[string]interface{} {

	var fields []map[string]string

	for _, item := range items {
		fields = append(fields, map[string]string{"value": item, "short": "false"})
	}

	return map[string]interface{}{
		"fallback": title + strings.Join(items, ", "),
		"color":    color,
		"title":    title,
		"fields":   fields,
	}
}

func (s SlackNotifier) NotifyContainerUpdate(successfulContainers, errorMessages []string) {

	var attachments []map[string]interface{}

	if len(successfulContainers) != 0 {
		attachments = append(attachments, buildAttachment(successfulContainers, SLACK_MESSAGE_SUCCESS, "good"))
	}

	if len(errorMessages) != 0 {
		attachments = append(attachments, buildAttachment(errorMessages, SLACK_MESSAGE_ERROR, "danger"))
	}

	// add a pretext to the first attachment
	attachments[0]["pretext"] = s.identity

	s.sendNotification(map[string]interface{}{
		"attachments": attachments,
	})
}
