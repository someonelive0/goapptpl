package main

import (
	"encoding/json"
	"time"

	"github.com/gofiber/fiber/v3"
	log "github.com/sirupsen/logrus"

	"goapptol/utils"
)

type ApiServer struct {
	Myconfig *MyConfig
	app      *fiber.App
	mysqlHdl *MysqlHandler
	minioHdl *MinioHandler
}

func (p *ApiServer) Start() error {
	app := fiber.New(fiber.Config{
		CaseSensitive: true,
		StrictRouting: true,
		Immutable:     true,
		ServerHeader:  "goapptpl",
		AppName:       "Test App v1.0.1",
		ReadTimeout:   30 * time.Second,
		WriteTimeout:  30 * time.Second,
		ProxyHeader:   fiber.HeaderXForwardedFor,
		UnescapePath:  false, // default false
	})

	log.Info("🚀 API server starting...")

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

	// API routes
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
	app.Get("/version", func(c fiber.Ctx) error {
		return c.SendString(utils.Version("goapptpl")) // => ✋ versoin
	})
	app.Get("/config", func(c fiber.Ctx) error {
		b, _ := json.Marshal(p.Myconfig)
		return c.Send(b)
	})

	// add MinioHandler
	minioHdl := MinioHandler{Minioconfig: &p.Myconfig.MinioConfig}
	minioHdl.AddRouter(app.Group("/minio"))

	// add MysqlHandler
	mysqlHdl := MysqlHandler{Dbconfig: &p.Myconfig.MysqlConfig}
	mysqlHdl.AddRouter(app.Group("/mysql"))

	// Or extend your config for customization
	// Assign the middleware to /metrics
	// and change the Title to `MyService Metrics Page`
	// app.Get("/metrics", monitor.New())

	// data, _ := json.MarshalIndent(app.Stack(), "", "  ")
	// log.Debug(string(data))
	// data, _ = json.MarshalIndent(app.Config(), "", "  ")
	// log.Debug("config: %s\n", data)

	p.app = app
	p.mysqlHdl = &mysqlHdl
	p.minioHdl = &minioHdl

	// 正常时阻塞在这里
	err := app.Listen("[::]:3000", fiber.ListenConfig{
		CertFile:    "etc/cert.pem",
		CertKeyFile: "etc/key.pem",
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
		err := p.app.ShutdownWithTimeout(2 * time.Second)
		// err := p.app.Shutdown()
		p.app = nil
		p.mysqlHdl.Close()
		p.mysqlHdl = nil
		return err
	}
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
