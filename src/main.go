package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/PCManiac/compile_vars"
	"github.com/caarlos0/env/v6"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	_ "github.com/lib/pq"
	"github.com/mrFokin/jrpc"
	"github.com/mrFokin/sessions"
)

// Config описывает параметры конфигурации получаемые из переменных окружения
type Config struct {
	Host           string `env:"HOST" envDefault:":8080"`
	LocationPrefix string `env:"LOCATION_PREFIX" envDefault:""`
	RefreshPostfix string `env:"REFRESH_POSTFIX" envDefault:""`
	DB             cfgDB
	JWT            cfgJWT
	Locals         cfgLocals
	FilesDir       string `env:"FILES_PATH,required"`
	AssetsDir      string `env:"ASSETS_PATH"  envDefault:"/assets"`
}

type cfgDB struct {
	Host            string        `env:"DB_HOST" envDefault:"db"`
	Port            string        `env:"DB_PORT" envDefault:"5432"`
	User            string        `env:"DB_USER,required"`
	Password        string        `env:"DB_PASSWORD,required"`
	Name            string        `env:"DB_NAME" envDefault:"bsight"`
	MaxOpenConns    int           `env:"DB_MAX_OPEN_CONNECTIONS" envDefault:"10"`
	MaxIdleConns    int           `env:"DB_MAX_IDLE_CONNECTIONS" envDefault:"10"`
	ConnMaxLifetime time.Duration `env:"DB_CONN_MAX_TIME" envDefault:"3m"`
}

type cfgJWT struct {
	Secret string `env:"JWT_SECRET,required"`
	//Domain string `env:"DOMAIN,required"`
}

type cfgLocals struct {
	Secret string `env:"LOCALS_SECRET,required"`
}

