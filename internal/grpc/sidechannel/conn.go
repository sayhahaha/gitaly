package sidechannel

import (
	"context"
	"fmt"
	"io"
	"net"

	"gitlab.com/gitlab-org/gitaly/v16/internal/git/pktline"
	"gitlab.com/gitlab-org/gitaly/v16/internal/grpc/client"
	"gitlab.com/gitlab-org/gitaly/v16/internal/log"
	"gitlab.com/gitlab-org/gitaly/v16/streamio"
	"google.golang.org/grpc"
)

// ServerConn and ClientConn implement an asymmetric framing protocol to
// exchange data between clients and servers in Sidechannel. A typical flow
// looks like following:
// - The client writes data into the connecton.
// - The client half-closes the connection. The server is aware of this event
// when reading operations return EOF.
// - The server writes the data back to the client, then close the connection.
// - The client read the data until EOF
//
// Half-close ability is important to signal the server that the client
// finishes data transformation. As sidechannel is built on top of Yamux
// stream, half-close ability is not supported. Therefore, we apply a
// length-prefix framing protocol, simiarly to Git pktline protocol, except we
// omit the band number. The close or half-close event are signaled by sending
// a flush packet.
//
// This is an example of the data written into the wire:
//
// | 4-byte length, including size of length itself.
// v
// 0009Hello0000
//          ^
//          | Flush packet signaling a half-close event
//
// Many methods in battle-tested pktline package are re-used to save us some
// times. At the moment, we don't need server-client half-closed ability. And
// it may affect the performance when wrapping huge data sent from the server.

const (
	// maxChunkSize is the maximum chunk size of data. The chunk size must include 4-byte
	// length prefix. This constant is different from MaxSidebandData because
	// we don't include the sideband number.
	maxChunkSize = pktline.MaxPktSize - 4
)

// ServerConn is a wrapper around net.Conn with the support of half-closed
// capacity for sidechannel. This struct is expected to be used by
// sidechannel's server only.
type ServerConn struct {
	conn net.Conn
	r    io.Reader
}

func newServerConn(c net.Conn) *ServerConn {
	scanner := pktline.NewScanner(c)
	reader := streamio.NewReader(func() ([]byte, error) {
		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				return nil, err
			}
			// If there is any error while scanning, scanner.Err() returns a
			// non-nil error. If scanner.Err() returns nil, the connection
			// reaches end-of-file. However, the effect of returning io.EOF is
			// that we allow two kinds of streams: "000fhello world0000" (with
			// trailing 0000) and "000fhello world" (without trialing 0000).
			// Having optional behaviors like this is a source of complexity.
			// We should not allow "000fhello world" without the trailing 0000.
			return nil, io.ErrUnexpectedEOF
		}

		if pktline.IsFlush(scanner.Bytes()) {
			return nil, io.EOF
		}

		data := scanner.Bytes()
		if len(data) < 4 {
			return nil, fmt.Errorf("sidechannel: invalid packet %q", data)
		}

		// pktline treats 0001, 0002, or 0003 as magic empty packets
		// They are irrelevant to sidechannel, hence should be rejected
		if len(data) == 4 {
			if s := string(data); s == "0001" || s == "0002" || s == "0003" {
				return nil, fmt.Errorf("sidechannel: invalid header %s", string(data[3]))
			}
		}

		return data[4:], nil
	})

	return &ServerConn{conn: c, r: reader}
}

// Read reads up to len(p) bytes into p. It returns the number of bytes read or
// any error encountered. This struct overrides Read() to extract the data
// wrapped in a frame generated by ClientConn.Write().
func (cc *ServerConn) Read(p []byte) (n int, err error) {
	return cc.r.Read(p)
}

// Write writes data to the connection. This method fallbacks to underlying
// connection without any modificiation.
func (cc *ServerConn) Write(b []byte) (n int, err error) {
	return cc.conn.Write(b)
}

// Close closes the connection. This method fallbacks to underlying
// connection without any modificiation.
func (cc *ServerConn) Close() error {
	return cc.conn.Close()
}

// ClientConn is a wrapper around net.Conn with the support of half-closed
// capacity for sidechannel. This struct is expected to use by sidechannel's
// client only.
type ClientConn struct {
	conn        net.Conn
	writeClosed bool
}

func newClientConn(c net.Conn) *ClientConn {
	return &ClientConn{conn: c}
}

// Read reads data from the connection. This method fallbacks to underlying
// connection without any modificiation.
func (cc *ClientConn) Read(b []byte) (n int, err error) {
	return cc.conn.Read(b)
}

// Write writes len(p) bytes from p to the underlying data stream.  It returns
// the number of bytes written from p and any error encountered that caused the
// write to stop early. This method overrides Write() to wrap the writing data
// into a frame. The frame is then extracted and read by ServerConn.Read().
func (cc *ClientConn) Write(p []byte) (int, error) {
	if cc.writeClosed {
		return 0, fmt.Errorf("sidechannel: write into a half-closed connection")
	}

	var n int

	for len(p) > 0 {
		chunk := maxChunkSize
		if len(p) < chunk {
			chunk = len(p)
		}

		if _, err := fmt.Fprintf(cc.conn, "%04x", chunk+4); err != nil {
			return n, err
		}
		if _, err := cc.conn.Write(p[:chunk]); err != nil {
			return n, err
		}
		n += chunk
		p = p[chunk:]
	}

	return n, nil
}

func (cc *ClientConn) close() error {
	return cc.conn.Close()
}

// CloseWrite shuts down the writing side of the connection. After this call,
// any read operations from the server return EOF. The reading side is still
// functional so that the server is still able to write back to the client. Any
// attempt to write into a half-closed connection returns an error.
func (cc *ClientConn) CloseWrite() error {
	if cc.writeClosed {
		return nil
	}

	cc.writeClosed = true
	if err := pktline.WriteFlush(cc.conn); err != nil {
		return err
	}

	return nil
}

// Dial configures the dialer to establish a Gitaly backchannel connection instead of a regular gRPC connection. It
// also injects sr as a sidechannel registry, so that Gitaly can establish sidechannels back to the client.
func Dial(ctx context.Context, registry *Registry, logger log.Logger, rawAddress string, connOpts []grpc.DialOption) (*grpc.ClientConn, error) {
	clientHandshaker := NewClientHandshaker(logger, registry)
	return client.Dial(ctx, rawAddress, client.WithGrpcOptions(connOpts), client.WithHandshaker(clientHandshaker))
}
