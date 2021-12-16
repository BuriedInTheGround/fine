package fine_test

import (
	"math/rand"
	"strconv"
	"sync"
	"testing"
	"time"

	"interrato.dev/fine"
)

const concurrentRuns = 200

func init() {
	rand.Seed(time.Now().UnixNano())
}

func TestState(t *testing.T) {
	machine := fine.Machine("a", fine.States{
		"a": {
			"next": "b",
		},
		"b": {
			"next": "c",
		},
		"c": {
			"next": "a",
		},
	})

	// Test that State() is always right: both after initializing the machine
	// and after a state change.
	for _, cur := range []string{"a", "b", "c", "a"} {
		if state := machine.State(); state != cur {
			t.Fatalf("wrong state: got %q, want %q", state, cur)
		}
		machine.Do("next")
	}

	// Concurrency test (run with `-race`).
	for i := 0; i < concurrentRuns; i++ {
		go func() string {
			return machine.State()
		}()
	}
}

func TestStates(t *testing.T) {
	states := fine.States{
		"a": {},
		"b": {},
		"c": {},
	}
	machine := fine.Machine("a", states)

	// Test that all and only the states that actually are in the FSM are
	// returned by States().
	got := machine.States()
	if len(got) != len(states) {
		t.Fatalf("wrong number of states: got %v, want %v", len(got), len(states))
	}
	for _, state := range got {
		if _, ok := states[state]; !ok {
			t.Fatalf("wrong state %q found", state)
		}
	}

	// Concurrency test (run with `-race`).
	for i := 0; i < concurrentRuns; i++ {
		go func() []string {
			return machine.States()
		}()
	}
}

func TestAdd(t *testing.T) {
	machine := fine.Machine("a", fine.States{
		"a": {
			"next": "b",
		},
		"b": {
			"next": "c",
		},
	})

	// This first call to Add() must return without errors.
	err := machine.Add("c", fine.Transitions{"next": "a"})
	if err != nil {
		t.Fatalf("no error expected, got: %v", err)
	}

	// This second call to Add() must return an error because of the duplicate
	// key (state).
	err = machine.Add("c", fine.Transitions{})
	if err == nil {
		t.Fatal("error expected, got <nil>")
	}

	// Concurrency test (run with `-race`).
	for i := 0; i < concurrentRuns; i++ {
		go func() {
			state := strconv.FormatInt(rand.Int63(), 10)
			machine.Add(state, fine.Transitions{})
		}()
	}
}

func TestAddOrReplace(t *testing.T) {
	machine := fine.Machine("a", fine.States{
		"a": {
			"next": "b",
		},
		"b": {
			"next": "c",
		},
	})

	// Test that adding with AddOrReplace() works correctly.
	next := make(chan string, 1)
	next <- "a"
	machine.AddOrReplace("c", fine.Transitions{
		"@enter": func() {
			<-next // Discard "a".
			go func() {
				next <- "b"
			}()
		},
		"next": func() string {
			return <-next
		},
	})
	machine.Do("next") // "a" -> "b" (next: ["a"])
	machine.Do("next") // "b" -> "c" (next: ["b"])
	machine.Do("next") // "c" -> "b" (next: [])
	if state := machine.State(); state != "b" {
		t.Fatalf("wrong state: got %q, want %q", state, "b")
	}

	// Test that replacing with AddOrReplace() works correctly.
	machine.AddOrReplace("c", fine.Transitions{"next": "a"})
	machine.Do("next") // "b" -> "c"
	machine.Do("next") // "c" -> "a"
	machine.Do("next") // "a" -> "b"
	if state := machine.State(); state != "b" {
		t.Fatalf("wrong state: got %q, want %q", state, "b")
	}

	// Concurrency test (run with `-race`).
	for i := 0; i < concurrentRuns; i++ {
		go func() {
			state := strconv.FormatInt(rand.Int63n(10), 10)
			machine.AddOrReplace(state, fine.Transitions{})
		}()
	}
}

