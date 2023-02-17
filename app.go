package main

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/docker/docker/client"
	cli "gopkg.in/urfave/cli.v1"
)

// App
type App struct {
	ctx    context.Context
	client *client.Client
}

// CopyToFile writes the content of the reader to the specified file
func CopyToFile(outfile string, r io.Reader) error {
	tmpFile, err := ioutil.TempFile(filepath.Dir(outfile), ".docker_temp_")
	if err != nil {
		return err
	}

	tmpPath := tmpFile.Name()

	_, err = io.Copy(tmpFile, r)
	tmpFile.Close()

	if err != nil {
		os.Remove(tmpPath)
		return err
	}

	if err = os.Rename(tmpPath, outfile); err != nil {
		os.Remove(tmpPath)
		return err
	}

	return nil
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

// serverResponse is a wrapper for http API responses.
type serverResponse struct {
	body       io.ReadCloser
	header     http.Header
	statusCode int
}

func (app *App) Run(args []string) {
	var cliApp = cli.NewApp()
	cliApp.Name = "Docker image tools"
	cliApp.Usage = ""
	cliApp.Flags = []cli.Flag{
		&cli.StringSliceFlag{
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
			Usage: "The output tar file",
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
