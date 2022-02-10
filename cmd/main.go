package main

import (
	"fmt"
	"os"

	"github.com/DesistDaydream/tar/pkg/handler"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
)

type tarFlags struct {
	logLevel  string
	logFile   string
	logFormat string
}

func (flags *tarFlags) AddYuqueExportFlags() {
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
	tarFlags := &tarFlags{}
	tarFlags.AddYuqueExportFlags()
	thFlags := &handler.TarHandlerFlags{}
	thFlags.AddFlag()

	pflag.Parse()

	// filepath.Walk(thFlags.ArchiveSrc, func(path string, info os.FileInfo, err error) error {
	// 	if err != nil {
	// 		return err
	// 	}

	// 	logrus.WithFields(logrus.Fields{
	// 		"mode": info.Mode(),
	// 		"name": info.Name(),
	// 		"sys":  info.Sys(),
	// 	}).Info()

	// 	return nil
	// })

	files, err := os.ReadDir(thFlags.ArchiveSrc)
	if err != nil {
		panic("获取待打包目录下的内容失败")
	}

	for _, file := range files {
		oneLayerArchiveSrcPath := fmt.Sprintf("%s\\%s", thFlags.ArchiveSrc, file.Name())

		logrus.WithFields(logrus.Fields{
			"oneLayerName": file.Name(),
			"oneLayerPath": oneLayerArchiveSrcPath,
		}).Info("检查第一层目录信息")

		archiveDestPath := fmt.Sprintf("tmp\\%s", file.Name())
		err = os.MkdirAll(archiveDestPath, 0775)
		if err != nil {
			panic(fmt.Sprintf("创建目录出错，%s", err))
		}

		files, _ := os.ReadDir(oneLayerArchiveSrcPath)
		for _, file := range files {
			twoLayerArchiveSrcPath := fmt.Sprintf("%s\\%s", oneLayerArchiveSrcPath, file.Name())

			logrus.WithFields(logrus.Fields{
				"twoLayerName": file.Name(),
				"twoLayerPath": twoLayerArchiveSrcPath,
			}).Info("检查第二层目录信息")

			// archiveDestName := fmt.Sprintf("%s\\%s.tar.gz", archiveDestPath, file.Name())
			archiveDestName := fmt.Sprintf("%s.tar.gz", file.Name())

			os.Chdir(oneLayerArchiveSrcPath)

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
