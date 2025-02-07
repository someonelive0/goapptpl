package main

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/gofiber/fiber/v3"
	log "github.com/sirupsen/logrus"

	"goapptol/utils"
)

const (
	MYSQL_MAX_TIMEOUT = 30 // mysql max timeout in seconds
)

type MysqlHandler struct {
	DbHandler
	cfg *mysql.Config // mysql config of dsn
}

// r := app.Group("/mysql")
func (p *MysqlHandler) AddRouter(r fiber.Router) error {
	log.Info("MysqlHandler AddRouter")

	r.Get("", p.homeHandler)
	r.Get("/", p.homeHandler)
	r.Get("/tables", p.tablesHandler)
	r.Get("/table/:table", p.tableHandler)
	r.Get("/table/:table/columns", p.columnsHandler)
	r.Get("/table/:table/indexes", p.indexesHandler)
	r.Get("/table/:table/constraints", p.constraintsHandler) // 表约束
	r.Get("/table/:table/keys", p.keysHandler)               // 表外键
	r.Get("/table/:table/references", p.referencesHandler)   // 表引用
	r.Get("/table/:table/triggers", p.tableTriggersHandler)  // 表触发器
	r.Get("/table/:table/stats", p.statsHandler)             // 表统计
	r.Get("/table/:table/describe", p.describeHandler)       // 表描述
	r.Get("/table/:table/ddl", p.ddlHandler)
	r.Get("/views", p.viewsHandler)
	r.Get("/view/:table", p.tableHandler)
	r.Get("/view/:table/columns", p.columnsHandler)
	r.Get("/view/:table/indexes", p.indexesHandler)
	r.Get("/view/:table/constraints", p.constraintsHandler) // 表约束
	r.Get("/view/:table/keys", p.keysHandler)               // 表外键
	r.Get("/view/:table/references", p.referencesHandler)   // 表引用
	r.Get("/view/:table/triggers", p.tableTriggersHandler)  // 表触发器
	r.Get("/view/:table/stats", p.statsHandler)             // 表统计
	r.Get("/view/:table/describe", p.describeHandler)       // 表描述
	r.Get("/view/:table/ddl", p.ddlHandler)
	r.Get("/procedures", p.proceduresHandler)
	r.Get("/procedure/:procedure", p.procedureHandler)
	r.Get("/events", p.eventsHandler)
	r.Get("/event/:event", p.eventHandler)
	r.Get("/triggers", p.triggersHandler)
	r.Get("/trigger/:trigger", p.triggerHandler)

	//解析DSN字符串
	cfg, err := mysql.ParseDSN(p.Dbconfig.Dsn[0])
	if err != nil {
		log.Errorf("parse mysql dsn '%s' failed: %v", p.Dbconfig.Dsn[0], err)
		return err
	}
	p.cfg = cfg
	// log.Debugf("mysql cfg: %#v", cfg)

	return nil
}

// GET /mysql
func (p *MysqlHandler) homeHandler(c fiber.Ctx) error {
	c.Response().Header.Set("Content-Type", "text/html")
	c.WriteString(`<html><body><h1>Mysql Information</h1>
	<a href="/mysql/tables?mime=json">tables</a><br>
	<a href="/mysql/table/:table?mime=json">table/:table_name/[columns|indexes|constraints|keys|references|triggers|stats|describe|ddl]</a><br>
	<a href="/mysql/views?mime=json">views</a><br>
	<a href="/mysql/view/:view?mime=json">view/:view_name/[columns|indexes|constraints|keys|references|triggers|stats|describe|ddl]</a><br>
	<a href="/mysql/procedures">procedures</a><br>
	<a href="/mysql/procedure/:procedure">procedure/:procedure_name</a><br>
	<a href="/mysql/events">events</a><br>
	<a href="/mysql/event/:event">event/:event_name</a><br>
	<a href="/mysql/triggers">triggers</a><br>
	<a href="/mysql/trigger/:trigger">trigger/:trigger_name</a><br>
	</body></html>`)
	return nil
}

