package main

import (
	"github.com/projectdiscovery/gologger"
	"github.com/secinto/elasticPusher/pusher"
)

func main() {
	// Parse the command line flags and read config files
	options := pusher.ParseOptions()

	newPusher, err := pusher.FromOptions(options)
	if err != nil {
		gologger.Fatal().Msgf("Could not create pusher: %s\n", err)
	}

	err = newPusher.PushFile()
	if err != nil {
		gologger.Fatal().Msgf("Could not push: %s\n", err)
	}
}
