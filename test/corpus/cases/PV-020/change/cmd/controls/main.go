package main

import (
	"flag"
	"fmt"
	"os"

	"pv020change/service"
)

func main() {
	group := flag.String("group", "", "in-process, description, or listener")
	description := flag.String("description", "", "path of the API description")
	limitRemoved := flag.Bool("limit-removed", false, "run the in-process controls with the limit removed")
	flag.Parse()

	var out service.Outcome
	switch *group {
	case "in-process":
		svc := service.Production()
		if *limitRemoved {
			svc = service.WithoutLimit()
		}
		out = service.RunInProcess(svc)
	case "description":
		out = service.RunDescription(*description)
	case "listener":
		out = service.RunListener(service.Production())
	default:
		fmt.Fprintln(os.Stderr, "group must be in-process, description, or listener")
		os.Exit(2)
	}
	fmt.Println(out.Line())
}