// GET /mysql/views?mime=excel|json
// mysql json_object() 不保证字段顺序，所以excel格式化时，需要按顺序
func (p *MysqlHandler) viewsHandler(c fiber.Ctx) error {
	return p.tablesViewsHandler(c, "VIEW")
}

// GET /mysql/tables?mime=excel|json
// mysql json_object() 不保证字段顺序，所以excel格式化时，需要按顺序
func (p *MysqlHandler) tablesHandler(c fiber.Ctx) error {
	return p.tablesViewsHandler(c, "BASE TABLE")
}

// table_type is "BASE TABLE" or "VIEW"
func (p *MysqlHandler) tablesViewsHandler(c fiber.Ctx, table_type string) error {
	sqltext := `
	select json_object(
		'table_catalog', table_catalog,
		'table_schema', table_schema,
		'table_name', table_name,
		'table_type', table_type,
		'table_rows', table_rows,
		'avg_row_length', avg_row_length,
		'data_length', data_length,
		'max_data_length', max_data_length,
		'index_length', index_length,
		'data_free', data_free,
		'create_time', create_time,
		'table_collation', table_collation,
		'table_comment', table_comment
		) as json
	from INFORMATION_SCHEMA.TABLES
	where table_schema = '` + p.cfg.DBName + `'` +
		`and table_type = '` + table_type + `'`

	mime := c.Query("mime", "json") // if Queries params mime is not set, default to json
	switch mime {
	case "json":
		// return p.sqlHandlerByJson(c, sqltext)
		// use local cache to reduce mysql load
		if b, found := p.Mycache.Get("mysql:tables"); found {
			c.Response().Header.Set("Content-Type", "application/json")
			c.Write(b.([]byte))
			return nil
		}

		ch := make(chan string, 100)
		go func() {
			if err := p.sql2chan(ch, sqltext); err != nil {
				// log.Errorf("sql2chan failed: %v", err)
				c.Status(fiber.StatusInternalServerError).SendString(err.Error()) // 500
				c.Response().ConnectionClose()
				close(ch)
			}
		}()

		c.Response().Header.Set("Content-Type", "application/json")
		c.WriteString("[")
		i := 0
		for jsonstr := range ch {
			if i > 0 {
				c.WriteString(",")
			}
			c.WriteString(jsonstr)
			i++
		}
		c.WriteString("]")

		p.Mycache.Set("mysql:tables", c.Response().Body(), 5*time.Second)
		return nil

	case "excel":
		filename := p.cfg.DBName + "-tables.xlsx"
		sheetname := p.cfg.DBName + " tables"
		ch := make(chan string, 100)
		go p.sql2chan(ch, sqltext)

		if err := utils.Json2excel(ch, sheetname, "log/"+filename); err != nil {
			return err
		}

		c.Attachment(filename)
		fp, err := os.Open("log/" + filename)
		if err != nil {
			return err
		}
		_, err = io.Copy(c, fp)
		fp.Close()
		os.Remove("log/" + filename)
		return err

	case "docx":
		filename := p.cfg.DBName + "-tables.docx"
		title := "数据库 Mysql - " + p.cfg.DBName + " tables"

		ch := make(chan string, 100)
		go p.sql2chan(ch, sqltext)

		if err := utils.Json2docx(ch, title, "log/"+filename); err != nil {
			return err
		}

		c.Attachment(filename)
		fp, err := os.Open("log/" + filename)
		if err != nil {
			return err
		}
		_, err = io.Copy(c, fp)
		fp.Close()
		os.Remove("log/" + filename)
		return err

	default:
		c.Status(400)
		c.SendString(fmt.Sprintf("mime '%s' not supported", mime))
		return nil
	}
}

