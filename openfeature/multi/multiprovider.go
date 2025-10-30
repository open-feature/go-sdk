// Package multi is an experimental implementation of a [of.FeatureProvider] that supports evaluating multiple feature flag
// providers together.
package multi

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"strings"
	"sync"

	of "github.com/open-feature/go-sdk/openfeature"
	"golang.org/x/sync/errgroup"
)

// Metadata Keys
const (
	MetadataProviderName                   = "multiprovider-provider-name"
	MetadataProviderType                   = "multiprovider-provider-type"
	MetadataSuccessfulProviderName         = "multiprovider-successful-provider-name"
	MetadataSuccessfulProviderNames        = MetadataSuccessfulProviderName + "s"
	MetadataStrategyUsed                   = "multiprovider-strategy-used"
	MetadataFallbackUsed                   = "multiprovider-fallback-used"
	MetadataIsDefaultValue                 = "multiprovider-is-result-default-value"
	MetadataEvaluationError                = "multiprovider-evaluation-error"
	MetadataComparisonDisagreeingProviders = "multiprovider-comparison-disagreeing-providers"
)

type (
	// Provider is an implementation of [of.FeatureProvider] that can execute multiple providers using various
	// strategies.
	Provider struct {
		providers          []NamedProvider
		metadata           of.Metadata
		initialized        bool
		overallStatus      of.State
		overallStatusLock  sync.RWMutex
		providerStatus     map[string]of.State
		providerStatusLock sync.Mutex
		strategyName       EvaluationStrategy    // the name of the strategy used for evaluation
		strategyFunc       StrategyFn[FlagTypes] // used for evaluating strategies
		logger             *slog.Logger
		outboundEvents     chan of.Event
		workerGroup        sync.WaitGroup
		shutdownFunc       context.CancelFunc
		globalHooks        []of.Hook
	}

	// NamedProvider extends [of.FeatureProvider] by adding a unique provider name. The Name method returns
	// the assigned provider name, while provider returns the underlying [of.FeatureProvider] instance.
	NamedProvider interface {
		of.FeatureProvider
		Name() string
		unwrap() of.FeatureProvider
	}

	// namedProvider allows for a unique name to be assigned to a provider during a multi-provider set up.
	// The name will be used when reporting errors & results to specify the provider associated with them.
	namedProvider struct {
		of.FeatureProvider
		name       string
		extraHooks []of.Hook
	}

	// Option function used for setting configuration via the options pattern
	Option func(*configuration)

	// Private Types
	namedEvent struct {
		of.Event
		providerName string
	}

	// configuration is the internal configuration of a [multi.Provider]
	configuration struct {
		useFallback      bool
		fallbackProvider of.FeatureProvider
		customStrategy   StrategyConstructor
		logger           *slog.Logger
		hooks            []of.Hook
		providers        []*namedProvider
		customComparator Comparator
	}

	// namedEventHandler is a wrapper around an [of.EventHandler] that includes the provider name.
	namedEventHandler struct {
		of.EventHandler
		name string
	}
)

func (n *namedProvider) Name() string {
	return n.name
}

func (n *namedProvider) unwrap() of.FeatureProvider {
	return n.FeatureProvider
}

var (
	stateValues      map[of.State]int
	stateTable       [3]of.State
	eventTypeToState map[of.EventType]of.State

	// Compile-time interface compliance checks
	_ of.FeatureProvider          = (*Provider)(nil)
	_ of.EventHandler             = (*Provider)(nil)
	_ of.ContextAwareStateHandler = (*Provider)(nil)
	_ of.Tracker                  = (*Provider)(nil)
	_ NamedProvider               = (*namedProvider)(nil)
)

// init Initialize "constants" used for event handling priorities and filtering.
func init() {
	// used for mapping provider event types & provider states to comparable values for evaluation
	stateValues = map[of.State]int{
		"":            -1, // Not a real state, but used for handling provider config changes
		of.ReadyState: 0,
		of.StaleState: 1,
		of.ErrorState: 2,
	}
	// used for mapping
	stateTable = [3]of.State{
		of.ReadyState, // 0
		of.StaleState, // 1
		of.ErrorState, // 2
	}
	eventTypeToState = map[of.EventType]of.State{
		of.ProviderConfigChange: "",
		of.ProviderReady:        of.ReadyState,
		of.ProviderStale:        of.StaleState,
		of.ProviderError:        of.ErrorState,
	}
}

