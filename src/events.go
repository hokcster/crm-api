package main

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	_ "github.com/PCManiac/logrus_init"
	"github.com/golang-jwt/jwt"
	"github.com/lib/pq"
	"github.com/mrFokin/jrpc"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type EventInfo struct {
	Id        string    `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	TeamId    int32     `json:"team_id" db:"team_id"`
	StartTime time.Time `json:"start_time" db:"start_time"`
	StopTime  time.Time `json:"stop_time" db:"stop_time"`
}

type SplitTags []string

func (a *SplitTags) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return nil
	}

	return json.Unmarshal(b, &a)
}

func (a SplitTags) Value() (driver.Value, error) {
	return json.Marshal(a)
}

type SliceOfMapInterface []map[string]interface{}

func (a *SliceOfMapInterface) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return nil
	}

	return json.Unmarshal(b, &a)
}

type SplitsInfo struct {
	Id        string              `json:"id" db:"id"`
	Event     string              `json:"event_id" db:"event_id"`
	StartTime *time.Time          `json:"start_time" db:"start_time"`
	StopTime  *time.Time          `json:"stop_time" db:"stop_time"`
	Tags      *SplitTags          `json:"tags" db:"tags"`
	Players   SliceOfMapInterface `json:"players" db:"players"`
}

func (h *handler) eventsList(c jrpc.Context) error {
	claims := c.EchoContext().Get("user").(*jwt.Token).Claims.(*UserClaims)
	club_id := claims.Data["club_id"]

	var params struct {
		StartTime time.Time `json:"start_time" db:"start_time"`
		StopTime  time.Time `json:"stop_time" db:"stop_time"`
	}

	if err := c.Bind(&params); err != nil {
		return errors.Wrap(err, "eventsList Bind error")
	}

	var data []EventInfo

	if err := h.DB.Select(&data, `select * from api_sight."eventList"($1, $2, $3);`, club_id, params.StartTime, params.StopTime); err != nil {
		log.WithFields(log.Fields{
			"proc":  "eventsList",
			"error": err,
		}).Error("SQL error")
		return errors.Wrap(err, "SQL error")
	}

	return c.Result(data)

}

func (h *handler) eventsGet(c jrpc.Context) error {
	claims := c.EchoContext().Get("user").(*jwt.Token).Claims.(*UserClaims)
	club_id := claims.Data["club_id"]

	var id string

	if err := c.Bind(&id); err != nil {
		return errors.Wrap(err, "eventsGet Bind error")
	}

	var data EventInfo

	if err := h.DB.Get(&data, `select * from api_sight."eventGet"($1, $2);`, club_id, id); err != nil {
		log.WithFields(log.Fields{
			"proc":     "eventsGet",
			"event_id": id,
			"error":    err,
		}).Error("SQL error")
		return errors.Wrap(err, "SQL error")
	}

	return c.Result(data)

}

func (h *handler) splitsPlayers(c jrpc.Context) error {
	claims := c.EchoContext().Get("user").(*jwt.Token).Claims.(*UserClaims)
	club_id := claims.Data["club_id"]

	var params struct {
		EventIds []string `json:"event_ids" db:"event_ids"`
		SplitIds []string `json:"split_ids" db:"split_ids"`
	}

	if err := c.Bind(&params); err != nil {
		return errors.Wrap(err, "eventsList Bind error")
	}

	var data []struct {
		Id   int32           `json:"player_id" db:"player_id"`
		Info json.RawMessage `json:"player_info" db:"player_info"`
	}

	if err := h.DB.Select(&data, `select * from api_sight."reportFilterPlayers"($1, $2, $3);`, club_id, pq.StringArray(params.EventIds), pq.StringArray(params.SplitIds)); err != nil {
		log.WithFields(log.Fields{
			"proc":  "eventsList",
			"error": err,
		}).Error("SQL error")
		return errors.Wrap(err, "SQL error")
	}

	return c.Result(data)

}

func (h *handler) splitsList(c jrpc.Context) error {
	claims := c.EchoContext().Get("user").(*jwt.Token).Claims.(*UserClaims)
	club_id := claims.Data["club_id"]

	var params []string

	if err := c.Bind(&params); err != nil {
		return errors.Wrap(err, "splitsList Bind error")
	}

	var data []SplitsInfo

	if err := h.DB.Select(&data, `select * from api_sight."splitsList"($1, $2);`, club_id, pq.StringArray(params)); err != nil {
		log.WithFields(log.Fields{
			"proc":  "splitsList",
			"error": err,
		}).Error("SQL error")
		return errors.Wrap(err, "SQL error")
	}

	return c.Result(data)

}
