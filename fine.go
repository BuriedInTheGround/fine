package fine

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
)

// Transitions is a mapping between events (or names of actions) and actions.
//
// An action can have one of the following types, or nil.
//
//     string
//     func() string
//     func(args ...interface{}) string
//     func()
//     func(args ...interface{})
//
// Trying to call an action that has a different type will panic.
//
// There are two special lifecycle functions, named "@enter" and "@exit",
// executed on entering and exiting a state, respectively. It is not possible
// to pass custom parameters to these functions. They receive an optional
// Metadata object and an optional pointer to the FSM itself. Thus, the
// possible types for lifecycle actions are the following, or nil.
//
//     func()
//     func(this *fine.FSM)
//     func(metadata fine.Metadata)
//     func(this *fine.FSM, metadata fine.Metadata)
type Transitions map[string]interface{}

// States are mappings from states to Transitions.
//
// A state has type string.
type States map[string]Transitions

// Metadata holds the information about a transition that changed the system
// state.
type Metadata struct {
	// The previous state from which the transition started.
	From string

	// The new state where the transition will end.
	To string

	// The name of the action that caused the system state change.
	Event string

	// The arguments that were passed to the action.
	Args []interface{}
}

// FSM is a finite-state machine that can be instantiated using the Machine
// function.
type FSM struct {
	current string
	states  States

	mu sync.RWMutex

	lastSubKey  int32
	subscribers map[int32]func(string)
}

// Machine instatiate a new FSM with the given initial state and the given set
// of possible states.
//
// Note: the given initial state must be within the given possible states.
func Machine(initialState string, states States) *FSM {
	// Check for the initial state being present.
	if _, ok := states[initialState]; !ok {
		panic("the initial state must exist")
	}

	// Instantiate the FSM object.
	m := &FSM{
		current:     initialState,
		states:      states,
		subscribers: make(map[int32]func(string)),
	}

	// Initialize the last subscriber key to zero.
	atomic.StoreInt32(&m.lastSubKey, 0)

	// Execute the first @enter lifecycle action on the initial state.
	m.doLifecycle("@enter", Metadata{To: m.current})

	return m
}

// State returns the current state of the FSM.
func (m *FSM) State() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.current
}

// States returns a slice with all the possible states of the FSM.
//
// Note: the order is not guaranteed.
func (m *FSM) States() []string {
	var states []string

	m.mu.RLock()
	for state := range m.states {
		states = append(states, state)
	}
	m.mu.RUnlock()

	return states
}

// Add allows to add a new state with its associated transitions. If a state
// with the same name is already present in the FSM a non-nil error is
// returned.
func (m *FSM) Add(state string, transitions Transitions) error {
	if m.Exists(state) {
		return fmt.Errorf("a state with name %q already exists", state)
	}

	m.mu.Lock()
	m.states[state] = transitions
	m.mu.Unlock()

	return nil
}

// AddOrReplace allows to add a new state with its associated transitions. If a
// state with the same name is already present in the FSM, its transitions will
// be completely overwritten.
func (m *FSM) AddOrReplace(state string, transitions Transitions) {
	m.mu.Lock()
	m.states[state] = transitions
	m.mu.Unlock()
}

// AddOrMerge allows to add a new state with its associated transitions. If a
// state with the same name is already present in the FSM, its transitions will
// be merged, keeping the newer ones in case of collisions.
func (m *FSM) AddOrMerge(state string, transitions Transitions) {
	if m.Exists(state) {
		m.mu.Lock()
		for k, v := range transitions {
			m.states[state][k] = v
		}
		m.mu.Unlock()
	} else {
		m.mu.Lock()
		m.states[state] = transitions
		m.mu.Unlock()
	}
}

// Exists returns whether the specified state is a possible state for the FSM.
func (m *FSM) Exists(state string) bool {
	m.mu.RLock()
	_, ok := m.states[state]
	m.mu.RUnlock()

	return ok
}

