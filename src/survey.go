package main

import (
	"encoding/json"
	"time"

	_ "github.com/PCManiac/logrus_init"
	"github.com/golang-jwt/jwt"
	"github.com/lib/pq"
	"github.com/mrFokin/jrpc"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type SurveyEvent struct {
	Rating *int16 `json:"rating" db:"rating"`
}

type SurveyEventInfo struct {
	Id        string    `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	TeamId    int32     `json:"team_id" db:"team_id"`
	StartTime time.Time `json:"start_time" db:"start_time"`
	StopTime  time.Time `json:"stop_time" db:"stop_time"`
	Rating    *int16    `json:"rating" db:"rating"`
}

func (h *handler) surveyEventsList(c jrpc.Context) error {
	claims := c.EchoContext().Get("user").(*jwt.Token).Claims.(*UserClaims)
	club_id := claims.Data["club_id"]
	user_id := claims.ID

	var params struct {
		Limit  int   `json:"limit" db:"limit"`
		Offset int64 `json:"offset" db:"offset"`
	}

	if err := c.Bind(&params); err != nil {
		return errors.Wrap(err, "surveyEventsList Bind error")
	}

	var data []SurveyEventInfo

	if err := h.DB.Select(&data, `select * from api_sight."surveyEventList"($1, $2, $3, $4);`, user_id, club_id, params.Limit, params.Offset); err != nil {
		log.WithFields(log.Fields{
			"proc":  "surveyEventsList",
			"error": err,
		}).Error("SQL error")
		return errors.Wrap(err, "SQL error")
	}

	return c.Result(data)

}

func (h *handler) surveyEventsGet(c jrpc.Context) error {
	claims := c.EchoContext().Get("user").(*jwt.Token).Claims.(*UserClaims)
	user_id := claims.ID

	var id string

	if err := c.Bind(&id); err != nil {
		return errors.Wrap(err, "surveyEventsGet Bind error")
	}

	var data json.RawMessage

	if err := h.DB.Get(&data, `select * from api_sight."surveyEventGet"($1, $2);`, user_id, id); err != nil {
		log.WithFields(log.Fields{
			"proc":     "surveyEventsGet",
			"event_id": id,
			"error":    err,
		}).Error("SQL error")
		return errors.Wrap(err, "SQL error")
	}

	return c.Result(data)
}

func (h *handler) surveyEventResponse(c jrpc.Context) error {
	claims := c.EchoContext().Get("user").(*jwt.Token).Claims.(*UserClaims)
	user_id := claims.ID

	var params struct {
		EventId  string      `json:"event_id" db:"event_id"`
		Response SurveyEvent `json:"response" db:"response"`
	}

	if err := c.Bind(&params); err != nil {
		return errors.Wrap(err, "surveyEventResponse Bind error")
	}

	resp, err := json.Marshal(params.Response)
	if err != nil {
		return errors.Wrap(err, "surveyEventResponse Marshal error")
	}

	if _, err := h.DB.Exec(`select * from api_sight."surveyEventResponse"($1, $2, $3);`, user_id, params.EventId, resp); err != nil {
		log.WithFields(log.Fields{
			"proc":   "surveyEventResponse",
			"params": params,
			"error":  err,
		}).Error("SQL error")
		return errors.Wrap(err, "SQL error")
	}

	return c.Result(true)
}

func (h *handler) surveyDailyList(c jrpc.Context) error {
	claims := c.EchoContext().Get("user").(*jwt.Token).Claims.(*UserClaims)
	club_id := claims.Data["club_id"]

	var params struct {
		Team int       `json:"team_id" db:"team_id"`
		Date time.Time `json:"date" db:"date"`
	}

	if err := c.Bind(&params); err != nil {
		return errors.Wrap(err, "surveyDailyList Bind error")
	}

	var data []struct {
		PlayerId   int        `json:"player_id" db:"player_id"`
		PlayerInfo ClubParams `json:"player_info" db:"player_info"`
		Response   ClubParams `json:"response" db:"response"`
	}

	if err := h.DB.Select(&data, `select * from api_sight."surveyDailyGetAll"($1, $2, $3);`, club_id, params.Team, params.Date); err != nil {
		log.WithFields(log.Fields{
			"proc":  "surveyDailyList",
			"error": err,
		}).Error("SQL error")
		return errors.Wrap(err, "SQL error")
	}

	return c.Result(data)
}

func (h *handler) surveyDailyGet(c jrpc.Context) error {
	claims := c.EchoContext().Get("user").(*jwt.Token).Claims.(*UserClaims)
	club_id := claims.Data["club_id"]
	user_id := claims.ID

	var params time.Time

	if err := c.Bind(&params); err != nil {
		return errors.Wrap(err, "surveyDailyGet Bind error")
	}

	var data []struct {
		PlayerId   int        `json:"player_id" db:"player_id"`
		PlayerInfo ClubParams `json:"player_info" db:"player_info"`
		Response   ClubParams `json:"response" db:"response"`
	}

	if err := h.DB.Select(&data, `select * from api_sight."surveyDailyGetOne"($1, $2, $3);`, club_id, user_id, params); err != nil {
		log.WithFields(log.Fields{
			"proc":  "surveyDailyGet",
			"error": err,
		}).Error("SQL error")
		return errors.Wrap(err, "SQL error")
	}

	return c.Result(data)
}

func (h *handler) surveyDailyResponse(c jrpc.Context) error {
	claims := c.EchoContext().Get("user").(*jwt.Token).Claims.(*UserClaims)
	user_id := claims.ID

	var params struct {
		Date     time.Time  `json:"date" db:"date"`
		Response ClubParams `json:"response" db:"response"`
	}

	if err := c.Bind(&params); err != nil {
		return errors.Wrap(err, "surveyDailyResponse Bind error")
	}

	resp, err := json.Marshal(params.Response)
	if err != nil {
		return errors.Wrap(err, "surveyEventResponse Marshal error")
	}

	if _, err := h.DB.Exec(`select * from api_sight."surveyDailyResponse"($1, $2, $3);`, user_id, params.Date, resp); err != nil {
		log.WithFields(log.Fields{
			"proc":   "surveyDailyResponse",
			"params": params,
			"error":  err,
		}).Error("SQL error")
		return errors.Wrap(err, "SQL error")
	}

	return c.Result(true)
}

func (h *handler) surveyDailyDays(c jrpc.Context) error {
	claims := c.EchoContext().Get("user").(*jwt.Token).Claims.(*UserClaims)
	user_id := claims.ID

	var params struct {
		Year  int `json:"year" db:"year"`
		Month int `json:"month" db:"month"`
	}

	if err := c.Bind(&params); err != nil {
		return errors.Wrap(err, "surveyDailyList Bind error")
	}

	var data []time.Time

	if err := h.DB.Select(&data, `select * from api_sight."surveyDailyPlayerDates"($1, $2, $3);`, user_id, params.Year, params.Month); err != nil {
		log.WithFields(log.Fields{
			"proc":  "surveyDailyList",
			"error": err,
		}).Error("SQL error")
		return errors.Wrap(err, "SQL error")
	}

	return c.Result(data)
}

func (h *handler) surveyPlayer10Days(c jrpc.Context) error {
	claims := c.EchoContext().Get("user").(*jwt.Token).Claims.(*UserClaims)
	club_id := claims.Data["club_id"]

	var params struct {
		EventIds []string `json:"event_ids"`
		SplitIds []string `json:"split_ids"`
		PlayerId int      `json:"player_id"`
	}

	if err := c.Bind(&params); err != nil {
		return errors.Wrap(err, "surveyPlayer10Days Bind error")
	}

	var data []struct {
		PlayerId   int        `json:"player_id" db:"player"`
		PlayerInfo ClubParams `json:"player_info" db:"player_info"`
		Response   ClubParams `json:"response" db:"response"`
	}

	if err := h.DB.Select(&data, `select * from api_sight."reportSurvey"($1, $2, $3, $4);`,
		pq.StringArray(params.EventIds), pq.StringArray(params.SplitIds), params.PlayerId, club_id); err != nil {
		log.WithFields(log.Fields{
			"proc":  "surveyPlayer10Days",
			"error": err,
		}).Error("SQL error")
		return errors.Wrap(err, "SQL error")
	}

	return c.Result(data)
}
