package main

import (
	"context"
	"fmt"

	"github.com/docker/docker/client"
	cli "gopkg.in/urfave/cli.v1"
)

// App
type App struct {
	ctx    context.Context
	client *client.Client
}

func NewApp(ctx context.Context) *App {
	dockerCli, err := client.NewClientWithOpts(client.FromEnv)
	dockerCli.NegotiateAPIVersion(context.Background())
	if err != nil {
		panic(err)
	}
	return &App{ctx: ctx,
		client: dockerCli,
	}
}

func (app *App) Run(args []string) {
	var cliApp = cli.NewApp()
	cliApp.Name = "Save docker image without layers in `exclude`"
	cliApp.Usage = ""
	cliApp.Flags = []cli.Flag{
		cli.StringSliceFlag{
			Name:  "image, i",
			Usage: "The image to save",
		},
		cli.StringSliceFlag{
			Name:  "exclude, e",
			Usage: "The existing image that does not need to export",
		},
		cli.StringFlag{
			Name:  "output, o",
			Value: "image.tgz",
			Usage: "The output tar.gz file",
		},
	}

	cliApp.Action = func(c *cli.Context) error {
		if err := app.process(c); err != nil {
			return cli.NewExitError(fmt.Sprintf("Failed to process: %v", err), 1)
		}
		return nil
	}
	cliApp.Run(args)
}

// Process command
func (app *App) process(c *cli.Context) error {
	images := ImageList(c.StringSlice("image"))
	excludeIDs := ImageList(c.StringSlice("exclude"))
	output := c.String("output")

	return SaveImage(app.ctx, images, excludeIDs, output, app.client)
}
