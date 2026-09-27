package network

// Connection represents an abstract, stream-oriented peer-to-peer transport session.
type Connection interface {
	// Send transmits a message frame to the remote peer.
	Send(data []byte) error

	// Receive reads the next message frame from the remote peer.
	Receive() ([]byte, error)

	// Close terminates the connection.
	Close() error

	// RemoteAddr returns the remote network address string.
	RemoteAddr() string
}
