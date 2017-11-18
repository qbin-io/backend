package main

import (
	"os"
	"path/filepath"
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
			Value:  qbin.Config["http"],
			Usage:  "HTTP listen address. Set to 'none' to disable.",
			EnvVar: "HTTP_LISTEN",
		},
		cli.StringFlag{
			Name:   "tcp",
			Value:  qbin.Config["tcp"],
			Usage:  "TCP (netcat API) listen address. Set to 'none' to disable.",
			EnvVar: "TCP_LISTEN",
		},
		cli.StringFlag{
			Name:   "frontend-path, p",
			Value:  qbin.Config["frontend-path"],
			Usage:  "Location of the frontend files.",
			EnvVar: "FRONTEND_PATH",
		},
		cli.StringFlag{
			Name:   "root, r",
			Value:  qbin.Config["root"] + qbin.Config["path"],
			Usage:  "The path under which the application will be reachable from the internet, with trailing slash.",
			EnvVar: "ROOT_URL",
		},
	}

	app.Action = func(c *cli.Context) error {
		qbin.Config["http"] = c.String("http")
		qbin.Config["tcp"] = c.String("tcp")

		frontendPath, err := filepath.Abs(c.String("frontend-path"))
		if err != nil {
			println("Frontend path couldn't be resolved.")
			panic(err)
		}
		qbin.Config["frontend-path"] = frontendPath

		qbin.Config["root"] = strings.TrimSuffix(c.String("root"), "/")
		// https://example.org/[grab this part]
		rootSplit := strings.SplitAfterN(qbin.Config["root"], "/", 4)
		relativeRoot := "/"
		if len(rootSplit) > 3 {
			relativeRoot = "/" + rootSplit[3]
		}
		qbin.Config["path"] = strings.TrimSuffix(relativeRoot, "/")

		if qbin.Config["http"] != "none" {
			qbin.StartHTTPServer()
		}

		// TODO: Any good way to make cli not exit immediately?
		for {
			time.Sleep(time.Hour)
		}
	}

	app.Run(os.Args)
}
