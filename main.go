package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os/exec"
	"strings"
)

func convert(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return nil
	}
	path := resp.Request.URL.Path
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	bodyBuf := bytes.NewBuffer(body)

	var command string
	if strings.HasSuffix(path, ".gif") {
		command = "gif2webp"
	} else {
		command = "cwebp"
	}
	var b bytes.Buffer
	var b_err bytes.Buffer
	convertCmd := exec.Command(command, "-o", "-", "--", "-")
	convertCmd.Stdin = bodyBuf
	convertCmd.Stdout = &b
	convertCmd.Stderr = &b_err

	err = convertCmd.Run()
	if err != nil {
		stderr, _ := b_err.ReadString(0)
		fmt.Println("Error converting: ", stderr)
		return err
	}
	resp.Body = io.NopCloser(&b)
	resp.Header.Del("Content-Length")
	resp.Header.Del("Accept-Ranges")
	resp.Header.Set("Transfer-Encoding", "chunked")
	resp.Header.Set("Content-Type", "image/webp")
	return nil
}

func main() {

	port := flag.Int("port", 3333, "port the image converter sever runs on")
	target := flag.String("target", "http://localhost:8080", "target that the request will be sent to")

	flag.Parse()

	t, err := url.Parse(*target)
	if err != nil {
		panic(err)
	}

	reverseProxy := httputil.NewSingleHostReverseProxy(t)
	reverseProxy.ModifyResponse = convert

	http.Handle("/", reverseProxy)
	fmt.Printf("starting webp image convertion server on http://localhost:%d\n", *port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", *port), nil); err != nil {
		panic(err)
	}
}
