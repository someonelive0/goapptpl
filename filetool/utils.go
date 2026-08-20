package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
)

func Chdir2PrgPath() (string, error) {
	file, err := exec.LookPath(os.Args[0])
	if err != nil {
		return "", err
	}
	path, err := filepath.Abs(file)
	if err != nil {
		return "", err
	}
	index := strings.LastIndex(path, string(os.PathSeparator))
	ret := path[:index]
	os.Chdir(ret)
	return ret, nil
}

/**
 * copy pwd .antiword to ${HOME}/.antiword when not existed
 * because antiword use as fonts map
 */
func antiword_init() error {
	currentUser, err := user.Current()
	if err != nil {
		fmt.Fprintln(os.Stderr, "can't get current user: ", err)
		return err
	}

	// fmt.Fprintf(os.Stderr, "current user uid=%s, username=%s, home=%s\n",
	// 	currentUser.Uid, currentUser.Username, currentUser.HomeDir)

	antiword_map_dir := currentUser.HomeDir + "/.antiword"
	if _, err := os.Stat(antiword_map_dir); os.IsNotExist(err) {
		fmt.Fprintln(os.Stderr, "antiword map dir is not existed: ", antiword_map_dir)

		pwd_map_dir := ".antiword"
		if _, err := os.Stat(pwd_map_dir); os.IsNotExist(err) {
			fmt.Fprintln(os.Stderr, "pwd antiword map dir is also not existed, can't copy dir to $HOME: ", pwd_map_dir)
			return err
		} else {
			fmt.Fprintf(os.Stderr, "cp antiword map dir '%s' to $HOME\n", pwd_map_dir)
			if err = CopyDir(pwd_map_dir, antiword_map_dir); err != nil {
				fmt.Fprintf(os.Stderr, "cp antiword map dir '%s' to $HOME failed: %s\n", pwd_map_dir, err)
				return err
			}
		}
	}

	return nil
}

// Manual Recursive Copy (Pre-Go 1.23)
func CopyDir(src, dst string) error {
	err := os.MkdirAll(dst, os.ModePerm)
	if err != nil {
		return err
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())
		if entry.IsDir() {
			if err = CopyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			data, err := os.ReadFile(srcPath)
			if err != nil {
				return err
			}
			if err = os.WriteFile(dstPath, data, 0644); err != nil {
				return err
			}
		}
	}
	return nil
}
