package main

import (
	"flag"
	"log"
	"github.com/nlopes/slack"
	"strings"
	"os"
	"io/ioutil"
	"encoding/json"
	"path/filepath"
	"markov"
)

type slackConnection struct {
	rtm *slack.RTM
}

var token = flag.String("apiToken", "Required", "Authentication token")
var useSeed = flag.Bool("useSeed", false, "Load Markov chain data from chat log json exports")
var seedDirectory = flag.String("seedDir", "", "Directory with json exports from Slack to seed the Markov chain generator")
var seedUser = flag.String("seedUser", "U03SW6XSU", "Use messages from this user ID to seed the Markov chain")
var markovFile = flag.String("markovFile", "markov.json", "JSON file of the Markov chain, defaults to markov.json")

func main() {
	flag.Parse()
	log.SetFlags(0)

	if *useSeed {
		log.Print("Loading Slack exports for seed data")
		if len(*seedDirectory) == 0 {
			loadSeedData()
		} else {
			log.Printf("Must specifiy a seed directory when loading data from seeds\n")
			os.Exit(1)
		}
	} else {
		chain = loadChainState(*markovFile)
	}

	runSlack(*token)
}

func runSlack(token string) {
	api := slack.New(token)

	conn := new(slackConnection)
	conn.rtm = api.NewRTM()

	saveChainState(chain, *markovFile)
	go conn.rtm.ManageConnection()

	Loop:

	for {
		select {
		case msg := <- conn.rtm.IncomingEvents:
			switch ev := msg.Data.(type) {
			case *slack.MessageEvent:
				messageReceived(ev.Channel, ev.Text, ev.User, conn)
			case *slack.InvalidAuthEvent:
				log.Fatal("Invalid Credentials")
				break Loop
			}
		}
	}
}

func saveChainState(chain *markov.Markov, filename string) {
	fo, err := os.Create(filename)
	if err != nil {
		log.Fatal("Uh oh, couldn't create the file for saving: %s", err.Error())
		os.Exit(1)
	}
	defer func() {
		if err := fo.Close(); err != nil {
			panic(err)
		}
	}()

	jsonParser := json.NewEncoder(fo)
	if err := jsonParser.Encode(&chain); err != nil {
		log.Fatal("Problem while persisting the markov chain to disk: %s", err.Error())
		os.Exit(1)
	}
}

func loadChainState(filename string) *markov.Markov {
	fo, err := os.Open(filename)
	if err != nil {
		log.Fatal("Couldn't open the markov chain: %s", err.Error())
		os.Exit(1)
	}
	defer func() {
		if err := fo.Close(); err != nil {
			panic(err)
		}
	}()

	jsonParser := json.NewDecoder(fo)

	chain := markov.New()
	if err = jsonParser.Decode(&chain); err != nil {
		log.Fatal("Problem in : %s", err.Error())
		os.Exit(1)
	}
	println(chain.GetLinkCount())
	println(chain.GetWordCount())
	return chain
}

func messageReceived(channel, text, user string, conn *slackConnection) {
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
			log.Printf("Channel: %s timestamp: %s: err: %s\n", channelID, timestamp, err)
		}
	}
}

func loadSeedData() {
	err := filepath.Walk("../data/seed", readMessages)
	if (err != nil) {
		log.Printf("File walk error: %v\n", err)
		os.Exit(1)
	}
	log.Println(chain.GetLinkCount())
	log.Println(chain.GetWordCount())
}

var chain *markov.Markov = markov.New()

func readMessages(path string, fileInfo os.FileInfo, _ error) error {
	if fileInfo.IsDir() {
		return nil
	}

	file, e := ioutil.ReadFile(path)
	if e != nil {
		log.Printf("File error: %v\n", e)
		os.Exit(1)
	}

	// the json exports from slack are arrays of json objects
	var data []map[string]interface{}

	err := json.Unmarshal(file, &data)
	if err != nil {
		log.Printf("Marshalling error: %v\n", err)
		os.Exit(1)
	}

	for _, element := range data {
		if message_type, ok := element["type"]; ok {
			if message_type == "message" {
				if user, ok := element["user"]; ok && user == *seedUser {
					log.Print(*seedUser)
					if text, ok := element["text"]; ok {
						var prev markov.Word = ""
						for _,word := range strings.Split(text.(string), " ") {
							if trimmed := strings.TrimSpace(word); len(trimmed) > 0 {
								if prev != "" {
									to := markov.Word(strings.ToLower(trimmed))
									chain.AddLink(prev, to)
								}
								prev = markov.Word(trimmed)
							}
						}
					}
				}
			}
		}
	}

	return nil
}