package openfeature

type API interface {
	SetProvider(provider FeatureProvider, async bool) error
	SetNamedProvider(clientDomain string, provider FeatureProvider, async bool) error
	SetEvaluationContext(apiCtx EvaluationContext)
	AddHooks(hooks ...Hook)
	Shutdown()
}

type Eventing interface {
	APIEvent
	ClientEvent
}

type APIEvent interface {
	AddHandler(eventType EventType, callback EventCallback)
	RemoveHandler(eventType EventType, callback EventCallback)
}

type ClientEvent interface {
	RegisterClientHandler(clientDomain string, t EventType, c EventCallback)
	RemoveClientHandler(name string, t EventType, c EventCallback)
}
