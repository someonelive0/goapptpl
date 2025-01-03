package main

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"strconv"
	"strings"

	clickhouse "github.com/ClickHouse/clickhouse-go/v2"
	"github.com/gofiber/fiber/v3"
	log "github.com/sirupsen/logrus"

	"goapptol/utils"
)

const (
	CLICKHOUSE_MAX_TIMEOUT = 30 // clickhouse max timeout in seconds
)

type ClickhouseHandler struct {
	DbHandler
	opt *clickhouse.Options // clickhouse config of dsn
}

// r := app.Group("/clickhouse")
func (p *ClickhouseHandler) AddRouter(r fiber.Router) error {
	log.Info("ClickhouseHandler AddRouter")

	r.Get("", p.homeHandler)
	r.Get("/", p.homeHandler)
	r.Get("/tables", p.tablesHandler)
	r.Get("/table/:table", p.tableHandler)
	r.Get("/table/:table/columns", p.columnsHandler)
	// r.Get("/table/:table/indexes", p.indexesHandler)
	r.Get("/views", p.viewsHandler)

	//解析DSN字符串
	opt, err := clickhouse.ParseDSN(p.Dbconfig.Dsn[0])
	if err != nil {
		log.Errorf("parse clickhouse dsn '%s' failed: %v", p.Dbconfig.Dsn[0], err)
		return err
	}
	p.opt = opt
	// log.Debugf("clickhouse options: %#v", opt)

	return nil
}

// GET /clickhouse
func (p *ClickhouseHandler) homeHandler(c fiber.Ctx) error {
	c.Context().SetContentType("text/html")
	c.WriteString(`<html><body><h1>Clickhouse Information</h1>
	<a href="/clickhouse/tables?mime=json">tables</a><br>
	<a href="/clickhouse/table/:table?mime=json">table/:table_name/[columns|indexes|constraints|keys|references|triggers|stats|describe|ddl]</a><br>
	<a href="/clickhouse/views?mime=json">views</a><br>
	<a href="/clickhouse/view/:view?mime=json">view/:view_name/[columns|indexes|constraints|keys|references|triggers|stats|describe|ddl]</a><br>
	<a href="/clickhouse/procedures">procedures</a><br>
	<a href="/clickhouse/procedure/:procedure">procedure/:procedure_name</a><br>
	<a href="/clickhouse/events">events</a><br>
	<a href="/clickhouse/event/:event">event/:event_name</a><br>
	<a href="/clickhouse/triggers">triggers</a><br>
	<a href="/clickhouse/trigger/:trigger">trigger/:trigger_name</a><br>
	</body></html>`)
	return nil
}

// GET /clickhouse/tables?mime=excel|json
func (p *ClickhouseHandler) tablesHandler(c fiber.Ctx) error {
	sqltext := fmt.Sprintf(`
		select toJSONString(map(
			'database', assumeNotNull(database)::String,
			'name', assumeNotNull(name)::String,
			'uuid', assumeNotNull(uuid)::String,
			'engine', assumeNotNull(engine)::String,
			'is_temporary', assumeNotNull(is_temporary)::String,
			'data_paths', assumeNotNull(data_paths)::String,
			'metadata_path', assumeNotNull(metadata_path)::String,
			'metadata_modification_time', assumeNotNull(metadata_modification_time)::String,
			'engine_full', assumeNotNull(engine_full)::String,
			'partition_key', assumeNotNull(partition_key)::String,
			'sorting_key', assumeNotNull(sorting_key)::String,
			'primary_key', assumeNotNull(primary_key)::String,
			'storage_policy', assumeNotNull(storage_policy)::String,
			'total_rows', assumeNotNull(total_rows)::String,
			'total_bytes', assumeNotNull(total_bytes)::String,
			'parts', assumeNotNull(parts)::String,
			'active_parts', assumeNotNull(active_parts)::String,
			'total_marks', assumeNotNull(total_marks)::String,
			'comment', assumeNotNull(comment)::String,
			'has_own_data', assumeNotNull(has_own_data)::String
			)) as json
		from system.tables 
		where database = '%s'
		format JSON`, p.opt.Auth.Database)

	mime := c.Query("mime", "json") // if Queries params mime is not set, default to json
	switch mime {
	case "json":
		return p.sqlHandlerByJson(c, sqltext)

	case "excel":
		filename := p.opt.Auth.Database + "-tables.xlsx"
		sheetname := p.opt.Auth.Database + " tables"
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

// GET /clickhouse/table/:table/columns?mime=excel|json
func (p *ClickhouseHandler) columnsHandler(c fiber.Ctx) error {
	table, _ := url.QueryUnescape(c.Params("table"))
	sqltext := fmt.Sprintf(`
	select toJSONString(map(
		'database', assumeNotNull(database)::String,
		'table', assumeNotNull(table)::String,
		'name', assumeNotNull(name)::String,
		'type', assumeNotNull(type)::String,
		'position', assumeNotNull(position)::String,
		'default_kind', assumeNotNull(default_kind)::String,
		'default_expression', assumeNotNull(default_expression)::String,
		'compression_codec', assumeNotNull(compression_codec)::String,
		'numeric_precision', assumeNotNull(numeric_precision)::String,
		'numeric_precision_radix', assumeNotNull(numeric_precision_radix)::String,
		'numeric_scale', assumeNotNull(numeric_scale)::String,
		'datetime_precision', assumeNotNull(datetime_precision)::String,
		'comment', assumeNotNull(comment)::String
		)) as json
	from system.columns
	where database = '%s' and table = '%s'
	order by position`, p.opt.Auth.Database, table)

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

// GET /clickhouse/table/:table?limit=10000&mime=excel|json
func (p *ClickhouseHandler) tableHandler(c fiber.Ctx) error {
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
	columnArray := strings.Split(columns, ",")

	sqltext := `
	select toJSONString(map( `

	for i, col := range columnArray {
		if i > 0 {
			sqltext += ","
		}
		//'database', assumeNotNull(database)::String
		sqltext += "'" + col + "', assumeNotNull(" + col + ")::String"
	}

	sqltext += `	)) as json 
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
			columnArray, sheetname, "log/"+filename); err != nil {
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

// GET /clickhouse/views?mime=excel|json
func (p *ClickhouseHandler) viewsHandler(c fiber.Ctx) error {
	sqltext := fmt.Sprintf(`
		select * from information_schema.views 
		where table_catalog = '%s'
		`, p.opt.Auth.Database)

	return p.sqlHandler2Json(c, sqltext)
}

// get columns of table to string with ',' split
func (p *ClickhouseHandler) getColumns(table string) (string, error) {
	if p.db == nil {
		if err := p.openDB(); err != nil {
			return "", err
		}
	}

	sqltext := fmt.Sprintf(`
	select groupArray(name)::String as columns from (
		select 
			name
		from system.columns
		where database = '%s' and table = '%s'
		order by position
	) b`, p.opt.Auth.Database, table)
	rows, err := p.db.Query(sqltext)
	if err != nil {
		log.Error("Error executing query:", err)
		return "", err
	}
	defer rows.Close()

	var columns string
	for rows.Next() {
		err = rows.Scan(&columns)
		if err != nil {
			log.Error("Error scanning row:", err)
			continue
		}
		break // just get one row
	}

	if err = rows.Err(); err != nil {
		log.Error("Error iterating through rows:", err)
		return "", err
	}

	// colums is ['col1','col2'], should erase '[],'
	columns = strings.ReplaceAll(columns, "'", "")
	columns = strings.Trim(columns, "[]")
	return columns, nil
}
