package main

import (
	"github.com/alecthomas/kong"
	_ "github.com/joho/godotenv/autoload"
)

var CLI struct {
	Firewall FirewallCmd `cmd:"" help:"firewall."`
}

type Context struct {
	Debug bool
}

func main() {
	ctx := kong.Parse(&CLI)
	// Call the Run() method of the selected parsed command.
	err := ctx.Run(&Context{Debug: false})
	ctx.FatalIfErrorf(err)
}
