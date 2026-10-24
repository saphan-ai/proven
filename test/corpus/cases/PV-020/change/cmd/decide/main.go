package main

import (
	"flag"
	"fmt"
	"os"

	"pv020change/service"
)

func main() {
	printLimit := flag.Bool("print-limit", false, "print the documented limit and exit")
	n := flag.Int("bytes", -1, "request body size in bytes")
	flag.Parse()
	if *printLimit {
		fmt.Println(service.DocumentedLimit)
		return
	}
	if *n < 0 {
		fmt.Fprintln(os.Stderr, "bytes is required and must be >= 0")
		os.Exit(2)
	}
	body := make([]byte, *n)
	for i := range body {
		body[i] = 'x'
	}
	svc := service.Production()
	for _, rt := range service.Routes {
		d := svc.Decide(rt.Method, rt.Path, body)
		acc := "refused"
		if d.Accepted {
			acc = "accepted"
		}
		fmt.Printf("%s %s|%s|%s\n", rt.Method, rt.Path, acc, d.Reason)
	}
}
