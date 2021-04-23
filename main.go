package main

import (
	"bytes"
	"flag"
	"fmt"
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

		var command string
		if strings.HasSuffix(path, ".gif") {
			command = "gif2webp"
		} else {
			command = "cwebp"
		}
		var b bytes.Buffer
		convertCmd := exec.Command(command, "-quiet", "-o", "-", "--", "-")
		convertCmd.Stdin = resp.Body
		convertCmd.Stdout = &b

		err = convertCmd.Run()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println(err)
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
