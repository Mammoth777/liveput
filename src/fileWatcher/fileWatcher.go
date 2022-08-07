package filewatcher

import (
	"io/fs"
	"liveput/src/transfer"
	"log"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

type FileWatcher struct {
	log          *log.Logger
	RootFileName string
	watcher      *fsnotify.Watcher
	client       *transfer.TransferClient
}

func NewFileWatcher(RootFileName string, client *transfer.TransferClient) *FileWatcher {
	return &FileWatcher{
		RootFileName: RootFileName,
		log:          log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile),
		client:       client,
	}
}

func (f *FileWatcher) Handler(op fsnotify.Op, fileName string) {
	f.log.Println("op:", op)
	var err error
	switch true {
	case fsnotify.Create == op || fsnotify.Write == op:
		f.log.Println("create or write: " + fileName)
		err = f.client.Transfer(fileName)
		f.HotAdd(fileName)
	case fsnotify.Rename == op || fsnotify.Remove == op:
		f.log.Println("rename or remove: " + fileName)
		err = f.client.RemoveFile(fileName)
		f.Remove(fileName)
	case fsnotify.Remove|fsnotify.Rename == op:
		f.log.Println("remove or rename DIR: " + fileName)
		err = f.client.RemoveDir(fileName)
		f.Remove(fileName)
	default:
		f.log.Println("unknown op: " + op.String())
	}
	if err != nil {
		f.log.Println("handler error: ", err)
	}
}

func (f *FileWatcher) Add(path string) error {
	return filepath.WalkDir(path, func(path string, dir fs.DirEntry, err error) error {
		if err != nil {
			f.log.Println("filepath.WalkDir error: ", err)
			return err
		}
		err = f.watcher.Add(path)
		if err != nil {
			f.log.Println("watcher.Add error: ", err)
			return err
		}
		return nil
	})
}

func (f *FileWatcher) HotAdd(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		f.log.Println("os.Stat error: ", err)
		return err
	}
	if info.IsDir() {
		err = f.Add(path)
		if err != nil {
			f.log.Println("watcher.Add error: ", err)
			return err
		}
	}
	return err
}

func (f *FileWatcher) Remove(path string) error {
	return f.watcher.Remove(path)
}

func (f *FileWatcher) Start() {
	f.log.Println("file watcher is running")
	var err error
	f.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}
	defer f.watcher.Close()
	done := make(chan bool)

	go func() {
		defer close(done)
		for {
			select {
			case event, ok := <-f.watcher.Events:
				if !ok {
					return
				}
				f.Handler(event.Op, event.Name)
			case err, ok := <-f.watcher.Errors:
				if !ok {
					return
				}
				f.log.Println(err)
			}
		}
	}()

	_ = f.client.CheckHash(f.RootFileName)
	err = f.Add(f.RootFileName)
	if err != nil {
		log.Fatalln(err)
	}
	<-done
}

func (f *FileWatcher) Stop() {
	f.log.Println("todo: file watcher stop")
}
