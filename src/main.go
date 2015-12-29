package main

import (
	"flag"
	"log"
	"strings"
	"os"
	"io/ioutil"
	"encoding/json"
	"path/filepath"
	"markov"
	"slackbot"
)

var token = flag.String("apiToken", "Required", "Authentication token")
var slackExportDir = flag.String("slackDir", "", "Directory with json exports from Slack to seed the Markov chain generator")
var seedUser = flag.String("seedUser", "U03SW6XSU", "Use messages from this user ID to seed the Markov chain")
var controlUser = flag.String("bossMan", "U0C01780Z", "Accept commands from this user")
var markovFile = flag.String("markovFile", "markov.json", "JSON file of the Markov chain, defaults to markov.json")
var responseProbability = flag.Int("responseProbability", 20, "The chance this bot will respond to a statement")

func main() {
	flag.Parse()
	log.SetFlags(0)

	chain := markov.New()
	if len(*slackExportDir) != 0 {
		log.Printf("Loading Slack messages from user %s to create a bot in their personality", *seedUser)
		importSlackData(chain)
	} else {
		chain.LoadChainState(*markovFile)
	}

	slackbot.RunSlack(*token, chain, *markovFile, *controlUser, *responseProbability)
}

func importSlackData(chain *markov.Markov) {
	readMessages := getReadMessageFunc(chain)

	err := filepath.Walk(*slackExportDir, readMessages)
	if (err != nil) {
		log.Printf("File walk error: %v\n", err)
		os.Exit(1)
	}
	log.Println(chain.GetLinkCount())
	log.Println(chain.GetWordCount())
}

func getReadMessageFunc(chain *markov.Markov) func(string, os.FileInfo, error) error {
	return func (path string, fileInfo os.FileInfo, _ error) error {
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
}
