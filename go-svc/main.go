package main

import (
	"flag"
	"fmt"
	"goapptol/utils"
	"os"
	"time"

	"github.com/gofiber/fiber/v3"

	log "github.com/sirupsen/logrus"
)

var (
	param_debug   = flag.Bool("D", false, "debug")
	param_version = flag.Bool("v", false, "version")
	param_config  = flag.String("f", "etc/goapptpl.toml", "config filename")
	START_TIME    = time.Now()
	myconfig      *MyConfig
)

func init() {
	// parse command line
	flag.Parse()
	if *param_version {
		fmt.Print(utils.APP_BANNER)
		fmt.Printf("%s\n", "1.0.0")
		os.Exit(0)
	}
	utils.ShowBannerForApp("goapptpl", utils.APP_VERSION, utils.BUILD_TIME)
	utils.Chdir2PrgPath()
	pwd, _ := utils.GetPrgDir()
	fmt.Println("pwd:", pwd)

	// load config
	config, err := LoadConfig(*param_config)
	if err != nil {
		fmt.Printf("loadConfig error %s\n", err)
		os.Exit(1)
	}
	myconfig = config

	// init log
	err = utils.InitLogRotate(myconfig.LogConfig.Path, myconfig.LogConfig.Filename,
		myconfig.LogConfig.Level,
		myconfig.LogConfig.Rotate_files,
		myconfig.LogConfig.Rotate_mbytes)
	if err != nil {
		fmt.Printf("InitLogRotate error %s\n", err)
		os.Exit(1)
	}

	log.Infof("BEGIN... %v, config=%v, debug=%v",
		START_TIME.Format("2006-01-02 15:04:05"), *param_config, *param_debug)
	log.Debugf("MyConfig: %s", myconfig.Dump())
}

func main() {

	app := fiber.New(fiber.Config{
		CaseSensitive: true,
		StrictRouting: true,
		Immutable:     true,
		ServerHeader:  "gofiber3",
		AppName:       "Test App v1.0.1",
		ReadTimeout:   30 * time.Second,
		WriteTimeout:  30 * time.Second,
		ProxyHeader:   fiber.HeaderXForwardedFor,
		UnescapePath:  false, // default false
	})

	log.Info("🚀 API server started")

	// Uer Middleware
	// Match any route
	app.Use(func(c fiber.Ctx) error {
		log.Trace("🥇 Any handler, 匹配任何路由: " + c.Path())
		return c.Next()
	})

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

	AddMinioHandler(app)
	// AddMinioHandler1(app.Group("/minio"))

	mysqlHdl := MysqlHandler{Dbconfig: myconfig.MysqlConfig}
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

	log.Fatal(app.Listen("[::]:3000", fiber.ListenConfig{
		CertFile:    "etc/cert.pem",
		CertKeyFile: "etc/key.pem",
	}))
}