// Configuration Options

// WithLogger sets a logger to be used with slog for internal logging. By default, all logs are discarded unless this is set.
func WithLogger(l *slog.Logger) Option {
	return func(conf *configuration) {
		conf.logger = l
	}
}

// WithFallbackProvider sets a fallback provider when using the [StrategyComparison] setting. The fallback provider is
// called when providers are not in agreement. If a fallback provider is not set and providers are not agreement, then
// the default result will be returned along with an error value.
func WithFallbackProvider(p of.FeatureProvider) Option {
	return func(conf *configuration) {
		conf.fallbackProvider = p
		conf.useFallback = true
	}
}

// WithCustomComparator sets a custom [Comparator] to use when using [StrategyComparison] when [of.FeatureProvider.ObjectEvaluation]
// is performed. This is required if the returned objects are not comparable, otherwise an error occur..
func WithCustomComparator(comparator Comparator) Option {
	return func(conf *configuration) {
		conf.customComparator = comparator
	}
}

// WithCustomStrategy sets a custom strategy function by defining a "constructor" that acts as closure over a slice of
// [NamedProvider] instances with your returned custom strategy function. This must be used in conjunction with [StrategyCustom]
func WithCustomStrategy(s StrategyConstructor) Option {
	return func(conf *configuration) {
		conf.customStrategy = s
	}
}

// WithGlobalHooks sets the global hooks for the provider. These are [of.Hook] instances that affect ALL [of.FeatureProvider]
// instances. For hooks that target specific providers make sure to attach them to that provider directly, or use the
// [WithProviderHooks] [Option] if that provider does not provide its own hook functionality.
func WithGlobalHooks(hooks ...of.Hook) Option {
	return func(conf *configuration) {
		conf.hooks = hooks
	}
}

// WithProvider registers a specific [of.FeatureProvider] instance under the given providerName. The providerName
// must be unique and correspond to the name used when creating the [multi.Provider]. Optional [of.Hook] instances
// may also be provided, which will execute only for this specific provider. This [Option] can be used multiple times
// with unique provider names to register multiple providers. The order in which options
// are provided determines the order in which the providers are registered and evaluated.
func WithProvider(providerName string, provider of.FeatureProvider, hooks ...of.Hook) Option {
	return func(conf *configuration) {
		conf.providers = append(conf.providers, &namedProvider{
			name:            providerName,
			FeatureProvider: provider,
			extraHooks:      hooks,
		})
	}
}

// Multiprovider Implementation
func buildMetadata(m []NamedProvider) of.Metadata {
	var separator string
	var metaName strings.Builder
	metaName.WriteString("MultiProvider {")
	for _, p := range m {
		metaName.WriteString(fmt.Sprintf("%s%s: %s", separator, p.Name(), p.Metadata().Name))
		if separator == "" {
			separator = ", "
		}
	}

	metaName.WriteRune('}')
	return of.Metadata{
		Name: metaName.String(),
	}
}

