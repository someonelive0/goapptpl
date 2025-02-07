package main

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"strconv"

	"github.com/gofiber/fiber/v3"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"

	"goapptol/utils"
)

const (
	PG_MAX_TIMEOUT = 30 // pg max timeout in seconds
)

type PgHandler struct {
	DbHandler
	u *url.URL // pg url of dsn
}

// r := app.Group("/postgresql")
func (p *PgHandler) AddRouter(r fiber.Router) error {
	log.Info("PgHandler AddRouter")

	r.Get("", p.homeHandler)
	r.Get("/", p.homeHandler)
	r.Get("/tables", p.tablesHandler)
	r.Get("/table/:table", p.tableHandler)
	r.Get("/table/:table/columns", p.columnsHandler)
	r.Get("/table/:table/indexes", p.indexesHandler)
	r.Get("/views", p.viewsHandler)
	r.Get("/view/:table", p.viewHandler)
	r.Get("/procedures", p.proceduresHandler)
	r.Get("/procedure/:procedure", p.procedureHandler)

	//解析DSN字符串
	u, err := url.Parse(p.Dbconfig.Dsn[0])
	if err != nil {
		log.Errorf("parse postgresql dsn '%s' failed: %v", p.Dbconfig.Dsn[0], err)
		return err
	}
	p.u = u
	log.Debugf("postgresql url: %#v", u)

	return nil
}

// GET /postgresql
func (p *PgHandler) homeHandler(c fiber.Ctx) error {
	c.Response().Header.Set("Content-Type", "text/html")
	c.WriteString(`<html><body><h1>Postgresql Information</h1>
	<a href="/postgresql/tables?mime=json">tables</a><br>
	<a href="/postgresql/table/:table?mime=json">table/:table_name/[columns|indexes|constraints|keys|references|triggers|stats|describe|ddl]</a><br>
	<a href="/postgresql/views?mime=json">views</a><br>
	<a href="/postgresql/view/:view?mime=json">view/:view_name/[columns|indexes|constraints|keys|references|triggers|stats|describe|ddl]</a><br>
	<a href="/postgresql/procedures">procedures</a><br>
	<a href="/postgresql/procedure/:procedure">procedure/:procedure_name</a><br>
	<a href="/postgresql/events">events</a><br>
	<a href="/postgresql/event/:event">event/:event_name</a><br>
	<a href="/postgresql/triggers">triggers</a><br>
	<a href="/postgresql/trigger/:trigger">trigger/:trigger_name</a><br>
	</body></html>`)
	return nil
}

