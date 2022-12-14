package cmd

import (
	filewatcher "liveput/src/fileWatcher"
	"liveput/src/transfer"
	"log"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "liveput",
	Short: "liveput is a tool for sync files in real-time",
	Long:  "liveput 通过tcp实时同步文件",
	Run: func(cmd *cobra.Command, args []string) {
		log.Println("hi, 笑一下【 ^_^ 今日笑脸 +1】")
	},
}

var (
	serverIp   string
	serverPort string
	clientCmd  = &cobra.Command{
		Use:   "client",
		Short: "watch files and sync to server",
		Long:  "client 监听文件, 并实时同步变化到server端",
		Aliases: []string{"c"},
		Run: func(cmd *cobra.Command, args []string) {
			// do stuff here
			client := transfer.NewTransferClient(serverIp, serverPort)
			fw := filewatcher.NewFileWatcher(rootWatchedDir, client)
			fw.Start()
		},
	}
	rootWatchedDir string
)

var (
	serverRootDir string
	serverCmd     = &cobra.Command{
		Use:   "server",
		Short: "sync files from client",
		Long:  "接收来自client端的文件",
		Aliases: []string{"s"},
		Run: func(cmd *cobra.Command, args []string) {
			server := transfer.NewTransferServer(serverRootDir)
			server.Start()
			// transfer.NewTransferServer("server-target/").Start()
		},
	}
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Println("execute err: ", err)
		os.Exit(1)
	}
}

func init() {
	clientCmd.PersistentFlags().StringVarP(&rootWatchedDir, "watch", "w", "", "要监听的相对路径(当前目录)或绝对路径")
	clientCmd.PersistentFlags().StringVarP(&serverIp, "ip", "i", "", "服务端ip地址(ipv4), 缺省则默认本机")
	clientCmd.PersistentFlags().StringVarP(&serverPort, "port", "p", "8080", "服务端端口号")

	serverCmd.PersistentFlags().StringVarP(&serverRootDir, "path", "p", "", "服务端存储的目标目录, 相对当前位置的相对路径或绝对路径")

	rootCmd.AddCommand(clientCmd)
	rootCmd.AddCommand(serverCmd)
}
