package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/PCManiac/logrus_init"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"github.com/mrFokin/jrpc"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type UserRolesRow struct {
	ID      string `json:"id" sql:",type:uuid" db:"id"`
	Caption string `json:"caption" db:"caption"`
}

type UsersRow struct {
	ID          string           `json:"id" sql:",type:uuid" db:"id"`
	Login       string           `json:"login" db:"login"`
	Password    string           `json:"password" db:"password"`
	Data        *json.RawMessage `json:"data" db:"data"`
	Active      bool             `json:"active" db:"active"`
	PrivateData *json.RawMessage `json:"private_data" db:"private_data"`
	Contractor  string           `json:"contractor" sql:",type:uuid" db:"contractor"`
}

type RolePermissionsRow struct {
	Role       string `json:"role" sql:",type:uuid" db:"role"`
	Permission int16  `json:"permission" db:"permission"`
}

type RoleMembershipRow struct {
	Role   string `json:"role" sql:",type:uuid" db:"role"`
	UserID string `json:"user_id" sql:",type:uuid" db:"user_id"`
}

type ContractorsRow struct {
	ID      string `json:"id" sql:",type:uuid" db:"id"`
	Caption string `json:"caption" db:"caption"`
}

type ClubsRow struct {
	ID         int              `json:"id" db:"id"`
	Name       string           `json:"name" db:"name"`
	UpdateTime *time.Time       `json:"updated_at" db:"updated_at"`
	CreateTime time.Time        `json:"created_at" db:"created_at"`
	Params     *json.RawMessage `json:"params" db:"params"`
	UID        string           `json:"uid" sql:",type:uuid" db:"uid"`
}

type TeamsRow struct {
	ID         int              `json:"id" db:"id"`
	Name       string           `json:"name" db:"name"`
	ClubID     int              `json:"club_id" db:"club_id"`
	UpdateTime *time.Time       `json:"updated_at" db:"updated_at"`
	CreateTime time.Time        `json:"created_at" db:"created_at"`
	Data       *json.RawMessage `json:"data" db:"data"`
}

type PositionsRow struct {
	ID     int    `json:"id" db:"id"`
	ClubID int    `json:"club_id" db:"club_id"`
	Name   string `json:"name" db:"name"`
	Alias  string `json:"alias" db:"alias"`
}

type PlayersRow struct {
	ID         int              `json:"id" db:"id"`
	ClubID     int              `json:"club_id" db:"club_id"`
	TeamID     *int             `json:"team_id" db:"team_id"`
	FName      *string          `json:"f_name" db:"f_name"`
	MName      *string          `json:"m_name" db:"m_name"`
	LName      *string          `json:"l_name" db:"l_name"`
	Gender     *int16           `json:"gender" db:"gender"`
	BirthDate  *time.Time       `json:"birth_date" db:"birth_date"`
	Jersey     *string          `json:"jersey" db:"jersey"`
	PositionID *int             `json:"position" db:"position"`
	Weight     *float32         `json:"current_weight" db:"current_weight"`
	Height     *float32         `json:"current_height" db:"current_height"`
	PhotoID    *int             `json:"photo_id" db:"photo_id"`
	Active     bool             `json:"active" db:"active"`
	MaxPulse   *int16           `json:"max_pulse" db:"max_pulse"`
	Data       *json.RawMessage `json:"data" db:"data"`
}

type FilesRow struct {
	ID   int             `json:"id" db:"id"`
	Data json.RawMessage `json:"file_data" db:"file_data"`
}

// ############### Типы данных для обратной репликации ###################
type EventSensorsRow struct {
	EventId  string `json:"event_id" db:"event_id"`
	SensorId int    `json:"sensor_id" db:"sensor_id"`
	PlayerID int    `json:"player_id" db:"player_id"`
}

type SplitRow struct {
	Id        string          `json:"id" db:"id"`
	EventId   string          `json:"event" db:"event"`
	StartTime time.Time       `json:"start_time" db:"start_time"`
	StopTime  time.Time       `json:"stop_time" db:"stop_time"`
	Tags      json.RawMessage `json:"tags" db:"tags"`
}

