package dnsresolver

import (
	"context"
	"net"
	"sync"
	"time"

	"gitlab.com/gitlab-org/gitaly/v16/internal/backoff"
	"gitlab.com/gitlab-org/gitaly/v16/internal/log"
	"gitlab.com/gitlab-org/gitaly/v16/internal/structerr"
	"google.golang.org/grpc/resolver"
)

type dnsResolver struct {
	logger log.Logger
	retry  backoff.Strategy

	ctx           context.Context
	cancel        context.CancelFunc
	cc            resolver.ClientConn
	host          string
	port          string
	refreshRate   time.Duration
	lookupTimeout time.Duration
	lookup        dnsLookuper
	reqs          chan struct{}
	wg            sync.WaitGroup
}

var defaultLookupTimeout = 15 * time.Second

type dnsLookuper interface {
	LookupHost(context.Context, string) ([]string, error)
}

// ResolveNow signals the resolver to perform a DNS resolution immediately. This method returns
// without waiting for the result. The resolver treats this as a hint rather than a command. The
// client connection receives the resolution result asynchronously via `clientconn.UpdateState`
// This method also skip resolver caching because it's likely the client calls this method after
// encounter an error with recent subchannels.
func (d *dnsResolver) ResolveNow(resolver.ResolveNowOptions) {
	select {
	case d.reqs <- struct{}{}:
	default:
	}
}

// Close cancels all activities of this dns resolver. It waits until the watch goroutine exits.
func (d *dnsResolver) Close() {
	d.cancel()
	d.wg.Wait()
}

func (d *dnsResolver) watch() {
	defer d.wg.Done()
	d.logger.Info("dns resolver: started")
	defer d.logger.Info("dns resolver: stopped")

	// Exponential retry after failed to resolve or client connection failed to update its state
	var retries uint
	for {
		state, err := d.resolve()
		if err != nil {
			d.logger.WithField("dns.retries", retries).WithField("dns.error", err).Error(
				"dns resolver: fail to lookup dns")
			d.cc.ReportError(err)
		} else {
			err = d.updateState(state)
		}

		var timer *time.Timer
		if err == nil {
			timer = time.NewTimer(d.refreshRate)
			retries = 0
		} else {
			timer = time.NewTimer(d.retry.Backoff(retries))
			retries++
		}

		select {
		case <-d.ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
			// Refresh timer expires, issue another DNS lookup.
			d.logger.Debug("dns resolver: refreshing")
			continue
		case <-d.reqs:
			// If the resolver is requested to resolve now, force notify the client
			// connection. Typically, client connection contacts the resolver when any
			// of the subchannels change its connectivity state.
			timer.Stop()
			d.logger.Debug("dns resolver: handle ResolveNow request")
		}
	}
}

func (d *dnsResolver) updateState(state *resolver.State) error {
	d.logger.WithField("dns.state", state).Info("dns resolver: updating state")
	return d.cc.UpdateState(*state)
}

func (d *dnsResolver) resolve() (*resolver.State, error) {
	timeout := d.lookupTimeout
	if timeout == 0 {
		timeout = defaultLookupTimeout
	}
	ctx, cancel := context.WithTimeout(d.ctx, timeout)
	defer cancel()

	addrs, err := d.lookup.LookupHost(ctx, d.host)
	if err != nil {
		err = handleDNSError(err)
		return &resolver.State{Addresses: []resolver.Address{}}, err
	}
	newAddrs := make([]resolver.Address, 0, len(addrs))
	for _, a := range addrs {
		addr, ok := tryParseIP(a, d.port)
		if !ok {
			return nil, structerr.New("dns: error parsing dns record IP address %v", a)
		}
		newAddrs = append(newAddrs, resolver.Address{Addr: addr})
	}

	return &resolver.State{Addresses: newAddrs}, nil
}

// handleDNSError massages the error to fit into expectations of the gRPC model:
//   - Timeouts and temporary errors should be communicated to gRPC to attempt another DNS query (with
//     Backoff).
//   - Other errors should be suppressed (they may represent the absence of a TXT record).
func handleDNSError(err error) error {
	if dnsErr, ok := err.(*net.DNSError); ok && !dnsErr.IsTimeout && !dnsErr.IsTemporary {
		return nil
	}

	return structerr.New("dns: record resolve error: %w", err)
}
