package main

import (
	"fmt"
	"os"
	"sync"

	"github.com/DesistDaydream/tar/pkg/handler"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
)

type tarFlags struct {
	logLevel   string
	logFile    string
	logFormat  string
	startAt    string
	count      int
	goroutines int
}

func (flags *tarFlags) TarFlags() {
	pflag.StringVar(&flags.logLevel, "log-level", "info", "日志级别:{debug, info, warn, error, fatal}")
	pflag.StringVar(&flags.logFile, "log-output", "", "日志输出位置, 默认标准输出")
	pflag.StringVar(&flags.logFormat, "log-format", "text", "日志格式:{json|text}")
	pflag.StringVar(&flags.startAt, "start-at", "", "指定要开始归档的目录")
	pflag.IntVar(&flags.count, "count", 10, "归档目录数量")
	pflag.IntVar(&flags.goroutines, "goroutines", 1, "并发数量")
}

// LogInit 日志功能初始化，若指定了 log-output 命令行标志，则将日志写入到文件中
func LogInit(level, file, format string) error {
	switch format {
	case "json":
		logrus.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat:   "2006-01-02 15:04:05",
			DisableTimestamp:  false,
			DisableHTMLEscape: false,
			DataKey:           "",
			// FieldMap:          map[logrus.fieldKey]string{},
			// CallerPrettyfier: func(*runtime.Frame) (string, string) {},
			PrettyPrint: false,
		})
	case "text":
		logrus.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
		})
	default:
		return fmt.Errorf("请指定正确的日志格式")
	}

	logLevel, err := logrus.ParseLevel(level)
	if err != nil {
		return err
	}
	logrus.SetLevel(logLevel)

	if file != "" {
		f, err := os.OpenFile(file, os.O_WRONLY|os.O_CREATE, 0755)
		if err != nil {
			return err
		}
		logrus.SetOutput(f)
	}

	return nil
}

var (
	// 程序所在的工作目录
	workPath string
)

func main() {
	// 设置命令行标志
	tarFlags := &tarFlags{}
	tarFlags.TarFlags()
	thFlags := &handler.TarHandlerFlags{}
	thFlags.AddFlag()

	pflag.Parse()

	// 初始化日志
	if err := LogInit(tarFlags.logLevel, tarFlags.logFile, tarFlags.logFormat); err != nil {
		logrus.Trace(errors.Wrap(err, "set log level error"))
	}

	// 处理各种目录的绝对路径变量
	// 工作目录的绝对路径
	workPath, _ = os.Getwd()
	// 归档源目录的绝对路径
	archiveSrcPath := workPath + string(os.PathSeparator) + thFlags.ArchiveSrc
	// 归档目目录的绝对路径
	archiveDestPath := workPath + string(os.PathSeparator) + thFlags.ArchiveDest
	logrus.WithFields(logrus.Fields{
		"工作路径":    workPath,
		"归档源,根路径": archiveSrcPath,
		"归档目,根路径": archiveDestPath,
	}).Info("运行前检查绝对路径")

	// 并发
	var wg sync.WaitGroup
	defer wg.Wait()
	// 控制并发
	concurrenceControl := make(chan bool, tarFlags.goroutines)

	// 用来判断是否开始循环的变量
	var count int = 1
	var isStart bool = false
	// 获取归档源目录下的文件列表
	dateDirFiles, err := os.ReadDir(archiveSrcPath)
	if err != nil {
		panic(fmt.Sprintf("获取日期目录 %s 的列表失败: %s", archiveSrcPath, err))
	}
	for _, dateFile := range dateDirFiles {
		// 控制并发
		concurrenceControl <- true
		// 并发
		wg.Add(1)

		dateDirName := dateFile.Name()

		// 判断是否开始循环的条件
		// 条件1：当前目录名称必须为指定的名称
		// 条件2：指定开始的目录名称不能为空
		// 条件3：循环必须已经开始
		// 当上述三个条件都成立，才会开始循环归档，否则跳过
		if dateDirName != tarFlags.startAt && tarFlags.startAt != "" && !isStart {
			logrus.Infof("跳过 %v 目录", dateDirName)
			continue
		}
		isStart = true

		// 计数器，循环指定次数即停止
		if count > tarFlags.count {
			break
		}

		// 日期目录的绝对路径
		datePath := fmt.Sprintf("%s%s%s", archiveSrcPath, string(os.PathSeparator), dateDirName)
		// 创建归档目标目录
		archiveDestDatePath := fmt.Sprintf("%s%s%s", archiveDestPath, string(os.PathSeparator), dateDirName)
		err = os.MkdirAll(archiveDestDatePath, 0775)
		if err != nil {
			panic(fmt.Sprintf("创建归档目标目录出错: %s", err))
		}

		logrus.WithFields(logrus.Fields{
			"归档源,日期名称": dateDirName,
			"归档源,日期路径": datePath,
			"归档目,日期路径": archiveDestDatePath,
		}).Trace("检查日期信息")

		go func(datePath, dateDirName, archiveDestPath string) {
			// 并发
			defer wg.Done()

			// 变更目录
			// go 并发与 os.Chdir 冲突，详见：https://github.com/golang/go/issues/27658
			err = os.Chdir(datePath)
			if err != nil {
				panic(fmt.Sprintf("切换目录出错: %s", err))
			}

			// 获取日期目录下的姓名列表
			nameDirFiles, err := os.ReadDir(datePath)
			if err != nil {
				panic(fmt.Sprintf("获取姓名目录 %s 的列表失败: %s", datePath, err))

			}
			for _, nameFile := range nameDirFiles {
				nameDirName := nameFile.Name()

				cwd, _ := os.Getwd()
				logrus.Debugf("当前在 %v 目录下操作 %v 用户目录\n", cwd, nameDirName)

				// 姓名目录的绝对路径
				namePath := fmt.Sprintf("%s%s%s", datePath, string(os.PathSeparator), nameDirName)

				// 归档目标文件名
				archiveDestPathFile := fmt.Sprintf("%s%s%s.%s", archiveDestDatePath, string(os.PathSeparator), nameDirName, thFlags.Extension)
				logrus.WithFields(logrus.Fields{
					"归档源,姓名名称": nameDirName,
					"归档源,姓名路径": namePath,
					"归档目,文件路径": archiveDestPathFile,
				}).Trace("检查姓名信息")

				// 开始归档
				// 第一个参数有两种选择
				// 1. 使用完整路径，那么归档后的文件中，包含所有路径上的目录
				// err = handler.Archiving(namePath, archiveDestPathFile, thFlags.Extension)
				// 2. 使用文件名称，那么归档后的文件中，只包含文件名目录。注意：只使用文件名的话，需要切换目录，切换目录又与go并发有冲突
				err = handler.Run(nameDirName, archiveDestPathFile, thFlags.Extension)
				if err != nil {
					logrus.Errorf("%s/%s 归档失败: %s", dateDirName, nameDirName, err)
				} else {
					logrus.Infof("%s/%s 归档成功", dateDirName, nameDirName)
				}
			}

			// 控制并发
			<-concurrenceControl
		}(datePath, dateDirName, archiveDestPath)

		count++
	}

	// 待提取的文件
	// var extractSrc = "./file_handle/tar_dir/test_tar.tar.gz"
	// // 提取后保存的路径，不写就是解压到当前目录
	// var extractDst = "./file_handle/tar_dir/"

	// Extracting(extractDst, extractSrc)
}
