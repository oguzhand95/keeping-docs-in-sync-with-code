package main

import (
	"log"

	"github.com/alecthomas/kong"

	"github.com/oguzhand95/keeping-docs-in-sync-with-code/internal/config"
	"github.com/oguzhand95/keeping-docs-in-sync-with-code/internal/engine"
	"github.com/oguzhand95/keeping-docs-in-sync-with-code/internal/server"
	"github.com/oguzhand95/keeping-docs-in-sync-with-code/internal/server/http"
)

type cli struct {
	Config string `help:"Path to config file" type:"existingfile" required:"" placeholder:"./config.yaml" env:"KEEPSYNC_CONFIG"`
}

func (c *cli) Run() error {
	// load configuration
	log.Printf("Loading configuration from %s\n", c.Config)
	if err := config.Load(c.Config); err != nil {
		log.Printf("Failed to load configuration: %v", err)
		return err
	}

	serverConf := &server.Conf{}
	if err := config.GetSection(serverConf); err != nil {
		return err
	}

	httpConf := &http.Conf{}
	if err := config.GetSection(httpConf); err != nil {
		return err
	}

	engineConf := &engine.Conf{}
	if err := config.GetSection(engineConf); err != nil {
		return err
	}

	return nil
}

func main() {
	var c cli
	ctx := kong.Parse(&c,
		kong.Name("keepsync"),
		kong.Description("Keeping documentation in sync with source code"),
		kong.UsageOnError(),
	)

	ctx.FatalIfErrorf(ctx.Run())
}
