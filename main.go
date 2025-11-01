package main

import (
	"flag"
	"fmt"
)

type Schedule struct {
}

func main() {
	schedule := flag.String("schedule", "", "path to schedule JSON file")
	overrides := flag.String("overrides", "", "path to overrides JSON file")
	from := flag.String("from", "", "start time for the schedule")
	until := flag.String("until", "", "end time for the schedule")
	flag.Parse()

	fmt.Println("overrides:", *overrides)
	fmt.Println("from:", *from)
	fmt.Println("until:", *until)

	if *schedule != "" {
		fmt.Println("schedule:", *schedule)
	}

	if *overrides != "" {
		fmt.Println("overrides:", *overrides)
	}

	if *from != "" {
		fmt.Println("from:", *from)
	}

	if *until != "" {
		fmt.Println("until:", *until)
	}
}
