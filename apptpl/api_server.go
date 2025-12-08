package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/adaptor"
	"github.com/gofiber/fiber/v3/middleware/redirect"
	"github.com/osamingo/gosh"
	"github.com/patrickmn/go-cache"
	log "github.com/sirupsen/logrus"

	"goapptol/utils"
)

type ApiServer struct {
	Myconfig    *MyConfig
	app         *fiber.App
	mysqlHdl    *MysqlHandler
	minioHdl    *MinioHandler
	redisHdl    *RedisHandler
	ckHdl       *ClickhouseHandler
	pgHdl       *PgHandler
	hardwareHdl *HardwareHandler
	hostHdl     *HostHandler

	mycache *cache.Cache
}

func (p *ApiServer) Start() error {
	p.mycache = cache.New(5*time.Minute, 10*time.Minute)

	log.Info("🚀 API server prepare...")
	app := fiber.New(fiber.Config{
		CaseSensitive: true,
		StrictRouting: true,
		Immutable:     true,
		ServerHeader:  "goapptpl",
		AppName:       "Go App Template v" + utils.APP_VERSION,
		ReadTimeout:   30 * time.Second,
		WriteTimeout:  30 * time.Second,
		ProxyHeader:   fiber.HeaderXForwardedFor,
		UnescapePath:  false, // default false
	})

	p.initRoute(app)

	// add MysqlHandler
	mysqlHdl := MysqlHandler{DbHandler: DbHandler{Dbconfig: &p.Myconfig.MysqlConfig, Mycache: p.mycache}}
	mysqlHdl.AddRouter(app.Group("/mysql"))

	// add MinioHandler
	minioHdl := MinioHandler{Minioconfig: &p.Myconfig.MinioConfig}
	minioHdl.AddRouter(app.Group("/minio"))

	// add RedisHandler
	redisHdl := RedisHandler{Redisconfig: &p.Myconfig.RedisConfig}
	redisHdl.AddRouter(app.Group("/redis"))

	// add ClickhouseHandler
	ckHdl := ClickhouseHandler{DbHandler: DbHandler{Dbconfig: &p.Myconfig.CkConfig, Mycache: p.mycache}}
	ckHdl.AddRouter(app.Group("/clickhouse"))

	// add PgHandler
	pgHdl := PgHandler{DbHandler: DbHandler{Dbconfig: &p.Myconfig.PgConfig, Mycache: p.mycache}}
	pgHdl.AddRouter(app.Group("/postgresql"))

	// add HardwareHandler
	hardwareHdl := HardwareHandler{Mycache: p.mycache}
	hardwareHdl.AddRouter(app.Group("/hardware"))

	// add HostHandler
	hostHdl := HostHandler{Mycache: p.mycache}
	hostHdl.AddRouter(app.Group("/host"))

	// data, _ := json.MarshalIndent(app.Stack(), "", "  ")
	// log.Debug(string(data))
	// data, _ = json.MarshalIndent(app.Config(), "", "  ")
	// log.Debug("config: %s\n", data)

	p.app = app
	p.mysqlHdl = &mysqlHdl
	p.minioHdl = &minioHdl
	p.redisHdl = &redisHdl
	p.ckHdl = &ckHdl
	p.pgHdl = &pgHdl
	p.hardwareHdl = &hardwareHdl
	p.hostHdl = &hostHdl

	// use CertFile and CertKeyFile to listen https
	// 正常时阻塞在这里
	// err := app.Listen("[::]:"+strconv.Itoa(int(p.Myconfig.Port)),
	// 	fiber.ListenConfig{
	// 		CertFile:              "etc/cert.pem",
	// 		CertKeyFile:           "etc/key.pem",
	// 		DisableStartupMessage: false,
	// 		EnablePrintRoutes:     false,
	// 		ListenerNetwork:       "tcp", // listen ipv4 and ipv6
	// 		BeforeServeFunc: func(app *fiber.App) error {
	// 			log.Info("🚀 API server starting...")
	// 			return nil
	// 		},
	// 	})

	// use inside tls config from utils/cert.go. not use cert and key files any more
	ln, err := net.Listen("tcp", "[::]:"+strconv.Itoa(int(p.Myconfig.Port)))
	if err != nil {
		log.Fatalf("net listen port %d error: %s", p.Myconfig.Port, err.Error())
		return err
	}
	if p.Myconfig.SslEnable {
		ln = tls.NewListener(ln, utils.TLSConfig())
	}

	// 正常时阻塞在这里
	err = app.Listener(ln,
		fiber.ListenConfig{
			DisableStartupMessage: false,
			EnablePrintRoutes:     false,
			ListenerNetwork:       "tcp", // listen ipv4 and ipv6
			BeforeServeFunc: func(app *fiber.App) error {
				log.Info("🚀 API server starting...")
				return nil
			},
		})
	if err != nil {
		log.Fatalf("api server start error: %s", err.Error())
		return err
	}

	log.Debug("api server stop")
	return nil
}

func (p *ApiServer) Stop() error {
	if p.app != nil {
		err := p.app.ShutdownWithTimeout(1 * time.Second)
		// err := p.app.Shutdown()
		p.app = nil
		p.mysqlHdl.Close()
		p.mysqlHdl = nil
		p.minioHdl = nil
		p.redisHdl.Close()
		p.redisHdl = nil
		p.ckHdl.Close()
		p.ckHdl = nil
		p.pgHdl.Close()
		p.pgHdl = nil
		return err
	}
	return nil
}

