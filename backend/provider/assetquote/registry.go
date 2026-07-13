package assetquote

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

type RegistryConfig struct {
	RequestTimeout   time.Duration
	MaxConcurrency   int
	MaxRetries       int
	RetryBackoff     time.Duration
	CacheTTL         time.Duration
	FXCacheTTL       time.Duration
	StaleTTL         time.Duration
	MinRequestGap    time.Duration
	FailureThreshold int
	CircuitOpenFor   time.Duration
}

func DefaultRegistryConfig() RegistryConfig {
	return RegistryConfig{
		RequestTimeout: 12 * time.Second, MaxConcurrency: 6, MaxRetries: 1, RetryBackoff: 150 * time.Millisecond,
		CacheTTL: 20 * time.Second, FXCacheTTL: 6 * time.Hour, StaleTTL: 24 * time.Hour, MinRequestGap: 120 * time.Millisecond,
		FailureThreshold: 3, CircuitOpenFor: 30 * time.Second,
	}
}

type cacheEntry struct {
	quote     Quote
	expiresAt time.Time
	staleAt   time.Time
}

type providerRuntime struct {
	provider QuoteProvider
	mu       sync.Mutex
	health   ProviderHealth
	nextCall time.Time
}

type Registry struct {
	config    RegistryConfig
	mu        sync.RWMutex
	providers []*providerRuntime
	cache     map[string]cacheEntry
	now       func() time.Time
}

func NewRegistry(config RegistryConfig, providers ...QuoteProvider) *Registry {
	config = normalizeRegistryConfig(config)
	r := &Registry{config: config, cache: make(map[string]cacheEntry), now: time.Now}
	for _, provider := range providers {
		r.Register(provider)
	}
	return r
}

func NewDefaultRegistry(config RegistryConfig) *Registry {
	return NewRegistry(config,
		NewEastmoneyProvider(nil),
		NewFundEstimateProvider(nil),
		NewFundNAVProvider(nil),
		NewSinaFundProvider(nil),
		NewSinaProvider(nil),
		NewTencentProvider(nil),
		NewSinaReferenceProvider(nil),
		NewSinaFXProvider(nil),
		NewFrankfurterProvider(nil),
		NewExchangeRateProvider(nil),
		NewBinanceProvider(nil),
		NewCoinGeckoProvider(nil),
	)
}

func normalizeRegistryConfig(config RegistryConfig) RegistryConfig {
	defaults := DefaultRegistryConfig()
	if config == (RegistryConfig{}) {
		return defaults
	}
	if config.RequestTimeout <= 0 {
		config.RequestTimeout = defaults.RequestTimeout
	}
	if config.MaxConcurrency <= 0 {
		config.MaxConcurrency = defaults.MaxConcurrency
	}
	if config.MaxRetries < 0 {
		config.MaxRetries = 0
	}
	if config.RetryBackoff <= 0 {
		config.RetryBackoff = defaults.RetryBackoff
	}
	if config.CacheTTL <= 0 {
		config.CacheTTL = defaults.CacheTTL
	}
	if config.FXCacheTTL <= 0 {
		config.FXCacheTTL = defaults.FXCacheTTL
	}
	if config.StaleTTL <= 0 {
		config.StaleTTL = defaults.StaleTTL
	}
	if config.MinRequestGap < 0 {
		config.MinRequestGap = 0
	}
	if config.FailureThreshold <= 0 {
		config.FailureThreshold = defaults.FailureThreshold
	}
	if config.CircuitOpenFor <= 0 {
		config.CircuitOpenFor = defaults.CircuitOpenFor
	}
	return config
}

