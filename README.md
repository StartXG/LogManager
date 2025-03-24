# LogManager 日志管理系统

一个高效的日志管理系统，支持多应用日志的定时压缩、归档和清理。

## 功能特点

- [x] 支持多应用日志管理
- [] 定时执行日志压缩
- [] 自动清理过期日志
- [x] 支持配置热重载
- [x] 灵活的日志目录配置
- [] 可设置存储空间上限

## 配置说明

配置文件位于 `config/config.yaml`，支持以下配置项：

### 全局配置

```yaml
global:
  # 压缩后的日志存储目录
  target_dir: "/path/to/target"
  # 目标目录最大容量（单位：GB）
  max_size: 2
  # 日志最大保存时间（单位：天）
  max_save_duration: 10
```

### 应用配置

```yaml
apps:
  # 应用名称
  - name: "app_name"
    # 应用日志目录
    log_dir: "/path/to/logs"
    # 需要压缩的日志文件列表
    log_files: 
      - "app.log"
    # 是否清空原始日志
    empty_origin: true
    # 执行时间配置
    exec_time:
      # 时区
      time_zone: "Asia/Shanghai"
      # 开始时间
      start_time: "00:01"
```

## 使用方法

1. 修改配置文件 `config/config.yaml`
2. 运行程序：
   ```bash
   go run cmd/main.go
   ```

## 注意事项

1. 确保程序对日志目录和目标目录有读写权限
2. 建议将 `target_dir` 设置在空间充足的磁盘分区
3. 合理设置 `max_size` 和 `max_save_duration` 避免磁盘空间耗尽
4. 建议将日志压缩任务设置在系统负载较低的时间段

## 开发环境

- Go `1.23+`
- 支持的操作系统：Linux, macOS