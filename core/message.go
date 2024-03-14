package core

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Rosa-Devs/Database/src/manifest"
	db "github.com/Rosa-Devs/Database/src/store"
	"github.com/Rosa-Devs/core/models"
	"github.com/Rosa-Devs/core/utils"
)

func (c *Core) newMessage(w http.ResponseWriter, r *http.Request) {

	req := new(models.NewMessageReq)

	err := utils.Read(r, &req)
	if err != nil {
		log.Println("Fail to read request")
		http.Error(w, "Fail to read request", http.StatusBadRequest)
	}

	msg_stuct := new(models.Message)

	currentTime := time.Now().UTC()

	// Format the time in the desired format
	timestamp := currentTime.Format("2006-01-02T15:04:05.000")

	msg_stuct.Data = req.Msg
	msg_stuct.Sender = c.profile.GetPublic()
	msg_stuct.Sender.Avatar = ""
	msg_stuct.SenderId = c.profile.Id
	msg_stuct.Time = timestamp + "1"
	msg_stuct.DataType = models.MessageType
	msg_stuct.Valid = false
	c.profile.Sign(msg_stuct)
	p_pub := c.profile.GetPublic()
	b := p_pub.ValidateMsg(*msg_stuct)
	log.Println("Self signed msg. Sigh verified:", b)

	log.Println(msg_stuct.Signature)

	db, ok := c.dbs[req.Manifest]
	if !ok {
		http.Error(w, "Failed to get database from dbs manager", http.StatusInternalServerError)
		log.Println("Failed to get database from dbs manager.")
		return
	}

	err = db.CreatePool(models.MsgPool)
	if err != nil {
		// http.Error(w, "Failed to create pool: "+err.Error(), http.StatusInternalServerError)
		log.Println("Not recreating pool:", err)
		// return
	}

	pool, err := db.GetPool(models.MsgPool)
	if err != nil {
		http.Error(w, "Failed to get pool: "+err.Error(), http.StatusInternalServerError)
		log.Println("Failed to get pool:", err)
		return
	}

	// SERIALIZE MSG
	msgBytes, err := json.Marshal(msg_stuct)
	if err != nil {
		http.Error(w, "Failed to serialize message: "+err.Error(), http.StatusInternalServerError)
		log.Println("Failed to serialize message:", err)
		return
	}

	pool.Record(msgBytes)

	utils.Send(w, 200)
}

func (Mgr *Core) messagesLit(w http.ResponseWriter, r *http.Request) {

	req := new(manifest.Manifest)

	err := utils.Read(r, &req)
	if err != nil {
		log.Println("Fail to read request")
		http.Error(w, "Fail to read request", http.StatusBadRequest)
	}

	db, ok := Mgr.dbs[*req]
	if !ok {
		log.Println("Failed to get database from dbs manager.")
		http.Error(w, "Failed to get database from dbs manager", http.StatusInternalServerError)
		return
	}

	err = db.CreatePool(models.MsgPool)
	if err != nil {
		// log.Println("Not recreating poo:", err)
		// http.Error(w, "Failed to create pool: "+err.Error(), http.StatusInternalServerError)
		// return
	}

	pool, err := db.GetPool(models.MsgPool)
	if err != nil {
		log.Println("Failed to get pool:", err)
		http.Error(w, "Failed to get pool: "+err.Error(), http.StatusInternalServerError)
		return
	}
	filter := map[string]interface{}{
		"datatype": models.MessageType,
	}

	data, err := pool.Filter(filter)
	if err != nil {
		fmt.Println("Data:", data)
		fmt.Println("Error filtering data:", err)
	}

	msg_data := convertToMessages(data, db)
	Mgr.Validator(&msg_data)

	// sort.Slice(msg_data, func(i, j int) bool {
	// 	timei, _ := time.Parse(time.RFC3339, msg_data[i].Time) // Assuming Data field contains timestamp in RFC3339 format
	// 	timej, _ := time.Parse(time.RFC3339, msg_data[j].Time)

	// 	// Compare timestamps to sort from latest to newest
	// 	return timei.After(timej)
	// })

	// log.Println(msg_data)

	utils.Send(w, msg_data)
}

func convertToMessages(data []map[string]interface{}, db *db.Database) []models.Message {
	messages := make([]models.Message, len(data))

	for i, item := range data {
		// Assuming your map contains fields like "ID" and "Text"
		// Adjust these according to your actual map structure
		d_type, _ := item["datatype"].(int)
		sender_map, _ := item["sender"].(map[string]interface{})
		sender := new(models.ProfileStorePublic)
		sender.ProfileFromMap(sender_map)

		filter := map[string]interface{}{
			"id": sender.Id, // Random integer between 0 and 100
		}

		pool, err := db.GetPool(models.UserPool)
		if err != nil {
			log.Println(err)
			return nil
		}

		profiles, err := pool.Filter(filter)
		if err != nil {
			log.Println("Fail to get pool:", err)
			return nil
		}

		if len(profiles) > 0 {
			p := new(models.ProfileStorePublic)
			p.ProfileFromMap(profiles[0])
			sender.Avatar = p.Avatar
		} else {
			sender.Avatar = ""
		}

		data, _ := item["data"].(string)
		time, _ := item["time"].(string)
		sign, _ := item["sign"].(string)
		s_id, _ := item["senderid"].(string)

		// Create a new Message and append it to the result slice
		messages[i] = models.Message{
			DataType:  d_type,
			Sender:    *sender,
			SenderId:  s_id,
			Data:      data,
			Time:      time,
			Signature: sign,
		}
	}

	return messages
}

func (c *Core) Validator(m *[]models.Message) {
	for i := range *m {
		user := c.FindUserById((*m)[i].Sender.Id)
		if user.Id != (*m)[i].Sender.Id {
			//log.Println("This account not is exsit:", user.Id)
		}

		msg := (*m)[i]

		if validated, ok := c.MessageValidateСache[msg.Data+msg.Time]; ok {
			//log.Println("Msg:", msg.Data, "Already Validated for user:", user.Name)
			(*m)[i].Valid = validated
			continue
		} else {
			if user.ValidateMsg(msg) {
				//log.Println("Msg:", (*m)[i].Data, "Validated:", "true", "With user:", user.Name)
				(*m)[i].Valid = true
				//.MessageValidateСache[msg.Data+msg.Time] = true
			} else {
				(*m)[i].Valid = false
				c.MessageValidateСache[msg.Data+msg.Time] = false
				//log.Println("Msg:", (*m)[i].Data, "Validated:", "false", "With user:", user.Name)
			}
		}
	}
}