// GET /mysql/table/:table/columns?mime=excel|json
func (p *MysqlHandler) columnsHandler(c fiber.Ctx) error {
	table, _ := url.QueryUnescape(c.Params("table"))
	sqltext := `
	select json_object(
		'table_catalog', table_catalog,
		'table_schema', table_schema,
		'table_name', table_name,
		'column_name', column_name,
		'ordinal_position', ordinal_position,
		'column_default', column_default,
		'is_nullable', is_nullable,
		'data_type', data_type,
		'column_type', column_type,
		'column_key', column_key,
		'collation_name', collation_name,
		'column_comment', column_comment
		) as json 
	from INFORMATION_SCHEMA.COLUMNS
	where table_schema = '` + p.cfg.DBName + `' and table_name = '` + table + `'`
	sqltext += ` order by ordinal_position`

	mime := c.Query("mime", "json") // if Queries params mime is not set, default to json
	switch mime {
	case "json":
		return p.sqlHandlerByJson(c, sqltext)

	case "excel":
		filename := table + "-columns.xlsx"
		sheetname := table + " columns"
		ch := make(chan string, 100)
		go p.sql2chan(ch, sqltext)

		if err := utils.Json2excel(ch, sheetname, "log/"+filename); err != nil {
			return err
		}

		c.Attachment(filename)
		fp, err := os.Open("log/" + filename)
		if err != nil {
			return err
		}
		_, err = io.Copy(c, fp)
		fp.Close()
		os.Remove("log/" + filename)
		return err

	default:
		c.Status(400)
		c.SendString(fmt.Sprintf("mime '%s' not supported", mime))
		return nil
	}
}

// GET /mysql/table/:table/columns?mime=excel|json
func (p *MysqlHandler) indexesHandler(c fiber.Ctx) error {
	table, _ := url.QueryUnescape(c.Params("table"))
	sqltext := fmt.Sprintf(`
	select json_object(
		'table_catalog', table_catalog,
		'table_schema', table_schema,
		'table_name', table_name,
		'index_name', index_name, 
		'index_type', index_type,
		'index_schema', index_schema,
		'nullable', nullable,
		'is_visible', is_visible,
		'non_unique', non_unique,
		'comment', comment,
		'index_comment', index_comment,
		'collation', collation,
		'comment', comment,
		'cardinality', cardinality
		) as json
	from (
		select
			table_catalog,
			table_schema,
			table_name,
			index_name,
			index_type,
			index_schema,
			nullable,
			is_visible,
			non_unique,
			comment,
			index_comment,
			collation,
			cardinality,
			GROUP_CONCAT(column_name ORDER BY seq_in_index) AS columns
			from information_schema.statistics
		where table_schema = '%s' and table_name = '%s'
		group by table_schema, table_name, index_name
	) as b`, p.cfg.DBName, table)

	mime := c.Query("mime", "json") // if Queries params mime is not set, default to json
	switch mime {
	case "json":
		return p.sqlHandlerByJson(c, sqltext)

	case "excel":
		filename := table + "-indexs.xlsx"
		sheetname := table + " indexs"
		ch := make(chan string, 100)
		go p.sql2chan(ch, sqltext)

		if err := utils.Json2excel(ch, sheetname, "log/"+filename); err != nil {
			return err
		}

		c.Attachment(filename)
		fp, err := os.Open("log/" + filename)
		if err != nil {
			return err
		}
		_, err = io.Copy(c, fp)
		fp.Close()
		os.Remove("log/" + filename)
		return err

	default:
		c.Status(400)
		c.SendString(fmt.Sprintf("mime '%s' not supported", mime))
		return nil
	}
}

