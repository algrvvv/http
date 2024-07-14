package internal

import (
	"fmt"
	"strings"
	"time"

	flag "github.com/spf13/pflag"
)

const (
	BANNER          = " _     _   _           _           \n| |   | | | |         | |          \n| |__ | |_| |_ _ __   | |__  _   _ \n| '_ \\| __| __| '_ \\  | '_ \\| | | |\n| | | | |_| |_| |_) | | |_) | |_| |\n|_| |_|\\__|\\__| .__/  |_.__/ \\__, |\n              | |             __/ |\n              |_|            |___/ \n               _                                \n   ____       | |                               \n  / __ \\  __ _| | __ _ _ ____   ____   ____   __\n / / _` |/ _` | |/ _` | '__\\ \\ / /\\ \\ / /\\ \\ / /\n| | (_| | (_| | | (_| | |   \\ V /  \\ V /  \\ V / \n \\ \\__,_|\\__,_|_|\\__, |_|    \\_/    \\_/    \\_/  \n  \\____/          __/ |                         \n                 |___/                          \n"
	VERSION         = "0.4.19"
	REPOSITORY_LINK = "https://github.com/algrvvv/http"
)

var (
	help        = flag.BoolP("help", "h", false, "show help message")
	version     = flag.BoolP("version", "v", false, "show version info")
	timeout     = flag.UintP("timeout", "t", 0, "number of seconds to wait for the request to complete")
	cookies     = flag.StringP("cookies", "c", "", "(TODO) cookies for request")
	proxy       = flag.StringP("proxy", "p", "", "(TODO) proxy for request")
	useragent   = flag.StringP("user-agent", "u", "", "(TODO) user-agent for request")
	redirect    = flag.BoolP("redirect", "r", false, "whether to follow redirects")
	headers     = flag.StringP("headers", "H", "", "(TODO) list of header names, separated by semicolons")
	AllHeaders  = flag.BoolP("all-headers", "A", false, "show all headers")
	WithoutBody = flag.BoolP("without-body", "W", false, "dont show request body")
)

type InvalidFlagOrOption struct {
	iflag string // флаг, который не был распознан
}

func (e InvalidFlagOrOption) Error() string {
	return fmt.Sprintf("Полученн недопустимый флаг: %s", e.iflag)
}

type MissingFlagOrOption struct {
	mflag string // недостающий флаг
}

func (e MissingFlagOrOption) Error() string {
	if e.mflag != "" {
		return fmt.Sprintf("Был пропущен недостающий флаг: %s", e.mflag)
	} else {
		return fmt.Sprintf("Пропущены критически важные для работы аргументы.\nИспользуетя -help, чтобы узнать о том, как пользоваться")
	}
}

func ParseAndGetRequest() (Request, error) {
	flag.Parse()

	if *version {
		printVersionMessage()
		return Request{}, nil
	} else if *help {
		printHelpMessage()
		return Request{}, nil
	}

	req := Request{
		Timeout:   time.Duration(*timeout),
		Cookies:   *cookies,
		Proxy:     *proxy,
		UserAgent: *useragent,
		Headers:   *headers,
		Redirect:  *redirect,
	}

	err := parseCommandLine(&req)
	if err != nil {
		return Request{}, err
	}

	return req, nil
}

func parseCommandLine(req *Request) error {
	args := flag.Args()
	if len(args) >= 2 {
		req.Method = strings.ToUpper(args[0])
		req.URL = checkAndGetURL(args[1])
		return nil
	} else {
		return MissingFlagOrOption{mflag: ""}
	}
}

func printHelpMessage() {
	fmt.Printf("Usage: http [method] [url] [flags...]\n\n")
	flag.PrintDefaults()
	// fmt.Printf("\t-t | --timeout\t\tspecify the number of seconds to wait for the request to complete\n")
}

func printVersionMessage() {
	fmt.Printf("%s\nCurrent version is %s\nRepository: %s\n", BANNER, VERSION, REPOSITORY_LINK)
}

func checkAndGetURL(url string) string {
	startWith := strings.Index(url, "://")
	if startWith == -1 {
		return fmt.Sprintf("http://%s", url)
	}

	return url
}
