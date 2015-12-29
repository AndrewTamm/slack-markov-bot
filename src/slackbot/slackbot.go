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
var controlUserId string
var random *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

func RunSlack(token string, chain *markov.Markov, file, controlUser string, responseProbability int) {
	api := slack.New(token)

	conn := new(slackConnection)
	conn.rtm = api.NewRTM()

	filename = file
	controlUserId = controlUser

	go conn.rtm.ManageConnection()

	Loop:

	for {
		select {
		case msg := <- conn.rtm.IncomingEvents:
			switch ev := msg.Data.(type) {
			case *slack.MessageEvent:
				if random.Intn(responseProbability) == 0 {
					messageReceived(chain, ev.Channel, ev.Text, ev.User, conn)
				}
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
		if seeds[0] == "marky" && user == controlUserId {
			command(seeds[1], chain)
			return
		}
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

func command(command string, chain *markov.Markov) {
	switch command {
	case "die":
		chain.SaveChainState(filename)
		os.Exit(0)
	}
}