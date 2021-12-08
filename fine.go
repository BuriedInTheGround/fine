package fine

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
)

// Transitions is a mapping between events (names of actions) and actions.
//
// An action can have one of the following types, or nil.
//
//     string
//
//     func()
//
//     func(...interface{})
//
//     func() string
//
//     func(...interface{}) string
//
// Trying to call an action that has a different type will panic.
//
// There are two special lifecycle functions, named "@enter" and "@exit",
// executed on entering and exiting a state, respectively. It is not possible
// to pass custom parameters to these functions. They receive four metadata
// arguments:
//
//     from  (string): the previous state from which the transition started
//     to    (string): the new state where the transition will end
//     event (string): the name of the action that caused the transition
//     args  ([]interface{}): the arguments that were passed to the action
type Transitions map[string]interface{}

// States are mappings from states to Transitions.
//
// A state has type string.
type States map[string]Transitions

// FSM is a Finite State Machine that can be instantiated using the Machine
// function.
type FSM struct {
	current string
	states  States
}

var (
	mu sync.RWMutex

	lastSubKey  int32
	subscribers map[int32]func(string)
)

func init() {
	atomic.StoreInt32(&lastSubKey, 0)

	mu.Lock()
	subscribers = make(map[int32]func(string))
	mu.Unlock()
}

// Machine instatiate a new FSM.
//
// Note that the given initial state must exist inside states.
func Machine(initialState string, states States) *FSM {
	// Check for the initial state being present.
	if _, ok := states[initialState]; !ok {
		panic("the initial state must exist")
	}

	// Instantiate the FSM object.
	m := &FSM{
		current: initialState,
		states:  states,
	}

	// Execute the first @enter lifecycle action on the initial state.
	m.do("@enter", nil, m.current, nil, nil)

	return m
}

// State returns the current state of the FSM.
func (m *FSM) State() string {
	mu.RLock()
	defer mu.RUnlock()

	return m.current
}

// States returns a slice with all the possible states of the FSM.
//
// Note that the order is not guaranteed.
func (m *FSM) States() []string {
	var states []string

	mu.RLock()
	for state := range m.states {
		states = append(states, state)
	}
	mu.RUnlock()

	return states
}

// Add allows to add a new state with its associated transitions. If a state
// with the same name is already present in the FSM a non-nil error is
// returned.
func (m *FSM) Add(state string, transitions Transitions) error {
	if m.Exists(state) {
		return fmt.Errorf("a state with name %q already exists", state)
	}

	mu.Lock()
	m.states[state] = transitions
	mu.Unlock()

	return nil
}

// AddOrReplace allows to add a new state with its associated transitions. If a
// state with the same name is already present in the FSM, its transitions will
// be completely overwritten.
func (m *FSM) AddOrReplace(state string, transitions Transitions) {
	mu.Lock()
	m.states[state] = transitions
	mu.Unlock()
}

// AddOrMerge allows to add a new state with its associated transitions. If a
// state with the same name is already present in the FSM, its transitions will
// be merged, keeping the newer ones in case of collisions.
func (m *FSM) AddOrMerge(state string, transitions Transitions) {
	if m.Exists(state) {
		mu.Lock()
		for k, v := range transitions {
			m.states[state][k] = v
		}
		mu.Unlock()
	} else {
		mu.Lock()
		m.states[state] = transitions
		mu.Unlock()
	}
}

// Exists returns whether the specified state is a possible state for the FSM.
func (m *FSM) Exists(state string) bool {
	mu.RLock()
	_, ok := m.states[state]
	mu.RUnlock()

	return ok
}

// Do executes the specified action on the FSM from the current state.
//
// The action parameter specifies the event, that is, the action name.
//
// It is possible to pass arguments to the action. If the action isn't a
// function or does not accept any parameter, the arguments will be ignored.
//
// Note that lifecycle actions cannot be manually executed.
func (m *FSM) Do(action string, args ...interface{}) (string, error) {
	// Prohibit the execution of lifecycle actions.
	if action == "@enter" || action == "@exit" {
		return "", errors.New("calling a lifecycle action manually is illegal")
	}

	// Check for the existence of the requested action.
	mu.RLock()
	if _, ok := m.states[m.current][action]; !ok {
		defer mu.RUnlock()
		return "", fmt.Errorf(
			"%q is not a valid action for the current state %q",
			action, m.current,
		)
	}

	// Execute the action, and evaluate what the new state will be.
	newState := m.do(action, args...)
	mu.RUnlock()

	mu.Lock()
	defer mu.Unlock()

	// If doing the action changes the state, execute the transition.
	if newState != m.current {
		// Metadata contains the following four parts, indicated with their
		// type.
		//
		// [from: string, to: string, action: string, args: []interface{}]
		metadata := []interface{}{m.current, newState, action, args}

		// Do the transition: execute the @exit lifecycle action, then update
		// the current state, and finally execute the @enter lifecycle action.
		m.do("@exit", metadata...)
		m.current = newState
		for _, callback := range subscribers {
			callback(m.current)
		}
		m.do("@enter", metadata...)
	}

	return m.current, nil
}

func (m *FSM) do(action string, args ...interface{}) string {
	// Execute the action based on the action type.
	switch next := m.states[m.current][action].(type) {
	case nil:
		return m.current

	case string:
		return next

	case func():
		next()

	case func(...interface{}):
		next(args...)

	case func() string:
		return next()

	case func(...interface{}) string:
		return next(args...)

	default:
		panic(fmt.Sprintf(
			"invalid type for action %q on state %q", action, m.current,
		))
	}

	return m.current
}

// Subscribe allows subscribing to state changes with a callback function. The
// FSM will call the callback function with the new state as a parameter every
// time the state changes.
//
// An unsubscribe function is returned.
func (m *FSM) Subscribe(callback func(state string)) func() {
	key := atomic.AddInt32(&lastSubKey, 1)

	mu.Lock()
	subscribers[key] = callback
	callback(m.current)
	mu.Unlock()

	return func() {
		mu.Lock()
		delete(subscribers, key)
		mu.Unlock()
	}
}
