package main

import (
	"fmt"

	"interrato.dev/fine"
)

func main() {
	turnstile := fine.Machine("locked", fine.States{
		"locked": {
			"pay":  "unlocked",
			"push": nil,
		},
		"unlocked": {
			"pay":  nil,
			"push": "locked",
		},
	})

	unsubscribe := turnstile.Subscribe(func(state string) {
		fmt.Printf("The turnstile is now %s.\n", state)
	})

	turnstile.Do("pay")
	turnstile.Do("push")

	unsubscribe()
}
