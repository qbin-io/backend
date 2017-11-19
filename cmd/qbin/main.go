package main

import (
	"os"
	"strings"

	"github.com/op/go-logging"
	"github.com/qbin-io/backend"
	"github.com/qbin-io/backend/http"
	"github.com/qbin-io/backend/tcp"
	"github.com/urfave/cli"
)

var flags = []cli.Flag{
	cli.StringFlag{
		Name: "database, d", EnvVar: "DATABASE", Value: "root:@/qbin",
		Usage: "MySQL/MariaDB connection string. It is recommended to pass this parameter as an environment variable."},
	cli.StringFlag{
		Name: "root, r", EnvVar: "ROOT_URL", Value: "http://127.0.0.1:8000",
		Usage: "The path under which the application will be reachable from the internet."},
	cli.StringFlag{
		Name: "wordlist", EnvVar: "WORD_LIST", Value: "eff_large_wordlist.txt",
		Usage: "Word list used for random slug generation."},
	cli.StringFlag{
		Name: "tcp", EnvVar: "TCP_LISTEN", Value: ":9000",
		Usage: "TCP (netcat API) listen address. Set to 'none' to disable."},
	cli.StringFlag{
		Name: "http", EnvVar: "HTTP_LISTEN", Value: ":8000",
		Usage: "HTTP listen address. Set to 'none' to disable."},
	cli.StringFlag{
		Name: "frontend-path, p", EnvVar: "FRONTEND_PATH", Value: "./frontend",
		Usage: "Location of the frontend files."},
	cli.BoolFlag{
		Name:  "debug",
		Usage: "Show (a lot) more output."},
	cli.BoolFlag{
		Name:  "help, h",
		Usage: "Shows this help, then exits."},
}

func main() {
	app := cli.NewApp()

	app.Name = "qbin"
	app.Version = "2.0.0a1"
	app.Usage = "a minimalist pastebin service"
	app.Flags = flags

	app.HideHelp = true
	cli.AppHelpTemplate = strings.Replace(cli.AppHelpTemplate, "GLOBAL OPTIONS:", "OPTIONS:", 1)

	app.Action = run

	app.Run(os.Args)
}

func run(c *cli.Context) error {
	if c.Bool("help") {
		cli.ShowAppHelp(c)
		return nil
	}

	if c.Bool("debug") {
		qbin.SetLogLevel(logging.DEBUG)
	}

	// Load words
	err := qbin.LoadWordsFile(c.String("wordlist"))
	if err != nil {
		qbin.Log.Errorf("Error loading word list from '%s': %s", c.String("wordlist"), err)
	}

	// Connect to database
	err = qbin.Connect(c.String("database"))
	if err != nil {
		qbin.Log.Criticalf("Error connecting to database: %s", err)
		panic(err)
	}

	// Serve HTTP
	if c.String("http") != "none" {
		go qbinHTTP.StartHTTP(c.String("http"), c.String("frontend-path"), c.String("root"))
	}

	// Serve TCP
	if c.String("tcp") != "none" {
		go qbinTCP.StartTCP(c.String("tcp"), c.String("root"))
	}

	// Sleep
	select {}
}
