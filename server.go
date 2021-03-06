package stats

import (
	"encoding/json"
	"fmt"
	"net"
	"strconv"
)

// ServerConfig is used to initialize a new ServerStats object
type ServerConfig struct {
	Domain           string
	Port             int
	Debug            bool
	CustomBufferSize int
}

// ServerStats is the object used to receive, store and send data for usage
type ServerStats struct {
	addr       string
	debug      bool
	bufferSize int
}

// NewServer create a new server object for use
func NewServer(config *ServerConfig) (*ServerStats, error) {

	bSize := config.CustomBufferSize
	if bSize == 0 {
		bSize = defaultBufferSize
	}

	return &ServerStats{
		addr:       config.Domain + ":" + strconv.Itoa(config.Port),
		debug:      config.Debug,
		bufferSize: bSize,
	}, nil
}

// Run starts receiving the profiling stats for storage and usage
func (s *ServerStats) Run() <-chan *Stats {

	results := make(chan *Stats)

	go func(results chan<- *Stats) {

		var serverAddr *net.UDPAddr
		var server *net.UDPConn
		var err error

		serverAddr, err = net.ResolveUDPAddr(udp, s.addr)
		if err != nil {
			panic(err)
		}

		server, err = net.ListenUDP(udp, serverAddr)
		if err != nil {
			panic(err)
		}
		defer server.Close()

		server.SetReadBuffer(s.bufferSize)

		var addr *net.UDPAddr
		var bytesRead int
		buff := make([]byte, s.bufferSize)
		stats := new(Stats)

		for {

			bytesRead, addr, err = server.ReadFromUDP(buff)
			if err != nil {
				fmt.Printf("Error:%s bytes read %d\n", err, bytesRead)
				continue
			}

			if s.debug {
				fmt.Printf("Recieved: %s from %s Read %d bytes\n", string(buff[0:bytesRead]), addr, bytesRead)
			}

			if err := json.Unmarshal(buff[0:bytesRead], &stats); err != nil {
				fmt.Println("Error Unmarshalling:", err)
			}

			results <- stats
		}
	}(results)

	return results
}
