package internal

import (
	"fmt"
	"strings"
	"time"

	flag "github.com/spf13/pflag"
)

const (
	VERSION = "0.0.1"
)

var (
	help       = flag.BoolP("help", "h", false, "show help message")
	version    = flag.BoolP("version", "v", false, "show version info")
	timeout    = flag.UintP("timeout", "t", 0, "timeout in seconds")
	cookies    = flag.StringP("cookies", "c", "", "cookies for request")
	proxy      = flag.StringP("proxy", "p", "", "proxy to use")
	useragent  = flag.StringP("user-agent", "u", "", "user-agent to use")
	redirect   = flag.BoolP("redirect", "r", false, "redirect after successful request")
	headers    = flag.StringP("headers", "H", "", "list of header names, separated by semicolons")
	AllHeaders = flag.BoolP("all-headers", "A", false, "show all header names")
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
		req.Method = args[0]
		req.URL = checkAndGetURL(args[1])
		return nil
	} else {
		return MissingFlagOrOption{mflag: ""}
	}
}

func printHelpMessage() {
	fmt.Println("http is util by golang and @algrvvv")
}

func printVersionMessage() {
	fmt.Println("version:", VERSION)
}

func checkAndGetURL(url string) string {
	startWith := strings.Index(url, "://")
	if startWith == -1 {
		return fmt.Sprintf("http://%s", url)
	}

	return url
}
