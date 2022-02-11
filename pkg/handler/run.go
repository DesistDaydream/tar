package handler

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"log"
	"os"
)

type ArchiveHandler interface {
	Archiving()
}

// TODO: 改成接口
var archivingWriter *tar.Writer

// var archivingWriter *zip.Writer

type tarWriter struct {
}

func (t *tarWriter) Archiving() {

}

type zipWriter struct {
}

func (t *zipWriter) Archiving() {

}

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
		zipWriter := zip.NewWriter(fileDescriptor)
		defer func() {
			// 检测一下是否成功关闭
			if err := zipWriter.Close(); err != nil {
				log.Fatalln(err)
			}
		}()

		return Zip(zipWriter, dst)
	case "tar.gz":
		// 将 tar 包使用 gzip 压缩，其实添加压缩功能很简单，
		// 只需要在 fw 和 archivingWriter 之前加上一层压缩就行了，和 Linux 的管道的感觉类似
		gzipWriter := gzip.NewWriter(fileDescriptor)
		defer gzipWriter.Close()

		// 创建 Tar.Writer 结构
		archivingWriter = tar.NewWriter(gzipWriter)

		defer archivingWriter.Close()

		return Targz(archivingWriter, src)

	default:
		panic("请指定正确的程序扩展名")

	}
}
