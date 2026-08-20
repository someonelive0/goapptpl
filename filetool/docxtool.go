package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"

	"code.sajari.com/docconv/v2"
)

func docxtool_file2txt(in_file, out_file string) error {
	res, err := docconv.ConvertPath(in_file)
	if err != nil {
		return err
	}

	// Create or open the target text file
	outfp := os.Stdout
	if len(out_file) > 0 {
		txtfp, err := os.Create(out_file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "docxtool_file2txt failed to create text file '%s': %v", out_file, err)
			return err
		}
		outfp = txtfp
	}
	defer outfp.Close()

	fmt.Fprintln(outfp, res)

	return nil
}

// 使用antiword转换器
func doctool_file2txt(in_file, out_file string) error {
	// Create a command to execute
	cmd := exec.Command("./antiword", "-r", "-s", "-w", "1200", "db", in_file)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	// Create or open the target text file
	outfp := os.Stdout
	if len(out_file) > 0 {
		txtfp, err := os.Create(out_file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "doctool_file2txt failed to create text file '%s': %v", out_file, err)
			return err
		}
		outfp = txtfp
	}
	defer outfp.Close()

	// Staart the subprocess
	if err = cmd.Start(); err != nil {
		return err
	}

	buf := make([]byte, 40960) // 每次读取40KB
	for {
		n, err := stdout.Read(buf)
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
