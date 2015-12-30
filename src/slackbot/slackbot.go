package slackbot
import (
	"github.com/nlopes/slack"
	"log"
	"strings"
	"markov"
	"os"
	"math/rand"
	"time"
)

type slackConnection struct {
	rtm *slack.RTM
}

var filename string
var random *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

func RunSlack(token string, chain *markov.Markov, file, seedUser, controlUser string, responseProbability int) {
	api := slack.New(token)

	conn := new(slackConnection)
	conn.rtm = api.NewRTM()

	filename = file

	go conn.rtm.ManageConnection()

	Loop:

	for {
		select {
		case msg := <- conn.rtm.IncomingEvents:
			switch ev := msg.Data.(type) {
			case *slack.MessageEvent:
				if ev.User != conn.rtm.GetInfo().User.ID {
					if strings.HasPrefix(ev.Text, "marky") && ev.User == controlUser {
						command(ev.Text, chain)
						return
					} else if refersToMe(ev.Text, conn) || random.Intn(responseProbability) == 0 {
						messageReceived(chain, ev.Channel, ev.Text, ev.User, conn)
					}

					if ev.User == seedUser  && !refersToMe(ev.Text, conn) {
						log.Printf("Learning '%s'", ev.Text)
						chain.LearnSentence(ev.Text)
					}
				}

			case *slack.InvalidAuthEvent:
				log.Fatal("Invalid Credentials")
				break Loop
			}
		}
	}
}

func messageReceived(chain *markov.Markov, channel, text, user string, conn *slackConnection) {
	log.Printf("channel %s user: %s text: %s", channel, user, text)
	seeds := strings.Split(text, " ")
	answer := ""
	for _, seed := range seeds {
		if possible := chain.Generate(seed, 15); len(possible) > len(answer) {
			answer = possible
		}
	}
	if len(answer) > 0 {
		params := slack.PostMessageParameters{}
		params.AsUser = true
		params.LinkNames = 1
		params.UnfurlLinks = true
		channelID, timestamp, err := conn.rtm.PostMessage(channel, answer, params)
		log.Printf("channel: %s timestamp: %s: err: %s\n", channelID, timestamp, err)
	}
}

func command(command string, chain *markov.Markov) {
	seeds := strings.Split(command, " ")
	switch seeds[1] {
	case "die":
		chain.SaveChainState(filename)
		os.Exit(0)
	}
}

func refersToMe(message string, conn *slackConnection) bool {
	return strings.Contains(message, conn.rtm.GetInfo().User.ID)
}