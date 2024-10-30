package utils

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"

	log "github.com/sirupsen/logrus"
	"github.com/xuri/excelize/v2"
)

// read json from chan string and write to excel with sheet name
// TODO 当数据行数很大时，会占用很大内存，应该改为流式处理
func Json2excel(ch chan string, sheetname, filename string) error {
	f := excelize.NewFile()
	defer f.Close()

	index, err := f.NewSheet(sheetname) // 创建一个工作表
	if err != nil {
		log.Errorf("Create sheet '%s' failed: %v", sheetname, err)
		return err
	}

	m := make(map[string]interface{})
	rows := 0
	keys := make([]string, 0) // json 的 keys

	for jsonstr := range ch {
		json.Unmarshal([]byte(jsonstr), &m)

		// 当第一行时，填写表头
		if rows == 0 {
			for k := range m {
				keys = append(keys, k)
			}
			sort.Strings(keys)

			// 填写表头，并设置表头样式
			for cols := range len(keys) {
				// 从A,B,C...开始,当超过26个字母后，从AA,AB,AC...开始，或BA,BB,BC...开始
				// cols%26 是取余数，cols/26 是取倍数
				// cell := string('A'+cols%26) + strconv.Itoa(rows+1)
				cell := fmt.Sprintf("%c%d", 'A'+cols%26, rows+1)
				if cols/26 > 0 {
					cell = string('A'+(cols/26-1)) + cell
				}
				// fmt.Printf("cols: %d:%d %s: %s\n", rows, cols, cell, keys[cols])
				if err = f.SetCellValue(sheetname, cell, keys[cols]); err != nil {
					log.Warnf("SetCellValue header failed: %v", err)
				}

				// when last cell, set style
				if cols == len(keys)-1 {
					if err := SetHeaderStyle(f, sheetname, "A1", cell); err != nil {
						log.Warnf("SetHeaderStyle failed: %v", err)
					}
				}
			}

			rows++ // 跳过第一行表头
		}

		// 填写数据
		for cols := range len(keys) {
			cell := string('A'+cols%26) + strconv.Itoa(rows+1)
			if cols/26 > 0 {
				cell = string('A'+(cols/26-1)) + cell
			}
			if err = f.SetCellValue(sheetname, cell, m[keys[cols]]); err != nil {
				log.Warnf("SetCellValue failed: %v", err)
			}
		}

		rows++
	}

	f.SetActiveSheet(index) // 设置工作簿的默认工作表
	if err = f.SaveAs(filename); err != nil {
		return err
	}

	return nil
}

// set style for header, from first cell to last cell
// such as "sheet1", from "A1" to "Z1"
func SetHeaderStyle(f *excelize.File, sheetname, firstcell, lastcell string) error {
	style, err := f.NewStyle(&excelize.Style{
		// 设置边框
		Border: []excelize.Border{
			// {Type: "left", Color: "000000", Style: 3},
			// {Type: "top", Color: "000000", Style: 4},
			{Type: "bottom", Color: "000000", Style: 5}, // 双细线
			{Type: "right", Color: "000000", Style: 6},  // 粗线
			// {Type: "diagonalDown", Color: "A020F0", Style: 7},
			// {Type: "diagonalUp", Color: "A020F0", Style: 8},
		},

		// 设置字体
		Font: &excelize.Font{
			Bold: true,
			// Italic: true,
			// Family: "Times New Roman",
			Size:  12,
			Color: "#000000",
		},

		// 设置填充，填充颜色为淡黄色（Light Yellow ）
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#FFFF99"}, // Light Yellow
			Pattern: 1,
		},
	})
	if err != nil {
		return err
	}

	err = f.SetCellStyle(sheetname, firstcell, lastcell, style)
	if err != nil {
		return err
	}
	return nil
}
