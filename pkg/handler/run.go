package handler

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"log"
	"os"

	"github.com/DesistDaydream/tar/pkg/archiving"
	"github.com/sirupsen/logrus"
)

func Run(src, dst, extension string) (err error) {
	// 创建文件
	fileDescriptor, err := os.Create(dst)
	if err != nil {
		return
	}
	defer fileDescriptor.Close()

	switch extension {
	case "zip":
		writer := zip.NewWriter(fileDescriptor)
		// 检测一下是否成功关闭
		defer func() {
			if err := writer.Close(); err != nil {
				log.Fatalln(err)
			}
		}()

		zipWriter := archiving.NewZipWriter(writer, src)
		return zipWriter.Archiving()

	case "tar.gz":
		gzipWriter := gzip.NewWriter(fileDescriptor)
		defer gzipWriter.Close()

		// 创建 Tar.Writer 结构
		writer := tar.NewWriter(gzipWriter)
		// 检测一下是否成功关闭
		defer func() {
			if err := writer.Close(); err != nil {
				logrus.Error(err)
			}
		}()

		tarWriter := archiving.NewTarWriter(writer, src)
		return tarWriter.Archiving()

	default:
		panic("请指定正确的程序扩展名")
	}
}
