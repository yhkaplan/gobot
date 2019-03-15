package gobot

import (
	"net/http"
	"bytes"
	"encoding/json"
	"os"

	"github.com/nlopes/slack"
	"github.com/nlopes/slack/slackevents"
)

type eventApiHandler struct {
	slackClient *slack.Client
	verificationToken string
	firstMessage string
	bot *Gobot
}

func NewEventApiHandler(verificationToken string, firstMessage string, bot *Gobot) eventApiHandler {
	client := slack.New(os.Getenv("SLACK_ACCESS_TOKEN"))
	return eventApiHandler{client, verificationToken, firstMessage, bot}
}

func (h eventApiHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	buf := new(bytes.Buffer)
	buf.ReadFrom(r.Body)
	body := buf.String()
	event, e := slackevents.ParseEvent(json.RawMessage(body), slackevents.OptionVerifyToken(&slackevents.TokenComparator{VerificationToken: h.verificationToken}))

	if e != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	switch event.Type {
	case slackevents.URLVerification:
		var r *slackevents.ChallengeResponse
		err := json.Unmarshal([]byte(body), &r)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		w.Header().Set("Content-Type", "text")
		w.Write([]byte(r.Challenge))
	case slackevents.CallbackEvent:
		mention, ok := event.InnerEvent.Data.(*slackevents.AppMentionEvent)
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		attachment := slack.Attachment{
			Text:       h.firstMessage,
			Color:      "#f9a41b",
			CallbackID: "selectingMachine",
			Actions: []slack.AttachmentAction{
				{
					Name:    "selectMachine",
					Type:    "select",
					Options: h.bot.GetMachines(),
				},
			},
		}

		h.slackClient.PostMessage(mention.Channel, slack.MsgOptionAttachments(attachment))
	}
}
