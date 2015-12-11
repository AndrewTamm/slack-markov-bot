package markov

import (
	"testing"
	"math/rand"
)

var bear Word = "bear"
var tree Word = "tree"
var is Word = "is"
var tall Word = "tall"
var short Word = "short"
var in Word = "in"
var the Word = "the"
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

func TestGenerate(t *testing.T) {
	markov := New()

	// use a constant seed to give the same result
	random = rand.New(rand.NewSource(10000))

	markov.AddLink(the, bear)
	markov.AddLink(bear, is)
	markov.AddLink(is, in)
	markov.AddLink(in, the)
	markov.AddLink(the, forest)

	sentence := markov.Generate("the", 6)

	if (sentence != "The bear is in the bear is") {
		t.Errorf("Didn't generate the right chain")
	}
}