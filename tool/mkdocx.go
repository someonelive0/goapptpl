package main

import (
	"fmt"
	"log"

	"github.com/gomutex/godocx"
)

func main() {
	// Create a new DOCX document
	document, err := godocx.NewDocument()
	if err != nil {
		log.Fatal(err)
	}

	document.AddHeading("Document Title，文档标题", 0)

	// Add a new paragraph to the document
	p := document.AddParagraph("A plain paragraph having some ")
	p.AddText("bold").Bold(true)
	p.AddText(" and some ")
	p.AddText("italic.").Italic(true)

	// Add a new paragraph to the document
	p = document.AddParagraph("A 文本表示 plain paragraph having some ")
	p.AddText("粗体").Bold(true)
	p.AddText(" and some ")
	p.AddText("斜体.").Italic(true)

	document.AddHeading("Heading, level 1", 1)
	document.AddParagraph("Intense quote").Style("Intense Quote")
	document.AddParagraph("first item in unordered list").Style("List Bullet")
	document.AddParagraph("first item in ordered list").Style("List Number")

	document.AddHeading("中文标题， level 2", 1)
	document.AddParagraph("哈哈 quote").Style("Intense Quote")
	document.AddParagraph("呵呵 item in unordered list").Style("List Bullet")
	document.AddParagraph("哦哦 item in ordered list").Style("List Number")

	records := []struct{ Qty, ID, Desc string }{{"5", "A001", "Laptop"}, {"10", "B202", "Smartphone"}, {"2", "E505", "Smartwatch"}}

	table := document.AddTable()
	table.Style("LightList-Accent4")
	hdrRow := table.AddRow()
	hdrRow.AddCell().AddParagraph("Qty")
	hdrRow.AddCell().AddParagraph("ID")
	hdrRow.AddCell().AddParagraph("Description")

	for _, record := range records {
		row := table.AddRow()
		row.AddCell().AddParagraph(record.Qty)
		row.AddCell().AddParagraph(record.ID)
		row.AddCell().AddParagraph(record.Desc)
	}

	fmt.Println(document.Path)
	fmt.Printf("body %#v\n", len(document.Document.Body.Children))
	for _, child := range document.Document.Body.Children {
		fmt.Printf("child %#v\n", child)

		if child.Para != nil {
			para_children := child.Para.GetCT().Children

			for _, para_child := range para_children {
				for _, run_child := range para_child.Run.Children {
					if run_child.Text != nil {
						fmt.Printf("run_child %#v\n", run_child.Text.Text)
						run_child.Text.Text += " 哈哈"
						fmt.Printf("run_child2 %#v\n", run_child.Text.Text)
					}
				}
			}
		}

	}

	// Save the modified document to a new file
	err = document.SaveTo("log/demo.docx")
	if err != nil {
		log.Fatal(err)
	}
}
