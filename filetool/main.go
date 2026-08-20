package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gabriel-vasile/mimetype"
)

var (
	param_version  = flag.Bool("v", false, "version")
	param_in_file  = flag.String("i", "", "set input file, eg file.docx or file.pdf")
	param_out_file = flag.String("o", "", "set output file to store txt contents, if empty output to stdout")
	APP_NAME       = "filetool"
	APP_VERSION    = "3.6.0"
	BUILD_TIME     = "unknown"
)

func init() {
	// parse command line
	flag.Parse()
	if *param_version {
		fmt.Fprintf(os.Stderr, "%s %s %s\n", APP_NAME, APP_VERSION, BUILD_TIME)
		os.Exit(0)
	}

	if len(*param_in_file) == 0 {
		fmt.Fprintf(os.Stderr, "input file can't be empty, please set it by -i\n")
		os.Exit(1)
	}

	Chdir2PrgPath()
}

func main() {
	in_file := *param_in_file
	out_file := *param_out_file

	// mimetype.SetLimit(1024 * 1024) // Set limit to 1MB.
	mimetype.SetLimit(0)
	mime, err := mimetype.DetectFile(in_file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "file '%s' get mime type failed: %s\n", in_file, err)
		os.Exit(1)
	}
	mimestr := mime.String()
	fmt.Fprintf(os.Stderr, "file '%s' mime: '%s', '%v'\n", in_file, mimestr, mime)

	switch mimestr {
	// .pdf
	case "application/pdf", "application/x-pdf":
		if err = pdftool_file2txt(in_file, out_file); err != nil {
			fmt.Fprintf(os.Stderr, "pdftool_file2txt failed: %s\n", err)
			os.Exit(1)
		}

	// .xlsx
	case "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":
		if err = xlsxtool_file2txt(in_file, out_file); err != nil {
			fmt.Fprintf(os.Stderr, "xlsxtool_file2txt failed: %s\n", err)
			os.Exit(1)
		}

	// .xls
	case "application/vnd.ms-excel", "application/msexcel":
		if err = xlstool_file2txt(in_file, out_file); err != nil {
			fmt.Fprintf(os.Stderr, "xlstool_file2txt failed: %s\n", err)
			os.Exit(1)
		}

	// .docx
	case "application/vnd.openxmlformats-officedocument.wordprocessingml.document":
		if err = docxtool_file2txt(in_file, out_file); err != nil {
			fmt.Fprintf(os.Stderr, "docxtool_file2txt failed: %s\n", err)
			os.Exit(1)
		}

	// .doc
	case "application/vnd.ms-word", "application/msword":
		if err = antiword_init(); err != nil {
			fmt.Fprintf(os.Stderr, "antiword_ini failed: %s\n", err)
			os.Exit(1)
		} else {
			if err = doctool_file2txt(in_file, out_file); err != nil {
				fmt.Fprintf(os.Stderr, "doctool_file2txt failed: %s\n", err)
				os.Exit(1)
			}
		}

	// .pptx
	case "application/vnd.openxmlformats-officedocument.presentationml.presentation":
		if err = pptxtool_file2txt(in_file, out_file); err != nil {
			fmt.Fprintf(os.Stderr, "pptxtool_file2txt failed: %s\n", err)
			os.Exit(1)
		}

	// .ppt
	case "application/vnd.ms-powerpoint", "application/mspowerpoint":
		if err = ppttool_file2txt(in_file, out_file); err != nil {
			fmt.Fprintf(os.Stderr, "ppttool_file2txt failed: %s\n", err)
			os.Exit(1)
		}

	case "application/x-ole-storage": // maybe doc or ppt
		ext := filepath.Ext(in_file)
		fmt.Fprintf(os.Stderr, "mime is application/x-ole-storage, check extention: %s\n", ext)
		if ext == ".ppt" {
			if err = ppttool_file2txt(in_file, out_file); err != nil {
				fmt.Fprintf(os.Stderr, "ppttool_file2txt failed: %s\n", err)
				os.Exit(1)
			}
		} else if ext == ".doc" {
			if err = doctool_file2txt(in_file, out_file); err != nil {
				fmt.Fprintf(os.Stderr, "doctool_file2txt failed: %s\n", err)
				os.Exit(1)
			}
		}

	default:
		// .csv .txt etc.
		if len(mimestr) > 5 && mimestr[:5] == "text/" {
			if err = texttool_file2txt(in_file, out_file); err != nil {
				fmt.Fprintf(os.Stderr, "texttool_file2txt failed: %s\n", err)
				os.Exit(1)
			}
		} else {
			fmt.Fprintf(os.Stderr, "unsupport file '%s' mime type '%s'\n", in_file, mime.String())
			os.Exit(1)
		}
	}
}
