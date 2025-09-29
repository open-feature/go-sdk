// Package multiprovider implements an OpenFeature provider that supports multiple feature flag providers.
package multiprovider

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"sync"
	"time"

	of "github.com/open-feature/go-sdk/openfeature"
	"golang.org/x/sync/errgroup"
)

// Metadata Keys
const (
	MetadataProviderName                   = "multiprovider-provider-name"
	MetadataProviderType                   = "multiprovider-provider-type"
	MetadataInternalError                  = "multiprovider-internal-error"
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

	// MultiProvider is an implementation of [of.FeatureProvider] that can execute multiple providers using various
	// strategies.
	MultiProvider struct {
		providers          ProviderMap
		metadata           of.Metadata
		initialized        bool
		totalStatus        of.State
		totalStatusLock    sync.RWMutex
		providerStatus     map[string]of.State
		providerStatusLock sync.Mutex
		strategyName       EvaluationStrategy    // the name of the strategy used for evaluation
		strategy           StrategyFn[FlagTypes] // used for custom strategies
		logger             *slog.Logger
		outboundEvents     chan of.Event
		inboundEvents      chan namedEvent
		workerGroup        sync.WaitGroup
		shutdownFunc       context.CancelFunc
		globalHooks        []of.Hook
	}

	// NamedProvider allows for a unique name to be assigned to a provider during a multi-provider set up.
	// The name will be used when reporting errors & results to specify the provider associated.
	NamedProvider struct {
		Name string
		of.FeatureProvider
	}

	// configuration is the internal configuration of a [MultiProvider]
	configuration struct {
		useFallback      bool
		fallbackProvider of.FeatureProvider
		customStrategy   StrategyFn[FlagTypes]
		logger           *slog.Logger
		timeout          time.Duration
		hooks            []of.Hook
		providerHooks    map[string][]of.Hook
		customComparator Comparator
	}

	// Option function used for setting configuration via the options pattern
	Option func(*configuration)

	// Private Types
	namedEvent struct {
		of.Event
		providerName string
	}
)

var (
	stateValues      map[of.State]int
	stateTable       [3]of.State
	eventTypeToState map[of.EventType]of.State

	// Compile-time interface compliance checks
	_ of.FeatureProvider = (*MultiProvider)(nil)
	_ of.EventHandler    = (*MultiProvider)(nil)
	_ of.StateHandler    = (*MultiProvider)(nil)
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

// WithLogger sets a logger to be used with slog for internal logging. By default, all logs are discarded
func WithLogger(l *slog.Logger) Option {
	return func(conf *configuration) {
		conf.logger = l
	}
}

// WithTimeout set a timeout for the total runtime for evaluation of parallel strategies
func WithTimeout(d time.Duration) Option {
	return func(conf *configuration) {
		conf.timeout = d
	}
}

// WithFallbackProvider sets a fallback provider when using the StrategyComparison
func WithFallbackProvider(p of.FeatureProvider) Option {
	return func(conf *configuration) {
		conf.fallbackProvider = p
		conf.useFallback = true
	}
}

// WithCustomComparator sets a custom [Comparator] to use when using [StrategyComparison] when [of.FeatureProvider.ObjectEvaluation]
// is performed. This is required if the returned objects are not comparable.
func WithCustomComparator(comparator Comparator) Option {
	return func(conf *configuration) {
		conf.customComparator = comparator
	}
}

// WithCustomStrategy sets a custom strategy. This must be used in conjunction with StrategyCustom
func WithCustomStrategy(s StrategyFn[FlagTypes]) Option {
	return func(conf *configuration) {
		conf.customStrategy = s
	}
}

// WithGlobalHooks sets the global hooks for the provider. These are [of.Hook] instances that affect ALL [of.FeatureProvider]
// instances. For hooks that target specific providers make sure to attach them to that provider directly, or use the
// [WithProviderHooks] Option if that provider does not provide its own hook functionality
func WithGlobalHooks(hooks ...of.Hook) Option {
	return func(conf *configuration) {
		conf.hooks = hooks
	}
}

// WithProviderHooks sets [of.Hook] instances that execute only for a specific [of.FeatureProvider]. The providerName
// must match the unique provider name set during [MultiProvider] creation. This should only be used if you need hooks
// that execute around a specific provider, but that provider does not currently accept a way to set hooks. This option
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

// NewMultiProvider returns the unified interface of multiple providers for interaction.
func NewMultiProvider(providerMap ProviderMap, evaluationStrategy EvaluationStrategy, options ...Option) (*MultiProvider, error) {
	if len(providerMap) == 0 {
		return nil, errors.New("providerMap cannot be nil or empty")
	}

	config := &configuration{
		logger:        slog.New(slog.DiscardHandler),
		providerHooks: make(map[string][]of.Hook),
		timeout:       5 * time.Second, // Default timeout
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
			wrappedProvider = IsolateProviderWithEvents(provider, config.providerHooks[name])
		} else {
			wrappedProvider = IsolateProvider(provider, config.providerHooks[name])
		}

		providers[name] = wrappedProvider
		collectedHooks = append(collectedHooks, wrappedProvider.Hooks()...)
	}

	multiProvider := &MultiProvider{
		providers:      providers,
		outboundEvents: make(chan of.Event),
		logger:         config.logger,
		metadata:       buildMetadata(providerMap),
		totalStatus:    of.NotReadyState,
		providerStatus: make(map[string]of.State, len(providers)),
		globalHooks:    append(config.hooks, collectedHooks...),
	}

	var strategy StrategyFn[FlagTypes]
	switch evaluationStrategy {
	case StrategyFirstMatch:
		strategy = NewFirstMatchStrategy(multiProvider.Providers())
	case StrategyFirstSuccess:
		strategy = NewFirstSuccessStrategy(multiProvider.Providers())
	case StrategyComparison:
		strategy = NewComparisonStrategy(multiProvider.Providers(), config.fallbackProvider, config.customComparator)
	default:
		if config.customStrategy == nil {
			return nil, fmt.Errorf("%s is an unknown evaluation strategy", evaluationStrategy)
		}
		strategy = config.customStrategy
	}
	multiProvider.strategy = strategy
	multiProvider.strategyName = evaluationStrategy

	return multiProvider, nil
}

