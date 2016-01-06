package slackbot
import (
	"github.com/nlopes/slack"
	"log"
	"strings"
	"markov"
	"os"
	"math/rand"
	"time"
	"regexp"
	"bytes"
	"unicode/utf8"
	"unicode"
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
	var answer []string
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

		for i, word := range answer {
			if c,w := capitalizeMentions(word); c {
				answer[i] = w
			}
		}

		if !isEmoji(answer[0]) {
			answer[0] = upperFirst(answer[0])
		}

		channelID, timestamp, err := conn.rtm.PostMessage(channel, strings.Join(answer, " "), params)
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

func capitalizeMentions(word string) (bool,string) {
	if r, err := regexp.Compile("@[Uu][A-z0-9]{8}"); err == nil {
		out := r.ReplaceAllFunc([]byte(word), bytes.ToUpper)
		return true, string(out)
	}
	return false, word
}

func isEmoji(word string) bool {
	if match, _ := regexp.Match(":[[:alnum:]]:", []byte(word)); match {
		return true
	}
	return false
}

func upperFirst(s string) string {
	if s == "" {
		return ""
	}
	r, n := utf8.DecodeRuneInString(s)
	return string(unicode.ToUpper(r)) + s[n:]
}