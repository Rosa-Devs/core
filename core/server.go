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

	corsMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Add CORS headers to allow requests from any origin
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

			// If it's a preflight request, respond with 200 OK
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			// Call the next handler in the chain
			next.ServeHTTP(w, r)
		})
	}

	// Create an HTTP server with the specified handler
	s.httpServer = &http.Server{
		Handler: corsMiddleware(s.router),
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
	// fmt.Println("Server listening on", addr.String())
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
	c.router.HandleFunc("POST /aunth/create", c.createNewAccount)
	c.router.HandleFunc("GET /aunth/autorized", c.autorized)
	c.router.HandleFunc("POST /aunth/trust", c.trust)

	//Channels
	c.router.HandleFunc("POST /channel/create", c.createNewManifest)
	c.router.HandleFunc("POST /channel/delete", c.deleteManifest)
	c.router.HandleFunc("POST /channel/add", c.addManifets)
	c.router.HandleFunc("GET /channel/list", c.listManifest)

	//EVENTS
	c.router.HandleFunc("POST /event/change", c.changeListeningDb)
	c.router.HandleFunc("GET /event/listen", c.listenEvents)

	//MESSAGE
	c.router.HandleFunc("POST /message/new", c.newMessage)
	c.router.HandleFunc("POST /message/list", c.messagesLit)
}
