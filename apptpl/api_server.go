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
	mysqlHdl := MysqlHandler{Dbconfig: &p.Myconfig.MysqlConfig, Mycache: p.mycache}
	mysqlHdl.AddRouter(app.Group("/mysql"))

	// add MinioHandler
	minioHdl := MinioHandler{Minioconfig: &p.Myconfig.MinioConfig}
	minioHdl.AddRouter(app.Group("/minio"))

	// add RedisHandler
	redisHdl := RedisHandler{Redisconfig: &p.Myconfig.RedisConfig}
	redisHdl.AddRouter(app.Group("/redis"))

	// add ClickhouseHandler
	ckHdl := ClickhouseHandler{Dbconfig: &p.Myconfig.CkConfig}
	ckHdl.AddRouter(app.Group("/clickhouse"))

	// add PgHandler
	pgHdl := PgHandler{Dbconfig: &p.Myconfig.PgConfig}
	pgHdl.AddRouter(app.Group("/postgresql"))

	// add HardwareHandler
	hardwareHdl := HardwareHandler{Mycache: p.mycache}
	hardwareHdl.AddRouter(app.Group("/hardware"))

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
	app.Post("/sit/apiData/sitBussinessSystem/apiData", p.buzHandler)

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

func (p *ApiServer) buzHandler(c fiber.Ctx) error {
	log.Debugf("headers: %v", c.GetReqHeaders())
	log.Debugf("body: %s", c.Body())

	m := make(map[string]int)
	if err := json.Unmarshal(c.Body(), &m); err != nil {
		log.Errorf("json.Unmarshal error: %#v", err)
		return err
	}
	log.Debugf("request pageNo: %d", m["pageNo"])
	log.Debugf("request pageSize: %d", m["pageSize"])

	var pageNo int = 1
	var pageSize int = 100
	if i, ok := m["pageNo"]; ok {
		if i > 0 {
			pageNo = i
		}
	}
	if i, ok := m["pageSize"]; ok {
		if i > 0 {
			pageSize = i
		}
	}

	buzs := make([]map[string]interface{}, 0)
	json.Unmarshal([]byte(zhihui_buz), &buzs)
	log.Debugf("zhihui_buz number: %d", len(buzs))
	// for i, buz := range buzs {
	// 	log.Debugf("zhihui_buz %d : %v", i, buz)
	// }

	offset0 := (pageNo - 1) * pageSize
	offset1 := pageNo * pageSize
	if offset0 > len(buzs) {
		return fiber.NewError(400, "pageNo is too large")
	}
	if offset1 > len(buzs) {
		offset1 = len(buzs)
	}
	log.Debugf("pageNo %d, pageSize %d", pageNo, pageSize)
	log.Debugf("offset from %d to %d", offset0, offset1)

	c.Context().SetContentType("application/json")

	b, _ := json.Marshal(buzs[offset0:offset1])
	resp := fmt.Sprintf(`{
		"message":"操作成功",
		"result":{
			"data": %s,
			"pageNo": %d,
			"pageSize": %d,
			"totalCount":%d,
			"totalPage":%d
		},
		"status":200
	}`, b,
		pageNo, pageSize, len(buzs), len(buzs)/pageSize+1)
	c.WriteString(resp)

	return nil
}

