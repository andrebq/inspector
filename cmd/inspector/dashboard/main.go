package dashboard

import (
	"errors"
	"io"

	"github.com/urfave/cli/v3"
)

func Cmd(stdout io.Writer) *cli.Command {
	mngApi := "http://localhost:8082/request-stream"
	return &cli.Command{
		Name:  "dashboard",
		Usage: "Inspector dashboard",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "endpoint",
				Usage:       "URL where Inspector Management API is running",
				Destination: &mngApi,
				Value:       mngApi,
			},
		},
		Action: func(ctx *cli.Context) error {
			//return dashboard.Run(ctx.Context)
			return errors.New("not implemented")

		},
	}
}
