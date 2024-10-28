package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gofiber/fiber/v3"
)

func AddMysqlHandler(r fiber.Router) error {
	println("0 ---------------->")

	// r := app.Group("/mysql")

	r.Get("/tables", tablesHandler)
	r.Get("/table/:table/columns", columnsHandler)
	r.Get("/table/:table", tableHandler)

	return nil
}

func tablesHandler(c fiber.Ctx) error {
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
	where table_schema = 'idss_dsmcp'`

	return sqlHandler(c, sqltext)
}

func columnsHandler(c fiber.Ctx) error {
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
	where table_schema = 'idss_dsmcp' and table_name = '` + table + `'`

	return sqlHandler(c, sqltext)
}

func tableHandler(c fiber.Ctx) error {
	table, _ := url.QueryUnescape(c.Params("table"))
	// columns := "id,api_id,app_id,hostname,buz_source,asset_name,api_method,api_endpoint,content_type,module_code,department_id,business_id,description,follow,monitor_cover,fever,asset_state,asset_value,sen_fever,discovery_time,risk_level,carrier_type,validate_time,ext_info,merge_state,check_state,tenant_id,create_user,create_time,update_user,update_time,api_no,pod,resource_pool,asset_code"
	columns, err := getColumns(table)
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

	return sqlHandler(c, sqltext)
}

func sqlHandler(c fiber.Ctx, sqltext string) error {
	log.Printf("/mysql/sql: %s\n", sqltext)
	db, err := sql.Open("mysql", "idss:BDsec2022,,@tcp(192.168.30.41:3306)/idss_dsmcp")
	if err != nil {
		fmt.Println("Error connecting to the database:", err)
		c.WriteString(err.Error())
		return err
	}
	defer db.Close()

	rows, err := db.Query(sqltext)
	if err != nil {
		fmt.Println("Error executing query:", err)
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
			fmt.Println("Error scanning row:", err)
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
		fmt.Println("Error iterating through rows:", err)
	}

	return nil
}

// get columns of table to string with ',' split
func getColumns(table string) (string, error) {
	db, err := sql.Open("mysql", "idss:BDsec2022,,@tcp(192.168.30.41:3306)/idss_dsmcp")
	if err != nil {
		fmt.Println("Error connecting to the database:", err)
		return "", err
	}
	defer db.Close()

	sqltext := `select group_concat(column_name) from INFORMATION_SCHEMA.COLUMNS
	where table_schema = 'idss_dsmcp' and table_name = '` + table + `'`
	rows, err := db.Query(sqltext)
	if err != nil {
		fmt.Println("Error executing query:", err)
		return "", err
	}
	defer rows.Close()

	var columns string
	for rows.Next() {
		err = rows.Scan(&columns)
		if err != nil {
			fmt.Println("Error scanning row:", err)
			continue
		}
		break // just get one row
	}

	if err = rows.Err(); err != nil {
		fmt.Println("Error iterating through rows:", err)
	}

	return columns, nil
}
