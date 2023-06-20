package openfeature

// Test Utils

type stateHandlerForTests struct {
	initF     func(e EvaluationContext)
	shutdownF func()
	State
}

func (s *stateHandlerForTests) Init(e EvaluationContext) {
	s.initF(e)
}

func (s *stateHandlerForTests) Shutdown() {
	s.shutdownF()
}

func (s *stateHandlerForTests) Status() State {
	return s.State
}
