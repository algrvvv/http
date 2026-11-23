package internal

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	flag "github.com/spf13/pflag"
)

const (
	BANNER          = " _     _   _           _           \n| |   | | | |         | |          \n| |__ | |_| |_ _ __   | |__  _   _ \n| '_ \\| __| __| '_ \\  | '_ \\| | | |\n| | | | |_| |_| |_) | | |_) | |_| |\n|_| |_|\\__|\\__| .__/  |_.__/ \\__, |\n              | |             __/ |\n              |_|            |___/ \n               _                                \n   ____       | |                               \n  / __ \\  __ _| | __ _ _ ____   ____   ____   __\n / / _` |/ _` | |/ _` | '__\\ \\ / /\\ \\ / /\\ \\ / /\n| | (_| | (_| | | (_| | |   \\ V /  \\ V /  \\ V / \n \\ \\__,_|\\__,_|_|\\__, |_|    \\_/    \\_/    \\_/  \n  \\____/          __/ |                         \n                 |___/                          \n"
	VERSION         = "0.7.0"
	REPOSITORY_LINK = "https://github.com/algrvvv/http"
)

type stringList []string

func (s *stringList) String() string {
	return strings.Join(*s, ",")
}

func (s *stringList) Set(str string) error {
	*s = append(*s, str)
	return nil
}

func (s *stringList) Type() string {
	return "stringList"
}

func (s *stringList) getHeaders() (http.Header, error) {
	headers := make(http.Header)
	for _, entry := range *s {
		name, value, ok := strings.Cut(entry, ":")
		if !ok || strings.TrimSpace(name) == "" {
			return nil, fmt.Errorf("invalid header %q: expected name:value", entry)
		}
		headers.Add(strings.TrimSpace(name), strings.TrimSpace(value))
	}
	return headers, nil
}

var (
	bodyOnly    = flag.Bool("body-only", false, "write only the raw response body")
	fail        = flag.Bool("fail", false, "return a nonzero exit code for HTTP 4xx and 5xx")
	pretty      = flag.Bool("pretty", false, "indent JSON response bodies")
	verbose     = flag.Bool("verbose", false, "show redirect hops and download timing on stderr")
	headers     stringList
	jsonHeaders = flag.BoolP(
		"json",
		"J",
		false,
		"use if you want to set header to content-type: application/json",
	)

	help    = flag.BoolP("help", "h", false, "show help message")
	version = flag.BoolP("version", "v", false, "show version info")
	timeout = flag.UintP(
		"timeout",
		"t",
		0,
		"number of seconds to wait for the request to complete",
	)
	cookies         = flag.StringP("cookies", "c", "", "cookies for request")
	proxy           = flag.StringP("proxy", "p", "", "proxy for request")
	useragent       = flag.StringP("user-agent", "u", "", "user-agent for request")
	redirect        = flag.BoolP("redirect", "r", false, "whether to follow redirects")
	AllHeaders      = flag.BoolP("all-headers", "A", false, "show all headers")
	WithoutBody     = flag.BoolP("without-body", "W", false, "dont show response body")
	ignoreCertCheck = flag.BoolP("ignore-cert-check", "I", false, "ignore certificate check")
	RequestBody     = flag.StringP("body", "b", "", "request body, @file or @-; example: '{\"key\":\"value\"}'")
)

func ParseAndGetRequest() (Request, error) {
	flag.VarP(&headers, "headers", "H", "request headers")
	flag.Parse()

	if *version {
		printVersionMessage()
		return Request{}, nil
	} else if *help {
		printHelpMessage()
		return Request{}, nil
	}

	args := flag.Args()
	if len(args) > 0 && args[0] == "set" {
		if len(args) != 3 {
			return Request{}, fmt.Errorf("usage: http set NAME URL")
		}
		return Request{}, saveHostAlias(args[1], args[2])
	}

	requestHeaders, err := headers.getHeaders()
	if err != nil {
		return Request{}, err
	}
	if *jsonHeaders && requestHeaders.Get("Content-Type") == "" {
		requestHeaders.Set("Content-Type", "application/json")
	}
	if *bodyOnly && (*WithoutBody || *pretty) {
		return Request{}, fmt.Errorf("--body-only cannot be combined with --without-body or --pretty")
	}

	req := Request{
		Timeout:         time.Duration(*timeout),
		Cookies:         *cookies,
		Proxy:           *proxy,
		UserAgent:       *useragent,
		Headers:         requestHeaders,
		Redirect:        *redirect,
		IgnoreCertCheck: *ignoreCertCheck,
	}

	err = parseCommandLine(&req)
	if err != nil {
		return Request{}, err
	}

	return req, nil
}

func parseCommandLine(req *Request) error {
	args := flag.Args()
	if len(args) < 2 {
		return fmt.Errorf("method and URL are required; use --help for usage")
	}
	req.Method = strings.ToUpper(args[0])
	resolvedURL, err := resolveHostAlias(args[1])
	if err != nil {
		return err
	}
	req.URL = resolvedURL
	body, err := parseBody(args[2:], *RequestBody, flag.CommandLine.Changed("body"), *jsonHeaders, os.Stdin)
	if err != nil {
		return err
	}
	req.Body = body
	return nil
}

func parseBody(args []string, body string, bodySet, jsonMode bool, stdin io.Reader) ([]byte, error) {
	if len(args) > 0 {
		if !jsonMode || bodySet {
			return nil, fmt.Errorf("positional JSON fields require -J and cannot be combined with -b")
		}
		fields := make(map[string]string, len(args))
		for _, arg := range args {
			key, value, ok := strings.Cut(arg, "=")
			if !ok || key == "" {
				return nil, fmt.Errorf("expected a JSON field in key=value format")
			}
			if _, exists := fields[key]; exists {
				return nil, fmt.Errorf("duplicate JSON field %q", key)
			}
			fields[key] = value
		}
		return json.Marshal(fields)
	}
	if body == "@-" {
		return io.ReadAll(stdin)
	}
	if path, ok := strings.CutPrefix(body, "@"); ok {
		return os.ReadFile(path)
	}
	return []byte(body), nil
}

func printHelpMessage() {
	fmt.Printf("Usage: http METHOD URL [flags...] [key=value ...]\n       http set NAME URL\n\n")
	flag.PrintDefaults()
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
