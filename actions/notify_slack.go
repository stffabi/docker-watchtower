package actions

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/mozillazg/request"
)

// SlackNotifier is used to make the https requests to a specified slack
// webhook URL
type SlackNotifier struct {
	slackURL string
	identity string
}

const (
	SlackMessageStartup = "Watchtower startup"
	SlackMessageError   = "Some errors while checking and redeployment (Please check logs):"
	SlackMessageSuccess = "Successfully redeployed images:"
)

func NewSlackNotifier(slackURL, identity string) *SlackNotifier {
	identity = strings.Trim(identity, " ")

	if len(identity) != 0 {
		identity = fmt.Sprintf("[%s]: ", identity)
	}

	return &SlackNotifier{
		slackURL: slackURL,
		identity: identity,
	}
}

func (s SlackNotifier) sendNotification(json map[string]interface{}) {
	c := new(http.Client)
	req := request.NewRequest(c)
	req.Json = json
	_, err := req.Post(s.slackURL)

	if err != nil {
		fmt.Println(err)
	}
}

func (s SlackNotifier) NotifyStartup() {
	s.sendNotification(map[string]interface{}{
		"text": fmt.Sprintf("%s%s", s.identity, SlackMessageStartup),
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
		attachments = append(attachments, buildAttachment(successfulContainers, SlackMessageSuccess, "good"))
	}

	if len(errorMessages) != 0 {
		attachments = append(attachments, buildAttachment(errorMessages, SlackMessageError, "danger"))
	}

	// add a pretext to the first attachment
	attachments[0]["pretext"] = s.identity

	s.sendNotification(map[string]interface{}{
		"attachments": attachments,
	})
}
