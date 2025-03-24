package command

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	mgr "LogManager/LogMgr"

	"github.com/spf13/cobra"
)

func LogManagerCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "log",
		Short: "log manager",
		Long:  "log manager",
		Run: func(cmd *cobra.Command, args []string) {
			// 创建一个channel用于接收信号
			sigChan := make(chan os.Signal, 1)
			// 注册要处理的信号
			signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

			fmt.Println("LogManager is running. Press Ctrl+C to exit.")
			// 在这里执行你的程序逻辑
			mgr.Run()

			// 阻塞等待信号
			sig := <-sigChan
			fmt.Printf("\nReceived signal %v, shutting down...\n", sig)

			// 停止定时任务管理器
			mgr.Stop()
		},
	}
}