// Providers Returns slice of providers wrapped in [NamedProvider] structs
func (mp *MultiProvider) Providers() []*NamedProvider {
	return toNamedProviderSlice(mp.providers)
}

// ProvidersByName Returns the internal [ProviderMap] of the [MultiProvider]
func (mp *MultiProvider) ProvidersByName() ProviderMap {
	return mp.providers
}

// EvaluationStrategy The current set strategy's name
func (mp *MultiProvider) EvaluationStrategy() string {
	return mp.strategyName
}

// Metadata provides the name "multiprovider" and the names of each provider passed.
func (mp *MultiProvider) Metadata() of.Metadata {
	return mp.metadata
}

// Hooks returns a collection [of.Hook] instances defined by this provider
func (mp *MultiProvider) Hooks() []of.Hook {
	// Hooks that should be included with the provider
	return []of.Hook{}
}

// BooleanEvaluation returns a boolean flag
func (mp *MultiProvider) BooleanEvaluation(ctx context.Context, flag string, defaultValue bool, flatCtx of.FlattenedContext) of.BoolResolutionDetail {
	res := mp.strategy(ctx, flag, defaultValue, flatCtx)
	return of.BoolResolutionDetail{
		Value:                    res.Value.(bool),
		ProviderResolutionDetail: res.ProviderResolutionDetail,
	}
}

// StringEvaluation returns a string flag
func (mp *MultiProvider) StringEvaluation(ctx context.Context, flag string, defaultValue string, flatCtx of.FlattenedContext) of.StringResolutionDetail {
	res := mp.strategy(ctx, flag, defaultValue, flatCtx)
	return of.StringResolutionDetail{
		Value:                    res.Value.(string),
		ProviderResolutionDetail: res.ProviderResolutionDetail,
	}
}

// FloatEvaluation returns a float flag
func (mp *MultiProvider) FloatEvaluation(ctx context.Context, flag string, defaultValue float64, flatCtx of.FlattenedContext) of.FloatResolutionDetail {
	res := mp.strategy(ctx, flag, defaultValue, flatCtx)
	return of.FloatResolutionDetail{
		Value:                    res.Value.(float64),
		ProviderResolutionDetail: res.ProviderResolutionDetail,
	}
}