type SplitPlayersRow struct {
	SplitId  string `json:"split_id" db:"split_id"`
	PlayerID int    `json:"player_id" db:"player_id"`
}

type SplitReportData struct {
	SplitId          string     `json:"split_id" db:"split_id"`
	PlayerId         int        `json:"player_id" db:"player_id"`
	SumLength        float32    `json:"sum_length" db:"sum_length"`
	LpsSeconds       float32    `json:"lps_seconds" db:"lps_seconds"`
	DopplerLen       float32    `json:"doppler_len" db:"doppler_len"`
	MaxSpeed         float32    `json:"max_speed" db:"max_speed"`
	MaxAcceleration  float32    `json:"max_acceleration" db:"max_acceleration"`
	LenInSpeedZones  [5]float32 `json:"len_in_speed_zones" db:"len_in_speed_zones"`
	JumpCount        int64      `json:"jump_count" db:"jump_count"`
	CountLoad        int64      `json:"count_load_data" db:"count_load_data"`
	MaxLoad          int32      `json:"max_load" db:"max_load"`
	SumLoad          int64      `json:"sum_load" db:"sum_load"`
	TimeInSpeedZones [5]float32 `json:"time_in_speed_zones" db:"time_in_speed_zones"`
	CountPulse       int64      `json:"count_pulse_values" db:"count_pulse_values"`
	SumPulse         int64      `json:"sum_pulse_values" db:"sum_pulse_values"`
	MaxPulse         int16      `json:"max_pulse" db:"max_pulse"`
	TimeInHrZones    [5]float32 `json:"time_in_hr_zones" db:"time_in_hr_zones"`
	ImpactCount      int64      `json:"impact_count" db:"impact_count"`
	AccelCount       int64      `json:"accel_count" db:"accel_count"`
	StopCount        int64      `json:"stop_count" db:"stop_count"`
	MaxAccelPow      float32    `json:"max_accel_pow" db:"max_accel_pow"`
	MaxStopPow       float32    `json:"max_stop_pow" db:"max_stop_pow"`
	AccCntByZones    [4]int64   `json:"acceleration_cnt_by_zones" db:"acceleration_cnt_by_zones"`
	AccLenByZones    [4]float32 `json:"acceleration_length_by_zones" db:"acceleration_length_by_zones"`
	StopCntByZones   [4]int64   `json:"stop_count_by_zones" db:"stop_count_by_zones"`
}

type ReverseRequest struct {
	Event           EventInfo         `json:"event"`
	EventSensors    []EventSensorsRow `json:"event_sensors"`
	Splits          []SplitRow        `json:"splits"`
	SplitPlayers    []SplitPlayersRow `json:"split_players"`
	SplitReportData []json.RawMessage `json:"split_report_data"`
}

