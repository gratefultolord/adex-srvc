package main

import (
	"flag"
	"fmt"
	"net/http"
	"time"
)

func main() {
	port := flag.Int("port", 9001, "server port")
	status := flag.Int("status", http.StatusOK, "response status")
	delay := flag.Duration("delay", 0, "response delay")
	flag.Parse()

	mux := http.NewServeMux()

	mux.HandleFunc("POST /bid", func(w http.ResponseWriter, r *http.Request) {
		if *delay > 0 {
			time.Sleep(*delay)
		}

		w.WriteHeader(*status)
	})

	addr := fmt.Sprintf(":%d", *port)

	fmt.Printf(
		"mock DSP started: addr=%s status=%d delay=%s\n",
		addr,
		*status,
		delay.String(),
	)

	if err := http.ListenAndServe(addr, mux); err != nil {
		panic(err)
	}
}
