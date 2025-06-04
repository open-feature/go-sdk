package multiprovider

import (
	"context"
	"errors"
	"fmt"
	of "github.com/open-feature/go-sdk/openfeature"
	"golang.org/x/sync/errgroup"
	"log/slog"
	"slices"
	"strings"
	"sync"
	"time"
)

const (

	//StrategyFirstMatch EvaluationStrategy = StrategyFirstMatch

	//StrategyFirstSuccess EvaluationStrategy = StrategyFirstSuccess

	// StrategyCustom allows for using a custom Strategy implementation. If this is set you MUST use the WithCustomStrategy
	// option to set it
	StrategyCustom EvaluationStrategy = "strategy-custom"

	// Metadata Keys

	MetadataProviderName                   = "multiprovider-provider-name"
	MetadataProviderType                   = "multiprovider-provider-type"
	MetadataInternalError                  = "multiprovider-internal-error"
	MetadataSuccessfulProviderName         = "multiprovider-successful-provider-name"
	MetadataStrategyUsed                   = "multiprovider-strategy-used"
	MetadataFallbackUsed                   = "multiprovider-fallback-used"
	MetadataIsDefaultValue                 = "multiprovider-is-result-default-value"
	MetadataEvaluationError                = "multiprovider-evaluation-error"
	MetadataComparisonDisagreeingProviders = "multiprovider-comparison-disagreeing-providers"
)

