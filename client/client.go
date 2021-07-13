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
	BaseURL string
	client  *http.Client
}

var transport *http.Transport

func CloseIdleConnections() {
	transport.CloseIdleConnections()
}

func Create(baseURL string) *Connection {
	cnx := Connection{}
	cnx.BaseURL = baseURL

	defaultTransportPointer := &http.Transport{Proxy: http.ProxyFromEnvironment}
	transport := *defaultTransportPointer // copy it
	transport.MaxIdleConns = 4000
	transport.MaxIdleConnsPerHost = 4000
	transport.MaxConnsPerHost = 4000

	log.Printf("transport=%#v", transport)

	client := http.Client{Transport: &transport}
	client.Jar, _ = cookiejar.New(nil)
	cnx.client = &client

	return &cnx
}

var defaultConnection *Connection

func CreateDefault(baseURL string) {
	defaultConnection = Create(baseURL)
}

func Default() *Connection {
	return defaultConnection
}

// creates a connection for the test
func CreateTest(baseURL string, client *http.Client) *Connection {
	cnx := Connection{}
	cnx.BaseURL = baseURL
	cnx.client = client

	client.Jar, _ = cookiejar.New(nil)

	defaultConnection = &cnx

	return &cnx
}

// The Submit form does not take a context, and is called from apiseq

func Submit(req Request) (*Response, error) {

	if defaultConnection == nil {
		log.Fatalf("DefaultConnection has not been set")
	}

	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()
	
	return _submit(ctx, defaultConnection, req)
}

func (cnx *Connection) Submit (req Request) (*Response, error) {

	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()

	return _submit(ctx, cnx, req)
}

// The Do version takes a context, and is called from Fornax-Test

func Do(ctx context.Context, req Request) (*Response, error) {

	if defaultConnection == nil {
		log.Fatalf("DefaultConnection has not been set")
	}

	return _submit(ctx, defaultConnection, req)
}

func (cnx *Connection) Do (ctx context.Context, req Request) (*Response, error) {
	return _submit(ctx, cnx, req)
}

func _submit(ctx context.Context, cnx *Connection, req Request) (*Response, error) {

	// set up the url with query params if needed

	urlString := req.Path

	log.Printf("req.BaseURL=%v", req.BaseURL)
	log.Printf("cnx.baseURL=%v", cnx.BaseURL)

	if len(req.BaseURL) > 0 {
		urlString = req.BaseURL + req.Path
	} else if len(cnx.BaseURL) > 0 {
		urlString = cnx.BaseURL + req.Path
	} 

	log.Printf("urlString=%v", urlString)

	urlp, err := url.Parse(urlString)
	if err != nil {
		return nil, fmt.Errorf("%s (1): %s", cnx.BaseURL, err)
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

	if err != nil {
		return nil, fmt.Errorf("%s (2): %s", cnx.BaseURL, err)
	}

	// add the headers
	if req.Headers != nil {
		for k, v := range req.Headers {
			hreq.Header.Set(k, v)
		}
	}

	// add the cookies
	for _, c := range cnx.client.Jar.Cookies(urlp) {
		hreq.AddCookie(c)
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

		return nil, fmt.Errorf("%s (3): %s", cnx.BaseURL, err)
	}

	defer resp.Body.Close()

	// save off returned cookies
	cnx.client.Jar.SetCookies(urlp, resp.Cookies())

	// pull body out
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%s (4): %s", cnx.BaseURL, err)
	}

	// return statusCode, returned body, and returned cookies
	rval := Response{}
	rval.StatusCode = resp.StatusCode
	rval.Status = resp.Status
	rval.Body = b
	rval.Cookies = resp.Cookies()

	return &rval, nil
}
