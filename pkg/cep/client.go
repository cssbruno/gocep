package cep

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"io"
	"net/http"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cssbruno/gocep/models"
	"github.com/cssbruno/gocep/pkg/util"
	"github.com/cssbruno/gocep/service/gocache"
	"golang.org/x/sync/singleflight"
)

// SearchStrategy controls provider execution behavior.
type SearchStrategy string

const (
	// SearchStrategyParallelFirstSuccess queries providers concurrently and returns the first success.
	SearchStrategyParallelFirstSuccess SearchStrategy = "parallel_first_success"
	// SearchStrategyOrderedFallback queries providers sequentially in policy order until one succeeds.
	SearchStrategyOrderedFallback SearchStrategy = "ordered_fallback"
)

// ProviderPolicy controls provider ordering, disabling, and timeouts.
type ProviderPolicy struct {
	Strategy SearchStrategy
	// PreferredSources are prioritized ahead of non-listed providers.
	PreferredSources []string
	// DisabledSources skips providers by source name.
	DisabledSources map[string]bool
	// SourceTimeouts sets per-provider request timeout by source name.
	SourceTimeouts map[string]time.Duration
}

// CacheEvent reports cache read/write outcomes.
type CacheEvent struct {
	CEP       string
	Operation string
	Hit       bool
	Error     error
}

// ProviderResultEvent reports provider request outcomes.
type ProviderResultEvent struct {
	CEP      string
	Source   string
	Duration time.Duration
	Success  bool
	Error    error
}

// Hooks allows callers to observe provider and cache events.
type Hooks struct {
	OnCacheEvent    func(CacheEvent)
	OnProviderEvent func(ProviderResultEvent)
}

// ClientOption customizes a Client at construction time.
type ClientOption func(*Client)

// Client performs CEP lookups with isolated configuration and state.
type Client struct {
	mu sync.RWMutex

	options       Options
	httpClient    *http.Client
	cacheProvider CacheProvider
	policy        ProviderPolicy
	hooks         Hooks

	endpoints       []models.Endpoint
	customEndpoints bool

	searchSingleflight singleflight.Group
}

type searchResult struct {
	JSON    string
	Address models.CEPAddress
	Err     error
}

type snapshot struct {
	options       Options
	httpClient    *http.Client
	cacheProvider CacheProvider
	policy        ProviderPolicy
	hooks         Hooks
	endpoints     []models.Endpoint
}

type globalCacheProvider struct{}

func (globalCacheProvider) SetAnyTTL(key string, value any, ttl time.Duration) bool {
	return gocache.SetAnyTTL(key, value, ttl)
}

func (globalCacheProvider) GetAny(key string) (any, bool) {
	return gocache.GetAny(key)
}

var defaultClient = NewClient()

// DefaultClient returns the package-level CEP client used by Search functions.
func DefaultClient() *Client {
	return defaultClient
}

// SetProviderPolicy updates provider policy for the package default client.
func SetProviderPolicy(policy ProviderPolicy) {
	defaultClient.SetProviderPolicy(policy)
}

// SetHooks updates observation hooks for the package default client.
func SetHooks(hooks Hooks) {
	defaultClient.SetHooks(hooks)
}

// NewClient creates an isolated CEP client.
func NewClient(opts ...ClientOption) *Client {
	c := &Client{
		options:       defaultOptions(),
		httpClient:    newDefaultHTTPClient(),
		cacheProvider: globalCacheProvider{},
		policy:        normalizeProviderPolicy(ProviderPolicy{}),
	}
	for _, opt := range opts {
		if opt != nil {
			opt(c)
		}
	}
	c.options = normalizeOptions(c.options)
	c.policy = normalizeProviderPolicy(c.policy)
	if c.httpClient == nil {
		c.httpClient = newDefaultHTTPClient()
	}
	return c
}

// WithOptions applies lookup options to a new client.
func WithOptions(opts Options) ClientOption {
	return func(c *Client) {
		c.options = normalizeOptions(opts)
	}
}

// WithHTTPClient sets a custom HTTP client for a new client.
func WithHTTPClient(client *http.Client) ClientOption {
	return func(c *Client) {
		if client == nil {
			c.httpClient = newDefaultHTTPClient()
			return
		}
		c.httpClient = client
	}
}

// WithCacheProvider sets a custom cache provider for a new client.
func WithCacheProvider(provider CacheProvider) ClientOption {
	return func(c *Client) {
		c.cacheProvider = provider
	}
}

