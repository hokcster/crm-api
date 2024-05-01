package main

import (
	"fmt"

	"github.com/golang-jwt/jwt"
	"github.com/jmoiron/sqlx"
	"github.com/mrFokin/jrpc"
	"github.com/pkg/errors"
)

type handler struct {
	DB  *sqlx.DB
	jwt cfgJWT
	cfg Config
}

type UserClaims struct {
	ID          string                 `json:"id"`
	Permissions []int32                `json:"permissions"`
	EventID     *int                   `json:"event_id"`
	Contractor  string                 `json:"contractor"`
	Data        map[string]interface{} `json:"data"`
	jwt.StandardClaims
}

func newHandler(cfg Config) (h handler, err error) {
	db := cfg.DB
	h.jwt = cfg.JWT
	h.cfg = cfg

	dbSource := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", db.Host, db.Port, db.User, db.Password, db.Name)
	h.DB, err = sqlx.Connect("postgres", dbSource)
	if err != nil {
		return h, errors.Wrap(err, "Connect SQL error")
	}

	h.DB.SetMaxOpenConns(db.MaxOpenConns)
	h.DB.SetMaxIdleConns(db.MaxIdleConns)
	h.DB.SetConnMaxLifetime(db.ConnMaxLifetime)
	return
}

func (h *handler) doCheckPermission(permissions []int32, permission int32) bool {
	for _, s := range permissions {
		if permission == s {
			return true
		}
	}
	return false
}

func (h *handler) checkPermissions(permissions []int32) jrpc.MiddlewareFunc {
	return func(next jrpc.HandlerFunc) jrpc.HandlerFunc {
		return func(c jrpc.Context) error {
			claims := c.EchoContext().Get("user").(*jwt.Token).Claims.(*UserClaims)

			for _, permission := range permissions {
				if !h.doCheckPermission(claims.Permissions, permission) {
					return ErrorForbidden
				}
			}

			return next(c)
		}
	}
}
