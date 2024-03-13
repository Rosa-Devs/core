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

	//For frontend
	c.StartManager("")

	utils.Send(w, nil)
	return
}
