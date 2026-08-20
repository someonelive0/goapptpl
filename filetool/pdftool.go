package main

import (
	"fmt"
	"os"

	"github.com/ledongthuc/pdf"
)

func pdftool_file2txt(in_file, out_file string) error {
	pdf.DebugOn = false

	f, r, err := pdf.Open(in_file)
	if err != nil {
		return err
	}
	defer f.Close()

	// Create or open the target text file
	outfp := os.Stdout
	if len(out_file) > 0 {
		txtfp, err := os.Create(out_file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "pdftool_file2txt failed to create text file '%s': %v", out_file, err)
			return err
		}
		outfp = txtfp
	}
	defer outfp.Close()

	totalPage := r.NumPage()

	// var textBuilder bytes.Buffer
	for pageIndex := 1; pageIndex <= totalPage; pageIndex++ {
		p := r.Page(pageIndex)
		if p.V.IsNull() {
			continue
		}
		pagetext, err := p.GetPlainText(nil)
		if err != nil {
			return err
		}
		// textBuilder.WriteString(pagetext)
		fmt.Fprintln(outfp, pagetext)
	}

	// text := textBuilder.String()
	// fmt.Fprintln(outfp, text)

	return nil
}

/*
func pdftool_file2txt1(in_file, out_file string) error {
	pdf.DebugOn = true

	f, r, err := pdf.Open(in_file)
	if err != nil {
		return err
	}
	defer f.Close()

	var buf bytes.Buffer
	b, err := r.GetPlainText()
	if err != nil {
		return err
	}
	buf.ReadFrom(b)

	// convert utf16le to utf8
	// UTF16LE解码器 // data: utf16le原始字节（不带BOM / 带BOM都兼容）
	// decoder := unicode.UTF16(unicode.LittleEndian, unicode.UseBOM).NewDecoder()
	// bs2, err := decoder.Bytes(buf.Bytes())
	// // bs2, _, err := transform.Bytes(decoder, buf.Bytes())
	// if err != nil {
	// 	return err
	// }
	// uint16arr := BytesToUint16s(buf.Bytes())
	// content := DecodeUTF16ToString(uint16arr)
	content := buf.String() // string(bs2) //
	fmt.Fprintln(os.stdout, "%s\n", content)

	return nil
}

func pdftool_file2txt_by_rows(in_file, out_file string) error {
	pdf.DebugOn = true

	f, r, err := pdf.Open(in_file)
	defer func() {
		_ = f.Close()
	}()
	if err != nil {
		return err
	}
	totalPage := r.NumPage()

	for pageIndex := 1; pageIndex <= totalPage; pageIndex++ {
		p := r.Page(pageIndex)
		if p.V.IsNull() || p.V.Key("Contents").Kind() == pdf.Null {
			continue
		}

		rows, _ := p.GetTextByRow()
		for _, row := range rows {
			println(">>>> row: ", row.Position)
			for _, word := range row.Content {
				fmt.Fprintln(os.stdout, word.S)
			}
		}
	}

	return nil
}
*/
