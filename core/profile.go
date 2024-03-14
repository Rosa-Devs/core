package core

import (
	"log"
	"net/http"

	"github.com/Rosa-Devs/core/models"
	"github.com/Rosa-Devs/core/utils"
)

func (c *Core) createNewAccount(w http.ResponseWriter, r *http.Request) {
	req := new(models.CreatePofileReq)

	if err := utils.Read(r, &req); err != nil {
		log.Println("Fail to read client request:", err)
		http.Error(w, "Failed to read client request", http.StatusBadRequest)
		return
	}

	profile, err := models.CreateProfile(req.Name, req.Avatar)
	if err != nil {
		log.Println(err)
		return
	}
	err = models.WriteToFile(c.Store.Profile, profile)
	if err != nil {
		log.Println(err)
		return
	}

	// For frontend
	c.StartManager()

	utils.Send(w, 200)
	return
}

func (c *Core) autorized(w http.ResponseWriter, r *http.Request) {
	if !c.Started {
		utils.Send(w, false)
	}
	if c.profile.Id == "UAUNT" {
		utils.Send(w, false)
	}

	utils.Send(w, true)
}

func (c *Core) trust(w http.ResponseWriter, r *http.Request) {

	req := new(models.TrustReq)

	if err := utils.Read(r, &req); err != nil {
		log.Println("Fail to read client request:", err)
		http.Error(w, "Failed to read client request", http.StatusBadRequest)
		return
	}

	if c.Started == false {
		log.Println("Db manager is not started")
		http.Error(w, "Db manager is not started", http.StatusBadRequest)
	}

	pool, err := c.Service_DB.GetPool("trust", true)
	if err != nil {
		log.Println(err)
		return
	}

	data, err := req.Serialize()
	if err != nil {
		log.Println("Fail to serialize public profile")
		return
	}

	err = pool.Record(data)
	if err != nil {
		log.Println(err)
		return
	}

	utils.Send(w, 200)
}
