package main

import (
	_ "github.com/PCManiac/logrus_init"
	"github.com/golang-jwt/jwt"
	"github.com/mrFokin/jrpc"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type PositionInfo struct {
	Id    int32  `json:"id" db:"id"`
	Name  string `json:"name" db:"name"`
	Alias string `json:"alias" db:"alias"`
}

func (h *handler) positionsList(c jrpc.Context) error {
	claims := c.EchoContext().Get("user").(*jwt.Token).Claims.(*UserClaims)
	club_id := claims.Data["club_id"]

	var data []PositionInfo

	if err := h.DB.Select(&data, `select * from api_sight."positionsList"($1);`, club_id); err != nil {
		log.WithFields(log.Fields{
			"proc":  "positionsList",
			"error": err,
		}).Error("SQL error")
		return errors.Wrap(err, "SQL error")
	}

	return c.Result(data)
}

func (h *handler) positionsGet(c jrpc.Context) error {
	claims := c.EchoContext().Get("user").(*jwt.Token).Claims.(*UserClaims)
	club_id := claims.Data["club_id"]

	var pos_id int

	if err := c.Bind(&pos_id); err != nil {
		return errors.Wrap(err, "teamsGet Bind error")
	}

	var data PositionInfo

	if err := h.DB.Get(&data, `select * from api_sight."positionsGet"($1, $2);`, club_id, pos_id); err != nil {
		log.WithFields(log.Fields{
			"proc":        "positionsGet",
			"position_id": pos_id,
			"error":       err,
		}).Error("SQL error")
		return errors.Wrap(err, "SQL error")
	}

	return c.Result(data)
}

func (h *handler) positionsAdd(c jrpc.Context) error {
	claims := c.EchoContext().Get("user").(*jwt.Token).Claims.(*UserClaims)
	club_id := claims.Data["club_id"]

	var params struct {
		Name  string `json:"name" db:"name"`
		Alias string `json:"alias" db:"alias"`
	}

	if err := c.Bind(&params); err != nil {
		return errors.Wrap(err, "positionsAdd Bind error")
	}

	var data int

	if err := h.DB.Get(&data, `select * from api_sight."positionsAdd"($1, $2, $3);`, club_id, params.Name, params.Alias); err != nil {
		log.WithFields(log.Fields{
			"proc":   "positionsAdd",
			"params": params,
			"error":  err,
		}).Error("SQL error")
		return errors.Wrap(err, "SQL error")
	}

	return c.Result(data)
}

func (h *handler) positionsUpdate(c jrpc.Context) error {
	claims := c.EchoContext().Get("user").(*jwt.Token).Claims.(*UserClaims)
	club_id := claims.Data["club_id"]

	var params struct {
		Id    int    `json:"id" db:"id"`
		Name  string `json:"name" db:"name"`
		Alias string `json:"alias" db:"alias"`
	}

	if err := c.Bind(&params); err != nil {
		return errors.Wrap(err, "positionsUpdate Bind error")
	}

	var data bool

	if err := h.DB.Get(&data, `select * from api_sight."positionsUpdate"($1, $2, $3, $4);`, club_id, params.Id, params.Name, params.Alias); err != nil {
		log.WithFields(log.Fields{
			"proc":   "positionsUpdate",
			"params": params,
			"error":  err,
		}).Error("SQL error")
		return errors.Wrap(err, "SQL error")
	}

	return c.Result(data)
}

func (h *handler) positionsDelete(c jrpc.Context) error {
	claims := c.EchoContext().Get("user").(*jwt.Token).Claims.(*UserClaims)
	club_id := claims.Data["club_id"]

	var id int

	if err := c.Bind(&id); err != nil {
		return errors.Wrap(err, "positionsDelete Bind error")
	}

	var data bool

	if err := h.DB.Get(&data, `select * from api_sight."positionsDelete"($1, $2);`, club_id, id); err != nil {
		log.WithFields(log.Fields{
			"proc":  "positionsDelete",
			"id":    id,
			"error": err,
		}).Error("SQL error")
		return errors.Wrap(err, "SQL error")
	}

	return c.Result(data)
}
