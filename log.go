package qbin

import (
	"os"

	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("example")

func init() {

	backend := logging.NewLogBackend(os.Stdout, "", 0)

	format := logging.MustStringFormatter(`%{color}%{time:15:04:05.000} %{shortfunc} › %{level:.4s} %{color:reset} %{message}`)
	formatter := logging.NewBackendFormatter(backend, format)

	logging.SetBackend(formatter)

}