// NewProvider returns a new [multi.Provider] that acts as a unified interface of multiple providers for interaction.
func NewProvider(evaluationStrategy EvaluationStrategy, options ...Option) (*Provider, error) {
	config := &configuration{
		logger:    slog.New(slog.DiscardHandler),
		providers: make([]*namedProvider, 0, 2),
	}

	for _, opt := range options {
		opt(config)
	}

	if len(config.providers) == 0 {
		return nil, errors.New("no providers configured: at least one provider must be registered using WithProvider()")
	}

	providers := make([]NamedProvider, 0, len(config.providers))
	collectedHooks := make([]of.Hook, 0, len(config.providers))
	for i, provider := range config.providers {
		// Validate Providers
		if provider.FeatureProvider == nil {
			return nil, fmt.Errorf("provider %s at %d cannot be nil", provider.name, i)
		}
		if provider.name == "" {
			return nil, fmt.Errorf("provider name at %d cannot be the empty string", i)
		}

		// Wrap any providers that include hooks
		if (len(provider.Hooks()) + len(provider.extraHooks)) == 0 {
			providers = append(providers, provider)
			continue
		}

		var wrappedProvider NamedProvider
		if _, ok := provider.FeatureProvider.(of.EventHandler); ok {
			wrappedProvider = isolateProviderWithEvents(provider, provider.extraHooks)
		} else {
			wrappedProvider = isolateProvider(provider, provider.extraHooks)
		}

		providers = append(providers, wrappedProvider)
		collectedHooks = append(collectedHooks, wrappedProvider.Hooks()...)
	}

	multiProvider := &Provider{
		providers:      providers,
		outboundEvents: make(chan of.Event, len(providers)),
		logger:         config.logger,
		metadata:       buildMetadata(providers),
		overallStatus:  of.NotReadyState,
		providerStatus: make(map[string]of.State, len(providers)),
		globalHooks:    append(config.hooks, collectedHooks...),
	}

	var strategy StrategyFn[FlagTypes]
	switch evaluationStrategy {
	case StrategyFirstMatch:
		strategy = newFirstMatchStrategy(multiProvider.Providers())
	case StrategyFirstSuccess:
		strategy = newFirstSuccessStrategy(multiProvider.Providers())
	case StrategyComparison:
		strategy = newComparisonStrategy(multiProvider.Providers(), config.fallbackProvider, config.customComparator)
	default:
		if config.customStrategy == nil {
			return nil, fmt.Errorf("%s is an unknown evaluation strategy", evaluationStrategy)
		}
		strategy = config.customStrategy(multiProvider.Providers())
	}
	multiProvider.strategyFunc = strategy
	multiProvider.strategyName = evaluationStrategy

	return multiProvider, nil
}

// Providers returns slice of providers wrapped in [NamedProvider] structs.
func (p *Provider) Providers() []NamedProvider {
	return p.providers
}

// EvaluationStrategy The name of the currently set [EvaluationStrategy].
func (p *Provider) EvaluationStrategy() string {
	return p.strategyName
}

// Metadata provides the name "multiprovider" along with the names and types of each internal [of.FeatureProvider].
func (p *Provider) Metadata() of.Metadata {
	return p.metadata
}

// Hooks returns a collection [of.Hook] instances configured to the provider using [WithGlobalHooks].
func (p *Provider) Hooks() []of.Hook {
	return p.globalHooks
}

// BooleanEvaluation evaluates the flag and returns a [of.BoolResolutionDetail].
func (p *Provider) BooleanEvaluation(ctx context.Context, flag string, defaultValue bool, flatCtx of.FlattenedContext) of.BoolResolutionDetail {
	res := p.strategyFunc(ctx, flag, defaultValue, flatCtx)
	return of.BoolResolutionDetail{
		Value:                    res.Value.(bool),
		ProviderResolutionDetail: res.ProviderResolutionDetail,
	}
}

// StringEvaluation evaluates the flag and returns a [of.StringResolutionDetail].
func (p *Provider) StringEvaluation(ctx context.Context, flag string, defaultValue string, flatCtx of.FlattenedContext) of.StringResolutionDetail {
	res := p.strategyFunc(ctx, flag, defaultValue, flatCtx)
	return of.StringResolutionDetail{
		Value:                    res.Value.(string),
		ProviderResolutionDetail: res.ProviderResolutionDetail,
	}
}

// FloatEvaluation evaluates the flag and returns a [of.FloatResolutionDetail].
func (p *Provider) FloatEvaluation(ctx context.Context, flag string, defaultValue float64, flatCtx of.FlattenedContext) of.FloatResolutionDetail {
	res := p.strategyFunc(ctx, flag, defaultValue, flatCtx)
	return of.FloatResolutionDetail{
		Value:                    res.Value.(float64),
		ProviderResolutionDetail: res.ProviderResolutionDetail,
	}
}

// IntEvaluation evaluates the flag and returns an [of.IntResolutionDetail].
func (p *Provider) IntEvaluation(ctx context.Context, flag string, defaultValue int64, flatCtx of.FlattenedContext) of.IntResolutionDetail {
	res := p.strategyFunc(ctx, flag, defaultValue, flatCtx)
	return of.IntResolutionDetail{
		Value:                    res.Value.(int64),
		ProviderResolutionDetail: res.ProviderResolutionDetail,
	}
}

