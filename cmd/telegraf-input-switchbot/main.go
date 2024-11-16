package main

import (
	"github.com/SlashNephy/telegraf-plugins/plugins/inputs"
	_ "github.com/SlashNephy/telegraf-plugins/plugins/inputs/switchbot"
)

func main() {
	inputs.Main()
}