func (r *Registry) Register(provider QuoteProvider) {
	if provider == nil || strings.TrimSpace(provider.Name()) == "" {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	name := strings.ToLower(strings.TrimSpace(provider.Name()))
	for index, runtime := range r.providers {
		if strings.ToLower(runtime.provider.Name()) == name {
			r.providers[index] = newProviderRuntime(provider)
			return
		}
	}
	r.providers = append(r.providers, newProviderRuntime(provider))
}

func newProviderRuntime(provider QuoteProvider) *providerRuntime {
	return &providerRuntime{provider: provider, health: ProviderHealth{Name: provider.Name(), State: HealthUnknown}}
}

func (r *Registry) Quote(ctx context.Context, request QuoteRequest) (QuoteResult, error) {
	request = request.Normalized()
	if err := request.Validate(); err != nil {
		return QuoteResult{}, err
	}
	now := r.now()
	if entry, ok := r.cached(request.CacheKey(), now, false); ok {
		health := r.healthFor(entry.quote.Provider)
		return QuoteResult{Quote: entry.quote, Health: health, Cached: true}, nil
	}
	runtimes := r.matchingProviders(request)
	if len(runtimes) == 0 {
		return r.staleOrError(request, nil, ErrUnsupported)
	}
	var providerErrors []error
	totalAttempts := 0
	for _, runtime := range runtimes {
		quote, attempts, err := r.fetch(ctx, runtime, request)
		totalAttempts += attempts
		if err != nil {
			providerErrors = append(providerErrors, err)
			continue
		}
		quote.AssetClass, quote.Symbol = request.AssetClass, request.Symbol
		if quote.Name == "" {
			quote.Name = request.Name
		}
		if quote.Exchange == "" {
			quote.Exchange = request.Exchange
		}
		if quote.Currency == "" {
			quote.Currency = request.Currency
		}
		quote.Provider, quote.Stale = runtime.provider.Name(), false
		if quote.FetchedAt.IsZero() {
			quote.FetchedAt = r.now()
		}
		if quote.AsOf.IsZero() {
			quote.AsOf = quote.FetchedAt
		}
		r.store(request.CacheKey(), quote)
		return QuoteResult{Quote: quote, Health: r.healthFor(runtime.provider.Name()), Attempts: totalAttempts}, nil
	}
	return r.staleOrError(request, providerErrors, errors.Join(providerErrors...))
}

func (r *Registry) Quotes(ctx context.Context, requests []QuoteRequest) ([]QuoteResult, error) {
	results := make([]QuoteResult, len(requests))
	errorsByIndex := make([]error, len(requests))
	semaphore := make(chan struct{}, r.config.MaxConcurrency)
	var waitGroup sync.WaitGroup
	for index, request := range requests {
		waitGroup.Add(1)
		go func(index int, request QuoteRequest) {
			defer waitGroup.Done()
			select {
			case semaphore <- struct{}{}:
				defer func() { <-semaphore }()
			case <-ctx.Done():
				errorsByIndex[index] = ctx.Err()
				return
			}
			result, err := r.Quote(ctx, request)
			if err != nil {
				errorsByIndex[index] = fmt.Errorf("%s: %w", request.Symbol, err)
				return
			}
			results[index] = result
		}(index, request)
	}
	waitGroup.Wait()
	quoteErrors := make([]error, 0)
	for _, err := range errorsByIndex {
		if err != nil {
			quoteErrors = append(quoteErrors, err)
		}
	}
	return results, errors.Join(quoteErrors...)
}

func (r *Registry) matchingProviders(request QuoteRequest) []*providerRuntime {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]*providerRuntime, 0, len(r.providers))
	for _, runtime := range r.providers {
		if runtime.provider.Supports(request) {
			result = append(result, runtime)
		}
	}
	if request.Provider != "" {
		sort.SliceStable(result, func(i, j int) bool {
			leftPreferred := strings.EqualFold(request.Provider, result[i].provider.Name())
			rightPreferred := strings.EqualFold(request.Provider, result[j].provider.Name())
			return leftPreferred && !rightPreferred
		})
	}
	return result
}

func (r *Registry) fetch(ctx context.Context, runtime *providerRuntime, request QuoteRequest) (Quote, int, error) {
	if health, open := runtime.circuitState(r.now()); open {
		return Quote{}, 0, fmt.Errorf("%s circuit open until %s", runtime.provider.Name(), health.CircuitOpenUntil.Format(time.RFC3339))
	}
	var lastErr error
	attempts := 0
	for attempt := 0; attempt <= r.config.MaxRetries; attempt++ {
		if attempt > 0 {
			if err := waitContext(ctx, r.config.RetryBackoff*time.Duration(attempt)); err != nil {
				return Quote{}, attempt, err
			}
		}
		if err := runtime.waitRateLimit(ctx, r.now, r.config.MinRequestGap); err != nil {
			return Quote{}, attempt, err
		}
		requestCtx, cancel := context.WithTimeout(ctx, r.config.RequestTimeout)
		startedAt := r.now()
		quotes, err := runtime.provider.FetchQuotes(requestCtx, []QuoteRequest{request})
		cancel()
		attempts++
		if err == nil && len(quotes) != 1 {
			err = &HTTPError{Provider: runtime.provider.Name(), Message: "provider returned an unexpected quote count"}
		}
		if err == nil {
			finishedAt := r.now()
			runtime.recordSuccess(finishedAt, finishedAt.Sub(startedAt))
			return quotes[0], attempts, nil
		}
		lastErr = err
		if !isRetryable(err) {
			break
		}
	}
	runtime.recordFailure(r.now(), lastErr, r.config.FailureThreshold, r.config.CircuitOpenFor)
	return Quote{}, attempts, fmt.Errorf("%s: %w", runtime.provider.Name(), lastErr)
}

