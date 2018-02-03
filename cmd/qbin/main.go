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
		Name: "database, d", EnvVar: "DATABASE", Value: "root:@tcp(localhost)/qbin",
		Usage: "MySQL/MariaDB connection string. It is recommended to pass this parameter as an environment variable."},
	cli.StringFlag{
		Name: "root, r", EnvVar: "ROOT_URL", Value: "http://127.0.0.1:8000",
		Usage: "The path under which the application will be reachable from the internet."},
	cli.BoolFlag{
		Name: "force-root", EnvVar: "FORCE_ROOT",
		Usage: "If this is set, requests that are not on the root URI will be redirected."},
	cli.StringFlag{
		Name: "wordlist", EnvVar: "WORD_LIST", Value: "eff_large_wordlist.txt",
		Usage: "Word list used for random slug generation."},
	cli.StringFlag{
		Name: "blacklist", EnvVar: "BLACKLIST", Value: "blacklist.regex",
		Usage: "Blacklist file containing one regular expression per line."},
	cli.StringSliceFlag{
		Name: "filters", EnvVar: "FILTERS", Value: &cli.StringSlice{"blacklist", "linkcount"},
		Usage: "Set the spam filters in use. Available filters: blacklist, linkcount"},
	cli.StringFlag{
		Name: "prism-server", EnvVar: "PRISM_SERVER", Value: "/tmp/prism-server.sock",
		Usage: "TCP address or unix socket path (when containing a /) to prism-server."},
	cli.StringFlag{
		Name: "tcp", EnvVar: "TCP_LISTEN", Value: ":9000",
		Usage: "TCP (netcat API) listen address. Set to 'none' to disable."},
	cli.StringFlag{
		Name: "http", EnvVar: "HTTP_LISTEN", Value: ":8000",
		Usage: "HTTP listen address. Set to 'none' to disable."},
	cli.StringFlag{
		Name: "https", EnvVar: "HTTPS_LISTEN", Value: "none",
		Usage: "HTTPS listen address, qbin will automatically get a Let's Encrypt certificate. Set to 'none' to disable."},
	cli.BoolFlag{
		Name: "hsts", EnvVar: "HSTS",
		Usage: "Send HSTS header with max-age=31536000 (1 year)."},
	cli.BoolFlag{
		Name: "hsts-preload", EnvVar: "HSTS_PRELOAD",
		Usage: "Send preload directive with the HSTS header. Requires --hsts."},
	cli.BoolFlag{
		Name: "hsts-subdomains", EnvVar: "HSTS_SUBDOMAINS",
		Usage: "Send includeSubDomains directive with the HSTS header. Requires --hsts."},
	cli.StringFlag{
		Name: "frontend-path, p", EnvVar: "FRONTEND_PATH", Value: "./frontend",
		Usage: "Location of the frontend files."},
	cli.BoolFlag{
		Name: "debug", EnvVar: "DEBUG",
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

	// Switch filters
	qbin.FilterEnable = qbin.Slice2map(c.StringSlice("filters"))

	// Load blacklist
	err = qbin.LoadBlacklistFile(c.String("blacklist"))
	if err != nil {
		qbin.Log.Errorf("Error loading blacklist from '%s': %s", c.String("blacklist"), err)
	}

	// Setup prism-server
	qbin.PrismServer = c.String("prism-server")

	// Connect to database
	err = qbin.Connect(c.String("database"))
	if err != nil {
		qbin.Log.Criticalf("Error connecting to database: %s", err)
		panic(err)
	}

	// Serve HTTP
	if c.String("http") != "none" || c.String("https") != "none" {
		hsts := ""
		if c.String("https") == "none" && c.Bool("hsts") {
			qbin.Log.Warning("You are using --hsts without --https. Ignoring and keeping HSTS off.")
		} else if c.Bool("hsts") {
			hsts = "max-age=31536000"
			if c.Bool("hsts-subdomains") {
				hsts += "; includeSubDomains"
			}
			if c.Bool("hsts-preload") {
				hsts += "; preload"
			}
		} else if c.Bool("hsts-subdomains") || c.Bool("hsts-preload") {
			qbin.Log.Warning("You are using --hsts-subdomains or --hsts-preload without --hsts. Ignoring and keeping HSTS off.")
		}

		go qbinHTTP.StartHTTP(qbinHTTP.Configuration{
			ListenHTTP:    c.String("http"),
			ListenHTTPS:   c.String("https"),
			FrontendPath:  c.String("frontend-path"),
			Root:          c.String("root"),
			CertWhitelist: c.Args(),
			ForceRoot:     c.Bool("force-root"),
			Hsts:          hsts,
		})
	}

	// Serve TCP
	if c.String("tcp") != "none" {
		go qbinTCP.StartTCP(c.String("tcp"), c.String("root"))
	}

	// Sleep
	select {}
}
