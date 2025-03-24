package logmgr

import (
	"LogManager/common"
	"bytes"
	"fmt"
	"os"
	"os/exec"
)

func RunShellCommand(name string, args ...string) (string, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := exec.Command(name, args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("%v: %s", err, stderr.String())
	}
	return stdout.String(), nil
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
