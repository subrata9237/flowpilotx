package client

import (
	"fmt"
	"time"
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/balancer/roundrobin"
)

// RequestOptions holds configuration options for individual gRPC requests
type RequestOptions struct {
	Timeout      time.Duration
	RetryPolicy  *RetryPolicy
}

// DefaultRequestOptions returns default options for individual requests
func DefaultRequestOptions() *RequestOptions {
	return &RequestOptions{
		Timeout:     30 * time.Second,  // Default request timeout
		RetryPolicy: DefaultRetryPolicy(),
	}
}

// WithRequestTimeout sets the timeout for this specific request
func (o *RequestOptions) WithRequestTimeout(timeout time.Duration) *RequestOptions {
	o.Timeout = timeout
	return o
}

// WithRequestRetryPolicy sets the retry policy for this specific request
func (o *RequestOptions) WithRequestRetryPolicy(policy *RetryPolicy) *RequestOptions {
	o.RetryPolicy = policy
	return o
}

// CreateRequestContext creates a context with the request-specific timeout
func (o *RequestOptions) CreateRequestContext(parent context.Context) (context.Context, context.CancelFunc) {
	if o.Timeout > 0 {
		return context.WithTimeout(parent, o.Timeout)
	}
	return context.WithCancel(parent)
}

// ClientOptions holds configuration options for gRPC clients
type ClientOptions struct {
	Addresses       []string        // Multiple server addresses for load balancing
	ConnTimeout     time.Duration   // Connection timeout
	RequestTimeout  time.Duration   // Default request timeout
	DialOptions     []grpc.DialOption
	UseInsecure     bool
	BlockingDial    bool
	RetryPolicy     *RetryPolicy    // Default retry policy
	LoadBalancing   *LoadBalancingConfig
}

// LoadBalancingConfig holds load balancing configuration
type LoadBalancingConfig struct {
	Enabled     bool
	Policy      string   // "round_robin", "pick_first", etc.
	HealthCheck bool     // Whether to enable health checking
}

// RetryPolicy defines the retry behavior for client operations
type RetryPolicy struct {
	MaxAttempts      int
	InitialBackoff   time.Duration
	MaxBackoff       time.Duration
	BackoffMultiplier float64
	NumRetries       int           // Tracks the number of retries performed
	RetryWindow      time.Duration // Time window for retries
	LastRetryTime    time.Time     // Last time a retry was attempted
	MinRetryInterval time.Duration // Minimum time between retries
}

// DefaultClientOptions returns the default client configuration
func DefaultClientOptions() *ClientOptions {
	return &ClientOptions{
		ConnTimeout:    5 * time.Second,
		RequestTimeout: 30 * time.Second,
		UseInsecure:    true,
		BlockingDial:   true,
		RetryPolicy:    DefaultRetryPolicy(),
		LoadBalancing:  DefaultLoadBalancingConfig(),
	}
}

// DefaultLoadBalancingConfig returns the default load balancing configuration
func DefaultLoadBalancingConfig() *LoadBalancingConfig {
	return &LoadBalancingConfig{
		Enabled:     false,
		Policy:      roundrobin.Name,
		HealthCheck: true,
	}
}

// DefaultRetryPolicy returns the default retry configuration
func DefaultRetryPolicy() *RetryPolicy {
	return &RetryPolicy{
		MaxAttempts:      3,
		InitialBackoff:   100 * time.Millisecond,
		MaxBackoff:       2 * time.Second,
		BackoffMultiplier: 1.5,
		NumRetries:       0,
		RetryWindow:      5 * time.Minute,    // Default 5 minute window for retries
		MinRetryInterval: time.Second,        // Minimum 1 second between retries
		LastRetryTime:    time.Time{},        // Zero time
	}
}

// WithAddresses sets multiple server addresses for load balancing
func (o *ClientOptions) WithAddresses(addresses []string) *ClientOptions {
	o.Addresses = addresses
	if len(addresses) > 1 {
		o.LoadBalancing.Enabled = true
	}
	return o
}

// WithAddress sets a single server address
func (o *ClientOptions) WithAddress(address string) *ClientOptions {
	o.Addresses = []string{address}
	o.LoadBalancing.Enabled = false
	return o
}

// WithConnTimeout sets the connection timeout
func (o *ClientOptions) WithConnTimeout(timeout time.Duration) *ClientOptions {
	o.ConnTimeout = timeout
	return o
}

// WithRequestTimeout sets the individual request timeout
func (o *ClientOptions) WithRequestTimeout(timeout time.Duration) *ClientOptions {
	o.RequestTimeout = timeout
	return o
}

// WithInsecure sets whether to use insecure credentials
func (o *ClientOptions) WithInsecure(useInsecure bool) *ClientOptions {
	o.UseInsecure = useInsecure
	return o
}

// WithBlockingDial sets whether to use blocking dial
func (o *ClientOptions) WithBlockingDial(blockingDial bool) *ClientOptions {
	o.BlockingDial = blockingDial
	return o
}

