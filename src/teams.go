package main

import (
	_ "github.com/PCManiac/logrus_init"
	"github.com/golang-jwt/jwt"
	"github.com/mrFokin/jrpc"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type TeamInfo struct {
	Id   int32  `json:"id" db:"id"`
	Name string `json:"name" db:"name"`
}

func (h *handler) teamsList(c jrpc.Context) error {
	claims := c.EchoContext().Get("user").(*jwt.Token).Claims.(*UserClaims)
	club_id := claims.Data["club_id"]

	var data []TeamInfo

	if err := h.DB.Select(&data, `select * from api_sight."teamsList"($1);`, club_id); err != nil {
		log.WithFields(log.Fields{
			"proc":  "teamsList",
			"error": err,
		}).Error("SQL error")
		return errors.Wrap(err, "SQL error")
	}

	return c.Result(data)

}

func (h *handler) teamsGet(c jrpc.Context) error {
	claims := c.EchoContext().Get("user").(*jwt.Token).Claims.(*UserClaims)
	club_id := claims.Data["club_id"]

	var team_id int

	if err := c.Bind(&team_id); err != nil {
		return errors.Wrap(err, "teamsGet Bind error")
	}

	var data TeamInfo

	if err := h.DB.Get(&data, `select * from api_sight."teamsGet"($1, $2);`, club_id, team_id); err != nil {
		log.WithFields(log.Fields{
			"proc":    "teamsGet",
			"team_id": team_id,
			"error":   err,
		}).Error("SQL error")
		return errors.Wrap(err, "SQL error")
	}

	return c.Result(data)

}

func (h *handler) teamsAdd(c jrpc.Context) error {
	claims := c.EchoContext().Get("user").(*jwt.Token).Claims.(*UserClaims)
	club_id := claims.Data["club_id"]

	var name string

	if err := c.Bind(&name); err != nil {
		return errors.Wrap(err, "teamsAdd Bind error")
	}

	var data int

	if err := h.DB.Get(&data, `select * from api_sight."teamsAdd"($1, $2);`, club_id, name); err != nil {
		log.WithFields(log.Fields{
			"proc":  "teamsAdd",
			"name":  name,
			"error": err,
		}).Error("SQL error")
		return errors.Wrap(err, "SQL error")
	}

	return c.Result(data)
}

func (h *handler) teamsUpdate(c jrpc.Context) error {
	claims := c.EchoContext().Get("user").(*jwt.Token).Claims.(*UserClaims)
	club_id := claims.Data["club_id"]

	var params struct {
		Id   int    `json:"id" db:"id"`
		Name string `json:"name" db:"name"`
	}

	if err := c.Bind(&params); err != nil {
		return errors.Wrap(err, "teamsUpdate Bind error")
	}

	var data bool

	if err := h.DB.Get(&data, `select * from api_sight."teamsUpdate"($1, $2, $3);`, club_id, params.Id, params.Name); err != nil {
		log.WithFields(log.Fields{
			"proc":   "teamsUpdate",
			"params": params,
			"error":  err,
		}).Error("SQL error")
		return errors.Wrap(err, "SQL error")
	}

	return c.Result(data)
}

func (h *handler) teamsDelete(c jrpc.Context) error {
	claims := c.EchoContext().Get("user").(*jwt.Token).Claims.(*UserClaims)
	club_id := claims.Data["club_id"]

	var team_id int

	if err := c.Bind(&team_id); err != nil {
		return errors.Wrap(err, "teamsDelete Bind error")
	}

	var data bool

	if err := h.DB.Get(&data, `select * from api_sight."teamsDelete"($1, $2);`, club_id, team_id); err != nil {
		log.WithFields(log.Fields{
			"proc":    "teamsDelete",
			"team_id": team_id,
			"error":   err,
		}).Error("SQL error")
		return errors.Wrap(err, "SQL error")
	}

	return c.Result(data)
}