func (p *ApiServer) initRoute(app *fiber.App) error {
	// Uer Middleware
	// Match any route
	// app.Use(func(c fiber.Ctx) error {
	// 	log.Trace("🥇 Any handler: " + c.Path())
	// 	return c.Next()
	// })
	app.Use(p.authMiddleware)

	// // Match all routes starting with /api
	// app.Use("/api", func(c fiber.Ctx) error {
	// 	log.Trace("🥈 Second handler")
	// 	return c.Next()
	// })

	// // Match request starting with /api
	// app.Use("/api", func(c fiber.Ctx) error {
	// 	log.Trace("🥈 third handler")
	// 	return c.Next()
	// })

	// // Match requests starting with /api or /home (multiple-prefix support)
	// app.Use([]string{"/api", "/home"}, func(c fiber.Ctx) error {
	// 	log.Trace("🥈 Fourth handler")
	// 	return c.Next()
	// })

	// // Attach multiple handlers
	// app.Use("/api", func(c fiber.Ctx) error {
	// 	c.Set("X-Custom-Header", random.String(32))
	// 	log.Trace("🥈 Fifth handler")
	// 	return c.Next()
	// }, func(c fiber.Ctx) error {
	// 	log.Trace("🥈 Fifth 2 handler")
	// 	return c.Next()
	// })

	// API meta routes
	// app.GET("/", handler.handleReadme)
	// app.GET("/api", handler.handleApi)
	// app.GET("/version", handler.handleVersion)
	// app.GET("/status", handler.handleStatus)
	// app.GET("/log", handler.handleLog)
	// app.GET("/dump", handler.handleDump)
	// app.GET("/errors", handler.handleErrors)
	// app.GET("/statistic", handler.handleStatistic)
	// app.GET("/config", handler.handleConfig)
	// app.GET("/health", handler.handleHealth)
	// app.GET("/cache", handler.handleCache)

	// Fiber redirect midlleware
	app.Use(redirect.New(redirect.Config{
		Rules: map[string]string{
			"/": "/meta",
		},
		StatusCode: 301,
	}))

	app.Get("/", func(c fiber.Ctx) error {
		return fiber.NewError(300, "please access /meta")
	})
	app.Get("/meta", func(c fiber.Ctx) error {
		c.Response().Header.Set("Content-Type", "text/html")
		return c.SendString(`<html><body><h1>Meta Information</h1>
		<a href="/meta/status">/meta/status</a><br>
		<a href="/meta/version">/meta/version</a><br>
		<a href="/meta/config">/meta/config</a><br>
		<h1>Sub modules</h1>
		<a href="/mysql">/mysql</a><br>
		<a href="/minio">/minio</a><br>
		<a href="/redis">/redis</a><br>
		<a href="/clickhouse">/clickhouse</a><br>
		<a href="/postgresql">/postgresql</a><br>
		<a href="/hardware">/hardware</a><br>
		<a href="/host">/host</a><br>
		</body></html>`)
	})
	app.Get("/meta/status", func(c fiber.Ctx) error {
		s := fmt.Sprintf(`{ "status": "%s", "runtime": "%s" }`,
			"running", START_TIME.Format(time.RFC3339)) // "2006-01-02 15:04:05"
		return c.SendString(s)
	})
	app.Get("/meta/version", func(c fiber.Ctx) error {
		return c.SendString(utils.Version("goapptpl")) // => ✋ versoin
	})
	app.Get("/meta/config", func(c fiber.Ctx) error {
		b, _ := json.Marshal(p.Myconfig)
		return c.Send(b)
	})

	// restart myself
	app.Get("/meta/restart", func(c fiber.Ctx) error {
		log.Warnf("RestartProcess... waiting 3 seconds")
		go func() {
			time.Sleep(3 * time.Second)
			err := utils.RestartProcess()
			if err != nil {
				log.Errorf("RestartProcess error: %v", err)
			}
		}()
		return c.SendString("process restarting... waiting 3 seconds, now is " + time.Now().Format(time.RFC3339))
	})

	// 增加运行时信息
	healthzHandler, err := gosh.NewStatisticsHandler(func(w io.Writer) gosh.JSONEncoder {
		return json.NewEncoder(w)
	})
	if err != nil {
		log.Warnf("new healthz handler error: %v", err)
	} else {
		app.Get("/meta/healthz", adaptor.HTTPHandler(healthzHandler))
	}

	// Or extend your config for customization
	// Assign the middleware to /metrics
	// and change the Title to `MyService Metrics Page`
	// app.Get("/metrics", monitor.New())

	return nil
}

func (p *ApiServer) authMiddleware(c fiber.Ctx) error {
	log.Trace("🥇 Auth handler: " + c.Path())

	// err := jwtware.New(jwtware.Config{
	// 	SigningKey: jwtware.SigningKey{Key: []byte("secret")},
	// })
	// if err != nil {
	// 	log.Error("new jwtware error: %#v", err)
	// }
	// log.Infof("jwt: %#v", err)

	// user := c.Locals("user").(*jwt.Token)
	// claims := user.Claims.(jwt.MapClaims)
	// log.Info(claims)

	return c.Next()
}
