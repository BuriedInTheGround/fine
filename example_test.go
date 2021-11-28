package fike_test

import (
	"fmt"

	"github.com/BuriedInTheGround/fike"
)

func ExampleMachine() {
	powerSwitch := fike.Machine("off", fike.States{
		"off": {"toggle": "on"},
		"on":  {"toggle": "off"},
	})

	fmt.Println("Current FSM state:", powerSwitch.State())
	// Output:
	// Current FSM state: off
}

func ExampleFSM_Do() {
	powerSwitch := fike.Machine("off", fike.States{
		"off": {"toggle": "on"},
		"on":  {"toggle": "off"},
	})

	fmt.Println("Current FSM state:", powerSwitch.State())
	fmt.Println("Toggling...")
	powerSwitch.Do("toggle")
	fmt.Println("Current FSM state:", powerSwitch.State())
	// Output:
	// Current FSM state: off
	// Toggling...
	// Current FSM state: on
}
