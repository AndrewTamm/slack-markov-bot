package slackbot
import (
	"github.com/nlopes/slack"
	"log"
	"strings"
	"markov"
)

type slackConnection struct {
	rtm *slack.RTM
}

func RunSlack(token string, chain *markov.Markov) {
	api := slack.New(token)

	conn := new(slackConnection)
	conn.rtm = api.NewRTM()


	go conn.rtm.ManageConnection()

	Loop:

	for {
		select {
		case msg := <- conn.rtm.IncomingEvents:
			switch ev := msg.Data.(type) {
			case *slack.MessageEvent:
				messageReceived(chain, ev.Channel, ev.Text, ev.User, conn)
			case *slack.InvalidAuthEvent:
				log.Fatal("Invalid Credentials")
				break Loop
			}
		}
	}
}

func messageReceived(chain *markov.Markov, channel, text, user string, conn *slackConnection) {
	if user != conn.rtm.GetInfo().User.ID {
		log.Printf("channel %s user: %s text: %s", channel, user, text)
		seeds := strings.Split(text, " ")
		var answer string = ""
		for _, seed := range seeds {
			if possible := chain.Generate(seed, 15); len(possible) > len(answer) {
				answer = possible
			}
		}
		if len(answer) > 0 {
			params := slack.PostMessageParameters{}
			params.AsUser = true
			params.LinkNames = 1
			channelID, timestamp, err := conn.rtm.PostMessage(channel, answer, params)
			log.Printf("channel: %s timestamp: %s: err: %s\n", channelID, timestamp, err)
		}
	}
}