// Do executes the specified action on the FSM from the current state.
//
// The action parameter specifies the event, that is, the action name.
//
// It is possible to pass arguments to the action. If the action isn't a
// function or does not accept any parameter, the arguments will be ignored.
//
// Note: lifecycle actions cannot be manually executed.
func (m *FSM) Do(action string, args ...interface{}) (string, error) {
	// Prohibit the execution of lifecycle actions.
	if action == "@enter" || action == "@exit" {
		return "", errors.New("calling a lifecycle action manually is illegal")
	}

	// Check for the existence of the requested action.
	m.mu.RLock()
	if _, ok := m.states[m.current][action]; !ok {
		defer m.mu.RUnlock()
		return "", fmt.Errorf(
			"%q is not a valid action for the current state %q",
			action, m.current,
		)
	}
	m.mu.RUnlock()

	// Execute the action, and evaluate what the new state will be.
	newState := m.do(action, args...)

	// Evaluate if the action changed the state.
	var stateChanged bool
	m.mu.RLock()
	if newState != m.current {
		stateChanged = true
	}
	m.mu.RUnlock()

	// If the state changed, execute the state transition.
	if stateChanged {
		m.mu.RLock()
		metadata := Metadata{
			From:  m.current,
			To:    newState,
			Event: action,
			Args:  args,
		}
		m.mu.RUnlock()

		// Execute the @exit lifecycle action.
		m.doLifecycle("@exit", metadata)

		// Update the current state.
		m.mu.Lock()
		m.current = newState
		m.mu.Unlock()

		// Notify the state change to all subscribers.
		m.mu.RLock()
		for _, callback := range m.subscribers {
			callback(newState)
		}
		m.mu.RUnlock()

		// And finally, execute the @enter lifecycle action.
		m.doLifecycle("@enter", metadata)
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.current, nil
}

func (m *FSM) do(action string, args ...interface{}) string {
	// Execute the action based on the action type.
	m.mu.RLock()
	switch next := m.states[m.current][action].(type) {
	case nil:
		defer m.mu.RUnlock()
		return m.current

	case string:
		m.mu.RUnlock()
		return next

	case func():
		m.mu.RUnlock()
		next()

	case func(...interface{}):
		m.mu.RUnlock()
		next(args...)

	case func() string:
		m.mu.RUnlock()
		return next()

	case func(...interface{}) string:
		m.mu.RUnlock()
		return next(args...)

	default:
		panic(fmt.Sprintf(
			"invalid type for action %q on state %q", action, m.current,
		))
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.current
}

func (m *FSM) doLifecycle(action string, metadata Metadata) {
	// Execute the action based on the action type.
	m.mu.RLock()
	switch lifecycle := m.states[m.current][action].(type) {
	case nil:
		m.mu.RUnlock()
		return

	case func():
		m.mu.RUnlock()
		lifecycle()

	case func(*FSM):
		m.mu.RUnlock()
		lifecycle(m)

	case func(Metadata):
		m.mu.RUnlock()
		lifecycle(metadata)

	case func(*FSM, Metadata):
		m.mu.RUnlock()
		lifecycle(m, metadata)

	default:
		panic(fmt.Sprintf(
			"invalid type for action %q on state %q", action, m.current,
		))
	}
}

// Subscribe allows subscribing to state changes with a callback function. The
// callback function will be executed every time the state changes and receives
// the new state as a parameter. The callback function also runs when
// subscribing and will receive the current state.
//
// An unsubscribe function is returned.
func (m *FSM) Subscribe(callback func(state string)) func() {
	key := atomic.AddInt32(&m.lastSubKey, 1)

	m.mu.Lock()
	m.subscribers[key] = callback
	callback(m.current)
	m.mu.Unlock()

	return func() {
		m.mu.Lock()
		delete(m.subscribers, key)
		m.mu.Unlock()
	}
}
