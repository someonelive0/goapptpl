package main

import (
	"fmt"
	"io"
	"os"
)

func texttool_file2txt(in_file, out_file string) error {
	infp, err := os.Open(in_file)
	if err != nil {
		return err
	}
	defer infp.Close()

	// Create or open the target text file
	outfp := os.Stdout
	if len(out_file) > 0 {
		txtfp, err := os.Create(out_file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "texttool_file2txt failed to create text file '%s': %v", out_file, err)
			return err
		}
		outfp = txtfp
	}
	defer outfp.Close()

	buf := make([]byte, 40960) // 每次读取40KB
	for {
		n, err := infp.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		fmt.Fprintf(outfp, "%s", string(buf[:n]))
	}

	return nil
}
