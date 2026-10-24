package main

import (
	"fmt"
	"os"
	"strings"
)

// usage: tool dest <setting> | tool emit <alert> | tool fields <alert>
func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: tool dest <setting> | emit <alert> | fields <alert>")
		os.Exit(2)
	}
	switch os.Args[1] {
	case "dest":
		id, err := ParseDestination(os.Args[2])
		if err != nil {
			fmt.Printf("refused: %v\n", err)
			return
		}
		fmt.Printf("accepted: %s\n", id)
	case "emit":
		res, err := Emit(os.Args[2], &Receiver{})
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Printf("considered=%d delivered=%d reason=%q\n", res.Considered, res.Delivered, res.Reason)
	case "fields":
		fs, err := FieldsOf(os.Args[2])
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Println(strings.Join(fs, ","))
	default:
		fmt.Fprintln(os.Stderr, "unknown operation")
		os.Exit(2)
	}
}
