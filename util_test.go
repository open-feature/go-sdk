package openfeature

// Test Utils

// event handlers
var h1 func(details EventDetails)
var h2 func(details EventDetails)
var h3 func(details EventDetails)
var h4 func(details EventDetails)

func init() {
	h1 = func(details EventDetails) {
		// noop
	}

	h2 = func(details EventDetails) {
		// noop
	}

	h3 = func(details EventDetails) {
		// noop
	}

	h4 = func(details EventDetails) {
		// noop
	}
}

// stateHandlerForTests is a StateHandler with callbacks
type stateHandlerForTests struct {
	initF     func(e EvaluationContext) error
	shutdownF func()
	State
}

func (s *stateHandlerForTests) Init(e EvaluationContext) error {
	if s.initF != nil {
		return s.initF(e)
	}
	return nil
}

func (s *stateHandlerForTests) Shutdown() {
	if s.shutdownF != nil {
		s.shutdownF()
	}
}

func (s *stateHandlerForTests) Status() State {
	return s.State
}

// ProviderEventing is an eventing implementation with invoke capability
type ProviderEventing struct {
	c chan Event
}

func (s ProviderEventing) Invoke(e Event) {
	s.c <- e
}

func (s ProviderEventing) EventChannel() <-chan Event {
	return s.c
}
