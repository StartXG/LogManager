package logmgr

import (
	"LogManager/common"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
)

type Scheduler struct {
	cron *cron.Cron
	mu   sync.Mutex
}

var scheduler *Scheduler

func InitScheduler() {
	scheduler = &Scheduler{
		cron: cron.New(cron.WithSeconds()),
	}
	scheduler.cron.Start()
}

func StopScheduler() {
	if scheduler != nil && scheduler.cron != nil {
		scheduler.cron.Stop()
	}
}

func (s *Scheduler) UpdateJobs() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 停止所有现有任务
	s.cron.Stop()
	s.cron = cron.New(cron.WithSeconds())

	// 为每个应用创建定时任务
	for _, app := range common.CONFIG.Apps {
		// 创建一个新的作用域来捕获app变量
		app := app

		// 解析时区
		loc, err := time.LoadLocation(app.ExecTime.TimeZone)
		if err != nil {
			fmt.Printf("Failed to load timezone for app %s: %v\n", app.Name, err)
			continue
		}

		// 解析开始时间
		startTime, err := time.ParseInLocation("15:04", app.ExecTime.StartTime, loc)
		if err != nil {
			fmt.Printf("Failed to parse start time for app %s: %v\n", app.Name, err)
			continue
		}

		// 创建cron表达式 (每天在指定时间执行)
		cronSpec := fmt.Sprintf("%d %d %d * * *", startTime.Second(), startTime.Minute(), startTime.Hour())

		// 添加任务
		_, err = s.cron.AddFunc(cronSpec, func() {
			fmt.Printf("Starting scheduled task for app: %s\n", app.Name)
			processApp(app)
		})

		if err != nil {
			fmt.Printf("Failed to schedule task for app %s: %v\n", app.Name, err)
			continue
		}
	}

	// 启动新的定时任务
	s.cron.Start()
}

// processApp 处理单个应用的日志
func processApp(app common.App) {
	// 检查目标目录大小
	output, err := RunShellCommand("du", "-s", common.CONFIG.Global.TargetDir)
	if err != nil {
		fmt.Printf("Failed to check target directory size: %v\n", err)
		return
	}

	// 解析du命令输出，获取目录大小（单位：KB）
	var size int64
	_, err = fmt.Sscanf(output, "%d", &size)
	if err != nil {
		fmt.Printf("Failed to parse directory size: %v\n", err)
		return
	}

	// 将KB转换为GB
	sizeGB := float64(size) / (1024 * 1024)

	// 检查是否超过阈值
	if sizeGB >= float64(common.CONFIG.Global.MaxSize) {
		fmt.Printf("Target directory size (%.2f GB) exceeds the limit (%d GB)\n", sizeGB, common.CONFIG.Global.MaxSize)
		if common.CONFIG.Global.CleanAuto {
			// 清理指定天数前的日志
			if _, err := RunShellCommand("find", common.CONFIG.Global.TargetDir, "-type", "d", "-mtime", "+"+common.CONFIG.Global.MaxSaveDuration, "-exec", "rm", "-rf", "{}", "+"); err != nil {
				fmt.Printf("Failed to clean target directory: %v\n", err)
				return
			}
			output, err := RunShellCommand("du", "-s", common.CONFIG.Global.TargetDir)
			if err!= nil {
				fmt.Printf("Failed to check target directory size: %v\n", err)
				return
			}
			_, err = fmt.Sscanf(output, "%d", &size)
			if err!= nil {
				fmt.Printf("Failed to parse directory size: %v\n", err)
				return
			}
			sizeGB = float64(size) / (1024 * 1024)
			fmt.Printf("Target directory size (%.2f GB) after clean\n", sizeGB)
			if sizeGB >= float64(common.CONFIG.Global.MaxSize) {
				fmt.Printf("Target directory size (%.2f GB) exceeds the limit (%d GB)\n", sizeGB, common.CONFIG.Global.MaxSize)
				if _, err := RunShellCommand("find", common.CONFIG.Global.TargetDir, "-type", "d", "-mtime", "+"+common.CONFIG.Global.MinSaveDuration, "-exec", "rm", "-rf", "{}", "+"); err!= nil {
					fmt.Printf("Failed to clean target directory: %v\n", err)
					return
				}
			}else{
				fmt.Println("Clean target directory successfully")
				return
			}
			fmt.Println("Clean target directory successfully")
			return
		} else {
			fmt.Println("Clean auto is false, please clean the target directory manually")
			return
		}
	}

	DateTimeStr := time.Now().Format("2006-01-02_15-04-05")
	logDir := fmt.Sprintf("%s/%s/%s", common.CONFIG.Global.TargetDir, DateTimeStr, app.Name)

	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		if err := os.MkdirAll(logDir, 0755); err != nil {
			fmt.Printf("Failed to create log directory for app %s: %v\n", app.Name, err)
			return
		}
	}

	for _, logFile := range app.LogFiles {
		logFilePath := fmt.Sprintf("%s/%s", app.LogDir, logFile)
		// 拷贝日志文件
		if _, err := RunShellCommand("cp", "-r", logFilePath, logDir); err != nil {
			fmt.Printf("Failed to copy log file for app %s: %v\n", app.Name, err)
			return
		}
		if app.EmptyOrigin {
			RunShellCommand("echo", "' '", ">", logFilePath)
		}
	}
}
