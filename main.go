package main

import (
	"ddns/ali"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// 配置文件路径
var config_path = filepath.Join(os.Getenv("HOME"), ".ddns_config.toml")
var record_path = filepath.Join(os.Getenv("HOME"), ".ddns_config_record.toml")

func main() {
	cfg := NewConfig()
	rec := NewRecord()

	nowIpv6 := get_ipv6()
	if nowIpv6 == "" {
		logf("获取本机ipv6地址失败")
		return
	}
	logf("成功获取ipv6地址成功, %s", nowIpv6)

	if rec.LastIpv6 == nowIpv6 {
		logf("ipv6地址未发生改变, 无需操作 >>>")
		return
	}

	// 获取, rr.domain 对应的recordId, 如果没有这条记录则创建一条
	recordId, err := rec.GetOrCreateRecordId(cfg)
	if err != nil {
		logf("获取record id 失败")
		return
	}
	logf("成功获取记录id, recordId = %s", recordId)

	err = ali.UpdateRecord(cfg.RR, recordId, nowIpv6, cfg.AccId, cfg.AccKey)
	if err != nil {
		logf("更新失败, %+v", err)
		return
	}
	rec.Write("last-ipv6", nowIpv6)
	logf("已更新ip地址, new Ipv6 = %s", nowIpv6)
}

type Record struct {
	v        *viper.Viper
	RecordId string
	LastIpv6 string
}

func NewRecord() *Record {
	v := viper.New()
	v.SetConfigFile(record_path)
	if _, err := os.Stat(record_path); os.IsNotExist(err) {
		v.WriteConfig()
	}
	v.ReadInConfig()
	return &Record{
		v:        v,
		RecordId: v.GetString("record-id"),
		LastIpv6: v.GetString("last-ipv6"),
	}
}

func (r *Record) Write(k string, v interface{}) {
	r.v.Set(k, v)
	r.v.WriteConfig()
}

func (r *Record) GetOrCreateRecordId(c *Config) (string, error) {
	id := r.v.GetString("record-id")
	if id == "" {
		recordId, err := ali.GetRecordsId(c.RR, c.Domain, c.AccId, c.AccKey)
		if err != nil {
			return "", err
		}
		if recordId != "" {
			r.Write("record-id", recordId)
			return recordId, nil
		} else {
			// 添加一个密匙
			recordId, err := ali.AddRecord(c.RR, c.Domain, "a:0:0:0:0:0:0:0", c.AccId, c.AccKey)
			if err != nil {
				return "", err
			} else {
				r.Write("record-id", recordId)
				return recordId, nil
			}
		}
	}
	return id, nil
}

type Config struct {
	RR     string
	Domain string
	AccId  string
	AccKey string
}

func NewConfig() *Config {
	if _, err := os.Stat(config_path); os.IsNotExist(err) {
		file, err := os.Create(config_path)
		if err != nil {
			panic("创建配置文件失败")
		}
		defer file.Close()
		temp := `# 阿里云的access key id
access-key-id = ""

# 阿里云的access key secret
access-key-secret = ""

# 阿里云注册的主域名
domain = ""

# 想要同步的主机域名的子域名 @代表空
rr = ""
`
		if n, err := file.WriteString(temp); err != nil || n <= 0 {
			panic("文件写入失败")
		}
		fmt.Println(" >> 检测到不存在配置文件,已创建该文件，使用：\n   vim ~/.ddns_config.toml \n 编辑配置文件后重新启动即可")
		os.Exit(0)
	}
	v := viper.New()
	v.SetConfigFile(config_path)
	v.ReadInConfig()
	rr := v.GetString("rr")
	domain := v.GetString("domain")
	id := v.GetString("access-key-id")
	key := v.GetString("access-key-secret")
	if rr == "" || domain == "" || id == "" || key == "" {
		logf("配置信息不完整")
		os.Exit(0)
	}
	return &Config{
		RR:     rr,
		Domain: domain,
		AccId:  id,
		AccKey: key,
	}
}

func logf(str string, vs ...interface{}) {
	t := time.Now()
	s := fmt.Sprintf("[%v]>> ", t)
	fmt.Printf(s+str+"\n", vs...)
}

func get_ipv6() string {
	s, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, a := range s {
		i := regexp.MustCompile(`(\w+:){7}\w+`).FindString(a.String())
		// 24开头的ipv6为公网ip
		if strings.Count(i, ":") == 7 && i[:2] == "24" {
			return i
		}
	}
	return ""
}