// ObjectEvaluation evaluates the flag and returns an [of.InterfaceResolutionDetail]. For the purposes of evaluation
// within strategies, the type of the default value is used as the assumed type of the returned responses from each provider.
// This is especially important when using the [StrategyComparison] configuration as an internal error will occur if this
// is not a comparable type unless the [WithCustomComparator] [Option] is configured.
func (p *Provider) ObjectEvaluation(ctx context.Context, flag string, defaultValue any, flatCtx of.FlattenedContext) of.InterfaceResolutionDetail {
	res := p.strategyFunc(ctx, flag, defaultValue, flatCtx)
	return of.InterfaceResolutionDetail{
		Value:                    res.Value,
		ProviderResolutionDetail: res.ProviderResolutionDetail,
	}
}

// Init will run the initialize method for all internal [of.FeatureProvider] instances and aggregate any errors.
func (p *Provider) Init(evalCtx of.EvaluationContext) error {
	return p.InitWithContext(context.Background(), evalCtx)
}

// InitWithContext will run the initialize method for all internal [of.FeatureProvider] instances and aggregate any errors.
func (p *Provider) InitWithContext(ctx context.Context, evalCtx of.EvaluationContext) error {
	var eg errgroup.Group
	// wrapper type used only for initialization of event listener workers
	p.logger.LogAttrs(ctx, slog.LevelDebug, "start initialization")
	handlers := make(chan namedEventHandler, len(p.providers))
	for _, provider := range p.providers {
		name := provider.Name()
		// Initialize each provider to not ready state. No locks required there are no workers running
		p.updateProviderState(name, of.NotReadyState)
		l := p.logger.With(slog.String(MetadataProviderName, name))
		prov := provider
		eg.Go(func() error {
			l.LogAttrs(ctx, slog.LevelDebug, "starting initialization")
			if stateHandle, ok := prov.unwrap().(of.StateHandler); ok {
				var err error
				if contextAwareHandle, ok := stateHandle.(of.ContextAwareStateHandler); ok {
					err = contextAwareHandle.InitWithContext(ctx, evalCtx)
				} else {
					err = stateHandle.Init(evalCtx)
				}

				if err != nil {
					l.LogAttrs(ctx, slog.LevelError, "initialization failed", slog.Any("error", err))
					p.updateProviderState(name, of.ErrorState)
					return &ProviderError{
						Err:          err,
						ProviderName: name,
					}
				}
			} else {
				l.LogAttrs(ctx, slog.LevelDebug, "StateHandle not implemented, skipping initialization")
			}
			l.LogAttrs(ctx, slog.LevelDebug, "initialization successful")
			if eventer, ok := prov.unwrap().(of.EventHandler); ok {
				l.LogAttrs(context.Background(), slog.LevelDebug, "detected EventHandler implementation")
				handlers <- namedEventHandler{eventer, name}
			}
			p.updateProviderState(name, of.ReadyState)
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		var pErr *ProviderError
		if !errors.As(err, &pErr) {
			pErr = &ProviderError{
				Err:          err,
				ProviderName: "unknown",
			}
		}

		p.setStatus(of.ErrorState)
		return pErr
	}
	close(handlers)
	workerCtx, shutdownFunc := context.WithCancel(context.Background())
	p.shutdownFunc = shutdownFunc

	if len(handlers) > 0 {
		p.workerGroup.Add(1)
		go p.forwardProviderEvents(workerCtx, handlers)
	} else {
		// we don't emit any events so we can just close the channel
		close(p.outboundEvents)
	}

	p.setStatus(of.ReadyState)
	p.initialized = true
	return nil
}

// forwardProviderEvents establishes an event forwarding pipeline that collects events from multiple provider
// event handlers and forwards them to the multiprovider's outbound event channel. It spawns a goroutine for
// each provider handler to listen for events, aggregates them through an internal pipe, and selectively forwards
// events that result in state changes. The function blocks until workerCtx is cancelled or all provider event
// channels are closed, ensuring proper cleanup by closing the outbound channel when complete.
func (p *Provider) forwardProviderEvents(workerCtx context.Context, handlers chan namedEventHandler) {
	defer p.workerGroup.Done()
	defer close(p.outboundEvents)

	workerLogger := p.logger.With(slog.String("multiprovider-worker", "event-forwarder-worker"))
	pipe := make(chan namedEvent)
	var wg sync.WaitGroup
	for ch := range handlers {
		wg.Add(1)
		go func(ctx context.Context, h of.EventHandler, name string, out chan<- namedEvent) {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case e, ok := <-h.EventChannel():
					if !ok {
						return
					}
					if e.EventMetadata == nil {
						e.EventMetadata = make(map[string]any)
					}
					e.EventMetadata[MetadataProviderName] = name
					if p, ok := h.(of.FeatureProvider); ok {
						e.EventMetadata[MetadataProviderType] = p.Metadata().Name
					}
					out <- namedEvent{
						Event:        e,
						providerName: name,
					}
				}
			}
		}(workerCtx, ch.EventHandler, ch.name, pipe)
	}

	go func() {
		wg.Wait()
		close(pipe)
	}()

	for e := range pipe {
		l := workerLogger.With(
			slog.String(MetadataProviderName, e.providerName),
			slog.String(MetadataProviderType, e.ProviderName),
		)
		l.LogAttrs(workerCtx, slog.LevelDebug, "received event from provider", slog.String("event-type", string(e.EventType)))
		if p.updateProviderStateFromEvent(e) {
			p.outboundEvents <- e.Event
			l.LogAttrs(workerCtx, slog.LevelDebug, "forwarded state update event")
		} else {
			l.LogAttrs(workerCtx, slog.LevelDebug, "total state not updated, inbound event will not be emitted")
		}
	}
}

