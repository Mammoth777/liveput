package transfer

import (
	"errors"
	"io"
	"log"
	"net"
	"os"
	"path"
	"path/filepath"
)

type TransferServer struct {
	log       *log.Logger
	Port      string
	Network   string
	TargetDir string
}

func NewTransferServer(TargetDir string) *TransferServer {
	return &TransferServer{
		log:       log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile),
		Network:   "tcp4",
		Port:      ":8080",
		TargetDir: TargetDir,
	}
}

func (s *TransferServer) Start() error {
	s.log.Println("tcp server start, listening on ", s.Port)
	listener, err := net.Listen(s.Network, s.Port)
	if err != nil {
		return err
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			s.log.Println("accept err: ", err)
			continue
		}
		go func() {
			defer func() {
				if err := recover(); err != nil {
					log.Println("server err: ", err)
				}
			}()
			s.handleConn(conn)
		}()
	}
}

// 处理连接
// 1. 接收文件名， 并返回ok
// 2. 接收全部文件，并存储在本地
func (s *TransferServer) handleConn(conn net.Conn) {
	defer func() {
		conn.Close()
		s.log.Println("server connection closed")
	}()
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		s.log.Println("read err: ", err)
		return
	}
	evtString := string(buf[:n])
	event, err := ParseEvent(evtString)
	if err != nil {
		s.log.Println("parse transfer event err: ", err)
	}
	s.log.Println("server receive file name: ", event.FileName)
	// 1. 接收文件名， 并返回ok
	conn.Write([]byte("ok"))
	// 2. 接收全部文件，并存储在本地
	// 2.1 创建或修改文件
	if event.EventType == Evt_Create {
		if event.IsDir {
			err := os.MkdirAll(path.Join(s.TargetDir, event.FileName), os.ModePerm)
			if err != nil {
				s.log.Println("create dir err: ", err)
			}
			return
		}
		target := path.Join(s.TargetDir, event.FileName)
		file, err := CreateFile(target)
		if err != nil {
			s.log.Println("create file err: ", err)
			return
		}
		defer file.Close()
		for {
			buf := make([]byte, 1024)
			n, err := conn.Read(buf)
			if err == io.EOF {
				s.log.Println("server receive end")
				break
			} else if err != nil {
				s.log.Println("read err: ", err)
				break
			}
			s.log.Println("server received: ", string(buf[:n]))
			_, err = file.Write(buf[:n])
			if err != nil {
				s.log.Println("write err: ", err)
				break
			}
		}
	} else {
		// 2.2 删除文件
		err := os.RemoveAll(path.Join(s.TargetDir, event.FileName))
		if err != nil {
			s.log.Println("remove file err: ", err)
		}
	}
}

func CreateFile(filename string) (*os.File, error) {
	if filename == "" {
		return nil, errors.New("filename is empty")
	}
	// 创建父目录
	pDir := filepath.Dir(filename)
	if pDir != "." {
		_, err := os.Stat(pDir)
		if err != nil {
			if os.IsNotExist(err) { // 父目录不存在
				err = os.MkdirAll(pDir, os.ModePerm)
				if err != nil {
					return nil, err
				}
				return os.Create(filename)
			} else {
				return nil, err
			}
		}
	}

	return os.Create(filename)
}
