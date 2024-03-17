package core

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Rosa-Devs/Database/src/manifest"
	db "github.com/Rosa-Devs/Database/src/store"
	"github.com/Rosa-Devs/core/models"
	"github.com/Rosa-Devs/core/utils"
)

func (c *Core) listenEvents(w http.ResponseWriter, r *http.Request) {
	// Set the response header to indicate SSE content type
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Create a channel to send events to the client
	fmt.Println("Client connected")
	eventChan := make(chan string)
	c.localApiClient[eventChan] = struct{}{} // Add the client to the clients map
	defer func() {
		delete(c.localApiClient, eventChan) // Remove the client when they disconnect
		close(eventChan)
	}()

	// Listen for client close and remove the client from the list
	notify := w.(http.CloseNotifier).CloseNotify()
	go func() {
		<-notify
		fmt.Println("Client disconnected")
	}()

	// Continuously send data to the client
	for {
		data := <-eventChan
		fmt.Fprintf(w, "event: update\n\n")
		w.(http.Flusher).Flush()
		fmt.Println("Sending data to client:", data)

		// Simulate some events being sent periodically
		time.Sleep(1 * time.Second)
	}

}

// broadcast sends an event to all connected clients
func (c *Core) broadcast(data string) {
	for client := range c.localApiClient {
		client <- data
	}
}

func (c *Core) databaseUpdateEventServer(ctx context.Context, m manifest.Manifest) {
	c.waitGrp.Add(1)

	go func(ctx context.Context, m manifest.Manifest, C *Core) {
		defer C.waitGrp.Done()

		// Get database
		database, ok := C.dbs[m]
		if !ok {
			log.Println("Fail to get db!")
			return
		}

		// Subscribe to event channel
		eventListener := make(chan db.Event)
		database.EventBus.Subscribe(db.DbUpdateEvent, eventListener)

		for {
			select {
			case <-ctx.Done():
				log.Println("DatabaseUpdateEventServer exiting.")
				return
			case <-eventListener:

				//Create own event system

				// Old event system using wails to update dialog widow
				// But we need a cross platform solution
				//runtime.EventsEmit(M.wailsctx, "update")
				log.Println("Reviced event from chanel")

				C.broadcast("update")

				time.Sleep(time.Second)
				//log.Println("Event")
			}
		}
	}(ctx, m, c)
}

func (c *Core) changeListeningDb(w http.ResponseWriter, r *http.Request) {

	req := new(models.ChangeEventDb_req)

	err := utils.Read(r, &req)
	if err != nil {
		log.Println("Fail to read request")
		http.Error(w, "Fail to read request", http.StatusBadRequest)
	}

	if c.Started == false {
		log.Println("Db manager is not started")
		http.Error(w, "Db manager is not started", http.StatusBadRequest)
	}

	// Create a cancelable context
	ctx, cancel := context.WithCancel(context.Background())

	// Replace the existing cancel function with the new one
	// to ensure that the old goroutine will exit.
	if c.cancelFunc != nil {
		c.cancelFunc()
	}

	// Set the new cancel function
	c.cancelFunc = cancel

	// Create new DatabaseUpdateEventServer with new data!
	log.Println("Recreating events server")
	c.databaseUpdateEventServer(ctx, req.Manifest)

	utils.Send(w, 200)
}