// IntEvaluation returns an int flag
func (mp *MultiProvider) IntEvaluation(ctx context.Context, flag string, defaultValue int64, flatCtx of.FlattenedContext) of.IntResolutionDetail {
	res := mp.strategy(ctx, flag, defaultValue, flatCtx)
	return of.IntResolutionDetail{
		Value:                    res.Value.(int64),
		ProviderResolutionDetail: res.ProviderResolutionDetail,
	}
}

// ObjectEvaluation returns an object flag
func (mp *MultiProvider) ObjectEvaluation(ctx context.Context, flag string, defaultValue any, flatCtx of.FlattenedContext) of.InterfaceResolutionDetail {
	res := mp.strategy(ctx, flag, defaultValue, flatCtx)
	return of.InterfaceResolutionDetail{
		Value:                    res.Value,
		ProviderResolutionDetail: res.ProviderResolutionDetail,
	}
}

// Init will run the initialize method for all internal [of.FeatureProvider] instances and aggregate any errors.
func (mp *MultiProvider) Init(evalCtx of.EvaluationContext) error {
	var eg errgroup.Group
	// wrapper type used only for initialization of event listener workers
	type namedEventHandler struct {
		of.EventHandler
		name string
	}
	mp.logger.LogAttrs(context.Background(), slog.LevelDebug, "start initialization")
	mp.inboundEvents = make(chan namedEvent, len(mp.providers))
	handlers := make(chan namedEventHandler, len(mp.providers))
	for name, provider := range mp.providers {
		// Initialize each provider to not ready state. No locks required there are no workers running
		mp.updateProviderState(name, of.NotReadyState)
		l := mp.logger.With(slog.String("multiprovider-provider-name", name))
		p := provider
		eg.Go(func() error {
			l.LogAttrs(context.Background(), slog.LevelDebug, "starting initialization")
			stateHandle, ok := p.(of.StateHandler)
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
				mp.updateProviderState(name, of.ReadyState)
			}
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		var pErr *ProviderError
		if errors.As(err, &pErr) {
			// Update provider status to error, no event needs to be emitted yet
			mp.updateProviderState(pErr.ProviderName, of.ErrorState)
		} else {
			pErr = &ProviderError{
				Err:          err,
				ProviderName: "unknown",
			}
			mp.setStatus(of.ErrorState)
		}

		return err
	}
	close(handlers)
	workerCtx, shutdownFunc := context.WithCancel(context.Background())
	for h := range handlers {
		go mp.startListening(workerCtx, h.name, h.EventHandler, &mp.workerGroup)
	}
	mp.shutdownFunc = shutdownFunc

	mp.workerGroup.Add(1)
	go func() {
		workerLogger := mp.logger.With(slog.String("multiprovider-worker", "event-forwarder-worker"))
		defer mp.workerGroup.Done()
		for e := range mp.inboundEvents {
			l := workerLogger.With(
				slog.String(MetadataProviderName, e.providerName),
				slog.String(MetadataProviderType, e.ProviderName),
			)
			l.LogAttrs(context.Background(), slog.LevelDebug, fmt.Sprintf("received %s event from provider", e.EventType))
			if mp.updateProviderStateFromEvent(e) {
				mp.outboundEvents <- e.Event
				l.LogAttrs(context.Background(), slog.LevelDebug, "forwarded state update event")
			} else {
				l.LogAttrs(context.Background(), slog.LevelDebug, "total state not updated, inbound event will not be emitted")
			}
		}
	}()

	mp.setStatus(of.ReadyState)
	mp.initialized = true
	return nil
}

// startListening is intended to be called on a per-provider basis as a goroutine to listen to events from a provider
// implementing [of.EventHandler].
func (mp *MultiProvider) startListening(ctx context.Context, name string, h of.EventHandler, wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()
	for {
		select {
		case e := <-h.EventChannel():
			e.EventMetadata[MetadataProviderName] = name
			e.EventMetadata[MetadataProviderType] = h.(of.FeatureProvider).Metadata().Name
			mp.inboundEvents <- namedEvent{
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
func (mp *MultiProvider) updateProviderState(name string, state of.State) bool {
	mp.providerStatusLock.Lock()
	defer mp.providerStatusLock.Unlock()
	mp.providerStatus[name] = state
	evalState := mp.evaluateState()
	if evalState != mp.Status() {
		mp.setStatus(evalState)
		return true
	}

	return false
}

// updateProviderStateFromEvent updates the state of an internal provider from an event emitted from it, and then
// re-evaluates the overall state of the multiprovider. If this method returns true the overall state changed.
func (mp *MultiProvider) updateProviderStateFromEvent(e namedEvent) bool {
	if e.EventType == of.ProviderConfigChange {
		mp.logger.LogAttrs(context.Background(), slog.LevelDebug, fmt.Sprintf("ProviderConfigChange event: %s", e.Message))
	}
	logProviderState(mp.logger, e, mp.providerStatus[e.providerName])
	return mp.updateProviderState(e.ProviderName, eventTypeToState[e.EventType])
}

// evaluateState Determines the overall state of the provider using the weights specified in Appendix A of the
// OpenFeature spec. This method should only be called if the provider state mutex is locked
func (mp *MultiProvider) evaluateState() of.State {
	maxState := stateValues[of.ReadyState] // initialize to the lowest state value
	for _, s := range mp.providerStatus {
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
			l.LogAttrs(context.Background(), slog.LevelInfo, fmt.Sprintf("provider %s has returned to ready state from %s", e.providerName, previousState))
			return
		}
		l.LogAttrs(context.Background(), slog.LevelDebug, fmt.Sprintf("provider %s is ready", e.providerName))
	case of.StaleState:
		l.LogAttrs(context.Background(), slog.LevelWarn, fmt.Sprintf("provider %s is stale: %s", e.providerName, e.Message))
	case of.ErrorState:
		l.LogAttrs(context.Background(), slog.LevelError, fmt.Sprintf("provider %s is in an error state: %s", e.providerName, e.Message))
	}
}

// Shutdown Shuts down all internal [of.FeatureProvider] instances and internal event listeners
func (mp *MultiProvider) Shutdown() {
	if !mp.initialized {
		// Don't do anything if we were never initialized
		return
	}
	// Stop all event listener workers, shutdown events should not affect overall state
	mp.shutdownFunc()
	// Stop forwarding worker
	close(mp.inboundEvents)
	mp.logger.LogAttrs(context.Background(), slog.LevelDebug, "triggered worker shutdown")
	// Wait for workers to stop
	mp.workerGroup.Wait()
	mp.logger.LogAttrs(context.Background(), slog.LevelDebug, "worker shutdown completed")
	mp.logger.LogAttrs(context.Background(), slog.LevelDebug, "starting provider shutdown")
	var wg sync.WaitGroup
	for _, provider := range mp.providers {
		wg.Add(1)

		go func(p of.FeatureProvider) {
			defer wg.Done()
			if stateHandle, ok := p.(of.StateHandler); ok {
				stateHandle.Shutdown()
			}
		}(provider)
	}

	mp.logger.LogAttrs(context.Background(), slog.LevelDebug, "waiting for provider shutdown completion")
	wg.Wait()
	mp.logger.LogAttrs(context.Background(), slog.LevelDebug, "provider shutdown completed")
	mp.setStatus(of.NotReadyState)
	close(mp.outboundEvents)
	mp.outboundEvents = nil
	mp.inboundEvents = nil
	mp.initialized = false
}

// Status the current state of the [MultiProvider]
func (mp *MultiProvider) Status() of.State {
	mp.totalStatusLock.RLock()
	defer mp.totalStatusLock.RUnlock()
	return mp.totalStatus
}

func (mp *MultiProvider) setStatus(state of.State) {
	mp.totalStatusLock.Lock()
	defer mp.totalStatusLock.Unlock()
	mp.totalStatus = state
	mp.logger.LogAttrs(context.Background(), slog.LevelDebug, "state updated", slog.String("state", string(state)))
}

// EventChannel the channel events are emitted on
func (mp *MultiProvider) EventChannel() <-chan of.Event {
	return mp.outboundEvents
}
