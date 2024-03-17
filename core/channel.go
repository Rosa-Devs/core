package core

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/Rosa-Devs/Database/src/manifest"

	"github.com/Rosa-Devs/core/models"
	"github.com/Rosa-Devs/core/utils"
)

func (c *Core) createNewManifest(w http.ResponseWriter, r *http.Request) {
	req := new(models.CreateManifestReq)

	err := utils.Read(r, &req)
	if err != nil {
		log.Println("Fail to read request")
		http.Error(w, "Fail to read request", http.StatusBadRequest)
	}

	m_json, err := manifest.GenereateManifest(req.Name, false, req.Opts).Serialize()
	if err != nil {
		log.Println("Fail to create manifest")
		http.Error(w, "Fail to create manifest", http.StatusInternalServerError)
	}

	res := new(models.CreateManifestRes)

	res.Manifest = string(m_json)

	utils.Send(w, res)
}

func (c *Core) addManifets(w http.ResponseWriter, r *http.Request) {
	req := new(models.AddManifest)

	err := utils.Read(r, &req)
	if err != nil {
		log.Println("Fail to read request")
		http.Error(w, "Fail to read request", http.StatusBadRequest)
	}

	if c.Started == false {
		log.Println("Db manager is not started")
		http.Error(w, "Db manager is not started", http.StatusBadRequest)
	}
	err = c.Service_DB.CreatePool("manifests")
	if err != nil {
		log.Println("Not recreating pool:", err)
	}
	pool, err := c.Service_DB.GetPool("manifests", true)
	if err != nil {
		log.Println("Fail to get pool", err)
		http.Error(w, "Fail to get pool", http.StatusBadRequest)
	}

	m := new(manifest.Manifest)
	m.Deserialize([]byte(req.Manifest))

	m_s := new(models.MStore)
	m_s.Data, err = m.Serialize()
	m_s.Type = models.MStore_TYPE_Manifet
	if err != nil {
		log.Panicln("Fail to serialize manifest", err)
		http.Error(w, "Fail to serialize manifest", http.StatusInternalServerError)
	}
	jsonData, err := json.Marshal(m_s)
	if err != nil {
		log.Println("Fail to marshal manifest", err)
		http.Error(w, "Fail to marshal manifest", http.StatusInternalServerError)
	}

	// Create new db
	// Try to create db
	err = c.Driver.CreateDb(*m)
	if err != nil {
		log.Println("Not recreating db db, err:", err)
	}
	db := c.Driver.GetDb(*m)
	db.StartWorker(60)
	c.registerUser(&db)
	c.dbs[*m] = &db

	err = pool.Record(jsonData)
	if err != nil {
		log.Println("Fail to update recod in pool", err)
		http.Error(w, "Fail to update recod in pool", http.StatusInternalServerError)
	}

	utils.Send(w, 200)

}

func (c *Core) listManifest(w http.ResponseWriter, r *http.Request) {
	if c.Started == false {
		log.Println("Db manager is not started")
		manifests := append([]manifest.Manifest{}, manifest.Manifest{Name: "Db Manager not started", PubSub: "0"})

		// Return the result
		utils.Send(w, manifests)
	}
	err := c.Service_DB.CreatePool("manifests")
	if err != nil {
		// log.Println("Not recreating pool:", err)
	}
	// READ MANIFET DB AND CREATE DBS
	pool, err := c.Service_DB.GetPool("manifests")
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

	var manifetss []manifest.Manifest
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

		manifetss = append(manifetss, *m)
	}

	utils.Send(w, manifetss)
}

func (c *Core) deleteManifest(w http.ResponseWriter, r *http.Request) {
	req := new(models.DeleteManifest)

	err := utils.Read(r, &req)
	if err != nil {
		log.Println("Fail to read request")
		http.Error(w, "Fail to read request", http.StatusBadRequest)
		return
	}

	if !c.Started {
		log.Println("Db manager is not started")
		http.Error(w, "Db manager is not started", http.StatusBadRequest)
		return
	}

	err = c.Service_DB.CreatePool("manifests")
	if err != nil {
		log.Println("Not recreating pool:", err)
		// http.Error(w, "Failed to create pool", http.StatusInternalServerError)
	}

	pool, err := c.Service_DB.GetPool("manifests", true)
	if err != nil {
		log.Println("Failed to get pool:", err)
		http.Error(w, "Failed to get pool", http.StatusInternalServerError)
		return
	}

	m_d, err := req.Serialize()
	if err != nil {
		log.Println("Failed to serialize manifest:", err)
		http.Error(w, "Failed to serialize manifest", http.StatusInternalServerError)
		return
	}
	encodedData := base64.StdEncoding.EncodeToString(m_d)

	filter := map[string]interface{}{
		"type": 1, // All manifests
		"data": encodedData,
	}

	data, err := pool.Filter(filter)
	if err != nil {
		log.Println("Error filtering data:", err)
		http.Error(w, "Error filtering data", http.StatusInternalServerError)
		return
	}

	log.Println(data)

	for _, record := range data {
		err := pool.Delete(record["_id"].(string))
		if err != nil {
			log.Println("Error deleting record:", err)
			http.Error(w, "Error deleting record", http.StatusInternalServerError)
			return
		}
	}

	utils.Send(w, 200)
}
