package transfer

import (
	"errors"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type State string

var ErrorFileIgnored error = errors.New("file is ignored")

type TransferClient struct {
	log        *log.Logger
	TargetIp   string
	TargetPort string
}

func (c *TransferClient) GetServerAddr() string {
	return c.TargetIp + ":" + c.TargetPort
}

func NewTransferClient(TargetIp string, TargetPort string) *TransferClient {
	return &TransferClient{
		log:        log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile),
		TargetIp:   TargetIp,
		TargetPort: TargetPort,
	}
}

func (c *TransferClient) Transfer(filepath string) error {
	// 0. 根据ignore file过滤文件
	err := c.checkIgnoreFile(filepath)
	if errors.Is(err, ErrorFileIgnored) {
		return nil
	}
	// 1. 验证文件存在
	fileInfo, err := os.Stat(filepath)
	if err != nil {
		c.log.Println("文件可能不存在或不可读: ", err)
		return err
	}
	// 2. 建立连接， 发送文件路径
	conn, err := net.Dial("tcp4", c.GetServerAddr())
	if err != nil {
		c.log.Println("net dial err: ", err)
		return err
	}
	defer conn.Close()
	event := NewTransferEvent(Evt_Create, filepath)
	event.IsDir = fileInfo.IsDir()
	_, err = conn.Write([]byte(event.String()))
	if err != nil {
		c.log.Println("Connection	write err: ", err)
		return err
	}
	// 3. 等待server确认接收
	buf := make([]byte, 1024)
	conn.SetReadDeadline(time.Now().Add(time.Second * 5)) // 5s超时
	n, err := conn.Read(buf)
	if err != nil {
		c.log.Println("Connection read err: ", err)
		return err
	}
	if string(buf[:n]) == "ok" {
		c.log.Println("server is ready to receive file")
	} else {
		c.log.Println("server is not ready to receive file")
		return err
	}
	// 3.0 发送目录/文件
	if fileInfo.IsDir() {
		// err = c.SendDir(conn, filepath)
		c.log.Println("send dir: ", filepath)
	} else {
		err = c.sendFile(conn, filepath)
	}
	// 4. server确认接收完成
	return err
}

func (c *TransferClient) checkIgnoreFile(filepath string) error {
	// viper.SetConfigFile("./config/config.yaml")
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	// 配置文件不存在
	if _, ok := err.(viper.ConfigFileNotFoundError); ok {
		log.Println("ignore file not exist")
		return nil
	}
	if err != nil {
		log.Println("err", err)
		return err
	}
	ignore := viper.GetStringSlice("ignore")
	log.Println("ignore: ", ignore)
	log.Println("file: ", filepath)
	for _, v := range ignore {
		if matchPath(v, filepath) {
			log.Println("此文件已更改但忽略上传: ", filepath)
			return ErrorFileIgnored
		}
	}
	return nil
}

func (c *TransferClient) sendFile(conn net.Conn, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		c.log.Println("open file err: ", err)
		return err
	}
	defer file.Close()
	for {
		buf := make([]byte, 1024)
		n, err := file.Read(buf)
		if err == io.EOF {
			c.log.Println("file read done")
			break
		} else if err != nil {
			c.log.Println("read file err: ", err)
			return err
		}
		_, err = conn.Write(buf[:n])
		if err != nil {
			c.log.Println("write err: ", err)
			return err
		}
	}
	return nil
}

func (c *TransferClient) RemoveFile(filename string) error {
	conn, err := net.Dial("tcp4", c.TargetIp)
	if err != nil {
		c.log.Println("net dial err: ", err)
		return err
	}
	defer conn.Close()
	event := NewTransferEvent(Evt_Remove, filename)
	// event.IsDir = false
	_, err = conn.Write([]byte(event.String()))
	return err
}

func (c *TransferClient) RemoveDir(filename string) error {
	conn, err := net.Dial("tcp4", c.TargetIp)
	if err != nil {
		c.log.Println("net dial err: ", err)
		return err
	}
	defer conn.Close()
	event := NewTransferEvent(Evt_RemoveDir, filename)
	event.IsDir = true
	_, err = conn.Write([]byte(event.String()))
	return err
}

func (c *TransferClient) CheckHash(filename string) error {
	c.log.Println("todo: 验证文件是否同步")
	return nil
}


// ---utils

func matchPath(origin string, target string) bool {
	var trimPath = func(s string) string {
		return strings.TrimRight(strings.TrimLeft(s, "./"), "/")
	}
	origin = trimPath(origin)
	target = trimPath(target)
	return origin == target || strings.HasPrefix(target, origin)
}