// GET /postgresql/tables?mime=excel|json
func (p *PgHandler) tablesHandler(c fiber.Ctx) error {
	sqltext := `select 
	json_build_object(
		'schemaname', tab.schemaname,
		'tablename', tab.tablename,
		'oid', cla."oid",
		'tableowner', tab.tableowner,
		'tablespace', tab.tablespace,
		'hasindexes', tab.hasindexes,
		'hasrules', tab.hasrules,
		'hastriggers', tab.hastriggers,
		'rowsecurity', tab.rowsecurity,
		'rows', stat.n_live_tup,
		'description', des.description
	) as json
from pg_tables tab
	left join pg_class cla on tab.tablename = cla.relname
	left join pg_description des on	des.objoid = cla.oid and objsubid = 0  --为0就是表的描述，其他是字段的描述
	left join pg_stat_user_tables stat on tab.tablename = stat.relname 
order by tab.schemaname, tab.tablename`

	mime := c.Query("mime", "json") // if Queries params mime is not set, default to json
	switch mime {
	case "json":
		return p.sqlHandlerByJson(c, sqltext)

	case "excel":
		filename := p.u.Path + "-tables.xlsx"
		sheetname := p.u.Path + " tables"
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

// GET /postgresql/table/:table/columns?mime=excel|json
func (p *PgHandler) columnsHandler(c fiber.Ctx) error {
	table, _ := url.QueryUnescape(c.Params("table"))
	sqltext := fmt.Sprintf(`select json_build_object(
		'ordinal_position', col.ordinal_position,
		'column_name', col.column_name,
		'table_schema', col.table_schema,
		'table_name', col.table_name,
		'data_type', col.data_type,
		'character_maximum_length', col.character_maximum_length,
		'numeric_precision', col.numeric_precision,
		'numeric_scale', col.numeric_scale,
		'is_nullable', col.is_nullable,
		'column_default', col.column_default,
		'description', des.description) as json
	from information_schema.columns col left join pg_description des
		on col.table_name::regclass = des.objoid
		and col.ordinal_position = des.objsubid
	where table_name = '%s'
	order by ordinal_position `, table)

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

// GET /postgresql/table/:table/columns?mime=excel|json
func (p *PgHandler) indexesHandler(c fiber.Ctx) error {
	table, _ := url.QueryUnescape(c.Params("table"))
	sqltext := fmt.Sprintf(`select json_build_object(
    'indexname', a.indexname,
    'schemaname', a.schemaname,
    'tablename', a.tablename,
    'tablespace', a.tablespace,
    'indexdef', a.indexdef,
    'amname', b.amname,
    'indexrelid', c.indexrelid,
    'indnatts', c.indnatts,
    'indisunique', c.indisunique,
    'indisprimary', c.indisprimary,
    'indisclustered', c.indisclustered,
    'description', d.description) as json
from
	pg_am b left join pg_class f on
	b.oid = f.relam left join pg_stat_all_indexes e on
	f.oid = e.indexrelid left join pg_index c on
	e.indexrelid = c.indexrelid left outer join pg_description d on
	c.indexrelid = d.objoid,
	pg_indexes a
where
	a.schemaname = e.schemaname
	and a.tablename = e.relname
	and a.indexname = e.indexrelname
	and e.relname = '%s'`, table)

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

// GET /postgresql/table/:table?limit=10000&mime=excel|json
func (p *PgHandler) tableHandler(c fiber.Ctx) error {
	table, _ := url.QueryUnescape(c.Params("table"))
	// TODO 当数据行数很大时，会占用很大内存，应该改为流式处理，这里限制最多10000行
	limit := c.Query("limit", "100")
	if i, err := strconv.Atoi(limit); err != nil || i > 10000 {
		limit = "100"
	}

	sqltext := fmt.Sprintf(`select row_to_json(%s) as json from %s limit %s`, table, table, limit)

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

		if err := utils.Json2excel(ch,
			sheetname, "log/"+filename); err != nil {
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

// GET /postgresql/views?mime=excel|json
func (p *PgHandler) viewsHandler(c fiber.Ctx) error {
	sqltext := `select 
		json_build_object(
			'schemaname', viw.schemaname,
			'viewname', viw.viewname,
			'oid', cla."oid",
			'viewowner', viw.viewowner,
			'reltablespace', cla.reltablespace,
			'reltype', cla.reltype,
			'definition', viw.definition,
			'relhasindex', cla.relhasindex,
			'relhasindex', cla.relhasindex,
			'relhastriggers', cla.relhastriggers,
			'relrowsecurity', cla.relrowsecurity,
			'rows', stat.n_live_tup,
			'description', des.description
		) as json
	from pg_views viw
	left join pg_class cla on viw.viewname = cla.relname
	left join pg_description des on	des.objoid = cla.oid and objsubid = 0  --为0就是表的描述，其他是字段的描述
	left join pg_stat_user_tables stat on viw.viewname = stat.relname 
	order by viw.schemaname, viw.viewname`

	mime := c.Query("mime", "json") // if Queries params mime is not set, default to json
	switch mime {
	case "json":
		return p.sqlHandlerByJson(c, sqltext)

	case "excel":
		filename := p.u.Path + "-tables.xlsx"
		sheetname := p.u.Path + " tables"
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

// GET /postgresql/view/:table?limit=10000&mime=excel|json
func (p *PgHandler) viewHandler(c fiber.Ctx) error {
	table, _ := url.QueryUnescape(c.Params("table"))
	// TODO 当数据行数很大时，会占用很大内存，应该改为流式处理，这里限制最多10000行
	limit := c.Query("limit", "100")
	if i, err := strconv.Atoi(limit); err != nil || i > 10000 {
		limit = "100"
	}

	sqltext := fmt.Sprintf(`select * from %s limit %s`, table, limit)

	mime := c.Query("mime", "json") // if Queries params mime is not set, default to json
	switch mime {
	case "json":
		// TODO 当表数据行数很大时，会占用很大内存，应该改为流式处理
		return p.sqlHandler2Json(c, sqltext)

	case "excel":
		filename := table + ".xlsx"
		sheetname := table
		ch := make(chan string, 100)
		go p.sql2chan(ch, sqltext)

		if err := utils.Json2excel(ch,
			sheetname, "log/"+filename); err != nil {
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

// GET /mysql/procedures
func (p *PgHandler) proceduresHandler(c fiber.Ctx) error {
	sqltext := `select 
		routine_catalog,
		routine_schema,
		routine_name,
		routine_type,
		routine_body,
		routine_definition,
		parameter_style,
		data_type
	from information_schema.routines`
	return p.sqlHandler2Json(c, sqltext)
}

// GET /mysql/procedure/:procedure
func (p *PgHandler) procedureHandler(c fiber.Ctx) error {
	procedure, _ := url.QueryUnescape(c.Params("procedure"))
	sqltext := fmt.Sprintf(`select * from pg_proc where proname = '%s'`, procedure)
	return p.sqlHandler2Json(c, sqltext)
}
