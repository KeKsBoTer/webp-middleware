package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os/exec"
	"strings"
)

func serve(target string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		resp, err := http.Get(target + path)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			log.Println(err)
			return
		}
		if resp.StatusCode != http.StatusOK {
			w.WriteHeader(resp.StatusCode)
			return
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			log.Println(err)
			return
		}
		bodyBuf := bytes.NewBuffer(body)

		var command string
		if strings.HasSuffix(path, ".gif") {
			command = "gif2webp"
		} else {
			command = "cwebp"
		}
		var b bytes.Buffer
		convertCmd := exec.Command(command, "-o", "-", "--", "-")
		convertCmd.Stdin = bodyBuf
		convertCmd.Stdout = &b

		h := w.Header()
		// something went wrong, use original image
		for k := range resp.Header {
			h.Set(k, h.Get(k))
		}

		err = convertCmd.Run()
		if err != nil {
			log.Println(err)
			w.Write(body)
			return
		}
		b.WriteTo(w)
	}
}

func main() {

	port := flag.Int("port", 3333, "port the image converter sever runs on")
	target := flag.String("target", "http://localhost:8080", "target that the request will be sent to")

	flag.Parse()

	http.HandleFunc("/", serve(*target))
	fmt.Printf("starting webp image convertion server on http://localhost:%d\n", *port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", *port), nil); err != nil {
		panic(err)
	}
}
