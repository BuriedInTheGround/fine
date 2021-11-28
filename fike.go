package fike

import (
	"fmt"
)

// Transitions is a mapping between action names and actual actions.
//
// An action can have one of the following types, or nil.
//
//     string
//
//     func()

//     func(...interface{})
//
//     func() string

//     func(...interface{}) string
//
// Trying to call an action that has a different type will panic.
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

// Machine instatiate a new FSM.
func Machine(initialState string, states States) *FSM {
	if _, ok := states[initialState]; !ok {
		panic("the initial state must exist")
	}
	m := &FSM{
		current: initialState,
		states:  states,
	}

	// TODO: execute the @enter lifecycle action

	return m
}

// State returns the current state of the FSM.
func (m *FSM) State() string {
	return m.current
}

// AddOrReplace allows to add a new state with its associated transitions. If a
// state with the same name is already present in the FSM a non-nil error is
// returned.
func (m *FSM) Add(state string, transitions Transitions) error {
	if m.Exists(state) {
		return fmt.Errorf("a state with name %q already exists", state)
	}
	m.states[state] = transitions
	return nil
}

// AddOrReplace allows to add a new state with its associated transitions. If a
// state with the same name is already present in the FSM, its transitions will
// be completely overwritten.
func (m *FSM) AddOrReplace(state string, transitions Transitions) {
	m.states[state] = transitions
}

// AddOrMerge allows to add a new state with its associated transitions. If a
// state with the same name is already present in the FSM, its transitions will
// be merged, keeping the newer ones in case of collisions.
func (m *FSM) AddOrMerge(state string, transitions Transitions) {
	if m.Exists(state) {
		for k, v := range transitions {
			m.states[state][k] = v
		}
	} else {
		m.states[state] = transitions
	}
}

// Exists returns whether the specified state is a possible state for the FSM.
func (m *FSM) Exists(state string) bool {
	_, ok := m.states[state]
	return ok
}

// Do executes the specified action on the FSM from the current state.
//
// It is possible to pass arguments to the action. If the action isn't a
// function, the arguments will be ignored.
func (m *FSM) Do(action string, args ...interface{}) (string, error) {
	do, ok := m.states[m.current][action]
	if !ok {
		return "", fmt.Errorf(
			"%q is not a valid action for the current state %q",
			action, m.current,
		)
	}

	if do == nil {
		return m.current, nil
	}

	previous := m.current

	switch next := do.(type) {
	case string:
		// TODO: execute the @exit lifecycle action
		m.current = next

	case func():
		next()

	case func(...interface{}):
		next(args...)

	case func() string:
		// TODO: execute the @exit lifecycle action
		m.current = next()

	case func(...interface{}) string:
		// TODO: execute the @exit lifecycle action
		m.current = next(args...)

	default:
		panic(fmt.Sprintf(
			"invalid type for action %q on state %q", action, m.current,
		))
	}

	if previous != m.current {
		// TODO: execute the @enter lifecycle action
	}

	return m.current, nil
}
