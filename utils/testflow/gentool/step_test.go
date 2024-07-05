package main

import (
	"log"
	"testing"
)

func TestQuoteBackticks(t *testing.T) {
	input := "`Hello ``world`"
	expected := "\"`\" + `Hello ` + \"``\" + `world` + \"`\""
	log.Println("Expected: ", expected)

	output := QuoteBackticks(input)

	if output != expected {
		t.Errorf("Expected %s, but got %q", expected, output)
	}
}
