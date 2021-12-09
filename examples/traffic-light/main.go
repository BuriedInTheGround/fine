package main

import (
	"fmt"
	"time"

	"interrato.dev/fine"
)

func main() {
	// Initialize the traffic light FSM. At every @enter, a new goroutine
	// starts, waits for some time, and finally requests the execution of the
	// change action.
	trafficLight := fine.Machine("red", fine.States{
		"green": {
			"@enter": func(this *fine.FSM) {
				go func() {
					time.Sleep(14 * time.Second)
					this.Do("change")
				}()
			},
			"change": "yellow",
		},
		"yellow": {
			"@enter": func(this *fine.FSM) {
				go func() {
					time.Sleep(3 * time.Second)
					this.Do("change")
				}()
			},
			"change": "red",
		},
		"red": {
			"@enter": func(this *fine.FSM) {
				go func() {
					time.Sleep(12 * time.Second)
					this.Do("change")
				}()
			},
			"change": "green",
		},
	})

	// Subscribe to the FSM, and every time the state changes print the
	// information on screen.
	t := time.Now()
	unsubscribe := trafficLight.Subscribe(func(state string) {
		fmt.Printf(
			"The %q light is now turned on. (took ~%.3fs)\n",
			state, time.Since(t).Seconds(),
		)
		t = time.Now()
	})

	// Unsubscribe after 40 seconds. Note that doing this also blocks the main
	// goroutine.
	<-time.After(40 * time.Second)
	unsubscribe()
	fmt.Println("unsubscribed")

	// Output:
	// The "red" light is now turned on. (took ~0.000s)
	// The "green" light is now turned on. (took ~12.011s)
	// The "yellow" light is now turned on. (took ~14.013s)
	// The "red" light is now turned on. (took ~3.000s)
	// unsubscribed
}
