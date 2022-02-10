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
}

func (flags *tarFlags) TarFlags() {
	pflag.StringVar(&flags.logLevel, "log-level", "info", "The logging level:[debug, info, warn, error, fatal]")
	pflag.StringVar(&flags.logFile, "log-output", "", "the file which log to, default stdout")
	pflag.StringVar(&flags.logFormat, "log-format", "text", "log format,one of: json|text")

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
	logrus.Debug("当前工作目录：", currentDir)
	tmpDir := currentDir + string(os.PathSeparator) + "tmp"

	thFlags.ArchiveSrc = currentDir + string(os.PathSeparator) + thFlags.ArchiveSrc

	dataDirFiles, err := os.ReadDir(thFlags.ArchiveSrc)
	if err != nil {
		panic("获取目录中的日期列表失败")
	}

	for _, file := range dataDirFiles {
		dataPath := fmt.Sprintf("%s%s%s", thFlags.ArchiveSrc, string(os.PathSeparator), file.Name())

		logrus.WithFields(logrus.Fields{
			"日期名称": file.Name(),
			"日期路径": dataPath,
		}).Debug("检查日期目录信息")

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
	}

	// 待提取的文件
	// var extractSrc = "./file_handle/tar_dir/test_tar.tar.gz"
	// // 提取后保存的路径，不写就是解压到当前目录
	// var extractDst = "./file_handle/tar_dir/"

	// Extracting(extractDst, extractSrc)
}
