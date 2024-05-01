package main

import (
	"encoding/json"
	"time"

	_ "github.com/PCManiac/logrus_init"
	"github.com/golang-jwt/jwt"
	"github.com/mrFokin/jrpc"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type PlayersInfo struct {
	Id         int32            `json:"id" db:"id"`
	TeamId     *int32           `json:"team_id" db:"team_id"`
	FName      *string          `json:"f_name" db:"f_name"`
	MName      *string          `json:"m_name" db:"m_name"`
	LName      *string          `json:"l_name" db:"l_name"`
	Gender     *int16           `json:"gender" db:"gender"`
	BirthDate  *time.Time       `json:"birth_date" db:"birth_date"`
	Jersey     *string          `json:"jersey" db:"jersey"`
	PositionId *int32           `json:"position" db:"position"`
	Weight     *float32         `json:"current_weight" db:"current_weight"`
	Height     *float32         `json:"current_height" db:"current_height"`
	MaxPulse   *int32           `json:"max_pulse" db:"max_pulse"`
	Data       *json.RawMessage `json:"data" db:"data"`
}

func (h *handler) playersList(c jrpc.Context) error {
	claims := c.EchoContext().Get("user").(*jwt.Token).Claims.(*UserClaims)
	club_id := claims.Data["club_id"]

	var data []PlayersInfo

	if err := h.DB.Select(&data, `select * from api_sight."playersList"($1);`, club_id); err != nil {
		log.WithFields(log.Fields{
			"proc":  "playersList",
			"error": err,
		}).Error("SQL error")
		return errors.Wrap(err, "SQL error")
	}

	return c.Result(data)

}

func (h *handler) playersGet(c jrpc.Context) error {
	claims := c.EchoContext().Get("user").(*jwt.Token).Claims.(*UserClaims)
	club_id := claims.Data["club_id"]

	var pl_id int

	if err := c.Bind(&pl_id); err != nil {
		return errors.Wrap(err, "playersGet Bind error")
	}

	var data PlayersInfo

	if err := h.DB.Get(&data, `select * from api_sight."playersGet"($1, $2);`, club_id, pl_id); err != nil {
		log.WithFields(log.Fields{
			"proc":    "playerssGet",
			"club_id": pl_id,
			"error":   err,
		}).Error("SQL error")
		return errors.Wrap(err, "SQL error")
	}

	return c.Result(data)
}

func (h *handler) playersAdd(c jrpc.Context) error {
	claims := c.EchoContext().Get("user").(*jwt.Token).Claims.(*UserClaims)
	club_id := claims.Data["club_id"]

	var params struct {
		TeamId     *int32           `json:"team_id" db:"team_id"`
		FName      *string          `json:"f_name" db:"f_name"`
		MName      *string          `json:"m_name" db:"m_name"`
		LName      *string          `json:"l_name" db:"l_name"`
		Gender     *int16           `json:"gender" db:"gender"`
		BirthDate  *time.Time       `json:"birth_date" db:"birth_date"`
		Jersey     *string          `json:"jersey" db:"jersey"`
		PositionId *int32           `json:"position" db:"position"`
		Weight     *float32         `json:"current_weight" db:"current_weight"`
		Height     *float32         `json:"current_height" db:"current_height"`
		MaxPulse   *int32           `json:"max_pulse" db:"max_pulse"`
		Data       *json.RawMessage `json:"data" db:"data"`
	}

	if err := c.Bind(&params); err != nil {
		return errors.Wrap(err, "playersAdd Bind error")
	}

	var data int

	if err := h.DB.Get(&data, `select * from api_sight."playersAdd"($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13);`,
		club_id, params.TeamId, params.FName, params.MName, params.LName, params.Gender, params.BirthDate,
		params.Jersey, params.PositionId, params.Weight, params.Height, params.MaxPulse, params.Data); err != nil {
		log.WithFields(log.Fields{
			"proc":   "playersAdd",
			"params": params,
			"error":  err,
		}).Error("SQL error")
		return errors.Wrap(err, "SQL error")
	}

	return c.Result(data)
}

func (h *handler) playersUpdate(c jrpc.Context) error {
	claims := c.EchoContext().Get("user").(*jwt.Token).Claims.(*UserClaims)
	club_id := claims.Data["club_id"]

	var params struct {
		Id         int32            `json:"id" db:"id"`
		TeamId     *int32           `json:"team_id" db:"team_id"`
		FName      *string          `json:"f_name" db:"f_name"`
		MName      *string          `json:"m_name" db:"m_name"`
		LName      *string          `json:"l_name" db:"l_name"`
		Gender     *int16           `json:"gender" db:"gender"`
		BirthDate  *time.Time       `json:"birth_date" db:"birth_date"`
		Jersey     *string          `json:"jersey" db:"jersey"`
		PositionId *int32           `json:"position" db:"position"`
		Weight     *float32         `json:"current_weight" db:"current_weight"`
		Height     *float32         `json:"current_height" db:"current_height"`
		MaxPulse   *int32           `json:"max_pulse" db:"max_pulse"`
		Data       *json.RawMessage `json:"data" db:"data"`
	}

	if err := c.Bind(&params); err != nil {
		return errors.Wrap(err, "playersUpdate Bind error")
	}

	var data bool

	if err := h.DB.Get(&data, `select * from api_sight."playersUpdate"($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14);`,
		club_id, params.Id, params.TeamId, params.FName, params.MName, params.LName, params.Gender, params.BirthDate,
		params.Jersey, params.PositionId, params.Weight, params.Height, params.MaxPulse, params.Data); err != nil {
		log.WithFields(log.Fields{
			"proc":   "playersUpdate",
			"params": params,
			"error":  err,
		}).Error("SQL error")
		return errors.Wrap(err, "SQL error")
	}

	return c.Result(data)
}

func (h *handler) playersDelete(c jrpc.Context) error {
	claims := c.EchoContext().Get("user").(*jwt.Token).Claims.(*UserClaims)
	club_id := claims.Data["club_id"]

	var id int

	if err := c.Bind(&id); err != nil {
		return errors.Wrap(err, "playersDelete Bind error")
	}

	var data bool

	if err := h.DB.Get(&data, `select * from api_sight."playersDelete"($1, $2);`, club_id, id); err != nil {
		log.WithFields(log.Fields{
			"proc":  "playersDelete",
			"id":    id,
			"error": err,
		}).Error("SQL error")
		return errors.Wrap(err, "SQL error")
	}

	return c.Result(data)
}

func (h *handler) playersResetPassword(c jrpc.Context) error {
	claims := c.EchoContext().Get("user").(*jwt.Token).Claims.(*UserClaims)
	user_id := claims.ID

	var params struct {
		Id       int32   `json:"id" db:"id"`
		Password *string `json:"new_password" db:"new_password"`
	}

	if err := c.Bind(&params); err != nil {
		return errors.Wrap(err, "playersResetPassword Bind error")
	}

	var data bool

	if err := h.DB.Get(&data, `select * from api_users."playerResetPassword"($1, $2, $3);`,
		user_id, params.Id, params.Password); err != nil {
		log.WithFields(log.Fields{
			"proc":   "playersResetPassword",
			"params": params,
			"error":  err,
		}).Error("SQL error")
		return errors.Wrap(err, "SQL error")
	}

	return c.Result(data)
}
