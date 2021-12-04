package fike_test

import (
	"fmt"

	"interrato.dev/fike"
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

func ExampleFSM_Do_lifecycle() {
	powerSwitch := fike.Machine("off", fike.States{
		"off": {
			"@enter": func(metadata ...interface{}) {
				fmt.Println("[INFO] from:", metadata[0])
				fmt.Println("[INFO] to:", metadata[1])
				fmt.Println("[INFO] event:", metadata[2])
				fmt.Println("[INFO] args:", metadata[3])
			},
			"@exit": func(metadata ...interface{}) {
				fmt.Println("[INFO] from:", metadata[0])
				fmt.Println("[INFO] to:", metadata[1])
				fmt.Println("[INFO] event:", metadata[2])
				fmt.Println("[INFO] args:", metadata[3])
			},
			"toggle": func(args ...interface{}) string {
				message := args[0].(string)
				fmt.Printf("message: %q\n", message)
				return "on"
			},
		},
		"on": {
			"@enter": func() {
				fmt.Println("Finally, light!")
			},
			"toggle": "off",
		},
	})

	fmt.Println("Current FSM state:", powerSwitch.State())
	fmt.Println("Toggling...")
	powerSwitch.Do("toggle", "Shine, step into the light")
	fmt.Println("Current FSM state:", powerSwitch.State())
	powerSwitch.Do("toggle")
	fmt.Println("Current FSM state:", powerSwitch.State())
	// Output:
	// [INFO] from: <nil>
	// [INFO] to: off
	// [INFO] event: <nil>
	// [INFO] args: <nil>
	// Current FSM state: off
	// Toggling...
	// message: "Shine, step into the light"
	// [INFO] from: off
	// [INFO] to: on
	// [INFO] event: toggle
	// [INFO] args: [Shine, step into the light]
	// Finally, light!
	// Current FSM state: on
	// [INFO] from: on
	// [INFO] to: off
	// [INFO] event: toggle
	// [INFO] args: []
	// Current FSM state: off
}
