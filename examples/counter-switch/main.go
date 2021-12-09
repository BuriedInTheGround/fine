package main

import (
	"fmt"
	"math/rand"
	"time"

	"interrato.dev/fine"
)

const maxRuns int = 50000

const lightbulbLifetime int = 10000

func main() {
	// Increment the counter every time the switch turns to ON.
	counter := 0
	toggleOn := func() string {
		counter += 1
		return "on"
	}

	// Initialize the switch.
	lightSwitch := fine.Machine("off", fine.States{
		"off": {"toggle": toggleOn},
		"on":  {"toggle": "off"},
	})

	// Select a random number of runs.
	rand.Seed(time.Now().UnixNano())
	t := rand.Intn(maxRuns)

	// Toggle the switch t times.
	for i := 0; i < t; i++ {
		lightSwitch.Do("toggle")
	}

	// Check if the lightbulb is still good.
	if counter > lightbulbLifetime {
		fmt.Println("Ops, you broke the lightbulb!")
	} else {
		fmt.Printf(
			"The lightbulb is fine, and currently turned %s.\n",
			lightSwitch.State(),
		)
	}
}
