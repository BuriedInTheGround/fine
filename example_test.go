package fine_test

import (
	"fmt"

	"interrato.dev/fine"
)

func ExampleMachine() {
	powerSwitch := fine.Machine("off", fine.States{
		"off": {"toggle": "on"},
		"on":  {"toggle": "off"},
	})

	fmt.Println("Current FSM state:", powerSwitch.State())
	// Output:
	// Current FSM state: off
}

func ExampleFSM_Add() {
	powerSwitch := fine.Machine("off", fine.States{
		"off": {"toggle": "on"},
	})

	// Here I add the "on" state.
	powerSwitch.Add("on", fine.Transitions{"toggle": "off"})

	// Trying to add the "off" state, which is already in the FSM, will fail.
	err := powerSwitch.Add("off", fine.Transitions{"smash": "broken"})
	fmt.Println("Adding \"off\" failed?", err != nil)

	powerSwitch.Do("toggle")
	fmt.Println("Current FSM state:", powerSwitch.State())
	// Output:
	// Adding "off" failed? true
	// Current FSM state: on
}

func ExampleFSM_AddOrReplace() {
	powerSwitch := fine.Machine("off", fine.States{
		"off": {"toggle": "on"},
	})

	// Here I add the "on" state. No difference with Add() here.
	powerSwitch.AddOrReplace("on", fine.Transitions{"toggle": "off"})

	// Here I try to add the "off" state, but, because it's already in the FSM,
	// its transitions are now replaced.
	powerSwitch.AddOrReplace("off", fine.Transitions{"smash": "broken"})

	// The "toggle" event for the "off" state does not exist anymore, so
	// nothing changes even if I try to invoke "toggle" many times.
	powerSwitch.Do("toggle")
	fmt.Println("Current FSM state:", powerSwitch.State())
	powerSwitch.Do("toggle")
	fmt.Println("Current FSM state still", powerSwitch.State())
	powerSwitch.Do("toggle")
	fmt.Println("Current FSM state is", powerSwitch.State(), "again")

	// So, let's break this switch.
	powerSwitch.Do("smash")
	fmt.Println("Current FSM state:", powerSwitch.State())
	// Output:
	// Current FSM state: off
	// Current FSM state still off
	// Current FSM state is off again
	// Current FSM state: broken
}

func ExampleFSM_AddOrMerge() {
	powerSwitch := fine.Machine("off", fine.States{
		"off": {"toggle": "on"},
	})

	// Here I add the "on" state. No difference with Add() here.
	powerSwitch.AddOrMerge("on", fine.Transitions{"toggle": "off"})

	// Here I try to add the "off" state, which is already in the FSM, and the
	// new transitions are merged to the old ones.
	powerSwitch.AddOrMerge("off", fine.Transitions{"smash": "broken"})

	// So I can still toggle it.
	powerSwitch.Do("toggle")
	fmt.Println("Current FSM state:", powerSwitch.State())

	// And toggle it again...
	powerSwitch.Do("toggle")
	fmt.Println("Current FSM state:", powerSwitch.State())

	// ...and now I smash it.
	powerSwitch.Do("smash")
	fmt.Println("Current FSM state:", powerSwitch.State())
	// Output:
	// Current FSM state: on
	// Current FSM state: off
	// Current FSM state: broken
}

func ExampleFSM_Do() {
	powerSwitch := fine.Machine("off", fine.States{
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
	powerSwitch := fine.Machine("off", fine.States{
		"off": {
			"@exit": func(metadata fine.Metadata) {
				fmt.Println("[INFO] from:", metadata.From)
				fmt.Println("[INFO] to:", metadata.To)
				fmt.Println("[INFO] event:", metadata.Event)
				fmt.Println("[INFO] args:", metadata.Args)
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
	// Output:
	// Current FSM state: off
	// Toggling...
	// message: "Shine, step into the light"
	// [INFO] from: off
	// [INFO] to: on
	// [INFO] event: toggle
	// [INFO] args: [Shine, step into the light]
	// Finally, light!
	// Current FSM state: on
}

func ExampleFSM_Subscribe() {
	powerSwitch := fine.Machine("off", fine.States{
		"off": {"toggle": "on"},
		"on":  {"toggle": "off"},
	})

	onSubscribe := true
	unsubscribe := powerSwitch.Subscribe(func(state string) {
		if onSubscribe {
			fmt.Printf("Subscribed with state set to %q\n", state)
			onSubscribe = false
		} else {
			fmt.Printf("State just changed to %q\n", state)
		}
	})

	powerSwitch.Do("toggle")
	powerSwitch.Do("toggle")
	powerSwitch.Do("toggle")

	unsubscribe()

	powerSwitch.Do("toggle")
	fmt.Println("Current FSM state:", powerSwitch.State())
	// Output:
	// Subscribed with state set to "off"
	// State just changed to "on"
	// State just changed to "off"
	// State just changed to "on"
	// Current FSM state: off
}
