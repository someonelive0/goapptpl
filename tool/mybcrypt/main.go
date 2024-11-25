package main

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const (
	// 传递给GenerateFromPassword的最小允许开销
	MinCost int = 4
	// 传递给GenerateFromPassword的最大允许开销
	MaxCost int = 31
	// 如果将低于MinCost的cost传递给GenerateFromPassword，则实际设置的cost
	DefaultCost int = 10
)

func main() {
	pwd := []byte("123456")
	hashed_pwd, _ := hash_pwd(pwd)
	fmt.Println(string(hashed_pwd))

	if verify(pwd, hashed_pwd) {
		fmt.Println("password is correct")
	}

	// =================
	fmt.Println("=================================")
	hashed_pwd = []byte("$2a$10$LC1ZhCKcSSHFoX6syEAmwOR0MELechH.il86dLbf3q9I5th.N.FD2")
	pwd = []byte("123456")
	if verify(pwd, hashed_pwd) {
		fmt.Println("second password is correct")
	} else {
		fmt.Println("second password is not correct")
	}
}

func hash_pwd(pwd []byte) ([]byte, error) {
	return bcrypt.GenerateFromPassword(pwd, bcrypt.DefaultCost)
}

func verify(pwd, hashed_pwd []byte) bool {

	err := bcrypt.CompareHashAndPassword(hashed_pwd, pwd)
	if err != nil {
		fmt.Println("CompareHashAndPassword failed: ", err)
	}
	return err == nil
}
