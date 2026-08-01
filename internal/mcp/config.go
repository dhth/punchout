package mcp

type Config struct {
	Transport Transport
	HTTPPort  uint
}

type Transport uint8

const (
	TransportStdio Transport = iota
	TransportHTTP
)

func (t Transport) String() string {
	switch t {
	case TransportStdio:
		return "stdio"
	case TransportHTTP:
		return "http"
	default:
		return "unknown"
	}
}

func ParseTransport(value string) (Transport, bool) {
	switch value {
	case "stdio":
		return TransportStdio, true
	case "http":
		return TransportHTTP, true
	default:
		return TransportStdio, false
	}
}