// WithEndpoints sets custom provider endpoints for a new client.
func WithEndpoints(endpoints []models.Endpoint) ClientOption {
	return func(c *Client) {
		c.endpoints = cloneEndpointsLocal(endpoints)
		c.customEndpoints = true
	}
}

// WithProviderPolicy sets provider policy for a new client.
func WithProviderPolicy(policy ProviderPolicy) ClientOption {
	return func(c *Client) {
		c.policy = normalizeProviderPolicy(policy)
	}
}

// WithHooks sets observation hooks for a new client.
func WithHooks(hooks Hooks) ClientOption {
	return func(c *Client) {
		c.hooks = hooks
	}
}

// Options returns a copy of current client options.
func (c *Client) Options() Options {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.options
}

// SetOptions replaces client options.
func (c *Client) SetOptions(next Options) {
	c.mu.Lock()
	c.options = normalizeOptions(next)
	c.mu.Unlock()
}

// HTTPClient returns the client used for provider requests.
func (c *Client) HTTPClient() *http.Client {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.httpClient == nil {
		return newDefaultHTTPClient()
	}
	return c.httpClient
}

// SetHTTPClient sets the HTTP client used for provider requests.
func (c *Client) SetHTTPClient(client *http.Client) {
	c.mu.Lock()
	if client == nil {
		c.httpClient = newDefaultHTTPClient()
	} else {
		c.httpClient = client
	}
	c.mu.Unlock()
}

// SetCacheProvider sets the cache provider used by this client.
func (c *Client) SetCacheProvider(provider CacheProvider) {
	c.mu.Lock()
	c.cacheProvider = provider
	c.mu.Unlock()
}

// SetEndpoints replaces provider endpoints for this client only.
func (c *Client) SetEndpoints(endpoints []models.Endpoint) {
	c.mu.Lock()
	c.endpoints = cloneEndpointsLocal(endpoints)
	c.customEndpoints = true
	c.mu.Unlock()
}

// UseGlobalEndpoints resets the client to read endpoints from models.GetEndpoints.
func (c *Client) UseGlobalEndpoints() {
	c.mu.Lock()
	c.endpoints = nil
	c.customEndpoints = false
	c.mu.Unlock()
}

// SetProviderPolicy sets provider strategy and filtering rules.
func (c *Client) SetProviderPolicy(policy ProviderPolicy) {
	c.mu.Lock()
	c.policy = normalizeProviderPolicy(policy)
	c.mu.Unlock()
}

// SetHooks sets cache/provider observation hooks.
func (c *Client) SetHooks(hooks Hooks) {
	c.mu.Lock()
	c.hooks = hooks
	c.mu.Unlock()
}

// Search runs SearchContext with context.Background.
func (c *Client) Search(cep string) (string, models.CEPAddress, error) {
	return c.SearchContext(context.Background(), cep)
}

// SearchContext looks up a CEP using this client's configuration.
func (c *Client) SearchContext(ctx context.Context, cep string) (string, models.CEPAddress, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	cfg := c.snapshot()
	normalizedCEP, err := util.NormalizeCEP(cep)
	if err != nil {
		return cfg.options.DefaultJSON, models.CEPAddress{}, ErrInvalidCEP
	}

	if cachedJSON, cachedAddress, found := c.readCachedResult(cfg, normalizedCEP); found {
		return cachedJSON, cachedAddress, nil
	}

	endpoints := applyProviderPolicy(cfg.endpoints, cfg.policy)
	if len(endpoints) == 0 {
		return cfg.options.DefaultJSON, models.CEPAddress{}, ErrNotFound
	}

	value, _, _ := c.searchSingleflight.Do(normalizedCEP, func() (any, error) {
		res := c.searchFromProviders(ctx, cfg, normalizedCEP, endpoints)
		return res, nil
	})

	result, ok := value.(searchResult)
	if !ok {
		return cfg.options.DefaultJSON, models.CEPAddress{}, ErrNotFound
	}
	return result.JSON, result.Address, result.Err
}

func (c *Client) snapshot() snapshot {
	c.mu.RLock()
	defer c.mu.RUnlock()

	cfg := snapshot{
		options:       normalizeOptions(c.options),
		httpClient:    c.httpClient,
		cacheProvider: c.cacheProvider,
		policy:        normalizeProviderPolicy(c.policy),
		hooks:         c.hooks,
	}
	if cfg.httpClient == nil {
		cfg.httpClient = newDefaultHTTPClient()
	}
	if c.customEndpoints {
		cfg.endpoints = cloneEndpointsLocal(c.endpoints)
	} else {
		cfg.endpoints = models.GetEndpoints()
	}
	return cfg
}

