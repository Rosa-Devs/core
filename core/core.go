package core

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/Rosa-Devs/Database/src/manifest"
	db "github.com/Rosa-Devs/Database/src/store"

	"github.com/Rosa-Devs/core/models"
	"github.com/Rosa-Devs/core/network"
	"github.com/Rosa-Devs/core/store"
)

const Randevuz = "RosaApp"

type Core struct {
	Icon    []byte
	Started bool
	DbPath  string

	// Api local endpoint
	httpServer    *http.Server
	router        *http.ServeMux
	localApiAddr  string
	localApiState bool

	// TEST
	Store   store.Store
	profile models.Profile

	ctx context.Context

	// Host
	host network.Host

	dbs                  map[manifest.Manifest]*db.Database
	MessageValidateСache map[string]bool

	Driver     *db.DB
	Service_DB db.Database

	// Event server
	stopCh     chan struct{}
	waitGrp    sync.WaitGroup
	cancelFunc context.CancelFunc
	eventCh    chan struct{}
}

func (d *Core) GetProfile() string {
	if d.Started == false {
		return ""
	}
	return d.profile.Name
}

func (d *Core) Start(localapi string) string {
	// Start local serivce api for client app
	var err error
	if !d.localApiState {
		d.localApiAddr, err = d.startLocalApi(localapi)
		if err != nil {
			log.Println("Fail to bind local api!", err)
			return ""
		}
		d.localApiState = true
	}

	d.stopCh = make(chan struct{})
	d.eventCh = make(chan struct{}, 1000)
	d.MessageValidateСache = make(map[string]bool)

	d.profile, err = models.LoadFromFile(d.Store.Profile)
	if err != nil {
		fmt.Println("Error loading profile:", err)
		d.profile = models.Profile{
			Id: "UAUNT",
		}
		return ""
	}

	if d.Started == true {
		log.Println("Dbs Manager already started..")
		return ""
	}
	d.Started = true

	d.DbPath = d.Store.Database
	d.dbs = make(map[manifest.Manifest]*db.Database)
	d.ctx = context.Background()

	// Create new Host instance with properties
	d.host = network.Host{
		MDnsServie: true,
		DhtService: true,
	}

	if d.host.InitHost(d.ctx) != nil {
		log.Println("Failt to init HOST module. Crytical error")
		return ""
	}

	d.Driver = &db.DB{
		H:  d.host.H,
		Pb: d.host.Ps,
	}
	d.Driver.Start(d.DbPath)

	m_db := manifest.Manifest{
		Name:   "Service",
		UId:    "1",
		PubSub: manifest.GenerateNoise(15),
		Chiper: manifest.GenerateNoise(32),
	}

	d.Driver.CreateDb(m_db)
	d.Service_DB = d.Driver.GetDb(m_db)

	err = d.Service_DB.CreatePool("manifests")
	if err != nil {
		// log.Println("Not recreating pool:", err)
	}
	err = d.Service_DB.CreatePool("trust")
	if err != nil {
		// log.Println("Not recreating pool:", err)
	}

	// READ MANIFET DB AND CREATE DBS
	pool, err := d.Service_DB.GetPool("manifests")
	if err != nil {
		log.Println("Failed to get pool")
	}

	filter := map[string]interface{}{
		"type": 1, // All manifests
	}

	data, err := pool.Filter(filter)
	if err != nil {
		fmt.Println("Data:", data)
		fmt.Println("Error filtering data:", err)
	}

	for _, record := range data {
		// log.Println(record)
		manifestData, ok := record["data"].(string)
		if !ok {
			fmt.Println("Data field not found in map")
			continue
		}

		decodedData, err := base64.StdEncoding.DecodeString(manifestData)
		if err != nil {
			log.Println("Error decoding base64 data:", err)
			continue
		}

		m := new(manifest.Manifest)
		err = m.Deserialize(decodedData)
		if err != nil {
			log.Println("Error deserializing manifest, err:", err)
			continue
		}

		// Try to create db
		err = d.Driver.CreateDb(*m)
		if err != nil {
			log.Println("Not recreating db db, err:", err)
		}
		// Get db by manifest
		db := d.Driver.GetDb(*m)
		db.StartWorker(35)
		d.dbs[*m] = &db
	}

	log.Println("All database are create and ready to use")

	return d.localApiAddr
}