func TestAddOrMerge(t *testing.T) {
	machine := fine.Machine("a", fine.States{
		"a": {
			"next": "b",
		},
		"b": {
			"next": "c",
		},
	})

	// Test that adding with AddOrMerge() works correctly.
	next := make(chan string, 1)
	next <- "a"
	machine.AddOrMerge("c", fine.Transitions{
		"@enter": func() {
			<-next // Discard "a".
			go func() {
				next <- "b"
			}()
		},
		"next": func() string {
			return <-next
		},
	})
	machine.Do("next") // "a" -> "b" (next: ["a"])
	machine.Do("next") // "b" -> "c" (next: ["a"] -> [] -> ["b"])
	machine.Do("next") // "c" -> "b" (next: ["b"] -> [])
	if state := machine.State(); state != "b" {
		t.Fatalf("wrong state: got %q, want %q", state, "b")
	}

	// Test that merging with AddOrMerge() works correctly.
	machine.AddOrMerge("c", fine.Transitions{
		"@enter": func() {
			go func() {
				next <- "a"
			}()
		},
	})
	machine.Do("next") // "b" -> "c" (next: [] -> ["a"])
	machine.Do("next") // "c" -> "a" (next: ["a"] -> [])
	machine.Do("next") // "a" -> "b" (next: [])
	if state := machine.State(); state != "b" {
		t.Fatalf("wrong state: got %q, want %q", state, "b")
	}

	// Concurrency test (run with `-race`).
	for i := 0; i < concurrentRuns; i++ {
		go func() {
			state := strconv.FormatInt(rand.Int63n(10), 10)
			machine.AddOrMerge(state, fine.Transitions{})
		}()
	}
}

func TestExists(t *testing.T) {
	machine := fine.Machine("a", fine.States{
		"a": {
			"next": "b",
		},
		"b": {
			"next": "c",
		},
		"c": {
			"next": "a",
		},
	})

	// Test that Exists() is right.
	for _, state := range machine.States() {
		if exists := machine.Exists(state); !exists {
			t.Fatalf("state %q exists", state)
		}
	}

	// Test that random states correctly do not exist.
	for base := 2; base <= 36; base++ {
		for i := 0; i < 20; i++ {
			state := strconv.FormatInt(rand.Int63n(256), base)
			// Skip obvious cases: we just want to make sure that nothing weird
			// may happen.
			if state == "a" || state == "b" || state == "c" {
				continue
			}
			if exists := machine.Exists(state); exists {
				t.Fatalf("state %q does not exist", state)
			}
		}
	}

	// Concurrency test (run with `-race`).
	for i := 0; i < concurrentRuns; i++ {
		go func() bool {
			state := strconv.FormatInt(rand.Int63(), 10)
			return machine.Exists(state)
		}()
	}
}

func TestDo(t *testing.T) {
	machine := fine.Machine("a", fine.States{
		"a": {
			"next": "b",
		},
		"b": {
			"next": "c",
		},
		"c": {
			"next": "a",
		},
	})

	// Test that Do() works correctly.
	for i := 0; i < 3; i++ {
		currentState, err := machine.Do("next")
		if err != nil {
			t.Fatalf("no error expected, got: %v", err)
		}
		if cur := machine.State(); currentState != cur {
			t.Fatalf("wrong state: got %q, want %q", currentState, cur)
		}
	}
	var err error
	_, err = machine.Do("non-existent-event")
	if err == nil {
		t.Fatal("error expected, got <nil>")
	}
	_, err = machine.Do("@enter")
	if err == nil {
		t.Fatal("error expected, got <nil>")
	}
	_, err = machine.Do("@exit")
	if err == nil {
		t.Fatal("error expected, got <nil>")
	}

	// Concurrency test (run with `-race`).
	for i := 0; i < concurrentRuns; i++ {
		go func() (string, error) {
			return machine.Do("next")
		}()
	}
}

func TestSubscribe(t *testing.T) {
	machine := fine.Machine("a", fine.States{
		"a": {
			"next": "b",
		},
		"b": {
			"next": "c",
		},
		"c": {
			"next": "a",
		},
	})

	// Test that all Subscribe() side effects are correct.
	history := make(chan string, 1)
	unsubscribe := machine.Subscribe(func(state string) {
		history <- state
	})
	if lastChange := <-history; lastChange != machine.State() {
		t.Fatalf("wrong state: got %q, want %q", lastChange, machine.State())
	}
	machine.Do("next")
	if lastChange := <-history; lastChange != machine.State() {
		t.Fatalf("wrong state: got %q, want %q", lastChange, machine.State())
	}
	unsubscribe()
	machine.Do("next")
	select {
	case lastChange := <-history:
		t.Fatalf("got unexpected state %q", lastChange)
	default:
		// This is the correct case: nothing must be received after the
		// unsubscribe.
	}

	// Concurrency test (run with `-race`).
	var wg sync.WaitGroup
	go func() {
		for {
			time.Sleep(time.Duration(rand.Intn(50)) * time.Microsecond)
			machine.Do("next")
		}
	}()
	for i := 0; i < concurrentRuns; i++ {
		wg.Add(1)
		go func() {
			unsubscribe := machine.Subscribe(func(state string) {
				_ = state
			})
			go func() {
				time.Sleep(time.Duration(rand.Intn(200)) * time.Microsecond)
				unsubscribe()
				wg.Done()
			}()
		}()
	}
	wg.Wait()
}
