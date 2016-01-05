package markov

import (
	"testing"
	"math/rand"
	"strings"
)

var bear Word = "bear"
var tree Word = "tree"
var is Word = "is"
var tall Word = "tall"
var short Word = "short"
var in Word = "in"
var the Word = "the"
var a Word = "a"
var forest Word = "forest"

var words []Word = []Word{bear, tree, is, tall, short, in, the, forest}

func TestAddWord(t *testing.T) {
	markov := New()

	checkedAddWord := func(word Word, expected bool, message string) {
		if markov.AddWord(word) != expected {
			t.Errorf(message, word)
		}
	}

	for _, word := range words {
		checkedAddWord(word, true, "Word %s was added")
	}
	for _, word := range words {
		checkedAddWord(word, false, "Word %s could not be")
	}
}

func TestAddLink(t *testing.T) {
	markov := New()

	checkedAddLink := func(from, to Word, expectedWeight int) {
		if markov.AddLink(from, to) != expectedWeight {
			t.Errorf("The weight was not as expected", from, to)
		}
	}

	checkedAddLink(the, bear, 1)
	checkedAddLink(bear, is, 1)
	checkedAddLink(is, in, 1)
	checkedAddLink(in, the, 1)
	checkedAddLink(the, forest, 1)

	checkedAddLink(the, bear, 2)
	checkedAddLink(bear, is, 2)
	checkedAddLink(is, in, 2)
	checkedAddLink(in, the, 2)
	checkedAddLink(the, tree, 1)

	checkedAddLink(the, tree, 2)
	checkedAddLink(tree, is, 1)
	checkedAddLink(is, in, 3)
	checkedAddLink(in, the, 3)
	checkedAddLink(the, forest, 2)
}

func TestLearnSentence(t *testing.T) {
	markov := New()

	learnedLinks := markov.LearnSentence("The quick brown fox jumps over a lazy dog.")

	if learnedLinks != 8 {
		t.Errorf("Didn't learn the right number of links. Learned %d", learnedLinks)
	}

	chain := markov.Generate("the", 9)
	sentence := strings.Join(chain, " ")

	if "the quick brown fox jumps over a lazy dog." != sentence {
		t.Errorf("Didn't learn links to give 'The quick brown fox jumps over a lazy dog. Got '%s' instead.", sentence)
	}
}

func TestGenerate(t *testing.T) {
	markov := New()

	// use a constant seed to give the same result
	random = rand.New(rand.NewSource(10000))

	markov.AddLink(the, bear)
	markov.AddLink(bear, is)
	markov.AddLink(is, in)
	markov.AddLink(in, a)
	markov.AddLink(a, forest)

	chain := markov.Generate("the", 7)
	sentence := strings.Join(chain, " ")

	if (sentence != "the bear is in a forest") {
		t.Errorf("Didn't generate the right chain. Expected 'the bear is in a forest' but got '%s' instead", sentence)
	}
}