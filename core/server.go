package core

import (
	"fmt"
	"net"
	"net/http"
)

// Server struct

// StartServer method initializes and starts the HTTP server
func (s *Core) startLocalApi(ip string) (string, error) {

	s.router = http.NewServeMux()

	s.registerLocalApi()

	// Create an HTTP server with the specified handler
	s.httpServer = &http.Server{
		Handler: s.router,
	}

	// Start the HTTP server

	// Manually bind to an available port
	var listener net.Listener
	var err error
	if ip != "" {
		listener, err = net.Listen("tcp", ip)
	} else {
		listener, err = net.Listen("tcp", ":0")
	}

	if err != nil {
		fmt.Println("Error binding to port:", err)
		return "", err
	}

	// Get the actual port that was assigned
	addr := listener.Addr().(*net.TCPAddr)
	//fmt.Println("Server listening on", addr.String())
	go func() {
		err = s.httpServer.Serve(listener)
		if err != nil {
			fmt.Println("Error starting server:", err)
			return
		}
	}()

	return addr.String(), nil

}

func (c *Core) registerLocalApi() {
	// c.router.HandleFunc("/", c.TestEndpoint)
}