type (
	// ProviderMap Alias for a map containing unique names for each included [of.FeatureProvider]
	ProviderMap = map[string]of.FeatureProvider

	MultiProvider struct {
		providers          ProviderMap
		metadata           of.Metadata
		initialized        bool
		totalStatus        of.State
		totalStatusLock    sync.RWMutex
		providerStatus     map[string]of.State
		providerStatusLock sync.Mutex
		strategy           Strategy
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

	// Configuration MultiProvider's internal configuration
	Configuration struct {
		useFallback               bool
		fallbackProvider          of.FeatureProvider
		customStrategy            Strategy
		logger                    *slog.Logger
		timeout                   time.Duration
		hooks                     []of.Hook
		providerHooks             map[string][]of.Hook
		customComparator          Comparator
		alwaysUseCustomComparator bool
	}

	// Option Function used for setting Configuration via the options pattern
	Option func(*Configuration)

	// Private Types
	namedEvent struct {
		of.Event
		providerName string
	}
)

var (
	_                of.FeatureProvider = (*MultiProvider)(nil)
	_                of.EventHandler    = (*MultiProvider)(nil)
	_                of.StateHandler    = (*MultiProvider)(nil)
	stateValues      map[of.State]int
	stateTable       [3]of.State
	eventTypeToState map[of.EventType]of.State
)

// init Initialize "constants" used for event handling priorities and filtering
func init() {
	// used for mapping provider event types & provider states to comparable values for evaluation
	stateValues = map[of.State]int{
		"":            -1, // Not a real state, but used for handling provider config changes
		of.ErrorState: 0,
		of.StaleState: 1,
		of.ReadyState: 2,
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

// WithLogger Sets a logger to be used with slog for internal logging. By default, all logs are discarded
func WithLogger(l *slog.Logger) Option {
	return func(conf *Configuration) {
		conf.logger = l
	}
}

// WithTimeout Set a timeout for the total runtime for evaluation of parallel strategies
func WithTimeout(d time.Duration) Option {
	return func(conf *Configuration) {
		conf.timeout = d
	}
}

// WithFallbackProvider Sets a fallback provider when using the StrategyComparison
func WithFallbackProvider(p of.FeatureProvider) Option {
	return func(conf *Configuration) {
		conf.fallbackProvider = p
		conf.useFallback = true
	}
}

// WithCustomStrategy sets a custom strategy. This must be used in conjunction with StrategyCustom
func WithCustomStrategy(s Strategy) Option {
	return func(conf *Configuration) {
		conf.customStrategy = s
	}
}

// WithGlobalHooks sets the global hooks for the provider. These are hooks that affect ALL providers. For hooks that
// target specific providers make sure to attach them to that provider directly, or use the WithProviderHook Option if
// that provider does not provide its own hook functionality
func WithGlobalHooks(hooks ...of.Hook) Option {
	return func(conf *Configuration) {
		conf.hooks = hooks
	}
}

// WithProviderHooks sets hooks that execute only for a specific provider. The providerName must match the unique provider
// name set during MultiProvider creation. This should only be used if you need hooks that execute around a specific
// provider, but that provider does not currently accept a way to set hooks. This option can be used multiple times using
// unique provider names. Using a provider name that is not known will cause an error.
func WithProviderHooks(providerName string, hooks ...of.Hook) Option {
	return func(conf *Configuration) {
		conf.providerHooks[providerName] = hooks
	}
}

// WithCustomComparatorFunc Set a custom comparison function along with an option to use it unconditionally with the
// EvaluateObject function of the provider. Primitive types are automatically handled, as well as objects that are
// comparable. If the EvaluateObject method is called and the result is not a comparable type, nor is this function set
// then the default value will be used and an error will be set in the resolution details. The [alwaysUse] parameter
// will only force using this function with the EvaluateObject. Setting this option without setting the strategy to
// StrategyComparison will have no effect on evaluation.
//
// If the nil pointer is set as the cmp function and the [alwaysUse] parameter is set a panic will occur.
func WithCustomComparatorFunc(cmpFunc Comparator, alwaysUse bool) Option {
	return func(conf *Configuration) {
		if alwaysUse && cmpFunc == nil {
			panic("invalid multiprovider state: comparison function cannot be nil with the alwaysUse flag set")
		}
		conf.customComparator = cmpFunc
		conf.alwaysUseCustomComparator = alwaysUse
	}
}

// Multiprovider Implementation

// AsNamedProviderSlice Converts the map into a slice of NamedProvider instances
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

	config := &Configuration{
		logger:        slog.New(DiscardHandler),
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
		providerStatus: make(map[string]of.State),
		globalHooks:    append(config.hooks, collectedHooks...),
	}

	var zeroDuration time.Duration
	if config.timeout == zeroDuration {
		config.timeout = 5 * time.Second
	}

	var strategy Strategy
	switch evaluationStrategy {
	case StrategyFirstMatch:
		strategy = NewFirstMatchStrategy(multiProvider.Providers())
	case StrategyFirstSuccess:
		strategy = NewFirstSuccessStrategy(multiProvider.Providers(), config.timeout)
	case StrategyComparison:
		strategy = NewComparisonStrategy(multiProvider.Providers(), config.fallbackProvider, nil, false)
	case StrategyCustom:
		if config.customStrategy != nil {
			strategy = config.customStrategy
		} else {
			return nil, fmt.Errorf("custom strategy must be set via an option if StrategyCustom is set")
		}
	default:
		return nil, fmt.Errorf("%s is an unknown evalutation strategy", strategy)
	}
	multiProvider.strategy = strategy

	return multiProvider, nil
}

// Providers Returns slice of providers wrapped in NamedProvider structs
func (mp *MultiProvider) Providers() []*NamedProvider {
	return toNamedProviderSlice(mp.providers)
}

// ProvidersByName Returns the internal ProviderMap of the MultiProvider
func (mp *MultiProvider) ProvidersByName() ProviderMap {
	return mp.providers
}

// EvaluationStrategy The current set strategy
func (mp *MultiProvider) EvaluationStrategy() string {
	return mp.strategy.Name()
}

// Metadata provides the name `multiprovider` and the names of each provider passed.
func (mp *MultiProvider) Metadata() of.Metadata {
	return mp.metadata
}

// Hooks returns a collection of.Hook defined by this provider
func (mp *MultiProvider) Hooks() []of.Hook {
	// Hooks that should be included with the provider
	return []of.Hook{}
}

// BooleanEvaluation returns a boolean flag
func (mp *MultiProvider) BooleanEvaluation(ctx context.Context, flag string, defaultValue bool, evalCtx of.FlattenedContext) of.BoolResolutionDetail {
	return mp.strategy.BooleanEvaluation(ctx, flag, defaultValue, evalCtx)
}

// StringEvaluation returns a string flag
func (mp *MultiProvider) StringEvaluation(ctx context.Context, flag string, defaultValue string, evalCtx of.FlattenedContext) of.StringResolutionDetail {
	return mp.strategy.StringEvaluation(ctx, flag, defaultValue, evalCtx)
}

// FloatEvaluation returns a float flag
func (mp *MultiProvider) FloatEvaluation(ctx context.Context, flag string, defaultValue float64, evalCtx of.FlattenedContext) of.FloatResolutionDetail {
	return mp.strategy.FloatEvaluation(ctx, flag, defaultValue, evalCtx)
}

// IntEvaluation returns an int flag
func (mp *MultiProvider) IntEvaluation(ctx context.Context, flag string, defaultValue int64, evalCtx of.FlattenedContext) of.IntResolutionDetail {
	return mp.strategy.IntEvaluation(ctx, flag, defaultValue, evalCtx)
}

// ObjectEvaluation returns an object flag
func (mp *MultiProvider) ObjectEvaluation(ctx context.Context, flag string, defaultValue interface{}, evalCtx of.FlattenedContext) of.InterfaceResolutionDetail {
	return mp.strategy.ObjectEvaluation(ctx, flag, defaultValue, evalCtx)
}

// Init will run the initialize method for all of provides and aggregate the errors.
func (mp *MultiProvider) Init(evalCtx of.EvaluationContext) error {
	var eg errgroup.Group
	// wrapper type used only for initialization of event listener workers
	type namedEventHandler struct {
		of.EventHandler
		name string
	}
	mp.logger.LogAttrs(context.Background(), slog.LevelDebug, "start initialization")
	mp.inboundEvents = make(chan namedEvent, len(mp.providers))
	handlers := make(chan namedEventHandler)
	for name, provider := range mp.providers {
		// Initialize each provider to not ready state. No locks required there are no workers running
		mp.providerStatus[name] = of.NotReadyState
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
				mp.providerStatusLock.Lock()
				defer mp.providerStatusLock.Unlock()
				mp.providerStatus[name] = of.ReadyState
			}
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		mp.setStatus(of.ErrorState)
		var pErr *ProviderError
		if errors.As(err, &pErr) {
			// Update provider status to error, no event needs to be emitted.
			// No locks needed as no workers are active at this point
			mp.providerStatus[pErr.ProviderName] = of.ErrorState
		} else {
			pErr = &ProviderError{
				Err:          err,
				ProviderName: "unknown",
			}
		}
		mp.outboundEvents <- of.Event{
			ProviderName: mp.Metadata().Name,
			EventType:    of.ProviderError,
			ProviderEventDetails: of.ProviderEventDetails{
				Message:     fmt.Sprintf("internal provider %s encountered an error during initialization: %+v", pErr.ProviderName, pErr.Err),
				FlagChanges: nil,
				EventMetadata: map[string]interface{}{
					MetadataProviderName:  pErr.ProviderName,
					MetadataInternalError: pErr.Error(),
				},
			},
		}
		return err
	}
	close(handlers)
	workerCtx, shutdownFunc := context.WithCancel(context.Background())
	for h := range handlers {
		go mp.startListening(workerCtx, h.name, h.EventHandler, &mp.workerGroup)
	}
	mp.shutdownFunc = shutdownFunc

	go func() {
		workerLogger := mp.logger.With(slog.String("multiprovider-worker", "event-forwarder-worker"))
		mp.workerGroup.Add(1)
		defer mp.workerGroup.Done()
		for e := range mp.inboundEvents {
			l := workerLogger.With(
				slog.String(MetadataProviderName, e.providerName),
				slog.String(MetadataProviderType, e.ProviderName),
			)
			l.LogAttrs(context.Background(), slog.LevelDebug, fmt.Sprintf("received %s event from provider", e.EventType))
			state := mp.updateProviderStateAndEvaluateTotalState(e, l)
			if state != mp.Status() {
				mp.setStatus(state)
				mp.outboundEvents <- e.Event
				l.LogAttrs(context.Background(), slog.LevelDebug, "forwarded state update event")
			} else {
				l.LogAttrs(context.Background(), slog.LevelDebug, "total state not updated, inbound event will not be emitted")
			}
		}
	}()

	mp.setStatus(of.ReadyState)
	mp.outboundEvents <- of.Event{
		ProviderName: mp.Metadata().Name,
		EventType:    of.ProviderReady,
		ProviderEventDetails: of.ProviderEventDetails{
			Message:     "all internal providers initialized successfully",
			FlagChanges: nil,
			EventMetadata: map[string]interface{}{
				MetadataProviderName: "all",
			},
		},
	}
	mp.initialized = true
	return nil
}

// startListening is intended to be
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

func (mp *MultiProvider) updateProviderStateAndEvaluateTotalState(e namedEvent, l *slog.Logger) of.State {
	if e.EventType == of.ProviderConfigChange {
		l.LogAttrs(context.Background(), slog.LevelDebug, fmt.Sprintf("ProviderConfigChange event: %s", e.Message))
		return mp.Status()
	}
	mp.providerStatusLock.Lock()
	defer mp.providerStatusLock.Unlock()
	logProviderState(l, e, mp.providerStatus[e.providerName])
	mp.providerStatus[e.providerName] = eventTypeToState[e.EventType]
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

// Shutdown Shuts down all internal providers
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

// Status the current state of the MultiProvider
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