func (h *handler) ReplicationMiddlewareAuth(username, password string, c echo.Context) (bool, error) {
	var data struct {
		secret  sql.NullString `db:"secret"`
		club_id sql.NullInt32  `db:"club_id"`
		hw      sql.NullString `db:"hw"`
	}

	if err := h.DB.QueryRow(`select * from api_replication."getBoardSecret"($1);`, username).Scan(&data.secret, &data.club_id, &data.hw); err != nil {
		log.WithFields(log.Fields{
			"username": username,
			"password": password,
			"error":    err,
			"proc":     "ReplicationMiddlewareAuth",
		}).Error("ReplicationMiddlewareAuth SQL error")
		return false, errors.Wrap(err, "ReplicationMiddlewareAuth SQL error")
	}
	if !data.secret.Valid {
		log.WithFields(log.Fields{
			"username": username,
			"password": password,
			"proc":     "ReplicationMiddlewareAuth",
		}).Error("ReplicationMiddlewareAuth no such username")
		return false, errors.New("ReplicationMiddlewareAuth no such username " + username + " " + password)
	}

	if data.secret.String != password {
		log.WithFields(log.Fields{
			"username": username,
			"password": password,
			"proc":     "ReplicationMiddlewareAuth",
		}).Error("ReplicationMiddlewareAuth passwords don't match")

		return false, nil
	}

	if data.hw.Valid {
		var req_id string = ""
		var dev_id string = ""
		var digest string = ""
		var in_signature string = ""
		for header_name, header_values := range c.Request().Header {
			if strings.ToLower(header_name) == "x-request-id" {
				for _, v := range header_values {
					req_id = req_id + v
				}
			}
			if strings.ToLower(header_name) == "x-device-id" {
				for _, v := range header_values {
					dev_id = dev_id + v
				}
			}
			if strings.ToLower(header_name) == "x-content-digest" {
				for _, v := range header_values {
					digest = digest + v
				}
			}
			if strings.ToLower(header_name) == "x-signature" {
				for _, v := range header_values {
					in_signature = in_signature + v
				}
			}
		}
		if username != dev_id {
			log.WithFields(log.Fields{
				"proc": "ReplicationMiddlewareAuth",
				"url":  c.Request().RequestURI,
			}).Error("Request host id did not match")
			return false, nil
		}
		hash_src := req_id + dev_id + digest + data.hw.String
		ha := sha256.New()
		ha.Write([]byte(hash_src))
		calculated := base64.StdEncoding.EncodeToString(ha.Sum(nil))
		if in_signature != calculated {
			log.WithFields(log.Fields{
				"proc": "ReplicationMiddlewareAuth",
				"url":  c.Request().RequestURI,
			}).Error("Request signature id did not match")
			return false, nil
		}
	}

	c.Set("club_id", data.club_id.Int32)
	c.Set("board_id", username)

	return true, nil
}

//############### Юзера ###################

func (h *handler) replicationContractorGet(c jrpc.Context) error {
	club_id := c.EchoContext().Get("club_id").(int32)

	var data ContractorsRow

	if err := h.DB.Get(&data, `select * from api_replication."getContractor"($1);`, club_id); err != nil {
		c.EchoContext().Echo().Logger.Errorj(map[string]interface{}{
			"error":   err,
			"proc":    "replicationContractorGet",
			"message": "SQL error",
		})
		return errors.Wrap(err, "replicationContractorGet Get error")
	}

	return c.Result(data)
}

func (h *handler) replicationRolesList(c jrpc.Context) error {
	roles := []UserRolesRow{}
	if err := h.DB.Select(&roles, `SELECT * FROM api_replication."listRoles"();`); err != nil {
		c.EchoContext().Echo().Logger.Errorj(map[string]interface{}{
			"error":   err,
			"proc":    "replicationRolesList",
			"message": "SQL error",
		})
		return err
	}
	return c.Result(roles)
}

func (h *handler) replicationUserList(c jrpc.Context) error {
	club_id := c.EchoContext().Get("club_id").(int32)

	users := []UsersRow{}
	if err := h.DB.Select(&users, `SELECT * FROM api_replication."listUsers"($1);`, club_id); err != nil {

		c.EchoContext().Echo().Logger.Errorj(map[string]interface{}{
			"error":   err,
			"proc":    "replicationUserList",
			"message": "SQL error",
		})

		return err
	}
	return c.Result(users)
}

func (h *handler) replicationRolePermissions(c jrpc.Context) error {
	roles := []RolePermissionsRow{}
	if err := h.DB.Select(&roles, `SELECT * FROM api_replication."listRolePermissions"();`); err != nil {
		c.EchoContext().Echo().Logger.Errorj(map[string]interface{}{
			"error":   err,
			"proc":    "replicationRolePermissions",
			"message": "SQL error",
		})
		return err
	}
	return c.Result(roles)
}

func (h *handler) replicationRoleMembership(c jrpc.Context) error {
	club_id := c.EchoContext().Get("club_id").(int32)

	roles := []RoleMembershipRow{}
	if err := h.DB.Select(&roles, `SELECT * FROM api_replication."listMembership"($1);`, club_id); err != nil {
		c.EchoContext().Echo().Logger.Errorj(map[string]interface{}{
			"error":   err,
			"proc":    "replicationRoleMembership",
			"message": "SQL error",
		})
		return err
	}
	return c.Result(roles)
}

//############### Спорт ###################

