package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/gofiber/fiber/v3"
	log "github.com/sirupsen/logrus"
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
	r.Get("/table/:table", p.tableHandler)

	return nil
}

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

	return p.sqlHandler(c, sqltext)
}

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

	return p.sqlHandler(c, sqltext)
}

func (p *MysqlHandler) tableHandler(c fiber.Ctx) error {
	table, _ := url.QueryUnescape(c.Params("table"))
	// columns := "id,api_id,app_id,hostname,buz_source,asset_name,api_method,api_endpoint,content_type,module_code,department_id,business_id,description,follow,monitor_cover,fever,asset_state,asset_value,sen_fever,discovery_time,risk_level,carrier_type,validate_time,ext_info,merge_state,check_state,tenant_id,create_user,create_time,update_user,update_time,api_no,pod,resource_pool,asset_code"
	columns, err := p.getColumns(table)
	if err != nil {
		c.WriteString(err.Error())
		return err
	}
	columnArray := strings.Split(columns, ",")
	limit := c.Query("limit", "100")

	sqltext := `
	select json_object(`

	for i, col := range columnArray {
		if i > 0 {
			sqltext += `,`
		}
		sqltext += `'` + col + `',` + col
	}

	sqltext += `	) as json 
	from ` + table + ` limit ` + limit

	return p.sqlHandler(c, sqltext)
}

func (p *MysqlHandler) sqlHandler(c fiber.Ctx, sqltext string) error {
	log.Tracef("/mysql sql: %s\n", sqltext)
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

	if err = rows.Err(); err != nil {
		log.Error("Error iterating through rows:", err)
		return err
	}

	return nil
}

// get columns of table to string with ',' split
func (p *MysqlHandler) getColumns(table string) (string, error) {
	if p.db == nil {
		if err := p.openDB(); err != nil {
			return "", err
		}
	}

	sqltext := `select group_concat(column_name) from INFORMATION_SCHEMA.COLUMNS
	where table_schema = '` + p.cfg.DBName + `' and table_name = '` + table + `'`
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

	return columns, nil
}

func (p *MysqlHandler) openDB() error {
	//将空闲时间字符串解析为time.Duration类型
	MaxIdleDuration, err := time.ParseDuration(p.Dbconfig.MaxIdleTime)
	if err != nil {
		return fmt.Errorf("parse dbconfig.maxidletime [%s] failed: %s", p.Dbconfig.MaxIdleTime, err)
	}

	//解析DSN字符串
	cfg, err := mysql.ParseDSN(p.Dbconfig.Dsn[0])
	if err != nil {
		log.Error("parse dsn failed:", err)
		return err
	}
	p.cfg = cfg
	// log.Debugf("mysql cfg: %#v", cfg)

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
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
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
