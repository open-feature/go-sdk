package openfeature

// Test Utils

type StateHandlerForTests struct {
	initF     func(e EvaluationContext)
	shutdownF func()
	State
}

func (s *StateHandlerForTests) Init(e EvaluationContext) {
	s.initF(e)
}

func (s *StateHandlerForTests) Shutdown() {
	s.shutdownF()
}

func (s *StateHandlerForTests) Status() State {
	return s.State
}

// ProviderEventing is an implementation of invoke capability
type ProviderEventing struct {
	c chan Event
}

func (s ProviderEventing) Invoke(e Event) {
	s.c <- e
}

func (s ProviderEventing) EventChannel() <-chan Event {
	return s.c
}