// GET /mysql/table/:table?limit=10000&mime=excel|json
func (p *MysqlHandler) tableHandler(c fiber.Ctx) error {
	table, _ := url.QueryUnescape(c.Params("table"))
	// TODO 当数据行数很大时，会占用很大内存，应该改为流式处理，这里限制最多10000行
	limit := c.Query("limit", "100")
	if i, err := strconv.Atoi(limit); err != nil || i > 10000 {
		limit = "100"
	}

	// columns := "id,api_id,app_id,hostname,buz_source,asset_name,api_method,api_endpoint,content_type,module_code,department_id,business_id,description,follow,monitor_cover,fever,asset_state,asset_value,sen_fever,discovery_time,risk_level,carrier_type,validate_time,ext_info,merge_state,check_state,tenant_id,create_user,create_time,update_user,update_time,api_no,pod,resource_pool,asset_code"
	columns, err := p.getColumns(table)
	if err != nil {
		c.WriteString(err.Error())
		return err
	}

	sqltext := `
	select json_object(`

	for i, col := range columns {
		if i > 0 {
			sqltext += ","
		}
		sqltext += "'" + col + "', `" + col + "`"
	}

	sqltext += `	) as json 
	from ` + table + ` limit ` + limit

	mime := c.Query("mime", "json") // if Queries params mime is not set, default to json
	switch mime {
	case "json":
		// TODO 当表数据行数很大时，会占用很大内存，应该改为流式处理
		return p.sqlHandlerByJson(c, sqltext)

	case "excel":
		filename := table + ".xlsx"
		sheetname := table
		ch := make(chan string, 100)
		go p.sql2chan(ch, sqltext)

		if err = utils.Json2excelWithColumn(ch,
			columns, sheetname, "log/"+filename); err != nil {
			return err
		}

		c.Attachment(filename)
		fp, err := os.Open("log/" + filename)
		if err != nil {
			return err
		}
		_, err = io.Copy(c, fp)
		fp.Close()
		os.Remove("log/" + filename)
		return err

	default:
		c.Status(400)
		c.SendString(fmt.Sprintf("mime '%s' not supported", mime))
		return nil
	}
}

// GET /mysql/table/:table/constraints 表约束
func (p *MysqlHandler) constraintsHandler(c fiber.Ctx) error {
	table, _ := url.QueryUnescape(c.Params("table"))
	sqltext := fmt.Sprintf(`
	SELECT CONSTRAINT_NAME, CONSTRAINT_TYPE, TABLE_NAME
	FROM information_schema.TABLE_CONSTRAINTS
	WHERE TABLE_SCHEMA = '%s' AND TABLE_NAME = '%s'`,
		p.cfg.DBName, table)

	return p.sqlHandler2Json(c, sqltext)
}

// GET /mysql/table/:table/keys 表外键
func (p *MysqlHandler) keysHandler(c fiber.Ctx) error {
	table, _ := url.QueryUnescape(c.Params("table"))
	sqltext := fmt.Sprintf(`
	select * from information_schema.key_column_usage
	where REFERENCED_TABLE_NAME != null
	and TABLE_SCHEMA = '%s' and TABLE_NAME = '%s'`,
		p.cfg.DBName, table)

	return p.sqlHandler2Json(c, sqltext)
}

// GET /mysql/table/:table/references 表引用
func (p *MysqlHandler) referencesHandler(c fiber.Ctx) error {
	table, _ := url.QueryUnescape(c.Params("table"))
	sqltext := fmt.Sprintf(`
	select * from information_schema.key_column_usage
	where REFERENCED_TABLE_SCHEMA = '%s' and REFERENCED_TABLE_NAME = '%s'`,
		p.cfg.DBName, table)

	return p.sqlHandler2Json(c, sqltext)
}

// GET /mysql/table/:table/triggers 表触发器
func (p *MysqlHandler) tableTriggersHandler(c fiber.Ctx) error {
	table, _ := url.QueryUnescape(c.Params("table"))
	sqltext := fmt.Sprintf(`
	select * from information_schema.triggers 
	where EVENT_OBJECT_SCHEMA = '%s' and EVENT_OBJECT_TABLE = '%s'`,
		p.cfg.DBName, table)

	return p.sqlHandler2Json(c, sqltext)
}

