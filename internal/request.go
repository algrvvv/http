package internal

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/algrvvv/http/internal/logger"
)

type Response struct {
	Protocol     string      // протокол
	Code         int         // код ответа
	Status       string      // статус ответа
	FinalLink    string      // итоговая ссылка, на которой оставновился запрос
	RedirectLink string      // ссылка переадресации, если есть
	Header       http.Header // заголовки
	Body         []byte      // тело ответа
	TotalTime    time.Duration
	Redirects    []string
	TimeLoad     time.Duration // время ответа
}

var shortListOfHeaders = []string{
	"Content-Type", "Accept", "Date", "Content-Length", "Connection",
}

func (r *Response) proto() string {
	return logger.LightBlue(r.Protocol)
}

func (r *Response) statusCode() string {
	if r.Code >= 200 && r.Code < 300 {
		return logger.LightGreen(r.Status)
	} else if r.Code >= 300 && r.Code < 400 {
		return logger.Orange(r.Status)
	} else {
		return logger.LightRed(r.Status)
	}
}

func (r *Response) FormatOutput() error {
	if *verbose {
		for _, hop := range r.Redirects {
			fmt.Fprintln(os.Stderr, "Redirect:", hop)
		}
		fmt.Fprintln(os.Stderr, "Total time:", r.TotalTime)
	}
	if *bodyOnly {
		_, err := os.Stdout.Write(r.Body)
		return err
	}
	fmt.Printf("%s %s\n", r.proto(), r.statusCode())
	fmt.Println("Final url: ", r.FinalLink)
	if r.RedirectLink != "" {
		fmt.Println("Redirect url: ", r.RedirectLink)
	}
	fmt.Println("Response time: ", r.TimeLoad)
	if *AllHeaders {
		for k, v := range r.Header {
			fmt.Println(k, ":", strings.Join(v, ","))
		}
	} else {
		for k, v := range r.Header {
			if slices.Index(shortListOfHeaders, k) != -1 {
				fmt.Println(k, ":", strings.Join(v, ","))
			}
		}
	}

	fmt.Println()

	if !*WithoutBody {
		if *pretty {
			var formatted bytes.Buffer
			if json.Indent(&formatted, r.Body, "", "  ") == nil {
				fmt.Println(formatted.String())
				return nil
			}
		}
		decodedBody, err := decodeUnicodeEscapes(r.Body)
		if err != nil {
			fmt.Println(string(r.Body) + "\n\n")
		} else {
			fmt.Println(decodedBody + "\n\n")
		}
	}
	return nil
}

func (r Response) StatusError() error {
	if *fail && r.Code >= 400 && r.Code < 600 {
		return fmt.Errorf("HTTP %s", r.Status)
	}
	return nil
}

type Request struct {
	Method          string        // метод запроса
	URL             string        // линк для запроса
	Body            []byte        // тело запроса
	Headers         http.Header   // заголовки
	UserAgent       string        // юзер агент
	Cookies         string        // куки
	Proxy           string        // прокси
	Timeout         time.Duration // время выделенное на запрос
	Redirect        bool          // Следовать ли за редиректами
	IgnoreCertCheck bool          // игнорирования проверки сертификатов (удобно для локального тестирования)
}

func (r Request) MakeRequest() (Response, error) {
	ctx := context.Background()
	if r.Timeout != 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, r.Timeout*time.Second)
		defer cancel()
	}

	req, err := http.NewRequestWithContext(ctx, r.Method, r.URL, bytes.NewReader(r.Body))
	if err != nil {
		return Response{}, err
	}
	if r.Headers != nil {
		req.Header = r.Headers.Clone()
	}
	if r.UserAgent != "" {
		req.Header.Set("User-Agent", r.UserAgent)
	}
	if r.Cookies != "" {
		req.Header.Add("Cookie", r.Cookies)
	}

	var redirects []string
	client := &http.Client{
		CheckRedirect: func(next *http.Request, via []*http.Request) error {
			if !r.Redirect {
				return http.ErrUseLastResponse
			}
			if len(via) >= 10 {
				return fmt.Errorf("stopped after 10 redirect attempts")
			}
			redirects = append(redirects, fmt.Sprintf("%s -> %s", via[len(via)-1].URL, next.URL))
			return nil
		},
	}
	if r.IgnoreCertCheck || r.Proxy != "" {
		transport := http.DefaultTransport.(*http.Transport).Clone()
		defer transport.CloseIdleConnections()
		if r.IgnoreCertCheck {
			transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		}
		if r.Proxy != "" {
			proxyURL, err := url.Parse(r.Proxy)
			if err != nil || proxyURL.Host == "" || proxyURL.Scheme == "" {
				return Response{}, fmt.Errorf("invalid proxy URL: expected scheme://host[:port]")
			}
			transport.Proxy = http.ProxyURL(proxyURL)
		}
		client.Transport = transport
	}

	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return Response{}, err
	}
	defer resp.Body.Close()

	loadTime := time.Since(start)

	var body []byte
	if !*WithoutBody {
		body, err = io.ReadAll(resp.Body)
		if err != nil {
			return Response{}, err
		}
	}
	totalTime := time.Since(start)

	var redirectURL string
	redirectLink, err := resp.Location()
	if err != nil {
		redirectURL = ""
	} else {
		redirectURL = redirectLink.String()
	}

	return Response{
		FinalLink:    resp.Request.URL.String(),
		RedirectLink: redirectURL,
		Code:         resp.StatusCode,
		TimeLoad:     loadTime,
		TotalTime:    totalTime,
		Redirects:    redirects,
		Header:       resp.Header,
		Body:         body,
		Protocol:     resp.Proto,
		Status:       resp.Status,
	}, nil
}

func decodeUnicodeEscapes(data []byte) (string, error) {
	str, err := strconv.Unquote(strings.Replace(strconv.Quote(string(data)), `\\u`, `\u`, -1))
	if err != nil {
		return "", err
	}
	return str, nil
}
