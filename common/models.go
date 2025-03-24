package common

type Config struct {
	Global Global `yaml:"global"`
	Apps   []App  `yaml:"apps"`
}

type Global struct {
	TargetDir       string `yaml:"target_dir"`
	MaxSize         int64  `yaml:"max_size"`
	MaxSaveDuration string `yaml:"max_save_duration"`
	MinSaveDuration string `yaml:"min_save_duration"`
	CleanAuto       bool   `yaml:"clean_auto"`
}

type ExecTime struct {
	TimeZone  string `yaml:"time_zone"`
	StartTime string `yaml:"start_time"`
}

type App struct {
	Name        string   `yaml:"name"`
	LogDir      string   `yaml:"log_dir"`
	LogFiles    []string `yaml:"log_files"`
	EmptyOrigin bool     `yaml:"empty_origin"`
	ExecTime    ExecTime `yaml:"exec_time"`
}
