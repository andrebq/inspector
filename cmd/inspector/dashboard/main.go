package dashboard

import (
	"io"

	"github.com/andrebq/inspector/internal/dashboard"
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
			return dashboard.Run(ctx.Context)
			// // TODO: implement something proper, for now,
			// // just dump requests to stdout
			// req, err := http.NewRequestWithContext(ctx.Context, "GET", mngApi, nil)
			// if err != nil {
			// 	return err
			// }
			// res, err := http.DefaultClient.Do(req)
			// if err != nil {
			// 	return err
			// }
			// if res.StatusCode != http.StatusOK {
			// 	return fmt.Errorf("unexpected response from server [%v - %v]", res.StatusCode, res.Status)
			// }
			// defer res.Body.Close()
			// dec := json.NewDecoder(res.Body)
			// for dec.More() {
			// 	var out manager.IOEvent
			// 	if err = dec.Decode(&out); errors.Is(err, io.EOF) {
			// 		return nil
			// 	} else if err != nil {
			// 		return err
			// 	}
			// 	fmt.Fprintf(stdout, "%#v\n", out)
			// }
			// return nil
		},
	}
}