func (h *handler) replicationClubGet(c jrpc.Context) error {
	club_id := c.EchoContext().Get("club_id").(int32)
	board_id := c.EchoContext().Get("board_id").(string)

	var data ClubsRow

	if err := h.DB.Get(&data, `select * from api_replication."getClubData"($1, $2);`, club_id, board_id); err != nil {
		c.EchoContext().Echo().Logger.Errorj(map[string]interface{}{
			"error":   err,
			"proc":    "replicationClubGet",
			"message": "SQL error",
		})
		return errors.Wrap(err, "replicationClubGet Get error")
	}

	return c.Result(data)
}

func (h *handler) replicationTeamsList(c jrpc.Context) error {
	club_id := c.EchoContext().Get("club_id").(int32)

	data := []TeamsRow{}
	if err := h.DB.Select(&data, `SELECT * FROM api_replication."listTeams"($1);`, club_id); err != nil {
		c.EchoContext().Echo().Logger.Errorj(map[string]interface{}{
			"error":   err,
			"proc":    "replicationTeamsList",
			"message": "SQL error",
		})
		return err
	}
	return c.Result(data)
}

func (h *handler) replicationPositionsList(c jrpc.Context) error {
	club_id := c.EchoContext().Get("club_id").(int32)

	data := []PositionsRow{}
	if err := h.DB.Select(&data, `SELECT * FROM api_replication."listPositions"($1);`, club_id); err != nil {
		c.EchoContext().Echo().Logger.Errorj(map[string]interface{}{
			"error":   err,
			"proc":    "replicationTeamsList",
			"message": "SQL error",
		})
		return err
	}
	return c.Result(data)
}

func (h *handler) replicationPlayersList(c jrpc.Context) error {
	club_id := c.EchoContext().Get("club_id").(int32)

	data := []PlayersRow{}
	if err := h.DB.Select(&data, `SELECT * FROM api_replication."listPlayers"($1);`, club_id); err != nil {
		c.EchoContext().Echo().Logger.Errorj(map[string]interface{}{
			"error":   err,
			"proc":    "replicationPlayersList",
			"message": "SQL error",
		})
		return err
	}
	return c.Result(data)
}

func (h *handler) replicationFilesList(c jrpc.Context) error {
	club_id := c.EchoContext().Get("club_id").(int32)

	data := []FilesRow{}
	if err := h.DB.Select(&data, `SELECT * FROM api_replication."listFiles"($1);`, club_id); err != nil {
		c.EchoContext().Echo().Logger.Errorj(map[string]interface{}{
			"error":   err,
			"proc":    "replicationPlayersList",
			"message": "SQL error",
		})
		return err
	}
	return c.Result(data)
}

func (h *handler) replicationGetFile(c echo.Context) error {
	file_id := c.Param("id")
	club_id := c.Get("club_id").(int32)

	file_info := new(PlayerPhotoFileInfo)
	err := h.DB.QueryRow(`SELECT * FROM api_replication."getFile"($1, $2);`, file_id, club_id).Scan(&file_info.Id, &file_info.FileData)
	if err != nil {
		c.Echo().Logger.Errorj(map[string]interface{}{
			"error":   err,
			"proc":    "replicationGetFile",
			"message": "SQL error",
		})

		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	filepath := strings.TrimSuffix(h.cfg.FilesDir, "/") + "/" + file_info.FileData.FileSystemName
	mime_type := file_info.FileData.MimeType

	f, err := os.Open(filepath)
	if err != nil {
		c.Echo().Logger.Errorj(map[string]interface{}{
			"error":    err,
			"proc":     "replicationGetFile",
			"filepath": filepath,
			"message":  "Open file error",
		})

		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}
	return c.Stream(http.StatusOK, mime_type, f)
}

func (h *handler) replicationGetClubLogo(c echo.Context) error {
	club_id := c.Get("club_id").(int32)

	var club_info ClubInfo

	if err := h.DB.Get(&club_info, `select id, "name", updated_at, created_at, params from api_replication."getClub"($1);`, club_id); err != nil {
		return errors.Wrap(err, "replicationGetClubLogo Get error")
	}

	filepath := strings.TrimSuffix(h.cfg.AssetsDir, "/") + "/404.png"
	logo_mime := "image/png"

	switch logo_path := club_info.Params["logo_path"].(type) {
	case string:
		filepath = strings.TrimSuffix(h.cfg.FilesDir, "/") + "/" + logo_path
	}

	switch mime := club_info.Params["logo_mime"].(type) {
	case string:
		logo_mime = mime
	}

	f, err := os.Open(filepath)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}
	return c.Stream(http.StatusOK, logo_mime, f)
}

