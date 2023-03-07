package main

import (
	"flag"
	"log"
	"time"

	_ "github.com/SlashNephy/telegraf-output-mackerel/plugins/outputs/mackerel"

	"github.com/influxdata/telegraf/plugins/common/shim"
)

var (
	configFile           = flag.String("config", "", "path to the config file for this plugin")
	pollInterval         = flag.Duration("poll_interval", 1*time.Second, "how often to send metrics")
	pollIntervalDisabled = flag.Bool(
		"poll_interval_disabled",
		false,
		"set to true to disable polling. You want to use this when you are sending metrics on your own schedule",
	)
)

func main() {
	flag.Parse()

	if *pollIntervalDisabled {
		*pollInterval = shim.PollIntervalDisabled
	}

	shimLayer := shim.New()
	if err := shimLayer.LoadConfig(configFile); err != nil {
		log.Fatalln(err)
	}
	if err := shimLayer.Run(*pollInterval); err != nil {
		log.Fatalln(err)
	}
}