func (c *Client) searchFromProviders(ctx context.Context, cfg snapshot, cep string, endpoints []models.Endpoint) searchResult {
	searchCtx, cancel := context.WithTimeout(ctx, cfg.options.SearchTimeout)
	defer cancel()

	switch cfg.policy.Strategy {
	case SearchStrategyOrderedFallback:
		return c.searchOrdered(searchCtx, cfg, cep, endpoints)
	default:
		return c.searchParallel(searchCtx, cancel, cfg, cep, endpoints)
	}
}

func (c *Client) searchOrdered(ctx context.Context, cfg snapshot, cep string, endpoints []models.Endpoint) searchResult {
	sawProviderTimeout := false

	for _, endpoint := range endpoints {
		select {
		case <-ctx.Done():
			return failureResult(ctx, cfg.options.DefaultJSON)
		default:
		}

		providerCtx, providerCancel := withProviderTimeout(ctx, cfg.policy, endpoint.Source)
		start := time.Now()
		result, err := c.queryEndpoint(providerCtx, cfg, cep, endpoint)
		providerCancel()
		c.emitProviderEvent(cfg.hooks, ProviderResultEvent{
			CEP:      cep,
			Source:   endpoint.Source,
			Duration: time.Since(start),
			Success:  err == nil,
			Error:    err,
		})
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				sawProviderTimeout = true
			}
			continue
		}
		jsonCep := string(result.Body)
		c.cacheSearchResult(cfg, cep, jsonCep, result.Address)
		return searchResult{JSON: jsonCep, Address: result.Address, Err: nil}
	}
	if sawProviderTimeout {
		return searchResult{JSON: cfg.options.DefaultJSON, Address: models.CEPAddress{}, Err: ErrTimeout}
	}
	return failureResult(ctx, cfg.options.DefaultJSON)
}

func (c *Client) searchParallel(ctx context.Context, cancel context.CancelFunc, cfg snapshot, cep string, endpoints []models.Endpoint) searchResult {
	results := make(chan Result, len(endpoints))
	var sawProviderTimeout atomic.Bool
	var wg sync.WaitGroup
	wg.Add(len(endpoints))

	for _, endpoint := range endpoints {
		endpoint := endpoint
		go func() {
			defer wg.Done()
			providerCtx, providerCancel := withProviderTimeout(ctx, cfg.policy, endpoint.Source)
			start := time.Now()
			result, err := c.queryEndpoint(providerCtx, cfg, cep, endpoint)
			providerCancel()
			c.emitProviderEvent(cfg.hooks, ProviderResultEvent{
				CEP:      cep,
				Source:   endpoint.Source,
				Duration: time.Since(start),
				Success:  err == nil,
				Error:    err,
			})
			if err != nil {
				if errors.Is(err, context.DeadlineExceeded) {
					sawProviderTimeout.Store(true)
				}
				return
			}

			select {
			case results <- result:
				cancel()
			case <-ctx.Done():
			}
		}()
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	for {
		select {
		case result, ok := <-results:
			if !ok {
				if sawProviderTimeout.Load() {
					return searchResult{JSON: cfg.options.DefaultJSON, Address: models.CEPAddress{}, Err: ErrTimeout}
				}
				return failureResult(ctx, cfg.options.DefaultJSON)
			}
			jsonCep := string(result.Body)
			c.cacheSearchResult(cfg, cep, jsonCep, result.Address)
			return searchResult{JSON: jsonCep, Address: result.Address, Err: nil}
		case <-ctx.Done():
			return failureResult(ctx, cfg.options.DefaultJSON)
		}
	}
}

func failureResult(ctx context.Context, defaultJSON string) searchResult {
	if err := ctx.Err(); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return searchResult{JSON: defaultJSON, Address: models.CEPAddress{}, Err: ErrTimeout}
		}
		return searchResult{JSON: defaultJSON, Address: models.CEPAddress{}, Err: err}
	}
	return searchResult{JSON: defaultJSON, Address: models.CEPAddress{}, Err: ErrNotFound}
}