func (r *Registry) staleOrError(request QuoteRequest, providerErrors []error, err error) (QuoteResult, error) {
	if entry, ok := r.cached(request.CacheKey(), r.now(), true); ok {
		entry.quote.Stale = true
		health := r.healthFor(entry.quote.Provider)
		r.incrementStale(entry.quote.Provider)
		health.StaleHits++
		return QuoteResult{Quote: entry.quote, Health: health, Cached: true, Stale: true}, nil
	}
	if err == nil {
		err = ErrUnsupported
	}
	if len(providerErrors) > 0 {
		return QuoteResult{}, fmt.Errorf("all quote providers failed: %w", err)
	}
	return QuoteResult{}, err
}

func (r *Registry) store(key string, quote Quote) {
	now := r.now()
	r.mu.Lock()
	ttl := r.config.CacheTTL
	if quote.AssetClass == AssetFX {
		ttl = r.config.FXCacheTTL
	}
	r.cache[key] = cacheEntry{quote: quote, expiresAt: now.Add(ttl), staleAt: now.Add(r.config.StaleTTL)}
	r.mu.Unlock()
}

func (r *Registry) cached(key string, now time.Time, allowStale bool) (cacheEntry, bool) {
	r.mu.RLock()
	entry, ok := r.cache[key]
	r.mu.RUnlock()
	if !ok || now.After(entry.staleAt) {
		return cacheEntry{}, false
	}
	if !allowStale && now.After(entry.expiresAt) {
		return cacheEntry{}, false
	}
	return entry, true
}

func (r *Registry) Health() []ProviderHealth {
	r.mu.RLock()
	runtimes := append([]*providerRuntime(nil), r.providers...)
	r.mu.RUnlock()
	health := make([]ProviderHealth, 0, len(runtimes))
	for _, runtime := range runtimes {
		snapshot, _ := runtime.circuitState(r.now())
		health = append(health, snapshot)
	}
	sort.Slice(health, func(i, j int) bool { return health[i].Name < health[j].Name })
	return health
}

func (r *Registry) healthFor(name string) ProviderHealth {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, runtime := range r.providers {
		if strings.EqualFold(runtime.provider.Name(), name) {
			health, _ := runtime.circuitState(r.now())
			return health
		}
	}
	return ProviderHealth{Name: name, State: HealthUnknown}
}

func (r *Registry) incrementStale(name string) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, runtime := range r.providers {
		if strings.EqualFold(runtime.provider.Name(), name) {
			runtime.mu.Lock()
			runtime.health.StaleHits++
			runtime.mu.Unlock()
			return
		}
	}
}

func (p *providerRuntime) circuitState(now time.Time) (ProviderHealth, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.health.State == HealthOpen && !now.Before(p.health.CircuitOpenUntil) {
		p.health.State, p.health.CircuitOpenUntil = HealthDegraded, time.Time{}
	}
	return p.health, p.health.State == HealthOpen
}

func (p *providerRuntime) waitRateLimit(ctx context.Context, now func() time.Time, gap time.Duration) error {
	p.mu.Lock()
	wait := time.Until(p.nextCall)
	if wait < 0 {
		wait = 0
	}
	base := now().Add(wait)
	p.nextCall = base.Add(gap)
	p.mu.Unlock()
	return waitContext(ctx, wait)
}

func (p *providerRuntime) recordSuccess(now time.Time, latency time.Duration) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.health.State, p.health.ConsecutiveFailures, p.health.LastError = HealthHealthy, 0, ""
	p.health.CircuitOpenUntil, p.health.LastSuccessAt = time.Time{}, now
	p.health.LastLatencyMs = latency.Milliseconds()
}

func (p *providerRuntime) recordFailure(now time.Time, err error, threshold int, openFor time.Duration) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.health.ConsecutiveFailures++
	p.health.LastFailureAt = now
	if err != nil {
		p.health.LastError = err.Error()
	}
	p.health.State = HealthDegraded
	if p.health.ConsecutiveFailures >= threshold {
		p.health.State, p.health.CircuitOpenUntil = HealthOpen, now.Add(openFor)
	}
}

func waitContext(ctx context.Context, duration time.Duration) error {
	if duration <= 0 {
		return nil
	}
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