/*
	From Mysql

select json_arrayagg(json_object(

	'buz_id', id,
	'id', custom_sys_code,
	'name', business_name,
	'orgId', dept_id,
	'twoOrgName', (select depart_name from sys_department sd where sd.id = dept_id),
	'threeOrgName', (select depart_name from sys_department sd where sd.id = dept_id),
	'respUserName', record_manager_name,
	'rankReportId', record_code,
	'rankReportName', record_name,
	'rank', record_level,
	'sys_admin', sys_admin,
	'system_label', '一般系统'
	)) as json

from sys_buz_system
order by id
*/
const zhihui_buz = `
[
    {
        "id": "zhihui_1",
        "name": "默认系统",
        "rank": "1",
        "orgId": 60000,
        "buz_id": 1,
        "sys_admin": null,
        "twoOrgName": "默认部门",
        "rankReportId": "1",
        "respUserName": null,
        "system_label": "一般系统",
        "threeOrgName": "默认部门",
        "rankReportName": "定级备案名称1"
    },
    {
        "id": "zhihui_sys_2",
        "name": "数管815(已删除)qqq",
        "rank": "-1",
        "orgId": 60002,
        "buz_id": 2,
        "sys_admin": "68BD7B0AF88CC1DFD719D0A9F4E627EA",
        "twoOrgName": "定开支持部",
        "rankReportId": "2",
        "respUserName": "2260DD6528E73D0EFDEEBDC2F20D2E14F67AC2BD695E52F706EE8C5C86588FAE",
        "system_label": "一般系统",
        "threeOrgName": "定开支持部",
        "rankReportName": "定级备案名称2"
    },
    {
        "id": "zhihui_3",
        "name": "数据成都老平台(已删除)",
        "rank": "1",
        "orgId": 60012,
        "buz_id": 3,
        "sys_admin": "",
        "twoOrgName": "测试部门",
        "rankReportId": "3",
        "respUserName": null,
        "system_label": "一般系统",
        "threeOrgName": "测试部门",
        "rankReportName": "数管平台开发测试系统"
    },
    {
        "id": "zhihui_4",
        "name": "视频准入系统",
        "rank": "1",
        "orgId": 60003,
        "buz_id": 4,
        "sys_admin": "E19A7D178827B56324AE6D0FD448D73E",
        "twoOrgName": "研发中心",
        "rankReportId": "4",
        "respUserName": null,
        "system_label": "一般系统",
        "threeOrgName": "研发中心",
        "rankReportName": "定级备案名称3"
    },
    {
        "id": "zhihui_5",
        "name": "审计系统",
        "rank": "2",
        "orgId": 60012,
        "buz_id": 5,
        "sys_admin": null,
        "twoOrgName": "测试部门",
        "rankReportId": "5",
        "respUserName": null,
        "system_label": "一般系统",
        "threeOrgName": "测试部门",
        "rankReportName": "定级备案名称5"
    },
    {
        "id": "zhihui_6",
        "name": "磐维数据库系统(已删除)",
        "rank": "3",
        "orgId": 60002,
        "buz_id": 6,
        "sys_admin": null,
        "twoOrgName": "定开支持部",
        "rankReportId": "6",
        "respUserName": null,
        "system_label": "一般系统",
        "threeOrgName": "定开支持部",
        "rankReportName": "定级备案名称5"
    },
    {
        "id": "zhihui_10",
        "name": "testdd(已删除)",
        "rank": "-1",
        "orgId": 60003,
        "buz_id": 10,
        "sys_admin": null,
        "twoOrgName": "研发中心",
        "rankReportId": "10",
        "respUserName": null,
        "system_label": "一般系统",
        "threeOrgName": "研发中心",
        "rankReportName": "1111"
    },
    {
        "id": "zhihui_12",
        "name": "准入系统(已删除)",
        "rank": "4",
        "orgId": 60003,
        "buz_id": 12,
        "sys_admin": "70C614759B62EC0E88CF6BAB67322CF3",
        "twoOrgName": "研发中心",
        "rankReportId": "12",
        "respUserName": null,
        "system_label": "一般系统",
        "threeOrgName": "研发中心",
        "rankReportName": "视频准入系统V1.0"
    },
    {
        "id": "zhihui_13",
        "name": "二级013333",
        "rank": "4",
        "orgId": 60005,
        "buz_id": 13,
        "sys_admin": "1280B0991810516F1556A0E71A16EB0E",
        "twoOrgName": "一级01",
        "rankReportId": "13",
        "respUserName": null,
        "system_label": "一般系统",
        "threeOrgName": "一级01",
        "rankReportName": "定级备案名称6"
    },
    {
        "id": "zhihui_14",
        "name": "三级01",
        "rank": "0",
        "orgId": 60006,
        "buz_id": 14,
        "sys_admin": null,
        "twoOrgName": "二级01-02",
        "rankReportId": "14",
        "respUserName": null,
        "system_label": "一般系统",
        "threeOrgName": "二级01-02",
        "rankReportName": "定级备案名称7"
    },
    {
        "id": "zhihui_15",
        "name": "二级02",
        "rank": "1",
        "orgId": 60007,
        "buz_id": 15,
        "sys_admin": "E19A7D178827B56324AE6D0FD448D73E",
        "twoOrgName": "一级02",
        "rankReportId": "15",
        "respUserName": null,
        "system_label": "一般系统",
        "threeOrgName": "一级02",
        "rankReportName": "定级备案名称8"
    },
    {
        "id": "zhihui_16",
        "name": "数管",
        "rank": "2",
        "orgId": 60008,
        "buz_id": 16,
        "sys_admin": null,
        "twoOrgName": "开发部",
        "rankReportId": "16",
        "respUserName": null,
        "system_label": "一般系统",
        "threeOrgName": "开发部",
        "rankReportName": "定级备案名称9"
    },
    {
        "id": "zhihui_17",
        "name": "人员管理系统",
        "rank": "-1",
        "orgId": 60009,
        "buz_id": 17,
        "sys_admin": "",
        "twoOrgName": "设计部",
        "rankReportId": "17",
        "respUserName": null,
        "system_label": "一般系统",
        "threeOrgName": "设计部",
        "rankReportName": "定级备案名称10"
    },
    {
        "id": "zhihui_19",
        "name": "03",
        "rank": "-1",
        "orgId": 60010,
        "buz_id": 19,
        "sys_admin": "",
        "twoOrgName": "一级03",
        "rankReportId": "19",
        "respUserName": null,
        "system_label": "一般系统",
        "threeOrgName": "一级03",
        "rankReportName": "定级备案名称11"
    },
    {
        "id": "zhihui_20",
        "name": "测试业务系统",
        "rank": "-1",
        "orgId": 60012,
        "buz_id": 20,
        "sys_admin": "",
        "twoOrgName": "测试部门",
        "rankReportId": "20",
        "respUserName": null,
        "system_label": "一般系统",
        "threeOrgName": "测试部门",
        "rankReportName": "中国移动辽宁公司CRM系统"
    },
    {
        "id": "zhihui_21",
        "name": "测试业务系统0002",
        "rank": "3",
        "orgId": 60012,
        "buz_id": 21,
        "sys_admin": "",
        "twoOrgName": "测试部门",
        "rankReportId": "21",
        "respUserName": null,
        "system_label": "一般系统",
        "threeOrgName": "测试部门",
        "rankReportName": "定级备案名称13"
    },
    {
        "id": "zhihui_23",
        "name": "模压飞洒房",
        "rank": "2",
        "orgId": 60010,
        "buz_id": 23,
        "sys_admin": "",
        "twoOrgName": "一级03",
        "rankReportId": "23",
        "respUserName": null,
        "system_label": "一般系统",
        "threeOrgName": "一级03",
        "rankReportName": "发撒方法"
    },
    {
        "id": "zhihui_24",
        "name": "1->1",
        "rank": "0",
        "orgId": 60000,
        "buz_id": 24,
        "sys_admin": "",
        "twoOrgName": "默认部门",
        "rankReportId": "24",
        "respUserName": null,
        "system_label": "一般系统",
        "threeOrgName": "默认部门",
        "rankReportName": "111111"
    },
    {
        "id": "zhihui_25",
        "name": "全国水土保持信息管理系统",
        "rank": "4",
        "orgId": 60015,
        "buz_id": 25,
        "sys_admin": "",
        "twoOrgName": "金水",
        "rankReportId": "25",
        "respUserName": null,
        "system_label": "一般系统",
        "threeOrgName": "金水",
        "rankReportName": "定级备案名称14"
    },
    {
        "id": "zhihui_26",
        "name": "2-1",
        "rank": "0",
        "orgId": 60000,
        "buz_id": 26,
        "sys_admin": "",
        "twoOrgName": "默认部门",
        "rankReportId": "26",
        "respUserName": null,
        "system_label": "一般系统",
        "threeOrgName": "默认部门",
        "rankReportName": "测试"
    },
    {
        "id": "zhihui_27",
        "name": "成都测试系统",
        "rank": "-1",
        "orgId": 60003,
        "buz_id": 27,
        "sys_admin": "70C614759B62EC0E88CF6BAB67322CF3",
        "twoOrgName": "研发中心",
        "rankReportId": "27",
        "respUserName": null,
        "system_label": "一般系统",
        "threeOrgName": "研发中心",
        "rankReportName": "中国电信上海公司支撑网新CRM系统"
    },
    {
        "id": "zhihui_28",
        "name": "1",
        "rank": "0",
        "orgId": 60013,
        "buz_id": 28,
        "sys_admin": "",
        "twoOrgName": "测试",
        "rankReportId": "28",
        "respUserName": null,
        "system_label": "一般系统",
        "threeOrgName": "测试",
        "rankReportName": "1"
    },
    {
        "id": "zhihui_32",
        "name": "yyyy",
        "rank": "",
        "orgId": 60002,
        "buz_id": 32,
        "sys_admin": "",
        "twoOrgName": "定开支持部",
        "rankReportId": "32",
        "respUserName": null,
        "system_label": "一般系统",
        "threeOrgName": "定开支持部",
        "rankReportName": "yyyy"
    },
    {
        "id": "zhihui_33",
        "name": "测试业务提供1",
        "rank": "",
        "orgId": 60000,
        "buz_id": 33,
        "sys_admin": "",
        "twoOrgName": "默认部门",
        "rankReportId": "33",
        "respUserName": null,
        "system_label": "一般系统",
        "threeOrgName": "默认部门",
        "rankReportName": "测试业务提供1"
    },
    {
        "id": "zhihui_34",
        "name": "测试002",
        "rank": "",
        "orgId": 60000,
        "buz_id": 34,
        "sys_admin": "",
        "twoOrgName": "默认部门",
        "rankReportId": "34",
        "respUserName": null,
        "system_label": "一般系统",
        "threeOrgName": "默认部门",
        "rankReportName": "测试002"
    },
    {
        "id": "zhihui_35",
        "name": "test11",
        "rank": "-1",
        "orgId": 60005,
        "buz_id": 35,
        "sys_admin": "B2DC9D6BEE35511022E4E11B025D31B9",
        "twoOrgName": "一级01",
        "rankReportId": "35",
        "respUserName": null,
        "system_label": "一般系统",
        "threeOrgName": "一级01",
        "rankReportName": "测试定级"
    },
    {
        "id": "zhihui_36",
        "name": "test",
        "rank": "0",
        "orgId": 60000,
        "buz_id": 36,
        "sys_admin": "70C614759B62EC0E88CF6BAB67322CF3",
        "twoOrgName": "默认部门",
        "rankReportId": "36",
        "respUserName": "2260DD6528E73D0EFDEEBDC2F20D2E14F530F91AC2F22D28B65544C1907D5930",
        "system_label": "一般系统",
        "threeOrgName": "默认部门",
        "rankReportName": "定级备案名称2"
    },
    {
        "id": "zhihui_sys_371",
        "name": "三七业务",
        "rank": "0",
        "orgId": 60003,
        "buz_id": 37,
        "sys_admin": "4406A9F81D1F60B668F51048CBF068A1",
        "twoOrgName": "研发中心",
        "rankReportId": "37",
        "respUserName": "2260DD6528E73D0EFDEEBDC2F20D2E14256ABAD90445BFD980960E94AE31F663",
        "system_label": "一般系统",
        "threeOrgName": "研发中心",
        "rankReportName": "定级备案名称2"
    },
    {
        "id": "zhihui_123",
        "name": "123",
        "rank": "3",
        "orgId": 60010,
        "buz_id": 38,
        "sys_admin": "235B0474B8CCBDA798A0327D2E49F7FD",
        "twoOrgName": "一级03",
        "rankReportId": "38",
        "respUserName": "3D70B22C02DDDDF19A35556873395F0D",
        "system_label": "一般系统",
        "threeOrgName": "一级03",
        "rankReportName": "测试业务系统"
    },
    {
        "id": "zhihui_39",
        "name": "测试业务系统",
        "rank": null,
        "orgId": 60002,
        "buz_id": 39,
        "sys_admin": null,
        "twoOrgName": "定开支持部",
        "rankReportId": "39",
        "respUserName": null,
        "system_label": "一般系统",
        "threeOrgName": "定开支持部",
        "rankReportName": "测试业务系统"
    },
    {
        "id": "zhihui_40",
        "name": "移动维护运营室_无归属资产",
        "rank": "-1",
        "orgId": 60002,
        "buz_id": 40,
        "sys_admin": "235B0474B8CCBDA798A0327D2E49F7FD",
        "twoOrgName": "定开支持部",
        "rankReportId": "40",
        "respUserName": null,
        "system_label": "一般系统",
        "threeOrgName": "定开支持部",
        "rankReportName": "移动维护运营室_无归属资产"
    },
    {
        "id": "zhihui_41",
        "name": "云和平台运营室_无归属资产",
        "rank": null,
        "orgId": 60002,
        "buz_id": 41,
        "sys_admin": null,
        "twoOrgName": "定开支持部",
        "rankReportId": "41",
        "respUserName": null,
        "system_label": "一般系统",
        "threeOrgName": "定开支持部",
        "rankReportName": "云和平台运营室_无归属资产"
    },
    {
        "id": "zhihui_42",
        "name": "CRM运维管理平台",
        "rank": "-1",
        "orgId": 60000,
        "buz_id": 42,
        "sys_admin": "E19A7D178827B56324AE6D0FD448D73E",
        "twoOrgName": "默认部门",
        "rankReportId": "42",
        "respUserName": null,
        "system_label": "一般系统",
        "threeOrgName": "默认部门",
        "rankReportName": "CRM运维管理平台"
    }
]
`
