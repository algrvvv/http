package internal

import (
	"flag"
	"fmt"
	"github.com/algrvvv/http/internal/http"
)

const (
	VERSION = "0.0.1"
)

var (
	waitingTime = flag.Int("waitingTime", -1, "waiting time in seconds")
	version     = flag.Bool("version", false, "show version")
	help        = flag.Bool("help", false, "show help message")
)

type InvalidFlags struct {
	iflag string // флаг, который не был распознан
}

func (e InvalidFlags) Error() string {
	return fmt.Sprintf("Полученн недопустимый флаг: %s", e.iflag)
}

type MissingFlag struct {
	mflag string // недостающий флаг
}

func (e MissingFlag) Error() string {
	return fmt.Sprintf("Был пропущен недостающий флаг: %s", e.mflag)
}

func ParseUtilOptions() (http.Request, error) {
	flag.Parse()

	if *version {
		printVersionMessage()
		return http.Request{}, nil
	} else if *help {
		printHelpMessage()
		return http.Request{}, nil
	}

	return http.Request{}, nil
}

func printHelpMessage() {
	fmt.Println("http is util by golang and @algrvvv")
}

func printVersionMessage() {
	fmt.Println("version:", VERSION)
}