// updateProviderState Updates the state of an internal provider and then re-evaluates the overall state of the
// multiprovider. If this method returns true the overall state changed.
func (p *Provider) updateProviderState(name string, state of.State) bool {
	p.providerStatusLock.Lock()
	defer p.providerStatusLock.Unlock()
	p.providerStatus[name] = state
	evalState := p.evaluateState()
	if evalState != p.Status() {
		p.setStatus(evalState)
		return true
	}

	return false
}

// updateProviderStateFromEvent updates the state of an internal provider from an event emitted from it, and then
// re-evaluates the overall state of the multiprovider. If this method returns true the overall state changed.
func (p *Provider) updateProviderStateFromEvent(e namedEvent) bool {
	if e.EventType == of.ProviderConfigChange {
		p.logger.LogAttrs(context.Background(), slog.LevelDebug, "ProviderConfigChange event", slog.String("event-message", e.Message))
	}
	p.providerStatusLock.Lock()
	previousState := p.providerStatus[e.providerName]
	p.providerStatusLock.Unlock()
	logProviderState(p.logger, e, previousState)
	return p.updateProviderState(e.providerName, eventTypeToState[e.EventType])
}

// evaluateState Determines the overall state of the provider using the weights specified in Appendix A of the
// OpenFeature spec. This method should only be called if the provider state mutex is locked
func (p *Provider) evaluateState() of.State {
	maxState := stateValues[of.ReadyState] // initialize to the lowest state value
	for _, s := range p.providerStatus {
		if stateValues[s] > maxState {
			// change in state due to higher priority
			maxState = stateValues[s]
		}
	}
	return stateTable[maxState]
}

func logProviderState(l *slog.Logger, e namedEvent, previousState of.State) {
	switch eventTypeToState[e.EventType] {
	case of.ReadyState:
		if previousState != of.NotReadyState {
			l.LogAttrs(context.Background(), slog.LevelInfo, "provider has returned to ready state",
				slog.String(MetadataProviderName, e.providerName), slog.String("previous-state", string(previousState)))
			return
		}
		l.LogAttrs(context.Background(), slog.LevelDebug, "provider is ready", slog.String(MetadataProviderName, e.providerName))
	case of.StaleState:
		l.LogAttrs(context.Background(), slog.LevelWarn, "provider is stale",
			slog.String(MetadataProviderName, e.providerName), slog.String("event-message", e.Message))
	case of.ErrorState:
		l.LogAttrs(context.Background(), slog.LevelError, "provider is in an error state",
			slog.String(MetadataProviderName, e.providerName), slog.String("event-message", e.Message))
	}
}

