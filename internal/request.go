package internal

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/algrvvv/http/internal/logger"
)

type Response struct {
	Protocol     string        // протокол
	Code         int           // код ответа
	Status       string        // статус ответа
	FinalLink    string        // итоговая ссылка, на которой оставновился запрос
	RedirectLink string        // ссылка переадресации, если есть
	Header       http.Header   // заголовки
	Body         []byte        // тело ответа
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

func (r *Response) FormatOutput() {
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

	fmt.Println("\n")

	if !*WithoutBody {
		decodedBody, err := decodeUnicodeEscapes(r.Body)
		if err != nil {
			fmt.Println(string(r.Body) + "\n\n")
		} else {
			fmt.Println(decodedBody + "\n\n")
		}
	}
}

type Request struct {
	Method          string            // метод запроса
	URL             string            // линк для запроса
	Body            []byte            // тело запроса
	Headers         map[string]string // заголовки
	UserAgent       string            // юзер агент
	Cookies         string            // куки
	Proxy           string            // TODO прокси
	Timeout         time.Duration     // время выделенное на запрос
	Redirect        bool              // Следовать ли за редиректами
	IgnoreCertCheck bool              // игнорирования проверки сертификатов (удобно для локального тестирования)
}

func (r Request) MakeRequest() (Response, error) {
	var (
		resp Response
		err  error
	)

	if r.Timeout != 0 {
		resp, err = r.getResponseWithTimeout()
	} else {
		resp, err = r.getResponse()
	}

	return resp, err
}

func (r Request) getResponse() (Response, error) {
	logger.Logger(string(r.Body), logger.SuccessLogType)
	req, err := http.NewRequest(r.Method, r.URL, bytes.NewReader(r.Body))
	if err != nil {
		return Response{}, err
	}
	return r.makeRequestWrapper(req)
}

func (r Request) getResponseWithTimeout() (Response, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.Timeout*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, r.Method, r.URL, bytes.NewReader(r.Body))
	if err != nil {
		return Response{}, err
	}
	return r.makeRequestWrapper(req)
}

func (r Request) makeRequestWrapper(req *http.Request) (Response, error) {
	// change headers
	for key, value := range r.Headers {
		req.Header.Add(key, value)
	}

	start := time.Now()
	client := &http.Client{}
	if r.Redirect {
		client = &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return nil
			},
		}
	} else {
		client = &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}
	}

	if r.IgnoreCertCheck {
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return Response{}, err
	}
	defer resp.Body.Close()

	loadTime := time.Since(start)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Response{}, err
	}

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
