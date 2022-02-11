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
	// TODO: 写个接口，分成两个
	case "zip":
		// 通过 fw 来创建 zip.Write
		writer := zip.NewWriter(fileDescriptor)
		defer func() {
			// 检测一下是否成功关闭
			if err := writer.Close(); err != nil {
				log.Fatalln(err)
			}
		}()

		z := archiving.NewZipWriter(writer, src)
		return z.Archiving()

	case "tar.gz":
		// 将 tar 包使用 gzip 压缩，其实添加压缩功能很简单，
		// 只需要在 fw 和 archivingWriter 之前加上一层压缩就行了，和 Linux 的管道的感觉类似
		gzipWriter := gzip.NewWriter(fileDescriptor)
		defer gzipWriter.Close()

		// 创建 Tar.Writer 结构
		writer := tar.NewWriter(gzipWriter)
		defer func() {
			// 检测一下是否成功关闭
			if err := writer.Close(); err != nil {
				logrus.Error(err)
			}
		}()

		defer writer.Close()

		tarWriter := archiving.NewTarWriter(writer, src)
		return tarWriter.Archiving()

	default:
		panic("请指定正确的程序扩展名")

	}
}
