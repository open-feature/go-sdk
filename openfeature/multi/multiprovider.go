// Package multi is an experimental implementation of a [of.FeatureProvider] that supports evaluating multiple feature flag
// providers together.
package multi

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
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
	// ProviderMap is an alias for a map containing unique names for each included [of.FeatureProvider]
	ProviderMap = map[string]of.FeatureProvider

	// Provider is an implementation of [of.FeatureProvider] that can execute multiple providers using various
	// strategies.
	Provider struct {
		providers          ProviderMap
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
		inboundEvents      chan namedEvent
		workerGroup        sync.WaitGroup
		shutdownFunc       context.CancelFunc
		globalHooks        []of.Hook
	}

	// NamedProvider allows for a unique name to be assigned to a provider during a multi-provider set up.
	// The name will be used when reporting errors & results to specify the provider associated with them.
	NamedProvider struct {
		Name string
		of.FeatureProvider
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
		providerHooks    map[string][]of.Hook
		customComparator Comparator
	}
)

var (
	stateValues      map[of.State]int
	stateTable       [3]of.State
	eventTypeToState map[of.EventType]of.State

	// Compile-time interface compliance checks
	_ of.FeatureProvider = (*Provider)(nil)
	_ of.EventHandler    = (*Provider)(nil)
	_ of.StateHandler    = (*Provider)(nil)
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

// WithProviderHooks sets [of.Hook] instances that execute only for a specific [of.FeatureProvider]. The providerName
// must match the unique provider name set during [multi.Provider] creation. This should only be used if you need hooks
// that execute around a specific provider, but that provider does not currently accept a way to set hooks. This [Option]
// can be used multiple times using unique provider names. Using a provider name that is not known will cause an error.
func WithProviderHooks(providerName string, hooks ...of.Hook) Option {
	return func(conf *configuration) {
		conf.providerHooks[providerName] = hooks
	}
}

// Multiprovider Implementation

// toNamedProviderSlice converts the provided [ProviderMap] into a slice of [NamedProvider] instances
func toNamedProviderSlice(m ProviderMap) []*NamedProvider {
	s := make([]*NamedProvider, 0, len(m))
	for name, provider := range m {
		s = append(s, &NamedProvider{Name: name, FeatureProvider: provider})
	}

	return s
}

func buildMetadata(m ProviderMap) of.Metadata {
	var separator string
	var metaName strings.Builder
	metaName.WriteString("MultiProvider {")
	names := make([]string, 0, len(m))
	for n := range m {
		names = append(names, n)
	}
	slices.Sort(names)
	for _, name := range names {
		metaName.WriteString(fmt.Sprintf("%s%s: %s", separator, name, m[name].Metadata().Name))
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
func NewProvider(providerMap ProviderMap, evaluationStrategy EvaluationStrategy, options ...Option) (*Provider, error) {
	if len(providerMap) == 0 {
		return nil, errors.New("providerMap cannot be nil or empty")
	}

	config := &configuration{
		logger:        slog.New(slog.DiscardHandler),
		providerHooks: make(map[string][]of.Hook),
	}

	for _, opt := range options {
		opt(config)
	}

	providers := providerMap
	collectedHooks := make([]of.Hook, 0, len(providerMap))
	for name, provider := range providerMap {
		// Validate Providers
		if name == "" {
			return nil, errors.New("provider name cannot be the empty string")
		}

		if provider == nil {
			return nil, fmt.Errorf("provider %s cannot be nil", name)
		}

		// Wrap any providers that include hooks
		if (len(provider.Hooks()) + len(config.providerHooks[name])) == 0 {
			continue
		}

		var wrappedProvider of.FeatureProvider
		if _, ok := provider.(of.EventHandler); ok {
			wrappedProvider = isolateProviderWithEvents(provider, config.providerHooks[name])
		} else {
			wrappedProvider = isolateProvider(provider, config.providerHooks[name])
		}

		providers[name] = wrappedProvider
		collectedHooks = append(collectedHooks, wrappedProvider.Hooks()...)
	}

	multiProvider := &Provider{
		providers:      providers,
		outboundEvents: make(chan of.Event),
		logger:         config.logger,
		metadata:       buildMetadata(providerMap),
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
func (p *Provider) Providers() []*NamedProvider {
	return toNamedProviderSlice(p.providers)
}

// ProvidersByName Returns the internal [ProviderMap].
func (p *Provider) ProvidersByName() ProviderMap {
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
	var eg errgroup.Group
	// wrapper type used only for initialization of event listener workers
	type namedEventHandler struct {
		of.EventHandler
		name string
	}
	p.logger.LogAttrs(context.Background(), slog.LevelDebug, "start initialization")
	p.inboundEvents = make(chan namedEvent, len(p.providers))
	handlers := make(chan namedEventHandler, len(p.providers))
	for name, provider := range p.providers {
		// Initialize each provider to not ready state. No locks required there are no workers running
		p.updateProviderState(name, of.NotReadyState)
		l := p.logger.With(slog.String(MetadataProviderName, name))
		prov := provider
		eg.Go(func() error {
			l.LogAttrs(context.Background(), slog.LevelDebug, "starting initialization")
			stateHandle, ok := prov.(of.StateHandler)
			if !ok {
				l.LogAttrs(context.Background(), slog.LevelDebug, "StateHandle not implemented, skipping initialization")
			} else if err := stateHandle.Init(evalCtx); err != nil {
				l.LogAttrs(context.Background(), slog.LevelError, "initialization failed", slog.Any("error", err))
				return &ProviderError{
					Err:          err,
					ProviderName: name,
				}
			}
			l.LogAttrs(context.Background(), slog.LevelDebug, "initialization successful")
			if eventer, ok := provider.(of.EventHandler); ok {
				l.LogAttrs(context.Background(), slog.LevelDebug, "detected EventHandler implementation")
				handlers <- namedEventHandler{eventer, name}
			} else {
				// Do not yet update providers that need event handling
				p.updateProviderState(name, of.ReadyState)
			}
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		var pErr *ProviderError
		if errors.As(err, &pErr) {
			// Update provider status to error, no event needs to be emitted yet
			p.updateProviderState(pErr.ProviderName, of.ErrorState)
		} else {
			pErr = &ProviderError{
				Err:          err,
				ProviderName: "unknown",
			}
			p.setStatus(of.ErrorState)
		}

		return err
	}
	close(handlers)
	workerCtx, shutdownFunc := context.WithCancel(context.Background())
	for h := range handlers {
		go p.startListening(workerCtx, h.name, h.EventHandler, &p.workerGroup)
	}
	p.shutdownFunc = shutdownFunc

	p.workerGroup.Add(1)
	go func() {
		workerLogger := p.logger.With(slog.String("multiprovider-worker", "event-forwarder-worker"))
		defer p.workerGroup.Done()
		for e := range p.inboundEvents {
			l := workerLogger.With(
				slog.String(MetadataProviderName, e.providerName),
				slog.String(MetadataProviderType, e.ProviderName),
			)
			l.LogAttrs(context.Background(), slog.LevelDebug, "received event from provider", slog.String("event-type", string(e.EventType)))
			if p.updateProviderStateFromEvent(e) {
				p.outboundEvents <- e.Event
				l.LogAttrs(context.Background(), slog.LevelDebug, "forwarded state update event")
			} else {
				l.LogAttrs(context.Background(), slog.LevelDebug, "total state not updated, inbound event will not be emitted")
			}
		}
	}()

	p.setStatus(of.ReadyState)
	p.initialized = true
	return nil
}

// startListening is intended to be called on a per-provider basis as a goroutine to listen to events from a provider
// implementing [of.EventHandler].
func (p *Provider) startListening(ctx context.Context, name string, h of.EventHandler, wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()
	for {
		select {
		case e := <-h.EventChannel():
			e.EventMetadata[MetadataProviderName] = name
			e.EventMetadata[MetadataProviderType] = h.(of.FeatureProvider).Metadata().Name
			p.inboundEvents <- namedEvent{
				Event:        e,
				providerName: name,
			}
		case <-ctx.Done():
			return
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
	logProviderState(p.logger, e, p.providerStatus[e.providerName])
	return p.updateProviderState(e.ProviderName, eventTypeToState[e.EventType])
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
	if !p.initialized {
		// Don't do anything if we were never initialized
		return
	}
	// Stop all event listener workers, shutdown events should not affect overall state
	p.shutdownFunc()
	// Stop forwarding worker
	close(p.inboundEvents)
	p.logger.LogAttrs(context.Background(), slog.LevelDebug, "triggered worker shutdown")
	// Wait for workers to stop
	p.workerGroup.Wait()
	p.logger.LogAttrs(context.Background(), slog.LevelDebug, "worker shutdown completed")
	p.logger.LogAttrs(context.Background(), slog.LevelDebug, "starting provider shutdown")
	var wg sync.WaitGroup
	for _, provider := range p.providers {
		wg.Add(1)

		go func(p of.FeatureProvider) {
			defer wg.Done()
			if stateHandle, ok := p.(of.StateHandler); ok {
				stateHandle.Shutdown()
			}
		}(provider)
	}

	p.logger.LogAttrs(context.Background(), slog.LevelDebug, "waiting for provider shutdown completion")
	wg.Wait()
	p.logger.LogAttrs(context.Background(), slog.LevelDebug, "provider shutdown completed")
	p.setStatus(of.NotReadyState)
	close(p.outboundEvents)
	p.outboundEvents = nil
	p.inboundEvents = nil
	p.initialized = false
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