// WithRetryPolicy sets the retry policy
func (o *ClientOptions) WithRetryPolicy(policy *RetryPolicy) *ClientOptions {
	o.RetryPolicy = policy
	return o
}

// WithLoadBalancing configures load balancing
func (o *ClientOptions) WithLoadBalancing(enabled bool, policy string) *ClientOptions {
	o.LoadBalancing.Enabled = enabled
	if policy != "" {
		o.LoadBalancing.Policy = policy
	}
	return o
}

// WithHealthCheck enables or disables health checking for load balancing
func (o *ClientOptions) WithHealthCheck(enabled bool) *ClientOptions {
	o.LoadBalancing.HealthCheck = enabled
	return o
}

// WithCustomDialOptions adds custom gRPC dial options
func (o *ClientOptions) WithCustomDialOptions(opts ...grpc.DialOption) *ClientOptions {
	o.DialOptions = append(o.DialOptions, opts...)
	return o
}

// BuildDialOptions constructs the final set of dial options
func (o *ClientOptions) BuildDialOptions() []grpc.DialOption {
	opts := make([]grpc.DialOption, 0)

	// Add custom options first
	opts = append(opts, o.DialOptions...)

	// Add standard options
	if o.UseInsecure {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}
	if o.BlockingDial {
		opts = append(opts, grpc.WithBlock())
	}
	if o.ConnTimeout > 0 {
		opts = append(opts, grpc.WithTimeout(o.ConnTimeout))
	}

	// Configure load balancing if enabled
	if o.LoadBalancing.Enabled {
		opts = append(opts, grpc.WithDefaultServiceConfig(
			fmt.Sprintf(`{"loadBalancingPolicy": "%s"}`, o.LoadBalancing.Policy),
		))
	}

	return opts
}

// WithRetryWindow sets the time window for retries
func (r *RetryPolicy) WithRetryWindow(window time.Duration) *RetryPolicy {
	r.RetryWindow = window
	return r
}

// WithMinRetryInterval sets the minimum interval between retries
func (r *RetryPolicy) WithMinRetryInterval(interval time.Duration) *RetryPolicy {
	r.MinRetryInterval = interval
	return r
}

// CanRetryNow checks if a retry can be performed at the current time
func (r *RetryPolicy) CanRetryNow() bool {
	if !r.HasRetriesRemaining() {
		return false
	}

	now := time.Now()

	// If this is the first retry attempt
	if r.LastRetryTime.IsZero() {
		return true
	}

	// Check if we're within the retry window
	if r.RetryWindow > 0 {
		windowStart := r.LastRetryTime.Add(-r.RetryWindow)
		if now.Before(windowStart) {
			return false
		}
	}

	// Check if minimum interval has elapsed
	return now.Sub(r.LastRetryTime) >= r.MinRetryInterval
}

// IncrementRetries increments the retry counter and updates the last retry time
func (r *RetryPolicy) IncrementRetries() {
	r.NumRetries++
	r.LastRetryTime = time.Now()
}

// GetNextRetryDelay calculates the next retry delay using exponential backoff
func (r *RetryPolicy) GetNextRetryDelay() time.Duration {
	if r.NumRetries == 0 {
		return r.InitialBackoff
	}

	// Calculate exponential backoff
	backoff := float64(r.InitialBackoff)
	for i := 0; i < r.NumRetries; i++ {
		backoff *= r.BackoffMultiplier
	}

	delay := time.Duration(backoff)
	if delay > r.MaxBackoff {
		delay = r.MaxBackoff
	}

	// Ensure we don't retry faster than MinRetryInterval
	if delay < r.MinRetryInterval {
		delay = r.MinRetryInterval
	}

	return delay
}

// ResetRetries resets the retry counter and last retry time
func (r *RetryPolicy) ResetRetries() {
	r.NumRetries = 0
	r.LastRetryTime = time.Time{}
}

// GetRetryCount returns the current number of retries
func (r *RetryPolicy) GetRetryCount() int {
	return r.NumRetries
}

// HasRetriesRemaining checks if more retries are available
func (r *RetryPolicy) HasRetriesRemaining() bool {
	return r.NumRetries < r.MaxAttempts
}

// NewRequestOptions creates a new RequestOptions instance with client defaults
func (o *ClientOptions) NewRequestOptions() *RequestOptions {
	return &RequestOptions{
		Timeout:     o.RequestTimeout,
		RetryPolicy: &RetryPolicy{
			MaxAttempts:      o.RetryPolicy.MaxAttempts,
			InitialBackoff:   o.RetryPolicy.InitialBackoff,
			MaxBackoff:       o.RetryPolicy.MaxBackoff,
			BackoffMultiplier: o.RetryPolicy.BackoffMultiplier,
			NumRetries:       0,
			RetryWindow:      5 * time.Minute,    // Default 5 minute window for retries
			MinRetryInterval: time.Second,        // Minimum 1 second between retries
			LastRetryTime:    time.Time{},        // Zero time
		},
	}
} 