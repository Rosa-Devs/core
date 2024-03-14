package core

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/Rosa-Devs/Database/src/manifest"
	db "github.com/Rosa-Devs/Database/src/store"
	"github.com/Rosa-Devs/core/models"
	"github.com/Rosa-Devs/core/utils"
)

func (c *Core) listenEvents(w http.ResponseWriter, r *http.Request) {
	// Set headers for Server-Sent Events
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// Listen for events
	for {
		select {
		case <-r.Context().Done():
			log.Println("Connection closed")
			return
		case <-c.EventCh:
			// Write event to client
			if _, err := w.Write([]byte("update\n\n")); err != nil {
				log.Println("Error writing to client:", err)
				return
			}
			w.(http.Flusher).Flush() // Flush response to ensure it's sent immediately
		}
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
				// Send event to event channel
				select {
				case C.EventCh <- struct{}{}:
				default:
					// If event channel is full, drop the event
					log.Println("Event channel is full, dropping event.")
				}

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