func main() {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	e.Logger.SetLevel(log.DEBUG)

	e.Pre(middleware.RemoveTrailingSlash())
	e.Use(middleware.Recover())
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: `{"time":"${time_rfc3339_nano}","id":"${id}","remote_ip":"${remote_ip}","host":"${host}",` +
			`"http method":"${method}","uri":"${uri}","jrpc method":"${header:JRPC-Method}","status":${status},"error":"${error}","latency":${latency},` +
			`"latency_human":"${latency_human}","bytes_in":${bytes_in},` +
			`"bytes_out":${bytes_out}}` + "\n",
	}))

	var config Config
	if err := env.Parse(&config); err != nil {
		e.Logger.Fatal(err)
	}

	h, err := newHandler(config)
	if err != nil {
		e.Logger.Fatal("Попытка подключения к БД", err)
	}

	//#########   Общие методы   #########
	e.Static("/", config.AssetsDir)

	public := jrpc.Endpoint(e, config.LocationPrefix+"/public")
	public.Method("ping", h.ping)
	public.Method("version", h.versionCheck)

	//#########   Интеграция с аутентификатором   #########
	locals := jrpc.Endpoint(e, config.LocationPrefix+"/locals", middleware.BodyDump(logJrpcRequest), h.InternalsValidator)
	locals.Method("claims.get", h.getClaims)

	//#########   Методы репликации   #########
	replication := jrpc.Endpoint(e, config.LocationPrefix+"/replication", middleware.BasicAuth(h.ReplicationMiddlewareAuth))
	replication.Method("get.contractor", h.replicationContractorGet)
	replication.Method("list.roles", h.replicationRolesList)
	replication.Method("list.users", h.replicationUserList)
	replication.Method("list.roles.permissions", h.replicationRolePermissions)
	replication.Method("list.roles.membership", h.replicationRoleMembership)

	replication.Method("get.club", h.replicationClubGet)
	replication.Method("list.teams", h.replicationTeamsList)
	replication.Method("list.positions", h.replicationPositionsList)
	replication.Method("list.players", h.replicationPlayersList)
	replication.Method("list.files", h.replicationFilesList)

	replication.Method("ping", h.replicationPing)

	e.GET(config.LocationPrefix+"/replication/files/:id", h.replicationGetFile, middleware.BasicAuth(h.ReplicationMiddlewareAuth))
	e.GET(config.LocationPrefix+"/replication/club_logo", h.replicationGetClubLogo, middleware.BasicAuth(h.ReplicationMiddlewareAuth))
	e.POST(config.LocationPrefix+"/replication/sensor_log", h.uploadLogFile, middleware.BasicAuth(h.ReplicationMiddlewareAuth))
	e.POST(config.LocationPrefix+"/replication/sensor_raw", h.uploadRawFile, middleware.BasicAuth(h.ReplicationMiddlewareAuth))
	e.GET(config.LocationPrefix+"/replication/update.yml", h.replicationGetCompose, middleware.BasicAuth(h.ReplicationMiddlewareAuth))

	replication.Method("event.save", h.saveCalculatedEvent)

	portableStub := jrpc.Endpoint(e, config.LocationPrefix+"/replication/portable", middleware.BasicAuth(h.ReplicationMiddlewareAuth))
	portableStub.Method("version.update.needed", h.portableStub)
	portableStub.Method("version.update", h.portableStub)
	portableStub.Method("ssh.up", h.portableStub)
	portableStub.Method("ssh.dn", h.portableStub)

	//#########   Методы api   #########
	web := jrpc.Endpoint(e, config.LocationPrefix+"/web", sessions.JWTWithRedirect("/auth/refresh"+config.RefreshPostfix, []byte(config.JWT.Secret), &UserClaims{}) /*, middleware.BodyDump(logJrpcRequest)*/)
	web.Method("clubs.get", h.clubsGet)

	web.Method("teams.list", h.teamsList)
	web.Method("teams.create", h.teamsAdd)
	web.Method("teams.get", h.teamsGet)
	web.Method("teams.update", h.teamsUpdate)
	web.Method("teams.delete", h.teamsDelete)

	web.Method("positions.list", h.positionsList)
	web.Method("positions.create", h.positionsAdd)
	web.Method("positions.get", h.positionsGet)
	web.Method("positions.update", h.positionsUpdate)
	web.Method("positions.delete", h.positionsDelete)

	web.Method("players.list", h.playersList)
	web.Method("players.create", h.playersAdd)
	web.Method("players.get", h.playersGet)
	web.Method("players.update", h.playersUpdate)
	web.Method("players.delete", h.playersDelete)
	web.Method("players.password.reset", h.playersResetPassword)

	web.Method("events.list", h.eventsList)
	web.Method("events.get", h.eventsGet)
	web.Method("splits.players", h.splitsPlayers)
	web.Method("splits.list", h.splitsList)

	web.Method("survey.events.list", h.surveyEventsList)
	web.Method("survey.events.get", h.surveyEventsGet)
	web.Method("survey.events.response", h.surveyEventResponse)

	web.Method("survey.daily.list", h.surveyDailyList)
	web.Method("survey.daily.get", h.surveyDailyGet)
	web.Method("survey.daily.response", h.surveyDailyResponse)
	web.Method("survey.daily.days", h.surveyDailyDays)
	web.Method("survey.personal", h.surveyPlayer10Days)

	//#########   Отчеты   #########
	api := jrpc.Endpoint(e, config.LocationPrefix+"/reports", sessions.JWTWithRedirect("/auth/refresh"+config.RefreshPostfix, []byte(config.JWT.Secret), &UserClaims{}) /*, middleware.BodyDump(logJrpcRequest)*/)
	api.Method("reports.workout", h.reportWorkout)
	api.Method("reports.match.table", h.reportMatchTable)
	api.Method("reports.match.graph", h.reportMatchGraph)
	api.Method("reports.personal", h.reportPersonal)
	api.Method("reports.survey.event", h.reportEventSurvey)
	api.Method("reports.injure.graph", h.reportEventGraph)

	//Отчёты PDF на бекенде
	e.GET(config.LocationPrefix+"/report/workout", h.reportPDFWorkout, middleware.BasicAuth(h.ReplicationMiddlewareAuth))

	//api.Method("reports.recalculate", h.reportRecalculate)
	/*	api.Method("events.create", h.eventsAdd, h.checkPermissions([]int32{103}))
		api.Method("events.get", h.eventsGet)
		api.Method("events.domains.update", h.eventDomainsUpdate, h.checkPermissions([]int32{99, 103}))*/

	//#########   Файлы   #########
	pg := e.Group(config.LocationPrefix + "/files")
	pg.Use(sessions.JWTWithRedirect("/auth/refresh"+config.RefreshPostfix, []byte(config.JWT.Secret), &UserClaims{}))
	pg.GET("/player/:id", h.getPlayerPhoto)
	pg.POST("/player/:id", h.uploadPlayerPhoto)
	pg.GET("/club", h.getClubLogo)

	e.Logger.Debug("Started. version: ", compile_vars.GetVersion(), " build_time: ", compile_vars.GetBuildTime(), " config: ", fmt.Sprintf("%+v", config))

	go func() {
		if err := e.Start(config.Host); err != nil {
			e.Logger.Info("shutting down the server", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
}
