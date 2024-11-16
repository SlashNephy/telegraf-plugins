package main

import (
	"github.com/SlashNephy/telegraf-plugins/plugins/outputs"
	_ "github.com/SlashNephy/telegraf-plugins/plugins/outputs/mackerel"
)

func main() {
	outputs.Main()
}
