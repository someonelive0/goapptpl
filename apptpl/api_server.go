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
	app.Route("/").Get(func(c fiber.Ctx) error {
		return fiber.NewError(400, "please access /meta/status")
	})
	app.Route("/meta/status").Get(func(c fiber.Ctx) error {
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

	app.Post("/ticket/v1/analysis", p.ticketHandler)
	app.Post("/datasecurity/analyzerisk/v2/batch", p.aiAnalyzeriskHandler)
	app.Post("/datasecurity/crucialdataonfly/identify/v1/batch", p.aiDataidentifyHandler)

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

func (p *ApiServer) ticketHandler(c fiber.Ctx) error {
	log.Debugf("headers: %v", c.GetReqHeaders())
	log.Debugf("body: %s", c.Body())

	m := make(map[string]string)
	if err := json.Unmarshal(c.Body(), &m); err != nil {
		log.Errorf("json.Unmarshal error: %#v", err)
		return err
	}
	log.Debugf("request appCode: %#v", m["appCode"])
	log.Debugf("request tenant: %#v", m["tenant"])
	log.Debugf("request ticket/data: %#v", m["data"])

	resp := `{
		"retCode": "1000",
		"msg": "",
		"token": "",
		"userInfo": {
			"accountID": "yewu-test",
			"name": "",
			"empNo": "",
			"idCardNum": "",
			"phone": "",
			"mobile": "",
			"email": "",
			"tenant": ""
		}
	}`
	c.Context().SetContentType("application/json")
	c.WriteString(resp)

	return nil
}

func (p *ApiServer) aiAnalyzeriskHandler(c fiber.Ctx) error {

	// test failed response
	// return fiber.NewError(400, `{ "detail": "AI模型分析失败" }`)

	// log.Debugf("headers: %v", c.GetReqHeaders())
	log.Debugf("body: %s", c.Body())

	m := make(map[string]interface{})
	if err := json.Unmarshal(c.Body(), &m); err != nil {
		log.Errorf("json.Unmarshal error: %#v", err)
		return err
	}
	if _, ok := m["inputs"]; !ok {
		return fiber.NewError(400, "inputs is required")
	}
	// log.Debugf("request inputs: %#v", m["inputs"])
	// log.Debugf("request tenant: %#v", m["tenant"])
	// log.Debugf("request ticket/data: %#v", m["data"])

	resp := `{
		"detail": "success",
		"output": [`
	inputs := m["inputs"].([]interface{})
	for i, input := range inputs {
		log.Debugf("request input: %#v", input)
		if i > 0 {
			resp += ", "
		}

		requestId := ""
		if _, ok := input.(map[string]interface{})["requestId"]; ok {
			requestId = input.(map[string]interface{})["requestId"].(string)
		}

		resp += fmt.Sprintf(`{
			"isAgree":  1,
			"reason": "符合风险AI模型",
			"requestId": "%s"
		}`, requestId)
	}

	resp += `
	] }`

	log.Debugf("response: |%s|", resp)
	if err := json.Unmarshal([]byte(resp), &m); err != nil {
		log.Errorf("json.Unmarshal resp error: %#v", err)
		return err
	}

	c.Context().SetContentType("application/json")
	c.WriteString(resp)

	return nil
}

func (p *ApiServer) aiDataidentifyHandler(c fiber.Ctx) error {

	// test failed response
	// return fiber.NewError(400, `{ "detail": "AI模型分析重要数据失败" }`)

	// log.Debugf("headers: %v", c.GetReqHeaders())
	log.Debugf("body: %s", c.Body())

	m := make(map[string]interface{})
	if err := json.Unmarshal(c.Body(), &m); err != nil {
		log.Errorf("json.Unmarshal error: %#v", err)
		return err
	}
	if _, ok := m["inputs"]; !ok {
		return fiber.NewError(400, "inputs is required")
	}
	// log.Debugf("request inputs: %#v", m["inputs"])
	// log.Debugf("request tenant: %#v", m["tenant"])
	// log.Debugf("request ticket/data: %#v", m["data"])

	resp := `{
		"detail": "ok",
		"output": [`
	inputs := m["inputs"].([]interface{})
	for i, input := range inputs {
		log.Debugf("request input: %#v", input)
		if i > 0 {
			resp += ", "
		}
		apiPattern := ""
		if _, ok := input.(map[string]interface{})["apiPattern"]; ok {
			apiPattern = input.(map[string]interface{})["apiPattern"].(string)
		}

		resp += fmt.Sprintf(`{
			"isCrucial":  1,
			"categories": ["身份鉴别信息", "A-2", "重要信息"],
			"apiPattern": "%s"
		}`, apiPattern)
	}

	resp += `
	] }`

	log.Debugf("response: |%s|", resp)
	if err := json.Unmarshal([]byte(resp), &m); err != nil {
		log.Errorf("json.Unmarshal resp error: %#v", err)
		return err
	}

	c.Context().SetContentType("application/json")
	c.WriteString(resp)

	return nil
}
