package utils

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/araddon/dateparse"
	"github.com/sirupsen/logrus"
)

const (
	TIME_HUMAN = "2006-01-02 15:04:05"
)

func GetPrgDir() (string, error) {
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
	return ret, nil
}

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

func ExistedOrCopy(dstfile, srcfile string) bool {
	if _, err := os.Stat(dstfile); err == nil {
		return true
	} else {
		if os.IsNotExist(err) {
			logrus.Warnf("File [%s] does not exist. Copy from [%s]", dstfile, srcfile)
			if _, err1 := CopyFile(srcfile, dstfile); err1 == nil {
				return true
			} else {
				logrus.Errorf("Copy file [%s] to [%s] failed: %s", srcfile, dstfile, err1)
			}
		}
	}
	return false
}

func CopyFile(srcfile, dstfile string) (int64, error) {
	fstat, err := os.Stat(srcfile)
	if err != nil {
		return 0, err
	}
	if !fstat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", srcfile)
	}

	f0, err := os.Open(srcfile)
	if err != nil {
		return 0, err
	}
	defer f0.Close()

	f1, err := os.Create(dstfile)
	if err != nil {
		return 0, err
	}
	defer f1.Close()
	nBytes, err := io.Copy(f1, f0)

	return nBytes, err
}

func ExistedPath(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}

	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func CheckMakeDir(path string) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	}

	if err := os.MkdirAll(path, 0755); err != nil {
		return err
	}

	return nil
}

// host can be hostname, ipv4, ipv6
// output such as [::]:8000
func MakeAddrPort(host string, port int) (string, error) {
	if len(host) == 0 {
		return fmt.Sprintf(":%d", port), nil
	}
	if ip := net.ParseIP(host); ip != nil { // it's ip
		if strings.IndexByte(host, '.') != -1 { // ipv4
			return fmt.Sprintf("%s:%d", host, port), nil
		} else { // ipv6, with ':'
			return fmt.Sprintf("[%s]:%d", host, port), nil
		}

	} else { // it's hostname
		if _, err := net.LookupIP(host); err != nil { // wrong hostname
			return fmt.Sprintf("%s:%d", host, port), fmt.Errorf("invalid ip or hostname %s", host)
		} else {
			return fmt.Sprintf("%s:%d", host, port), nil
		}
	}
}

func GetLocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}
	for _, addr := range addrs {
		ipAddr, ok := addr.(*net.IPNet)
		if !ok {
			continue
		}
		if ipAddr.IP.IsLoopback() {
			continue
		}
		if !ipAddr.IP.IsGlobalUnicast() {
			continue
		}
		return ipAddr.IP.String(), nil
	}
	return "", fmt.Errorf("not found valid interface address")
}

func GetOutBoundIP4() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:53")
	if err != nil {
		return "", err
	}
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	ip := strings.Split(localAddr.String(), ":")[0]
	return ip, nil
}

// valid datetime is rfc3339 format or not. then convert utc to cst
func DatetimeIsvalid(s string) (string, bool) {
	var err error
	var t time.Time
	var ok = false
	if t, err = time.Parse(time.RFC3339Nano, s); err == nil {
		ok = true
	} else {
		if t, err = time.Parse(time.RFC3339, s); err == nil {
			ok = true
		} else {
			if t, err = time.Parse("2006-01-02 15:04:05", s); err == nil {
				s = t.Format(time.RFC3339)
				ok = true
			}
		}
	}

	// check time zone, if it is utc, convert it to cst
	// sometime return "UTC" and 0, When return "CST", should return 28800=8*3600 seconds
	zone, seconds := t.Zone()
	if zone != "CST" {
		timestamp := t.Add(-8*time.Hour + time.Second*time.Duration(seconds)) // (-8 * time.Hour)
		loc := time.FixedZone("CST", 8*3600)
		s = timestamp.In(loc).Format(time.RFC3339Nano)
	}
	return s, ok
}

func JsonPretty(b []byte) ([]byte, error) {
	var out bytes.Buffer
	err := json.Indent(&out, b, "", "  ")
	return out.Bytes(), err
}

// use library dateparse.ParseAny to parse any datetime format
func ParseAnyDatetime(s string) (time.Time, error) {
	t, err := dateparse.ParseAny(s)
	if err != nil {
		return t, err
	}

	if t.Year() == 0 {
		t0 := time.Now()
		t1 := time.Date(t0.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), 0, time.Local)
		if t1.After(t0) { // maybe t1 after t0, that means t1.year is wrong, should be t0.yeas-1
			if t1.Sub(t0) > 8*time.Hour {
				t1 = time.Date(t0.Year()-1, t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), 0, time.Local)
			}
		}
		return t1, nil
	} else {
		return t, nil
	}
}

func MD5(v []byte) string {
	//d := []byte(v)
	m := md5.New()
	m.Write(v)
	return hex.EncodeToString(m.Sum(nil))
}
