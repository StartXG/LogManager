package common

import (
	"log"
	"os"
	"sync"
	"time"

	"gopkg.in/fsnotify.v1"
	"gopkg.in/yaml.v3"
)

var (
	configLock sync.RWMutex
)

// InitConfigManager 初始化配置管理器
func InitConfigManager() error {
	// 首次加载配置
	if err := reloadConfig(); err != nil {
		return err
	}

	// 启动配置文件监控
	go watchConfig()

	return nil
}

// reloadConfig 重新加载配置
func reloadConfig() error {
	fileByte, err := os.ReadFile("./config/config.yaml")
	if err != nil {
		return err
	}

	var newConfig Config
	if err := yaml.Unmarshal(fileByte, &newConfig); err != nil {
		return err
	}

	configLock.Lock()
	CONFIG = &newConfig
	configLock.Unlock()

	return nil
}

// watchConfig 监控配置文件变化
func watchConfig() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Printf("Failed to create config watcher: %v", err)
		return
	}
	defer watcher.Close()

	configPath := "./config/config.yaml"
	if err = watcher.Add(configPath); err != nil {
		log.Printf("Failed to add config file to watcher: %v", err)
		return
	}

	debounceTimer := time.NewTimer(0)
	<-debounceTimer.C // 消耗初始事件

	for {
		select {
		case event := <-watcher.Events:
			// 检查文件是否存在
			if _, err := os.Stat(configPath); os.IsNotExist(err) {
				log.Printf("Config file does not exist: %v", err)
				continue
			}

			// 处理文件变更事件
			if event.Op&(fsnotify.Write|fsnotify.Create) != 0 {
				debounceTimer.Reset(100 * time.Millisecond)
				go func() {
					<-debounceTimer.C
					if err := reloadConfig(); err != nil {
						log.Printf("Failed to reload config: %v", err)
					} else {
						log.Println("Config reloaded successfully")
					}
				}()
			}
		case err := <-watcher.Errors:
			log.Printf("Config watcher error: %v", err)
		}
	}
}

// GetConfig 获取配置的线程安全方法
func GetConfig() *Config {
	configLock.RLock()
	defer configLock.RUnlock()
	return CONFIG
}
