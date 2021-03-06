package client

import (
	"context"
	"log"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNormal(t *testing.T) {

	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()

	f := func(rw http.ResponseWriter, req *http.Request) {
		log.Printf("handler called, path=%s", req.URL.String())
	}

	server := httptest.NewServer(http.HandlerFunc(f))

	log.Printf("server.URL is %s", server.URL)

	CreateTest(server.URL, server.Client())

	req := Request{}

	req.Method = "GET"
	req.Path = "/xyzzy"

	resp, err := Do(ctx, &req)
	log.Printf("err is %#v", err)
	log.Printf("err is %T", err)
	log.Printf("resp is %#v", resp)
}

func TestConnectRefused(t *testing.T) {

	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()

	f := func(rw http.ResponseWriter, req *http.Request) {
		log.Printf("handler called, path=%s", req.URL.String())
	}

	server := httptest.NewServer(http.HandlerFunc(f))

	server.Close()

	log.Printf("server.URL is %s", server.URL)

	CreateTest(server.URL, server.Client())

	req := Request{}

	req.Method = "GET"
	req.Path = "/xyzzy"

	resp, err := Do(ctx, &req)
	log.Printf("err is %#v", err)
	log.Printf("err is %T", err)
	log.Printf("resp is %#v", resp)
}

func TestConnectionClosed(t *testing.T) {

	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()

	var server *httptest.Server

	f := func(rw http.ResponseWriter, req *http.Request) {

		log.Printf("handler called, path=%s", req.URL.String())
		log.Print("sleeping for 30 seconds")
		time.Sleep(30 * time.Second)
	}

	server = httptest.NewServer(http.HandlerFunc(f))

	go func() {
		time.Sleep(3 * time.Second)
		log.Printf("closing server")
		server.Listener.Close()
	}()

	log.Printf("server.URL is %s", server.URL)

	CreateTest(server.URL, server.Client())

	req := Request{}

	req.Method = "GET"
	req.Path = "/xyzzy"

	resp, err := Do(ctx, &req)
	log.Printf("err is %#v", err)
	log.Printf("err is %q", err)
	log.Printf("err is %T", err)
	log.Printf("resp is %#v", resp)

}

func TestRequestTimeout(t *testing.T) {

	timeout := 10 * time.Second

	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()

	f := func(rw http.ResponseWriter, req *http.Request) {

		log.Printf("handler called, path=%s", req.URL.String())
		log.Printf("sleeping 30 seconds")
		time.Sleep(timeout * 2)
		log.Printf("done sleeping")
	}

	ts := httptest.NewUnstartedServer(http.HandlerFunc(f))

	defer ts.Close()

	ts.Config.WriteTimeout = timeout
	ts.Start()

	cx := Create(ts.URL)

	req := Request{}

	req.Method = "GET"
	req.Path = "/xyzzy"

	resp, err := cx.Do(ctx, &req)

	log.Printf("err is %#v", err)
	log.Printf("err is %q", err)
	log.Printf("err is %T", err)
	log.Printf("resp is %#v", resp)
}

func TestProxy(t *testing.T) {
	log.Printf("Running test proxy")

	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()

	req := Request{}

	req.Method = "POST"

	cx := Create("https://api-devintever3.mountain.siriusxm.com")

	b, err := ioutil.ReadFile("testdata/auth.json")
	if err != nil {
		log.Fatal(err)
	}

	req.ReqBody = string(b)
	req.Path = "/rest/v2/experience/modules/modify/authentication"

	resp, err := cx.Do(ctx, &req)
	
	log.Printf("err is %#v", err)
	log.Printf("err is %q", err)
	log.Printf("err is %T", err)
	log.Printf("resp is %#v", string(resp.Body))

	b, err = ioutil.ReadFile("testdata/resume.json")
	if err != nil {
		log.Fatal(err)
	}

	req.ReqBody = string(b)
	req.Path = "/rest/v2/experience/modules/resume"

	resp, err = cx.Do(ctx, &req)
	
	log.Printf("err is %#v", err)
	log.Printf("err is %q", err)
	log.Printf("err is %T", err)
	log.Printf("resp is %#v", string(resp.Body))
}
