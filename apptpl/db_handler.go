package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/patrickmn/go-cache"
	log "github.com/sirupsen/logrus"
)

const (
	MAX_TIMEOUT = 30 // db max timeout in seconds
)

type DbHandler struct {
	Dbconfig *DBConfig
	Mycache  *cache.Cache
	db       *sql.DB // dbpool
}

// write sql result from colums record to fiber response
func (p *DbHandler) sqlHandler2Json(c fiber.Ctx, sqltext string) error {
	log.Tracef("%s SQL: %s\n", p.Dbconfig.Dbtype, sqltext)

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

	// c.Context().SetContentType("text/x-sql;charset=UTF-8") // text/plain;charset=UTF-8
	c.Response().Header.Set("Content-Type", "application/json")

	columns, err := rows.Columns()
	if err != nil {
		log.Error("Error getting columns:", err)
		return err
	}
	column_num := len(columns)

	// 返回值 Map切片
	// records := make([]map[string]interface{}, 0)
	// 一条数据的各列的值（需要指定长度为列的个数，以便获取地址）
	values := make([]interface{}, column_num)
	// 一条数据的各列的值的地址
	values_ptr := make([]interface{}, column_num)

	c.WriteString("[")
	i := 0
	for rows.Next() {
		// 获取各列的值的地址
		for i := 0; i < column_num; i++ {
			values_ptr[i] = &values[i]
		}

		// 扫描一行数据到值数组中
		err = rows.Scan(values_ptr...)
		if err != nil {
			log.Error("Error scanning row:", err)
			continue
		}

		// 一条数据的Map (列名和值的键值对)
		entry := make(map[string]interface{})

		// Map 赋值，将列名和值对应起来
		for i, col := range columns {
			var v interface{}

			val := values[i] // 值复制给val(所以Scan时指定的地址可重复使用)
			b, ok := val.([]byte)
			if ok {
				v = string(b) // 字符切片转为字符串
			} else {
				v = val
			}
			entry[col] = v
		}

		// records = append(records, entry)
		if i > 0 {
			c.WriteString(",")
		}
		b, _ := json.Marshal(entry)
		c.Write(b)
		i++
	}
	c.WriteString("]")
	log.Tracef("%s query rows: %d", p.Dbconfig.Dbtype, i)

	if err = rows.Err(); err != nil {
		log.Error("Error iterating through rows:", err)
		return err
	}

	return nil
}

// write sql result from json object to fiber response
func (p *DbHandler) sqlHandlerByJson(c fiber.Ctx, sqltext string) error {
	log.Tracef("%s SQL: %s\n", p.Dbconfig.Dbtype, sqltext)
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

	c.Response().Header.Set("Content-Type", "application/json")

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
	log.Tracef("%s query rows: %d", p.Dbconfig.Dbtype, i)

	if err = rows.Err(); err != nil {
		log.Error("Error iterating through rows:", err)
		return err
	}

	return nil
}

// write sql result to channel
func (p *DbHandler) sql2chan(ch chan string, sqltext string) error {
	log.Tracef("%s sql: %s\n", p.Dbconfig.Dbtype, sqltext)
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

func (p *DbHandler) openDB() error {
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
	ctx, cancel := context.WithTimeout(context.Background(), MAX_TIMEOUT*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		log.Errorf("ping db [%s] failed: %s", p.Dbconfig.Dsn[0], err)
		return err
	}

	log.Infof("ping db [%s] success", p.Dbconfig.Dsn[0])
	p.db = db
	return nil
}

func (p *DbHandler) Close() error {
	if p.db != nil {
		return p.db.Close()
	}
	return nil
}
