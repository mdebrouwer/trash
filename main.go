package main

import (
	"log"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"net/http"
	"net"
	"time"
	"fmt"
)

func NewSingleHostReverseProxy(target *url.URL) *httputil.ReverseProxy {
	director := func(req *http.Request) {
		fmt.Println("CALLING DIRECTOR")
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.URL.Path = target.Path
	}
	return &httputil.ReverseProxy{
		Director: director,
		Transport: &http.Transport{
			Proxy: func(req *http.Request) (*url.URL, error) {
				fmt.Println("CALLING PROXY")
				return http.ProxyFromEnvironment(req)
			},
			Dial: func(network, addr string) (net.Conn, error) {
				fmt.Println("CALLING DIAL")
				conn, err := (&net.Dialer{
					Timeout:   30 * time.Second,
					KeepAlive: 30 * time.Second,
				}).Dial(network, addr)
				if err == nil {
					return conn, err
				}
				fmt.Printf("Error during Dial: %v", err.Error())
				return net.Dial(network, "localhost:7070")
			},
			TLSHandshakeTimeout: 10 * time.Second,
		},
	}
}

func main() {
	filePath := "D:/data/"
	proxyUrl := "localhost:8080"

	fileServer := http.FileServer(http.Dir(filePath))
	go http.ListenAndServe(":7070", fileServer)

	stashProxy := NewSingleHostReverseProxy(&url.URL{
		Scheme: "http",
		Host:   proxyUrl,
	})
	go http.ListenAndServe(":9090", stashProxy)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	log.Println("Shutting down...")
}
