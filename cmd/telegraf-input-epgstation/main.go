package main

import (
	"github.com/SlashNephy/telegraf-plugins/plugins/inputs"
	_ "github.com/SlashNephy/telegraf-plugins/plugins/inputs/epgstation"
)

func main() {
	inputs.Main()
}