// ############### Обратная репликация ###################
func (h *handler) saveCalculatedEvent(c jrpc.Context) error {
	club_id := c.EchoContext().Get("club_id").(int32)

	var params ReverseRequest

	if err := c.Bind(&params); err != nil {
		return errors.Wrap(err, "saveCalculatedEvent Bind error")
	}

	TX, err := h.DB.Beginx()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"proc":  "saveCalculatedEvent",
		}).Error("Beginx error")
		return errors.Wrap(err, "Beginx error")
	}

	defer func(TX *sqlx.Tx) {
		r := recover()
		if r != nil {
			if err = TX.Rollback(); err != nil {
				log.WithFields(log.Fields{
					"error": err,
					"proc":  "saveCalculatedEvent defer",
				}).Error("Rollback error")
			}
		} else {
			if err = TX.Commit(); err != nil {
				log.WithFields(log.Fields{
					"error": err,
					"proc":  "saveCalculatedEvent defer",
				}).Error("Commit error")
			}
			log.WithFields(log.Fields{
				"proc": "saveCalculatedEvent defer",
			}).Trace("Finished")
		}
	}(TX)

	if _, err := TX.Exec(`select * from api_replication."prepareCalculatedEvent"($1, $2);`, club_id, params.Event.Id); err != nil {
		log.WithFields(log.Fields{
			"proc":  "saveCalculatedEvent",
			"SQL":   "prepareCalculatedEvent",
			"error": err,
		}).Error("SQL error")
		return errors.Wrap(err, "SQL error")
	}

	if _, err := TX.Exec(`select * from api_replication."eventsAdd"($1, $2, $3, $4, $5, $6);`,
		club_id, params.Event.Id, params.Event.Name, params.Event.TeamId, params.Event.StartTime, params.Event.StopTime); err != nil {
		log.WithFields(log.Fields{
			"proc":   "saveCalculatedEvent",
			"SQL":    "eventsAdd",
			"params": params,
			"error":  err,
		}).Error("SQL error")

		if strings.Contains(err.Error(), "Splits overlapped") {
			return ErrorSplitsOverlapped
		}
		return errors.Wrap(err, "SQL error")
	}

	for _, eventSensor := range params.EventSensors {
		if _, err := TX.Exec(`select * from api_replication."eventsSetPlayerSensor"($1, $2, $3, $4);`,
			club_id, eventSensor.EventId, eventSensor.PlayerID, eventSensor.SensorId); err != nil {
			log.WithFields(log.Fields{
				"proc":   "saveCalculatedEvent",
				"SQL":    "eventsSetPlayerSensor",
				"params": params,
				"error":  err,
			}).Error("SQL error")
			return errors.Wrap(err, "SQL error")
		}
	}

	for _, split := range params.Splits {
		if _, err := TX.Exec(`select * from api_replication."splitsAdd"($1, $2, $3, $4, $5, $6);`,
			club_id, split.Id, split.EventId, split.StartTime, split.StopTime, split.Tags); err != nil {
			log.WithFields(log.Fields{
				"proc":   "saveCalculatedEvent",
				"SQL":    "splitsAdd",
				"params": params,
				"error":  err,
			}).Error("SQL error")
			return errors.Wrap(err, "SQL error")
		}
	}

	for _, player := range params.SplitPlayers {
		if _, err := TX.Exec(`select * from api_replication."splitsPlayersAdd"($1, $2, $3);`,
			club_id, player.SplitId, player.PlayerID); err != nil {
			log.WithFields(log.Fields{
				"proc":   "saveCalculatedEvent",
				"SQL":    "splitsPlayersAdd",
				"params": params,
				"error":  err,
			}).Error("SQL error")

			if strings.Contains(err.Error(), "Splits overlapped") {
				return ErrorSplitsOverlapped
			}
			return errors.Wrap(err, "SQL error")
		}
	}

	for _, reportData := range params.SplitReportData {
		if _, err := TX.Exec(`select * from api_replication."splitsReportDataAdd"($1, $2);`,
			club_id, reportData); err != nil {
			log.WithFields(log.Fields{
				"proc":   "saveCalculatedEvent",
				"SQL":    "splitsReportDataAdd",
				"params": params,
				"error":  err,
			}).Error("SQL error")
			return errors.Wrap(err, "SQL error")
		}
	}

	return c.Result(true)

}

