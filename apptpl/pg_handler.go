package main

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"

	"goapptol/utils"
)

const (
	PG_MAX_TIMEOUT = 30 // pg max timeout in seconds
)

type PgHandler struct {
	Dbconfig *DBConfig
	db       *sql.DB  // pg dbpool
	u        *url.URL // pg url of dsn
}

// r := app.Group("/postgresql")
func (p *PgHandler) AddRouter(r fiber.Router) error {
	log.Info("PgHandler AddRouter")

	r.Get("/tables", p.tablesHandler)
	r.Get("/table/:table/columns", p.columnsHandler)
	r.Get("/table/:table/indexes", p.indexesHandler)
	r.Get("/table/:table", p.tableHandler)

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
		return p.sqlHandler(c, sqltext)

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
		return p.sqlHandler(c, sqltext)

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
		return p.sqlHandler(c, sqltext)

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
		return p.sqlHandler(c, sqltext)

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
		return err

	default:
		c.Status(400)
		c.SendString(fmt.Sprintf("mime '%s' not supported", mime))
		return nil
	}
}

// write sql result to fiber response
func (p *PgHandler) sqlHandler(c fiber.Ctx, sqltext string) error {
	log.Tracef("postgresql sql: %s\n", sqltext)
	if p.db == nil {
		if err := p.openDB(); err != nil {
			return err
		}
	}

	rows, err := p.db.Query(sqltext)
	if err != nil {
		log.Error("Error executing query:", err)
		c.WriteString(err.Error())
		return err
	}
	defer rows.Close()

	c.Context().SetContentType("application/json")

	c.WriteString("[")
	i := 0
	for rows.Next() {
		var jsonstr string
		err = rows.Scan(&jsonstr)
		if err != nil {
			log.Error("Error scanning row:", err)
			continue
		}
		if i > 0 {
			c.WriteString(",")
		}
		c.WriteString(jsonstr)
		i++
	}
	c.WriteString("]")
	log.Tracef("pg query rows: %d", i)

	if err = rows.Err(); err != nil {
		log.Error("Error iterating through rows:", err)
		return err
	}

	return nil
}

// write sql result to channel
func (p *PgHandler) sql2chan(ch chan string, sqltext string) error {
	log.Tracef("/postgresql sql: %s\n", sqltext)
	if p.db == nil {
		if err := p.openDB(); err != nil {
			return err
		}
	}

	rows, err := p.db.Query(sqltext)
	if err != nil {
		log.Error("Error executing query:", err)
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var jsonstr string
		err = rows.Scan(&jsonstr)
		if err != nil {
			log.Error("Error scanning row:", err)
			continue
		}
		ch <- jsonstr
	}

	if err = rows.Err(); err != nil {
		log.Error("Error iterating through rows:", err)
		return err
	}

	close(ch)
	return nil
}

// get columns of table to string with ',' split. sort by ordinal_position
// func (p *PgHandler) getColumns(table string) ([]string, error) {
// 	if p.db == nil {
// 		if err := p.openDB(); err != nil {
// 			return nil, err
// 		}
// 	}

// 	sqltext := fmt.Sprintf(`select column_name
// 	from information_schema.columns
// 	where table_name='%s'
// 	order by ordinal_position `, table)
// 	rows, err := p.db.Query(sqltext)
// 	if err != nil {
// 		log.Error("Error executing query:", err)
// 		return nil, err
// 	}
// 	defer rows.Close()

// 	var columns []string = make([]string, 0)
// 	var column string
// 	for rows.Next() {
// 		err = rows.Scan(&column)
// 		if err != nil {
// 			log.Error("Error scanning row:", err)
// 			continue
// 		}
// 		columns = append(columns, column)
// 	}

// 	if err = rows.Err(); err != nil {
// 		log.Error("Error iterating through rows:", err)
// 		return nil, err
// 	}

// 	return columns, nil
// }

func (p *PgHandler) openDB() error {
	//将空闲时间字符串解析为time.Duration类型
	MaxIdleDuration, err := time.ParseDuration(p.Dbconfig.MaxIdleTime)
	if err != nil {
		return fmt.Errorf("parse dbconfig.maxidletime [%s] failed: %s", p.Dbconfig.MaxIdleTime, err)
	}

	//打开数据库连接
	db, err := sql.Open(p.Dbconfig.Dbtype, p.Dbconfig.Dsn[0])
	if err != nil {
		log.Error("open database failed:", err)
		return err
	}

	//设置最大开放连接数，注意该值为小于0或等于0指的是无限制连接数
	db.SetMaxOpenConns(p.Dbconfig.MaxOpenConns)

	//设置空闲连接数，将此值设置为小于或等于0将意味着不保留空闲连接，即立即关闭连接
	db.SetMaxIdleConns(p.Dbconfig.MaxIdleConns)

	//设置最大空闲超时
	db.SetConnMaxIdleTime(MaxIdleDuration)
	ctx, cancel := context.WithTimeout(context.Background(), PG_MAX_TIMEOUT*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		log.Errorf("ping postgresql [%s] failed: %s", p.Dbconfig.Dsn[0], err)
		return err
	}

	log.Infof("ping postgresql [%s:%s] success", p.u.Host, p.u.Port())
	p.db = db
	return nil
}

func (p *PgHandler) Close() error {
	if p.db != nil {
		return p.db.Close()
	}
	return nil
}
