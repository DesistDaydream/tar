package main

import (
	"fmt"
	"os"

	"github.com/DesistDaydream/tar/pkg/handler"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
)

type tarFlags struct {
	logLevel  string
	logFile   string
	logFormat string
	startAt   string
	count     int
}

func (flags *tarFlags) TarFlags() {
	pflag.StringVar(&flags.logLevel, "log-level", "info", "日志级别:{debug, info, warn, error, fatal}")
	pflag.StringVar(&flags.logFile, "log-output", "", "日志输出位置, 默认标准输出")
	pflag.StringVar(&flags.logFormat, "log-format", "text", "日志格式:{json|text}")
	pflag.StringVar(&flags.startAt, "start-at", "", "指定要开始归档的目录")
	pflag.IntVar(&flags.count, "count", 10, "归档目录数量")

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

func main() {
	// 设置命令行标志
	tarFlags := &tarFlags{}
	tarFlags.TarFlags()
	thFlags := &handler.TarHandlerFlags{}
	thFlags.AddFlag()

	pflag.Parse()

	// 初始化日志
	if err := LogInit(tarFlags.logLevel, tarFlags.logFile, tarFlags.logFormat); err != nil {
		logrus.Fatal(errors.Wrap(err, "set log level error"))
	}

	currentDir, _ := os.Getwd()
	logrus.Debug("当前工作目录:", currentDir)

	// 保存归档文件的绝对路径
	tmpDir := currentDir + string(os.PathSeparator) + thFlags.ArchiveDest
	// 归档源的绝对路径
	archiveSrcPath := currentDir + string(os.PathSeparator) + thFlags.ArchiveSrc

	// 获取归档源目录下的文件列表
	dataDirFiles, err := os.ReadDir(archiveSrcPath)
	if err != nil {
		panic("获取目录中的日期列表失败")
	}

	// 用来判断是否开始循环的变量
	var count int = 1
	var isStart bool = false

	for _, file := range dataDirFiles {
		// 判断是否开始循环的条件
		// 条件1：当前目录名称必须为指定的名称
		// 条件2：指定开始的目录名称不能为空
		// 条件3：循环必须已经开始
		// 当上述三个条件都成立，才会开始循环归档，否则跳过
		if file.Name() != tarFlags.startAt && tarFlags.startAt != "" && !isStart {
			logrus.Infof("跳过 %v 目录", file.Name())
			continue
		}
		isStart = true

		// 计数器，循环指定次数即停止
		if count > tarFlags.count {
			break
		}
		logrus.WithFields(logrus.Fields{
			"当前工作次数": count,
			"总工作次数":  tarFlags.count,
		}).Debug("检查循环次数")

		dataPath := fmt.Sprintf("%s%s%s", thFlags.ArchiveSrc, string(os.PathSeparator), file.Name())

		logrus.WithFields(logrus.Fields{
			"日期名称": file.Name(),
			"日期路径": dataPath,
		}).Debug("检查日期目录信息")

		// 创建待归档目录
		archiveDestPath := fmt.Sprintf("%s%s%s", tmpDir, string(os.PathSeparator), file.Name())
		logrus.Debug("检查归档目标目录", archiveDestPath)

		err = os.MkdirAll(archiveDestPath, 0775)
		if err != nil {
			panic(fmt.Sprintf("检查归档目标目录出错，%s", err))
		}

		err = os.Chdir(dataPath)
		if err != nil {
			panic(fmt.Sprintf("切换目录出错：%s", err))
		}

		nameDirFiles, err := os.ReadDir(dataPath)
		if err != nil {
			panic(fmt.Sprintf("获取 %s 目录中的名称列表失败:%s", dataPath, err))

		}
		for _, file := range nameDirFiles {
			twoLayerArchiveSrcPath := fmt.Sprintf("%s%s%s", dataPath, string(os.PathSeparator), file.Name())

			logrus.WithFields(logrus.Fields{
				"姓名名称": file.Name(),
				"姓名路径": twoLayerArchiveSrcPath,
			}).Debug("检查第二层目录信息")

			// archiveDestName := fmt.Sprintf("%s%s%s.tar.gz", archiveDestPath,string(os.PathSeparator), file.Name())
			archiveDestName := fmt.Sprintf("%s%s%s.tar.gz", archiveDestPath, string(os.PathSeparator), file.Name())
			logrus.Debug("检查归档目标名称", archiveDestName)

			// 打包
			err = handler.Archiving(file.Name(), archiveDestName)
			if err != nil {
				logrus.Error("归档失败：", err)
			}
		}

		count++
	}

	// 待提取的文件
	// var extractSrc = "./file_handle/tar_dir/test_tar.tar.gz"
	// // 提取后保存的路径，不写就是解压到当前目录
	// var extractDst = "./file_handle/tar_dir/"

	// Extracting(extractDst, extractSrc)
}
