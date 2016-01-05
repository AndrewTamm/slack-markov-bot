package markov
import (
	"strings"
	"math/rand"
	"time"
	"os"
	"log"
	"encoding/json"
)

var random *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

type Word string

type Link struct {
	Weight int	`json:"weight"`
	Target Word	`json:"target"`
}

type adjacencyList map[Word][]Link

type Markov struct {
	Chain                adjacencyList	`json:"chain"`
	LinkCount, WordCount int
}

func New() *Markov {
	return &Markov{Chain: make(adjacencyList)}
}

// find a word in the linked list chain by value, returning true and
// the element if found, and false and nil or the last element otherwise
func find(theChain []Link, word Word) (bool, int) {
	for i := range theChain {
		if theChain[i].Target == word {
			return true, i
		}
	}
	return false, -1
}

// AddLink creates a link from 'from' and to 'to'.
// from and to are added to the markov chain if they are not already in it
// return the new weight of the link
func (m *Markov) AddLink(from, to Word) int {
	m.AddWord(from)
	m.AddWord(to)
	fromLinks := m.Chain[from]
	found, index := find(fromLinks, to)
	if found {
		// increase the weight of this link
		fromLinks[index].Weight++
		return fromLinks[index].Weight
	} else {
		// create a new link with a weight of 1
		m.Chain[from] = append(fromLinks, Link{Target: to, Weight: 1})
		m.LinkCount++
		return 1
    }
}

// AddWord adds a word to the markov chain with no links.
// Returns true if the word was a new word we've never seen and false if it was
// already in the chain
func (m *Markov) AddWord(word Word) bool {
	// if there's no chain for this word
	if m.Chain[word] == nil {
		// create a new chain
		m.Chain[word] = []Link{}
		m.WordCount++
		// and return true
		return true
	}
	// if we've already seen it return false
	return false
}

// LearnSentence causes the markov chain to add an entire sentence (one or more words separated by whitespace)
// Returns the number of new links between words added to the chain, less than or equal to the number of words.
func (m *Markov) LearnSentence(sentence string) int {
	var prev Word = ""
	newLinkCount := 0
	for _,word := range strings.Split(sentence, " ") {
		if trimmed := strings.TrimSpace(word); len(trimmed) > 0 {
			to := Word(strings.ToLower(trimmed))
			if prev != "" {
				var weight = m.AddLink(prev, to)
				if weight == 1 {
					newLinkCount++
				}
			}
			prev = to
		}
	}
	return newLinkCount
}

// Generate creates a string (if one exists) starting with the start word and
// continuing for at most maxWords (to avoid loops). The returned string will
// have the first word capitalized.
func (m *Markov) Generate(start string, maxWords int) []string {
	startingWord := Word(strings.ToLower(start))
	if m.Chain[startingWord] == nil {
		return []string{}
	}
	markovChain := []string{start}

	links := m.Chain[startingWord]
	for wordCount := 1; wordCount <= maxWords; wordCount++ {
		if len(links) == 0 {
			break
		}
		linkWeight := totalWeight(links)
		targetWeight := random.Intn(linkWeight)
		var seenWeight int
		for i := range links {
			seenWeight += links[i].Weight
			if seenWeight > targetWeight {
				target := links[i].Target
				links = m.Chain[target]
				markovChain = append(markovChain, string(target))
				break
			}
		}
	}

	return markovChain
}

// GetWordCount returns the total number of unique words in the markov chain.
func (m *Markov) GetWordCount() int {
	return m.WordCount
}

// GetLinkCount returns the total number of links between words in the chain.
// This is likely greater than the number of unique words (but far less than
// the number of total words) due to the interconnectedness.
func (m *Markov) GetLinkCount() int {
	return m.LinkCount
}

func (m *Markov) SaveChainState(filename string) {
	fo, err := os.Create(filename)
	if err != nil {
		log.Fatal("Uh oh, couldn't create the file for saving", err.Error())
		os.Exit(1)
	}
	defer func() {
		if err := fo.Close(); err != nil {
			panic(err)
		}
	}()

	jsonParser := json.NewEncoder(fo)
	if err := jsonParser.Encode(m); err != nil {
		log.Fatal("Problem while persisting the markov chain to disk", err.Error())
		os.Exit(1)
	}
}

func (m *Markov) LoadChainState(filename string) {
	fo, err := os.Open(filename)
	if err != nil {
		log.Fatal("Couldn't open the markov chain file", err.Error())
		os.Exit(1)
	}
	defer func() {
		if err := fo.Close(); err != nil {
			panic(err)
		}
	}()

	jsonParser := json.NewDecoder(fo)

	if err = jsonParser.Decode(m); err != nil {
		log.Fatal("Problem in decoding json file", err.Error())
		os.Exit(1)
	}
	println(m.GetLinkCount())
	println(m.GetWordCount())
}

func totalWeight(theChain []Link) int {
	var sum int
	for i := range theChain {
		sum += theChain[i].Weight
	}
	return sum
}