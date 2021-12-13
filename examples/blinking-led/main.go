package main

import (
	"fmt"
	"strings"
	"time"

	"interrato.dev/fine"
)

const blinkPace = 100 * time.Millisecond

func main() {
	blink := make(chan bool, 1)

	// Initialize the FSM for the LED.
	led := fine.Machine("off", fine.States{
		"off": {"toggle": "on"},
		"on":  {"toggle": "off"},
	})

	// Add the @enter state for every state of the LED: if blinking is turned
	// on, wait some time, toggle, and feedback the blinking state.
	for _, state := range led.States() {
		led.AddOrMerge(state, fine.Transitions{
			"@enter": func(this *fine.FSM) {
				go func() {
					if feedback := <-blink; feedback {
						time.Sleep(blinkPace)
						this.Do("toggle")
						blink <- feedback
					}
				}()
			},
		})
	}

	// Initialize the switch for toggling the blinking state of the LED.
	blinkSwitch := fine.Machine("off", fine.States{
		"off": {
			"toggle": "on",
		},
		"on": {
			"@enter": func() {
				fmt.Println("Blinking is turned on.")
				go func() {
					blink <- true
				}()
				// Here, do a first manual toggle so that the feedback loop
				// starts.
				led.Do("toggle")
			},
			"@exit": func() {
				blink <- false
				go func() {
					// Here, wait for the last LED toggling before signaling
					// that blinking is now turned off.
					time.Sleep(blinkPace)
					fmt.Println("Blinking is turned off.")
				}()
			},
			"toggle": "off",
		},
	})

	unsubLed := led.Subscribe(func(state string) {
		fmt.Printf("The led is %s\n", strings.ToUpper(state))
	})

	// Toggle the switch three times: blinking on, then off, then on again.
	for i := 0; i < 3; i++ {
		blinkSwitch.Do("toggle")
		<-time.After(2 * time.Second)
	}

	unsubLed()
	fmt.Println("Exiting...")
}
