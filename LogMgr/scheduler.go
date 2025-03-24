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

// checkDirectorySize 检查目录大小并返回大小（GB）
func checkDirectorySize(dirPath string) (float64, error) {
	output, err := RunShellCommand("du", "-s", dirPath)
	if err != nil {
		return 0, fmt.Errorf("failed to check directory size: %v", err)
	}

	var size int64
	_, err = fmt.Sscanf(output, "%d", &size)
	if err != nil {
		return 0, fmt.Errorf("failed to parse directory size: %v", err)
	}

	return float64(size) / (1024 * 1024), nil
}

// cleanDirectory 清理指定目录中超过指定天数的文件
func cleanDirectory(dirPath string, days string) error {
	_, err := RunShellCommand("find", dirPath, "-type", "d", "-mtime", "+"+days, "-exec", "rm", "-rf", "{}", "+")
	return err
}

// processApp 处理单个应用的日志
func processApp(app common.App) {
	// 检查目标目录大小
	sizeGB, err := checkDirectorySize(common.CONFIG.Global.TargetDir)
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}

	// 检查是否超过阈值
	if sizeGB >= float64(common.CONFIG.Global.MaxSize) {
		fmt.Printf("Target directory size (%.2f GB) exceeds the limit (%d GB)\n", sizeGB, common.CONFIG.Global.MaxSize)
		if !common.CONFIG.Global.CleanAuto {
			fmt.Println("Clean auto is false, please clean the target directory manually")
			return
		}

		// 首先尝试清理超过MaxSaveDuration天数的文件
		if err := cleanDirectory(common.CONFIG.Global.TargetDir, common.CONFIG.Global.MaxSaveDuration); err != nil {
			fmt.Printf("Failed to clean target directory: %v\n", err)
			return
		}

		// 检查清理后的目录大小
		sizeGB, err = checkDirectorySize(common.CONFIG.Global.TargetDir)
		if err != nil {
			fmt.Printf("%v\n", err)
			return
		}
		fmt.Printf("Target directory size (%.2f GB) after initial clean\n", sizeGB)

		// 如果仍然超过阈值，尝试清理超过MinSaveDuration天数的文件
		if sizeGB >= float64(common.CONFIG.Global.MaxSize) {
			fmt.Printf("Target directory size still exceeds the limit, performing deeper clean...\n")
			if err := cleanDirectory(common.CONFIG.Global.TargetDir, common.CONFIG.Global.MinSaveDuration); err != nil {
				fmt.Printf("Failed to perform deeper clean: %v\n", err)
				return
			}
		}
		fmt.Println("Clean target directory successfully")
		return
	}

	// 创建新的日志目录
	DateTimeStr := time.Now().Format("2006-01-02_15-04-05")
	logDir := fmt.Sprintf("%s/%s/%s", common.CONFIG.Global.TargetDir, DateTimeStr, app.Name)

	if err := os.MkdirAll(logDir, 0755); err != nil {
		fmt.Printf("Failed to create log directory for app %s: %v\n", app.Name, err)
		return
	}

	// 拷贝日志文件
	for _, logFile := range app.LogFiles {
		logFilePath := fmt.Sprintf("%s/%s", app.LogDir, logFile)
		if _, err := RunShellCommand("cp", "-r", logFilePath, logDir); err != nil {
			fmt.Printf("Failed to copy log file for app %s: %v\n", app.Name, err)
			return
		}
		if app.EmptyOrigin {
			// 使用更安全的方式清空原始日志文件
			if _, err := RunShellCommand("truncate", "-s", "0", logFilePath); err != nil {
				fmt.Printf("Failed to empty log file %s: %v\n", logFilePath, err)
			}
		}
	}
}