func (h *handler) uploadFile(c echo.Context, path string) error {
	club_id := c.Get("club_id").(int32)
	board_id := c.Get("board_id").(string)

	file, err := c.FormFile("file")
	if err != nil {
		log.WithFields(log.Fields{
			"proc":  "uploadLogFile",
			"error": err,
		}).Error("FormFile error")

		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	src, err := file.Open()
	if err != nil {
		log.WithFields(log.Fields{
			"proc":  "uploadLogFile",
			"error": err,
		}).Error("Open error")

		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	defer src.Close()

	newDirPath := strings.TrimSuffix(h.cfg.FilesDir, "/") + "/" + path + "/" + strconv.Itoa(int(club_id)) + "/" + board_id
	err = os.MkdirAll(newDirPath, os.ModePerm)
	if err != nil {
		log.WithFields(log.Fields{
			"proc":  "uploadLogFile",
			"error": err,
		}).Error("MkdirAll error")

		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	newFilePath := newDirPath + "/" + file.Filename

	dst, err := os.Create(newFilePath)
	if err != nil {
		log.WithFields(log.Fields{
			"proc":  "uploadLogFile",
			"error": err,
		}).Error("Create error")

		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	defer dst.Close()

	if _, err = io.Copy(dst, src); err != nil {
		log.WithFields(log.Fields{
			"proc":  "uploadLogFile",
			"error": err,
		}).Error("Copy error")

		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, "")
}

func (h *handler) uploadLogFile(c echo.Context) error {
	return h.uploadFile(c, "logs")
}

func (h *handler) uploadRawFile(c echo.Context) error {
	return h.uploadFile(c, "raw")
}

func (h *handler) replicationGetCompose(c echo.Context) error {
	club_id := c.Get("club_id").(int32)
	board_id := c.Get("board_id").(string)

	filePath1 := strings.TrimSuffix(h.cfg.FilesDir, "/") + "/update"
	filePath2 := strings.TrimSuffix(h.cfg.FilesDir, "/") + "/update/" + strconv.Itoa(int(club_id))
	filePath3 := strings.TrimSuffix(h.cfg.FilesDir, "/") + "/update/" + strconv.Itoa(int(club_id)) + "/" + board_id

	var composePath string
	if _, err := os.Stat(filePath3 + "/docker-compose.yml"); err == nil {
		composePath = filePath3 + "/docker-compose.yml"
	} else if _, err := os.Stat(filePath2 + "/docker-compose.yml"); err == nil {
		composePath = filePath2 + "/docker-compose.yml"
	} else if _, err := os.Stat(filePath1 + "/docker-compose.yml"); err == nil {
		composePath = filePath1 + "/docker-compose.yml"
	} else {
		return echo.NewHTTPError(http.StatusNotFound, "file not found")
	}

	f, err := os.Open(composePath)

	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	return c.Stream(http.StatusOK, "application/yml", f)
}

func (h *handler) replicationPing(c jrpc.Context) error {
	club_id := c.EchoContext().Get("club_id").(int32)
	board_id := c.EchoContext().Get("board_id").(string)

	log.WithFields(log.Fields{
		"proc":     "replicationPing",
		"club_id":  club_id,
		"board_id": board_id,
	}).Error("Board is up")

	return c.Result(true)
}

func (h *handler) portableStub(c jrpc.Context) error {

	return c.Result(false)
}
