package main

import (
	"time"

	"github.com/gofiber/fiber/v3"
	log "github.com/sirupsen/logrus"
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

	// // API routes
	// // GET /api/register
	// app.Get("/api/*", func(c fiber.Ctx) error {
	// 	msg := fmt.Sprintf("✋ %s", c.Params("*"))
	// 	return c.SendString(msg) // => ✋ register
	// })

	// AddMinioHandler(app)
	minioHdl := MinioHandler{Minioconfig: &p.Myconfig.MinioConfig}
	minioHdl.AddRouter(app.Group("/minio"))

	mysqlHdl := MysqlHandler{Dbconfig: &p.Myconfig.MysqlConfig}
	mysqlHdl.AddRouter(app.Group("/mysql"))

	// // GET /flights/LAX-SFO
	// app.Get("/flights/:from-:to", func(c fiber.Ctx) error {
	// 	msg := fmt.Sprintf("💸 From: %s, To: %s", c.Params("from"), c.Params("to"))
	// 	return c.SendString(msg) // => 💸 From: LAX, To: SFO
	// })

	// // GET /dictionary.txt
	// app.Get("/:file.:ext", func(c fiber.Ctx) error {
	// 	msg := fmt.Sprintf("📃 %s.%s", c.Params("file"), c.Params("ext"))
	// 	return c.SendString(msg) // => 📃 dictionary.txt
	// })

	// // GET /john/75
	// app.Get("/:name/:age/:gender?", func(c fiber.Ctx) error {
	// 	msg := fmt.Sprintf("👴 %s is %s years old", c.Params("name"), c.Params("age"))
	// 	return c.SendString(msg) // => 👴 john is 75 years old
	// })

	// // GET /john
	// app.Get("/:name", func(c fiber.Ctx) error {
	// 	msg := fmt.Sprintf("Hello, %s 👋!", c.Params("name"))
	// 	return c.SendString(msg) // => Hello john 👋!
	// })

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
