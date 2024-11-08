package utils

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gomutex/godocx"
	log "github.com/sirupsen/logrus"
)

// read json from chan string and write to docx
// TODO 当数据行数很大时，会占用很大内存，应该改为流式处理
func Json2docx(ch chan string, title, filename string) error {

	// Create a new DOCX document
	document, err := godocx.NewDocument()
	if err != nil {
		log.Errorf("new docx error: %v", err)
		return err
	}

	document.AddHeading(title, 0)

	// Add a new paragraph to the document
	document.AddParagraph("产生时间: " + time.Now().Format("2006-01-02 15:04:05"))

	document.AddHeading("数据表格", 1)
	document.AddParagraph("下面是数据内容").Style("List Bullet")

	table := document.AddTable()
	table.Style("LightList-Accent4")

	m := make(map[string]interface{})
	columns := make([]string, 0)
	rows := 0
	for jsonstr := range ch {
		json.Unmarshal([]byte(jsonstr), &m)

		if rows == 0 {
			hdrRow := table.AddRow()
			for k := range m {
				columns = append(columns, k)
				hdrRow.AddCell().AddParagraph(k)
			}
		}

		// 填写数据
		row := table.AddRow()
		for cols := range len(columns) {
			text := fmt.Sprintf("%v", m[columns[cols]])
			row.AddCell().AddParagraph(text)
		}

		rows++
	}

	total := fmt.Sprintf("共 %d 行", rows)
	document.AddParagraph(total).Style("Intense Quote")

	// Save the modified document to a new file
	err = document.SaveTo(filename)
	if err != nil {
		log.Errorf("save docx '%s' error: %v", filename, err)
		return err
	}

	return nil
}

// read json from chan string and write to docx
// TODO 当数据行数很大时，会占用很大内存，应该改为流式处理
func Json2docxWithColumn(ch chan string, columns []string,
	title, filename string) error {

	// Create a new DOCX document
	document, err := godocx.NewDocument()
	if err != nil {
		log.Errorf("new docx error: %v", err)
		return err
	}

	document.AddHeading(title, 0)

	// Add a new paragraph to the document
	document.AddParagraph("产生时间: " + time.Now().Format("2006-01-02 15:04:05"))

	document.AddHeading("数据表格", 1)
	document.AddParagraph("下面是数据内容").Style("List Bullet")

	table := document.AddTable()
	table.Style("LightList-Accent4")
	hdrRow := table.AddRow()
	for _, column := range columns {
		hdrRow.AddCell().AddParagraph(column)
	}

	m := make(map[string]interface{})
	rows := 0
	for jsonstr := range ch {
		json.Unmarshal([]byte(jsonstr), &m)

		// 填写数据
		row := table.AddRow()
		for cols := range len(columns) {
			text := fmt.Sprintf("%v", m[columns[cols]])
			row.AddCell().AddParagraph(text)
		}

		rows++
	}

	total := fmt.Sprintf("共 %d 行", rows)
	document.AddParagraph(total).Style("Intense Quote")

	// Save the modified document to a new file
	err = document.SaveTo(filename)
	if err != nil {
		log.Errorf("save docx '%s' error: %v", filename, err)
		return err
	}

	return nil
}