func withProviderTimeout(ctx context.Context, policy ProviderPolicy, source string) (context.Context, context.CancelFunc) {
	timeout, ok := policy.SourceTimeouts[source]
	if !ok || timeout <= 0 {
		return ctx, func() {}
	}
	return context.WithTimeout(ctx, timeout)
}

func applyProviderPolicy(endpoints []models.Endpoint, policy ProviderPolicy) []models.Endpoint {
	filtered := make([]models.Endpoint, 0, len(endpoints))
	for _, endpoint := range endpoints {
		if policy.DisabledSources[endpoint.Source] {
			continue
		}
		filtered = append(filtered, endpoint)
	}

	if len(policy.PreferredSources) == 0 || len(filtered) < 2 {
		return filtered
	}

	rank := make(map[string]int, len(policy.PreferredSources))
	for i, source := range policy.PreferredSources {
		rank[source] = i
	}
	maxRank := len(policy.PreferredSources) + 1

	sort.SliceStable(filtered, func(i, j int) bool {
		ri, ok := rank[filtered[i].Source]
		if !ok {
			ri = maxRank
		}
		rj, ok := rank[filtered[j].Source]
		if !ok {
			rj = maxRank
		}
		return ri < rj
	})

	return filtered
}

func (c *Client) queryEndpoint(ctx context.Context, cfg snapshot, cep string, endpoint models.Endpoint) (Result, error) {
	if endpoint.Source == models.SourceCorreio {
		return queryCorreioEndpoint(ctx, cfg.httpClient, cfg.options.MaxProviderBody, cep, endpoint)
	}
	return queryJSONEndpoint(ctx, cfg.httpClient, cfg.options.MaxProviderBody, cep, endpoint)
}

