package inspector

import (
	"io"

	"github.com/andrebq/inspector/cmd/inspector/dashboard"
	"github.com/andrebq/inspector/cmd/inspector/proxy"
	"github.com/urfave/cli/v3"
)

func App(stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:  "inspector",
		Usage: "Reverse proxy to help you inspect HTTP requests",
		Commands: []*cli.Command{
			proxy.Cmd(),
			dashboard.Cmd(stdout),
		},
	}
}
