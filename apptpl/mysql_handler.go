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

	"github.com/go-sql-driver/mysql"
	"github.com/gofiber/fiber/v3"
	log "github.com/sirupsen/logrus"

	"goapptol/utils"
)

const (
	MYSQL_MAX_TIMEOUT = 30 // mysql max timeout in seconds
)

type MysqlHandler struct {
	Dbconfig *DBConfig
	db       *sql.DB       // mysql dbpool
	cfg      *mysql.Config // mysql config of dsn
}

// r := app.Group("/mysql")
func (p *MysqlHandler) AddRouter(r fiber.Router) error {
	log.Info("MysqlHandler AddRouter")

	r.Get("/tables", p.tablesHandler)
	r.Get("/table/:table/columns", p.columnsHandler)
	r.Get("/table/:table/indexes", p.indexesHandler)
	r.Get("/table/:table", p.tableHandler)

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

// GET /mysql/tables?mime=excel|json
// mysql json_object() 不保证字段顺序，所以excel格式化时，需要按顺序
func (p *MysqlHandler) tablesHandler(c fiber.Ctx) error {
	sqltext := `
	select json_object(
		'table_catalog', table_catalog,
		'table_schema', table_schema,
		'table_name', table_name,
		'table_type', table_type,
		'table_rows', table_rows,
		'table_rows', table_rows,
		'avg_row_length', avg_row_length,
		'data_length', data_length,
		'index_length', index_length,
		'create_time', create_time,
		'table_collation', table_collation,
		'table_comment', table_comment
		) as json
	from INFORMATION_SCHEMA.TABLES
	where table_schema = '` + p.cfg.DBName + `'`

	mime := c.Query("mime", "json") // if Queries params mime is not set, default to json
	switch mime {
	case "json":
		return p.sqlHandler(c, sqltext)

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
		return p.sqlHandler(c, sqltext)

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
		return err

	default:
		c.Status(400)
		c.SendString(fmt.Sprintf("mime '%s' not supported", mime))
		return nil
	}
}

// write sql result to fiber response
func (p *MysqlHandler) sqlHandler(c fiber.Ctx, sqltext string) error {
	log.Tracef("mysql sql: %s\n", sqltext)
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
	log.Tracef("mysql query rows: %d", i)

	if err = rows.Err(); err != nil {
		log.Error("Error iterating through rows:", err)
		return err
	}

	return nil
}

// write sql result to channel
func (p *MysqlHandler) sql2chan(ch chan string, sqltext string) error {
	log.Tracef("/mysql sql: %s\n", sqltext)
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

func (p *MysqlHandler) openDB() error {
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
	ctx, cancel := context.WithTimeout(context.Background(), MYSQL_MAX_TIMEOUT*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		log.Errorf("ping mysql [%s] failed: %s", p.Dbconfig.Dsn[0], err)
		return err
	}

	log.Infof("ping mysql [%s] success", p.cfg.Addr)
	p.db = db
	return nil
}

func (p *MysqlHandler) Close() error {
	if p.db != nil {
		return p.db.Close()
	}
	return nil
}
