package comm

import (
	"fmt"
	"github.com/go-redis/redis"
	"github.com/evenzhang/gbusd/gserv/log"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func FatalPrintAndDie(args ...interface{}) {
	fmt.Println("Fatal Error", args)
	log.Fatalln(args)
}
func PrintAndExit(args ...interface{}) {
	fmt.Println(args)
	os.Exit(0)
}
func AssertErr(err error, args ...interface{}) {
	if err != nil {
		fmt.Println(args, err)
		log.Fatalln(args, err)
		os.Exit(0)
	}
}

func GetLocalAddr() string {
	addrs, err := net.InterfaceAddrs()

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	for _, address := range addrs {
		// 检查ip地址判断是否回环地址
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}

func GetCurrentDirectory() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		fmt.Println(err)
	}
	return strings.Replace(dir, "\\", "/", -1) + "/"
}

func NewRedisClient(addr string) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "",
		DB:       0,
	})
	_, err := client.Ping().Result()
	return client, err
}

func Unix(datetime string) int64 {
	local, _ := time.LoadLocation("Asia/Shanghai")
	date, _ := time.ParseInLocation("2006-01-02 15:04:05", "2017-10-16 17:38:21", local)
	return date.Unix()
}

func ReadFileContent(filePath string) (string, error) {
	fi, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer fi.Close()
	data, err := ioutil.ReadAll(fi)
	if err != nil {
		return "", err
	}

	return string([]byte(data)), nil
}
