package main

import (
	"fmt"
	"os"

	"code.sajari.com/docconv/v2"
	"github.com/KSpaceer/goppt"
)

func pptxtool_file2txt(in_file, out_file string) error {
	file, err := os.Open(in_file)
	if err != nil {
		return err
	}
	defer file.Close()

	// Create or open the target text file
	outfp := os.Stdout
	if len(out_file) > 0 {
		txtfp, err := os.Create(out_file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "pptxtool_file2txt failed to create text file '%s': %v", out_file, err)
			return err
		}
		outfp = txtfp
	}
	defer outfp.Close()

	// ConvertPptx handles the internal XML stream parsing
	res, err := docconv.Convert(file, "application/vnd.openxmlformats-officedocument.presentationml.presentation", true)
	if err != nil {
		return err
	}

	// fmt.Fprintln(outfp, res.Body) // Extracted plain text content
	fmt.Fprintln(outfp, res)

	return nil
}

func ppttool_file2txt(in_file, out_file string) error {
	f, err := os.Open(in_file)
	if err != nil {
		return err
	}

	// Create or open the target text file
	outfp := os.Stdout
	if len(out_file) > 0 {
		txtfp, err := os.Create(out_file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ppttool_file2txt failed to create text file '%s': %v", out_file, err)
			return err
		}
		outfp = txtfp
	}
	defer outfp.Close()

	text, err := goppt.ExtractText(f)
	if err != nil {
		return err
	}

	fmt.Fprintln(outfp, text)

	return nil
}
