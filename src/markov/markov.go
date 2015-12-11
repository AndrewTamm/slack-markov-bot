package markov
import (
	"strings"
	"math/rand"
	"time"
)

var random *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

type Word string

type Link struct {
	weight int
	target Word
}

type adjacencyList map[Word][]Link

type Markov struct {
	chain 				 adjacencyList
	linkCount, wordCount int
}

func New() *Markov {
	return &Markov{chain: make(adjacencyList)}
}

// find a word in the linked list chain by value, returning true and
// the element if found, and false and nil or the last element otherwise
func find(theChain []Link, word Word) (bool, int) {
	for i := range theChain {
		if theChain[i].target == word {
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
	fromLinks := m.chain[from]
	found, index := find(fromLinks, to)
	if found {
		// increase the weight of this link
		fromLinks[index].weight++
		return fromLinks[index].weight
	} else {
		// create a new link with a weight of 1
		m.chain[from] = append(fromLinks, Link{target: to, weight: 1})
		m.linkCount++
		return 1
    }
}

// AddWord adds a word to the markov chain with no links.
// Returns true if the word was a new word we've never seen and false if it was
// already in the chain
func (m *Markov) AddWord(word Word) bool {
	// if there's no chain for this word
	if m.chain[word] == nil {
		// create a new chain
		m.chain[word] = []Link{}
		m.wordCount++
		// and return true
		return true
	}
	// if we've already seen it return false
	return false
}

// Generate creates a string (if one exists) starting with the start word and
// continuing for at most maxWords (to avoid loops). The returned string will
// have the first word capitalized.
func (m *Markov) Generate(start string, maxWords int) string {
	startingWord := Word(start)
	if m.chain[startingWord] == nil {
		return ""
	}
	markovChain := []string{strings.Title(start)}

	links := m.chain[startingWord]
	for wordCount := 1; wordCount <= maxWords; wordCount++ {
		if len(links) == 0 {
			break
		}
		linkWeight := totalWeight(links)
		targetWeight := random.Intn(linkWeight)
		var seenWeight int
		for i := range links {
			seenWeight += links[i].weight
			if seenWeight > targetWeight {
				target := links[i].target
				links = m.chain[target]
				markovChain = append(markovChain, string(target))
				break
			}
		}
	}

	return strings.Join(markovChain, " ")
}

// GetWordCount returns the total number of unique words in the markov chain.
func (m *Markov) GetWordCount() int {
	return m.wordCount
}

// GetLinkCount returns the total number of links between words in the chain.
// This is likely greater than the number of unique words (but far less than
// the number of total words) due to the interconnectedness.
func (m *Markov) GetLinkCount() int {
	return m.linkCount
}

func totalWeight(theChain []Link) int {
	var sum int
	for i := range theChain {
		sum += theChain[i].weight
	}
	return sum
}