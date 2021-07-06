// Package client implements a simple HTTP client to handle POST and
// GET commands.
package client

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
)

// Response carries return info
type Response struct {
	StatusCode int
	Status     string
	Body       []byte
	Cookies    []*http.Cookie
}

type Request struct {
	BaseURL     string
	Method      string
	Path        string
	ReqBody     string
	QueryParams map[string]string
	Headers     map[string]string
}

type Connection struct {
	baseURL string
	client  *http.Client
}

var transport *http.Transport

func CloseIdleConnections() {
	transport.CloseIdleConnections()
}

func Create(baseURL string) *Connection {
	cnx := Connection{}
	cnx.baseURL = baseURL

	defaultTransportPointer := &http.Transport{Proxy: http.ProxyFromEnvironment}
	transport := *defaultTransportPointer // copy it
	transport.MaxIdleConns = 4000
	transport.MaxIdleConnsPerHost = 4000
	transport.MaxConnsPerHost = 4000

	client := http.Client{Transport: &transport}
	client.Jar, _ = cookiejar.New(nil)
	cnx.client = &client

	return &cnx
}

// creates a connection for the test
func CreateTest(baseURL string, client *http.Client) *Connection {
	cnx := Connection{}
	cnx.baseURL = baseURL
	cnx.client = client

	return &cnx
}

func (cnx Connection) Do(ctx context.Context, req Request) (*Response, error) {

	// set up the url with query params if needed

	urlString := req.Path

	if len(req.BaseURL) > 0 {
		urlString = req.BaseURL + req.Path
	} else if len(cnx.baseURL) > 0 {
		urlString = cnx.baseURL + req.Path
	} 

	urlp, err := url.Parse(urlString)
	if err != nil {
		return nil, fmt.Errorf("%s (1): %s", cnx.baseURL, err)
	}

	if req.QueryParams != nil {
		q := urlp.Query()
		for k, v := range req.QueryParams {
			q.Set(k, v)
		}
		urlp.RawQuery = q.Encode()
	}

	reqBody := strings.NewReader(req.ReqBody)

	hreq, err := http.NewRequest(req.Method, urlp.String(), reqBody)
	//hreq = hreq.WithContext(httptrace.WithClientTrace(hreq.Context(), trace))

	if err != nil {
		return nil, fmt.Errorf("%s (2): %s", cnx.baseURL, err)
	}

	// add the headers
	if req.Headers != nil {
		for k, v := range req.Headers {
			hreq.Header.Set(k, v)
		}
	}

	// submit it
	resp, err := cnx.client.Do(hreq)
	if err != nil {

		uerr, ok := err.(*url.Error)

		if ok {
			nerr, ok := uerr.Err.(*net.OpError)
			if ok {
				log.Printf("nerr %#v", nerr)
			}
		}

		//serr := nerr.Err.(*os.SyscallError)
		//log.Printf("serr %#v, %q", serr, serr.Error())

		return nil, fmt.Errorf("%s (3): %s", cnx.baseURL, err)
	}

	defer resp.Body.Close()

	// pull body out
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%s (4): %s", cnx.baseURL, err)
	}

	// return statusCode, returned body, and returned cookies
	rval := Response{}
	rval.StatusCode = resp.StatusCode
	rval.Status = resp.Status
	rval.Body = b
	rval.Cookies = resp.Cookies()

	return &rval, nil
}