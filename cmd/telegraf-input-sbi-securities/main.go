package main

import (
	"github.com/SlashNephy/telegraf-plugins/plugins/inputs"
	_ "github.com/SlashNephy/telegraf-plugins/plugins/inputs/sbisecurities"
)

func main() {
	inputs.Main()
}
