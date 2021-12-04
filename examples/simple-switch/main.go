package main

import (
	"fmt"

	"interrato.dev/fike"
)

func main() {
	lightSwitch := fike.Machine("off", fike.States{
		"off": {"toggle": "on"},
		"on":  {"toggle": "off"},
	})

	lightSwitch.Do("toggle") // => "on", nil
	lightSwitch.Do("toggle") // => "off", nil

	switch lightSwitch.State() {
	case "off":
		fmt.Println("Come on, turn the lights on!")
	case "on":
		fmt.Println("Oh, finally I see you!")
	}

	// Output:
	// C'mon, turn the lights on!
}
