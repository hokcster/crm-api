package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/mrFokin/jrpc"
)

const (
	queryClubGetUUID string = `SELECT * FROM api_sight."clubGet"($1::uuid);`
)

func (h *handler) InternalsValidator(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Request
		reqBody := []byte{}
		if c.Request().Body != nil { // Read
			reqBody, _ = ioutil.ReadAll(c.Request().Body)
		}
		c.Request().Body = ioutil.NopCloser(bytes.NewBuffer(reqBody)) // Reset

		passed_key := c.Request().Header.Get("x-access-token")
		if passed_key == "" {
			c.Echo().Logger.Errorj(map[string]interface{}{
				"reqBody": string(reqBody),
				"proc":    "InternalsValidator",
				"message": "No access token passed",
			})

			return echo.NewHTTPError(http.StatusForbidden)
		}

		contractor_secret := h.cfg.Locals.Secret

		calc_src := append([]byte(contractor_secret), reqBody...)
		hash_s := sha256.Sum256(calc_src)
		hash_src := hash_s[:]
		if hex.EncodeToString(hash_src) != passed_key {
			c.Echo().Logger.Errorj(map[string]interface{}{
				"passed_token": passed_key,
				"calc_token":   hex.EncodeToString(hash_src),
				"reqBody":      string(reqBody),
				"proc":         "B2BValidator",
				"message":      "Tokens dont match",
			})
			return echo.NewHTTPError(http.StatusForbidden)
		}

		return next(c)
	}
}

func (h *handler) getClaims(c jrpc.Context) error {
	var params struct {
		User       string `json:"user_id"`
		Contractor string `json:"contractor_id"`
		Domain     string `json:"domain"`
	}

	if err := c.Bind(&params); err != nil {
		c.EchoContext().Echo().Logger.Errorj(map[string]interface{}{
			"error":   err,
			"proc":    "eventGetClaims",
			"message": "Bind error",
		})
		return err
	}

	var data struct {
		Roles  *json.RawMessage `json:"roles" db:"roles"`
		Claims *json.RawMessage `json:"claims" db:"claims"`
	}

	roles := json.RawMessage("[]")
	data.Roles = &roles

	var clubData struct {
		ID         int              `json:"id" db:"id"`
		Name       string           `json:"name" db:"name"`
		UpdateTime time.Time        `json:"updated_at" db:"updated_at"`
		CreateTime time.Time        `json:"created_at" db:"created_at"`
		Params     *json.RawMessage `json:"params" db:"params"`
	}

	if err := h.DB.Get(&clubData, queryClubGetUUID, params.Contractor); err != nil {
		c.EchoContext().Echo().Logger.Errorj(map[string]interface{}{
			"error":   err,
			"proc":    "getClaims",
			"message": "SQL error",
			"club_id": params.Contractor,
		})

		return err
	}

	claims := map[string]interface{}{
		"id": params.User,
		"data": map[string]interface{}{
			"club_id": clubData.ID,
		},
	}

	claims_data, err := json.Marshal(claims)
	if err != nil {
		c.EchoContext().Echo().Logger.Errorj(map[string]interface{}{
			"error":      err,
			"proc":       "eventGetClaims",
			"message":    "SQL error",
			"domain":     params.Domain,
			"user":       params.User,
			"contractor": params.Contractor,
		})
	}

	data.Claims = (*json.RawMessage)(&claims_data)

	return c.Result(data)
}
