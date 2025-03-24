package logmgr

import (
	"LogManager/common"
	"fmt"
	"os"
	"os/exec"
)

func RunShellCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...) // 创建命令
	cmd.Stdout = os.Stdout             // 标准输出
	cmd.Stderr = os.Stderr             // 标准错误输出
	return cmd.Run()                   // 执行命令
}

func Run() {
	// 初始化定时任务管理器
	InitScheduler()

	// 确保目标目录存在
	if _, err := os.Stat(common.CONFIG.Global.TargetDir); os.IsNotExist(err) {
		if err := os.MkdirAll(common.CONFIG.Global.TargetDir, 0755); err != nil {
			fmt.Println("Failed to create target directory:", err)
			return
		}
	}

	// 更新定时任务
	scheduler.UpdateJobs()
}

func Stop() {
	// 停止定时任务管理器
	StopScheduler()
}