// GET /mysql/table/:table/stats 表统计
func (p *MysqlHandler) statsHandler(c fiber.Ctx) error {
	table, _ := url.QueryUnescape(c.Params("table"))
	sqltext := fmt.Sprintf(`
	select * from information_schema.tables
	where TABLE_SCHEMA = '%s' and TABLE_NAME = '%s'`,
		p.cfg.DBName, table)

	return p.sqlHandler2Json(c, sqltext)
}

// GET /mysql/table/:table/describe 表描述
func (p *MysqlHandler) describeHandler(c fiber.Ctx) error {
	table, _ := url.QueryUnescape(c.Params("table"))
	sqltext := fmt.Sprintf(`
	describe %s`, table)

	return p.sqlHandler2Json(c, sqltext)
}

// GET /mysql/table/:table/ddl
func (p *MysqlHandler) ddlHandler(c fiber.Ctx) error {
	table, _ := url.QueryUnescape(c.Params("table"))
	sqltext := fmt.Sprintf(`
	show create table %s`, table)

	return p.sqlHandler2Json(c, sqltext)
}

// GET /mysql/procedures
func (p *MysqlHandler) proceduresHandler(c fiber.Ctx) error {
	sqltext := fmt.Sprintf(`SHOW PROCEDURE STATUS where db = '%s'`, p.cfg.DBName)

	return p.sqlHandler2Json(c, sqltext)
}

// GET /mysql/procedure/:procedure
func (p *MysqlHandler) procedureHandler(c fiber.Ctx) error {
	procedure, _ := url.QueryUnescape(c.Params("procedure"))
	sqltext := fmt.Sprintf(`SHOW CREATE PROCEDURE %s`, procedure)

	return p.sqlHandler2Json(c, sqltext)
}

// GET /mysql/events
func (p *MysqlHandler) eventsHandler(c fiber.Ctx) error {
	sqltext := fmt.Sprintf(`SHOW EVENTS from %s`, p.cfg.DBName)

	return p.sqlHandler2Json(c, sqltext)
}

// GET /mysql/event/:event
func (p *MysqlHandler) eventHandler(c fiber.Ctx) error {
	event, _ := url.QueryUnescape(c.Params("event"))
	sqltext := fmt.Sprintf(`SHOW CREATE EVENT %s`, event)

	return p.sqlHandler2Json(c, sqltext)
}

// GET /mysql/triggers
func (p *MysqlHandler) triggersHandler(c fiber.Ctx) error {
	sqltext := fmt.Sprintf(`SHOW triggers from %s`, p.cfg.DBName)

	return p.sqlHandler2Json(c, sqltext)
}

// GET /mysql/trigger/:trigger
func (p *MysqlHandler) triggerHandler(c fiber.Ctx) error {
	trigger, _ := url.QueryUnescape(c.Params("trigger"))
	sqltext := fmt.Sprintf(`SHOW CREATE trigger %s`, trigger)

	return p.sqlHandler2Json(c, sqltext)
}

// get columns of table to string with ',' split. sort by ordinal_position
func (p *MysqlHandler) getColumns(table string) ([]string, error) {
	if p.db == nil {
		if err := p.openDB(); err != nil {
			return nil, err
		}
	}

	sqltext := fmt.Sprintf(`select column_name
		from INFORMATION_SCHEMA.COLUMNS
		where table_schema = '%s' and table_name = '%s'
		order by ordinal_position`, p.cfg.DBName, table)
	rows, err := p.db.Query(sqltext)
	if err != nil {
		log.Error("Error executing query:", err)
		return nil, err
	}
	defer rows.Close()

	var columns []string = make([]string, 0)
	var column string
	for rows.Next() {
		err = rows.Scan(&column)
		if err != nil {
			log.Error("Error scanning row:", err)
			continue
		}
		columns = append(columns, column)
	}

	if err = rows.Err(); err != nil {
		log.Error("Error iterating through rows:", err)
		return nil, err
	}

	return columns, nil
}
