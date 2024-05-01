package main

import (
	"encoding/json"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/mrFokin/jrpc"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type ClubParams map[string]interface{}

func (a *ClubParams) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return nil
	}

	return json.Unmarshal(b, &a)
}

type ClubInfo struct {
	Id         int32      `json:"id" db:"id"`
	Name       string     `json:"name" db:"name"`
	UpdateTime time.Time  `json:"updated_at" db:"updated_at"`
	CreateTime time.Time  `json:"created_at" db:"created_at"`
	Params     ClubParams `json:"params" db:"params"`
}

func (h *handler) clubsGet(c jrpc.Context) error {

	claims := c.EchoContext().Get("user").(*jwt.Token).Claims.(*UserClaims)
	club_id := claims.Data["club_id"]

	var data ClubInfo

	if err := h.DB.Get(&data, `select * from api_sight."clubGetById"($1);`, club_id); err != nil {
		log.WithFields(log.Fields{
			"proc":    "clubsGet",
			"club_id": club_id,
			"error":   err,
		}).Error("SQL error")
		return errors.Wrap(err, "SQL error")
	}

	return c.Result(data)

}
