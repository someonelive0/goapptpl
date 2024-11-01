package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/BurntSushi/toml"

	"goapptol/utils"
)

type DBConfig struct {
	Dbtype       string   `toml:"dbtype" json:"dbtype"`
	MaxOpenConns int      `toml:"maxopenconns" json:"maxopenconns"`
	MaxIdleConns int      `toml:"maxidleconns" json:"maxidleconns"`
	MaxIdleTime  string   `toml:"maxidletime" json:"maxidletime"`
	Dsn          []string `toml:"dsn" json:"dsn"`
}

type MinioConfig struct {
	Addr     string `toml:"addr" json:"addr"`
	User     string `toml:"user" json:"user"`
	Password string `toml:"password" json:"-"`
	Ssl      bool   `toml:"ssl" json:"ssl"`
	Timeout  uint   `toml:"timeout" json:"timeout"`
}

type RedisConfig struct {
	Addr     string `toml:"addr" json:"addr"`
	Password string `toml:"password" json:"-"`
	Db       uint   `toml:"db" json:"db"`
	Timeout  uint   `toml:"timeout" json:"timeout"`
}

type LogConfig struct {
	Level         string `toml:"level" json:"level"`
	Path          string `toml:"path" json:"path"`
	Filename      string `toml:"filename" json:"filename"`
	Rotate_files  uint   `toml:"rotate_files" json:"rotate_files"`
	Rotate_mbytes uint   `toml:"rotate_mbytes" json:"rotate_mbytes"`
}

type MyConfig struct {
	Filename string    `toml:"filename" json:"filename" xml:"filename,attr"`
	LoadTime time.Time `toml:"load_time" json:"load_time" xml:"load_time,attr"`

	Version string `toml:"version" json:"version"`
	Host    string `toml:"host" json:"host"`
	Port    uint   `toml:"port" json:"port"`

	MysqlConfig DBConfig    `toml:"mysql" json:"mysql"`
	MinioConfig MinioConfig `toml:"minio" json:"minio"`
	RedisConfig RedisConfig `toml:"redis" json:"redis"`
	CkConfig    DBConfig    `toml:"clickhouse" json:"clickhouse"`
	LogConfig   LogConfig   `toml:"log" json:"log"`
}

func (p *MyConfig) Dump() []byte {
	b, _ := json.MarshalIndent(p, "", " ")
	return b
}

func LoadConfig(filename string) (*MyConfig, error) {
	if !utils.ExistedOrCopy(filename, filename+".tpl") {
		return nil, fmt.Errorf("config file [%s] or template file are not found", filename)
	}

	myconfig := &MyConfig{
		Filename: filename,
		LoadTime: time.Now(),
	}
	_, err := toml.DecodeFile(filename, myconfig)
	if err != nil {
		return nil, fmt.Errorf("config file [%s] unmarshal toml failed: %s", filename, err)
	}

	return myconfig, nil
}
