package core

import (
	"log"
	"net/http"

	"github.com/Rosa-Devs/core/models"
	"github.com/Rosa-Devs/core/utils"
)

func (c *Core) metrics(w http.ResponseWriter, r *http.Request) {

	metric := models.Metrics{
		State: models.State{
			Db:  c.Started,
			Api: c.localApiState,
		},
		Status: models.Status{
			DhtNodes:    len(c.host.H.Peerstore().Peers()),
			Connections: len(c.host.H.Network().Conns()),
			Channels:    len(c.dbs),
		},
	}

	log.Println(&metric)

	utils.Send(w, metric)
}
