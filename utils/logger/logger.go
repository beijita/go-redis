package logger

import (
	"fmt"
	"go-redis/common"
	"go-redis/utils/file"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

type Settings struct {
	Path       string `yml:"path"`
	Name       string `yml:"name"`
	Ext        string `yml:"ext"`
	TimeFormat string `yml:"time-format"`
}

var (
	logFile            *os.File
	defaultPrefix      = ""
	defaultCallerDepth = 2
	logger             *log.Logger
	mu                 sync.Mutex
	logPrefix          = ""
	levelFlags         = []string{"DEBUG", "INFO", "WARN", "ERROR", "FATAL"}
)

const flags = log.LstdFlags

func init() {
	logger = log.New(os.Stdout, defaultPrefix, flags)
}

func Setup(settings *Settings) {
	var err error
	dir := settings.Path
	filename := fmt.Sprintf("%s-%s.%s", settings.Name, time.Now().Format(settings.TimeFormat), settings.Ext)
	logFile, err := file.MustOpen(filename, dir)
	if err != nil {
		log.Fatalf("logging.Setup err:%v", err)
	}
	writer := io.MultiWriter(os.Stdout, logFile)
	logger = log.New(writer, defaultPrefix, flags)
}

func setPrefix(logLevel common.LogLevel) {
	_, file, line, ok := runtime.Caller(defaultCallerDepth)
	if ok {
		logPrefix = fmt.Sprintf("[%s][%s:%d] ", levelFlags[logLevel], filepath.Base(file), line)
	} else {
		logPrefix = fmt.Sprintf("[%s] ", levelFlags[logLevel])
	}

	logger.SetPrefix(logPrefix)
}

func Debug(v ...interface{}) {
	mu.Lock()
	defer mu.Unlock()
	setPrefix(common.DEBUG)
	logger.Println(v)
}

func Info(v ...interface{}) {
	mu.Lock()
	defer mu.Unlock()
	setPrefix(common.INFO)
	logger.Println(v)
}

func Warn(v ...interface{}) {
	mu.Lock()
	defer mu.Unlock()
	setPrefix(common.WARN)
	logger.Println(v)
}

func Error(v ...interface{}) {
	mu.Lock()
	defer mu.Unlock()
	setPrefix(common.ERROR)
	logger.Println(v)
}

func Fatal(v ...interface{}) {
	mu.Lock()
	defer mu.Unlock()
	setPrefix(common.FATAL)
	logger.Println(v)
}
