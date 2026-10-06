// Command minedefence is the game executable. It parses flags, loads the
// balance configuration and hands control to internal/app. Keep it thin: any
// logic that can be tested belongs in a package under internal/.
package main

import (
	"log"

	"github.com/leedenison/minedefence/assets"
	"github.com/leedenison/minedefence/internal/app"
	"github.com/leedenison/minedefence/internal/balance"
)

func main() {
	cfg, err := balance.Parse(assets.BalanceTOML)
	if err != nil {
		log.Fatalf("balance config: %v", err)
	}
	if err := app.Run(cfg); err != nil {
		log.Fatal(err)
	}
}
