package main

import (
	"os"
	"strings"
	"time"

	"github.com/qbin-io/backend"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()

	app.Name = "qbin"
	app.Version = "2.0.0a1"
	app.Usage = "a minimalist pastebin service"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "http",
			Value:  ":3000",
			Usage:  "HTTP listen address. Set to 'none' to disable.",
			EnvVar: "HTTP_LISTEN",
		},
		cli.StringFlag{
			Name:   "tcp",
			Value:  ":90",
			Usage:  "TCP (netcat API) listen address. Set to 'none' to disable.",
			EnvVar: "TCP_LISTEN",
		},
		cli.StringFlag{
			Name:   "frontend-path, p",
			Value:  "/usr/share/qbin/frontend",
			Usage:  "Location of the frontend files.",
			EnvVar: "FRONTEND_PATH",
		},
		cli.StringFlag{
			Name:   "root, r",
			Value:  "https://qbin.io/",
			Usage:  "The path under which the application will be reachable from the internet, with trailing slash.",
			EnvVar: "ROOT_URL",
		},
	}

	app.Action = func(c *cli.Context) error {
		qbin.Config["http"] = c.String("http")
		qbin.Config["tcp"] = c.String("tcp")
		qbin.Config["frontend-path"] = c.String("frontend-path")
		qbin.Config["root"] = strings.TrimSuffix(c.String("root"), "/")

		// https://example.org/[grab this part]
		rootSplit := strings.SplitAfterN(qbin.Config["root"], "/", 4)
		relativeRoot := "/"
		if len(rootSplit) > 3 {
			relativeRoot = "/" + rootSplit[3]
		}
		qbin.Config["relative-root"] = strings.TrimSuffix(relativeRoot, "/")

		qbin.StartHTTPServer()

		// TODO: Any good way to make cli not exit immediately?
		for {
			time.Sleep(time.Hour)
		}
	}

	app.Run(os.Args)
}
