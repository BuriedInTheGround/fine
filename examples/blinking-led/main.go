package main

import (
	"fmt"
	"strings"
	"time"

	"interrato.dev/fine"
)

const blinkPace = 300 * time.Millisecond

func main() {
	blink := make(chan bool, 1)

	// Initialize the FSM for the LED.
	led := fine.Machine("off", fine.States{
		"off": {"toggle": "on"},
		"on":  {"toggle": "off"},
	})

	// Add the @enter state for every state of the LED: wait some time, then,
	// if blinking is turned on, toggle and feedback the blinking state.
	for _, state := range led.States() {
		led.AddOrMerge(state, fine.Transitions{
			"@enter": func(this *fine.FSM) {
				go func() {
					time.Sleep(blinkPace)
					if feedback := <-blink; feedback {
						this.Do("toggle")
						blink <- feedback
					}
				}()
			},
		})
	}

	// Initialize the switch for toggling the blinking state of the LED.
	blinkSwitch := fine.Machine("off", fine.States{
		"off": {"toggle": "on"},
		"on": {
			"@enter": func() {
				// First, toggle manually so that the feedback loop starts.
				led.Do("toggle")
				// Then, activate the blinking.
				blink <- true
			},
			"@exit": func() {
				// Disactivate the blinking.
				blink <- false
			},
			"toggle": "off",
		},
	})

	unsubLed := led.Subscribe(func(state string) {
		fmt.Printf("The led is %s\n", strings.ToUpper(state))
	})

	unsubSwitch := blinkSwitch.Subscribe(func(state string) {
		fmt.Printf("Blinking is now turned %s.\n", state)
	})

	// Toggle the switch three times: blinking on, then off, then on again.
	for i := 0; i < 3; i++ {
		blinkSwitch.Do("toggle")
		<-time.After(2 * time.Second)
	}

	unsubLed()
	unsubSwitch()
	fmt.Println("Exiting...")

	// Output:
	// The led is OFF
	// Blinking is now turned off.
	// Blinking is now turned on.
	// The led is ON
	// The led is OFF
	// The led is ON
	// The led is OFF
	// The led is ON
	// The led is OFF
	// The led is ON
	// The led is OFF
	// Blinking is now turned off.
	// Blinking is now turned on.
	// The led is ON
	// The led is OFF
	// The led is ON
	// The led is OFF
	// The led is ON
	// The led is OFF
	// The led is ON
	// The led is OFF
	// Exiting...
}