func queryJSONEndpoint(ctx context.Context, httpClient *http.Client, maxBody int64, cep string, endpoint models.Endpoint) (Result, error) {
	queryCEP := cep
	if endpoint.Source == models.SourceCdnApiCep && len(queryCEP) > 7 {
		queryCEP = addHyphen(queryCEP)
	}

	url := strings.Replace(endpoint.URL, "%s", queryCEP, 1)
	req, err := http.NewRequestWithContext(ctx, endpoint.Method, url, nil)
	if err != nil {
		return Result{}, err
	}

	response, err := executeRequestWithClient(httpClient, req)
	if err != nil {
		return Result{}, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(io.LimitReader(response.Body, maxBody+1))
	if err != nil {
		return Result{}, err
	}
	if int64(len(body)) > maxBody {
		return Result{}, ErrNotFound
	}
	if len(body) == 0 {
		return Result{}, ErrNotFound
	}

	address, err := ParseCEPAddress(endpoint.Source, body)
	if err != nil || !isCompleteAddress(address) {
		if err != nil {
			return Result{}, err
		}
		return Result{}, ErrNotFound
	}

	normalizedJSON, err := marshalAddressJSON(address)
	if err != nil {
		return Result{}, err
	}

	return Result{Body: normalizedJSON, Address: address}, nil
}

func queryCorreioEndpoint(ctx context.Context, httpClient *http.Client, maxBody int64, cep string, endpoint models.Endpoint) (Result, error) {
	payload := strings.Replace(endpoint.Body, "%s", cep, 1)
	req, err := http.NewRequestWithContext(ctx, endpoint.Method, endpoint.URL, strings.NewReader(payload))
	if err != nil {
		return Result{}, err
	}
	req.Header.Set("Content-Type", "text/xml; charset=utf-8")

	response, err := executeRequestWithClient(httpClient, req)
	if err != nil {
		return Result{}, err
	}
	defer response.Body.Close()

	rawBody, err := io.ReadAll(io.LimitReader(response.Body, maxBody+1))
	if err != nil {
		return Result{}, err
	}
	if int64(len(rawBody)) > maxBody {
		return Result{}, ErrNotFound
	}

	correio := new(models.Correio)
	if err := xml.Unmarshal(rawBody, correio); err != nil {
		return Result{}, err
	}

	responseAddress := correio.Body.LookupCEPResponse.Return
	address := models.CEPAddress{
		City:         responseAddress.City,
		StateCode:    responseAddress.StateCode,
		Street:       responseAddress.Address,
		Neighborhood: responseAddress.Neighborhood,
	}
	if !isCompleteAddress(address) {
		return Result{}, ErrNotFound
	}

	normalizedJSON, err := marshalAddressJSON(address)
	if err != nil {
		return Result{}, err
	}

	return Result{Body: normalizedJSON, Address: address}, nil
}

func (c *Client) readCachedResult(cfg snapshot, cep string) (jsonCep string, address models.CEPAddress, found bool) {
	if !cfg.options.CacheEnabled || cfg.cacheProvider == nil {
		return "", models.CEPAddress{}, false
	}

	value, ok := cfg.cacheProvider.GetAny(cep)
	if !ok {
		c.emitCacheEvent(cfg.hooks, CacheEvent{CEP: cep, Operation: "read", Hit: false})
		return "", models.CEPAddress{}, false
	}

	switch cached := value.(type) {
	case cachedResult:
		if cached.JSON == "" || !isCompleteAddress(cached.Address) {
			c.emitCacheEvent(cfg.hooks, CacheEvent{CEP: cep, Operation: "read", Hit: false, Error: ErrNotFound})
			return "", models.CEPAddress{}, false
		}
		c.emitCacheEvent(cfg.hooks, CacheEvent{CEP: cep, Operation: "read", Hit: true})
		return cached.JSON, cached.Address, true
	case string:
		var parsedAddress models.CEPAddress
		if err := json.Unmarshal([]byte(cached), &parsedAddress); err != nil {
			c.emitCacheEvent(cfg.hooks, CacheEvent{CEP: cep, Operation: "read", Hit: false, Error: err})
			return "", models.CEPAddress{}, false
		}
		if !isCompleteAddress(parsedAddress) {
			c.emitCacheEvent(cfg.hooks, CacheEvent{CEP: cep, Operation: "read", Hit: false, Error: ErrNotFound})
			return "", models.CEPAddress{}, false
		}
		_ = cfg.cacheProvider.SetAnyTTL(cep, cachedResult{
			JSON:    cached,
			Address: parsedAddress,
		}, cfg.options.CacheTTL)
		c.emitCacheEvent(cfg.hooks, CacheEvent{CEP: cep, Operation: "read", Hit: true})
		return cached, parsedAddress, true
	default:
		c.emitCacheEvent(cfg.hooks, CacheEvent{CEP: cep, Operation: "read", Hit: false, Error: ErrNotFound})
		return "", models.CEPAddress{}, false
	}
}

func (c *Client) cacheSearchResult(cfg snapshot, cep, jsonCep string, address models.CEPAddress) {
	if !cfg.options.CacheEnabled || cfg.cacheProvider == nil || !isCompleteAddress(address) {
		return
	}

	ok := cfg.cacheProvider.SetAnyTTL(cep, cachedResult{
		JSON:    jsonCep,
		Address: address,
	}, cfg.options.CacheTTL)
	c.emitCacheEvent(cfg.hooks, CacheEvent{CEP: cep, Operation: "write", Hit: ok})
}

func (c *Client) emitCacheEvent(hooks Hooks, event CacheEvent) {
	if hooks.OnCacheEvent == nil {
		return
	}
	defer func() { _ = recover() }()
	hooks.OnCacheEvent(event)
}

func (c *Client) emitProviderEvent(hooks Hooks, event ProviderResultEvent) {
	if hooks.OnProviderEvent == nil {
		return
	}
	defer func() { _ = recover() }()
	hooks.OnProviderEvent(event)
}

func normalizeProviderPolicy(policy ProviderPolicy) ProviderPolicy {
	out := ProviderPolicy{
		Strategy:         policy.Strategy,
		PreferredSources: cloneStringSlice(policy.PreferredSources),
		DisabledSources:  make(map[string]bool, len(policy.DisabledSources)),
		SourceTimeouts:   make(map[string]time.Duration, len(policy.SourceTimeouts)),
	}
	if out.Strategy == "" {
		out.Strategy = SearchStrategyParallelFirstSuccess
	}
	for source, disabled := range policy.DisabledSources {
		if disabled {
			out.DisabledSources[source] = true
		}
	}
	for source, timeout := range policy.SourceTimeouts {
		if timeout > 0 {
			out.SourceTimeouts[source] = timeout
		}
	}
	return out
}

func cloneStringSlice(in []string) []string {
	if len(in) == 0 {
		return nil
	}
	out := make([]string, len(in))
	copy(out, in)
	return out
}

func cloneEndpointsLocal(in []models.Endpoint) []models.Endpoint {
	if len(in) == 0 {
		return nil
	}
	out := make([]models.Endpoint, len(in))
	copy(out, in)
	return out
}
