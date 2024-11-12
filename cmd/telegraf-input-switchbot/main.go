package main

import (
	"log"
	"os"
	"time"

	_ "github.com/SlashNephy/telegraf-plugins/plugins/inputs/switchbot"
	"github.com/influxdata/telegraf/plugins/common/shim"
	"github.com/jessevdk/go-flags"
)

var options struct {
	ConfigPath   *string       `long:"config" description:"path to the config file for this plugin"`
	PollInterval time.Duration `long:"poll-interval" description:"how often to send metrics" default:"0s"`
}

func main() {
	log.SetOutput(os.Stderr)

	parser := flags.NewParser(&options, flags.Default)
	if _, err := parser.Parse(); err != nil {
		log.Fatalf("failed to parse flags: %s", err)
	}

	shimLayer := shim.New()
	if err := shimLayer.LoadConfig(options.ConfigPath); err != nil {
		log.Fatalf("failed to load config: %s", err)
	}

	if err := shimLayer.RunInput(options.PollInterval); err != nil {
		log.Fatalf("failed to run: %s", err)
	}
}