// Shutdown Shuts down all internal [of.FeatureProvider] instances and internal event listeners
func (p *Provider) Shutdown() {
	ctx := context.Background()
	err := p.ShutdownWithContext(ctx)
	if err != nil {
		p.logger.LogAttrs(ctx, slog.LevelWarn, "error during shutdown", slog.Any("error", err))
	}
}

// ShutdownWithContext shuts down all internal [of.FeatureProvider] instances and internal event listeners
func (p *Provider) ShutdownWithContext(ctx context.Context) error {
	if !p.initialized {
		// Don't do anything if we were never initialized
		p.logger.LogAttrs(ctx, slog.LevelDebug, "provider not initialized, skipping shutdown")
		return nil
	}

	p.logger.LogAttrs(ctx, slog.LevelDebug, "starting provider shutdown")
	// Stop all event listener workers, shutdown events should not affect overall state
	p.shutdownFunc()

	var wg sync.WaitGroup
	var errs []error
	for _, provider := range p.providers {
		name := provider.Name()
		if stateHandle, ok := provider.unwrap().(of.StateHandler); ok {
			wg.Add(1)
			go func(p of.StateHandler) {
				defer wg.Done()
				if contextAwareHandle, ok := p.(of.ContextAwareStateHandler); ok {
					if err := contextAwareHandle.ShutdownWithContext(ctx); err != nil {
						errs = append(errs, fmt.Errorf("provider %s shutdown failed: %w", name, err))
					}
				} else {
					p.Shutdown()
				}
			}(stateHandle)
		}
	}

	p.logger.LogAttrs(context.Background(), slog.LevelDebug, "waiting for provider shutdown completion")
	wg.Wait()
	// Stop forwarding worker
	p.logger.LogAttrs(context.Background(), slog.LevelDebug, "triggered worker shutdown")
	// Wait for workers to stop
	p.workerGroup.Wait()
	p.logger.LogAttrs(context.Background(), slog.LevelDebug, "worker shutdown completed")
	p.setStatus(of.NotReadyState)
	p.initialized = false
	p.logger.LogAttrs(context.Background(), slog.LevelDebug, "provider shutdown completed")

	return errors.Join(errs...)
}

// Status provides the current state of the [multi.Provider].
func (p *Provider) Status() of.State {
	p.overallStatusLock.RLock()
	defer p.overallStatusLock.RUnlock()
	return p.overallStatus
}

func (p *Provider) setStatus(state of.State) {
	p.overallStatusLock.Lock()
	defer p.overallStatusLock.Unlock()
	p.overallStatus = state
	p.logger.LogAttrs(context.Background(), slog.LevelDebug, "state updated", slog.String("state", string(state)))
}

// EventChannel is the channel that all events are emitted on.
func (p *Provider) EventChannel() <-chan of.Event {
	return p.outboundEvents
}

// Track implements the [of.Tracker] interface by forwarding tracking calls to all internal providers that
// are in ready state and implement the [of.Tracker] interface.
func (p *Provider) Track(ctx context.Context, trackingEventName string, evaluationContext of.EvaluationContext, details of.TrackingEventDetails) {
	if !p.initialized {
		// Don't do anything if we were never initialized
		p.logger.LogAttrs(ctx, slog.LevelDebug, "provider not initialized, skipping tracking", slog.String("tracking-event", trackingEventName))
		return
	}
	p.providerStatusLock.Lock()
	statuses := maps.Clone(p.providerStatus)
	p.providerStatusLock.Unlock()
	providers := make([]NamedProvider, 0, len(p.providers))
	for _, p := range p.providers {
		if statuses[p.Name()] == of.ReadyState {
			providers = append(providers, p)
		}
	}
	for _, provider := range providers {
		if tracker, ok := provider.unwrap().(of.Tracker); ok {
			tracker.Track(ctx, trackingEventName, evaluationContext, details)
		}
	}
}
