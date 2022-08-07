package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"time"
)

func main() {
	var serverReady = make(chan bool)
	done := make(chan bool)
	// mock receiver
	go func() {
		listener, err := net.Listen("tcp4", ":8081")
		if err != nil {
			log.Fatalln(err)
		}
		defer listener.Close()
		serverReady <- true
		for {
			log.Println("ready to accept")
			conn, err := listener.Accept()
			if err != nil {
				log.Fatalln(err)
			}
			for {
				buf := make([]byte, 1024)
				n, err := conn.Read(buf)
				if err == io.EOF {
					log.Println("client closed")
					break
				} else if err != nil {
					log.Fatalln(err)
				}
				// conn.Write([]byte("ok"))
				log.Println("server received: ", string(buf[:n]))
				log.Println("server receive again: ", string(buf[:n]))
			}
		}
	}()
	// mock sender
	go func() {
		<-serverReady
		conn, err := net.Dial("tcp4", ":8081")
		if err != nil {
			log.Fatalln(err)
		}
		conn.SetReadDeadline(time.Now().Add(time.Second * 2))
		for i := 0; i < 5; i++ {
			time.Sleep(time.Second)
			_, err := conn.Write([]byte(fmt.Sprintf("hello-%d", i)))
			if err != nil {
				log.Fatalln(err)
			}
			responseBuf := make([]byte, 1024)
			log.Println("ready to read")
			n, err := conn.Read(responseBuf)
			if err != nil {
				log.Fatalln(err)
			}
			if string(responseBuf[:n]) == "ok" {
				log.Println("client received: ", string(responseBuf[:n]))
			}
			log.Printf("client send %d bytes", n)
		}
		conn.Close()
	}()
	<-done
}
