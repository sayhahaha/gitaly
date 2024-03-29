// Package backchannel implements connection multiplexing that allows for invoking
// gRPC methods from the server to the client.
//
// gRPC allows only for invoking RPCs from client to the server. Invoking
// RPCs from the server to the client can be useful in some cases such as
// tunneling through firewalls. While implementing such a use case would be
// possible with plain bidirectional streams, the approach has various limitations
// that force additional work on the user. All messages in a single stream are ordered
// and processed sequentially. If concurrency is desired, this would require the user
// to implement their own concurrency handling. Request routing and cancellations would also
// have to be implemented separately on top of the bidirectional stream.
//
// To do away with these problems, this package provides a multiplexed transport for running two
// independent gRPC sessions on a single connection. This allows for dialing back to the client from
// the server to establish another gRPC session where the server and client roles are switched.
//
// The server side uses listenmux to support clients that are unaware of the multiplexing.
//
// Usage:
//  1. Implement a ServerFactory, which is simply a function that returns a Server that can serve on the backchannel
//     connection. Plug in the ClientHandshake to the Clientconn via grpc.WithTransportCredentials when dialing.
//     This ensures all connections established by gRPC work with a multiplexing session and have a backchannel Server serving.
//  2. Create a *listenmux.Mux and register a *ServerHandshaker with it.
//  3. Pass the *listenmux.Mux into the grpc Server using grpc.Creds.
//     The Handshake method is called on each newly established connection that presents the backchannel magic bytes. It dials back to the client's backchannel server. Server
//     makes the backchannel connection's available later via the Registry's Backchannel method. The ID of the
//     peer associated with the current RPC handler can be fetched via GetPeerID. The returned ID can be used
//     to access the correct backchannel connection from the Registry.
package backchannel

import (
	"net"
	"sync"

	"github.com/hashicorp/yamux"
	"gitlab.com/gitlab-org/gitaly/v16/internal/log"
)

// magicBytes are sent by the client to server to identify as a multiplexing aware client.
var magicBytes = []byte("backchannel")

// muxConfig returns a new config to use with the multiplexing session.
func muxConfig(logger log.Logger, cfg Configuration) *yamux.Config {
	yamuxCfg := yamux.DefaultConfig()
	yamuxCfg.Logger = logger
	yamuxCfg.LogOutput = nil
	// gRPC is already configured to send keep alives so we don't need yamux to do this for us.
	// gRPC is a better choice as it sends the keep alives also to non-multiplexed connections.
	yamuxCfg.EnableKeepAlive = false
	yamuxCfg.AcceptBacklog = cfg.AcceptBacklog
	yamuxCfg.MaxStreamWindowSize = cfg.MaximumStreamWindowSizeBytes
	yamuxCfg.StreamCloseTimeout = cfg.StreamCloseTimeout

	return yamuxCfg
}

// connCloser wraps a net.Conn and calls the provided close function instead when Close
// is called.
type connCloser struct {
	net.Conn
	// once ensures the close function is called only once. gRPC may invoke Close() on connCloser
	// multiple times.
	once  sync.Once
	close func() error
	// closeErr records the error from the first call to close.
	closeErr error
}

// Close calls the provided close function. The close function is only executed once and
// further calls to Close return the error from the first invocation.
func (cc *connCloser) Close() error {
	cc.once.Do(func() { cc.closeErr = cc.close() })
	return cc.closeErr
}
