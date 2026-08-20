package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/pbnjay/grate"
	_ "github.com/pbnjay/grate/simple" // tsv and csv support
	_ "github.com/pbnjay/grate/xls"
	_ "github.com/pbnjay/grate/xlsx"
	"github.com/xuri/excelize/v2"
)

func xlsxtool_file2txt(in_file, out_file string) error {
	f, err := excelize.OpenFile(in_file)
	if err != nil {
		return err
	}
	defer f.Close()

	// Create or open the target text file
	outfp := os.Stdout
	if len(out_file) > 0 {
		txtfp, err := os.Create(out_file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "xlsxtool_file2txt failed to create text file '%s': %v", out_file, err)
			return err
		}
		outfp = txtfp
	}
	defer outfp.Close()

	sheet_count := f.SheetCount

	for i := range sheet_count {
		// Get all rows from the first sheet
		sheetName := f.GetSheetName(i)
		fmt.Fprintf(outfp, "--- %s ---\n", sheetName)
		rows, err := f.GetRows(sheetName)
		if err != nil {
			return err
		}

		// Print rows as plain text
		for _, row := range rows {
			for _, col := range row {
				fmt.Fprintf(outfp, "%s\t", col)
			}
			fmt.Fprintln(outfp)
		}
	}
	return nil
}

/**
 * grate support both xlsx and xls
 */
func xlstool_file2txt(in_file, out_file string) error {
	// Open the legacy XLS file
	wb, err := grate.Open(in_file)
	if err != nil {
		log.Fatalf("Failed to open XLS file: %v", err)
	}
	defer wb.Close()

	// Create or open the target text file
	outfp := os.Stdout
	if len(out_file) > 0 {
		txtfp, err := os.Create(out_file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "xlstool_file2txt failed to create text file '%s': %v", out_file, err)
			return err
		}
		outfp = txtfp
	}
	defer outfp.Close()

	// Get a list of all sheets in the file
	sheets, err := wb.List()
	if err != nil {
		return fmt.Errorf("xlstool_file2txt failed to list sheets: %v", err)
	}

	// Process each sheet and extract text
	for _, sheetName := range sheets {
		sheet, err := wb.Get(sheetName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "xlstool_file2txt failed to open sheet %s: %v", sheetName, err)
			continue
		}

		fmt.Fprintf(outfp, "--- %s ---\n", sheetName)

		// Iterate through all rows and columns
		for sheet.Next() {
			row := sheet.Strings()
			// Join cell values with a tab delimiter for clean text scanning
			line := strings.Join(row, "\t")
			fmt.Fprintln(outfp, line)
		}
	}

	return nil
}
