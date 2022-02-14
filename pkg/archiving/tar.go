package archiving

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
)

type TarWriter struct {
	Writer *tar.Writer
	Src    string
}

func NewTarWriter(writer *tar.Writer, src string) *TarWriter {
	return &TarWriter{
		Writer: writer,
		Src:    src,
	}
}

func (t *TarWriter) Archiving() error {
	return filepath.WalkDir(t.Src, func(filePath string, dirEntry os.DirEntry, err error) error {
		// 因为这个闭包会返回个 error ，所以先要处理一下这个
		if err != nil {
			return err
		}
		fileInfo, _ := dirEntry.Info()
		hdr, err := tar.FileInfoHeader(fileInfo, "")
		if err != nil {
			return err
		}

		// 去掉 filePath 的最左侧的 / 去掉
		hdr.Name = strings.TrimPrefix(filePath, string(filepath.Separator))

		// 写入文件信息
		if err := t.Writer.WriteHeader(hdr); err != nil {
			return err
		}

		// 判断下文件是否是标准文件，如果不是就不处理了，
		// 如： 目录，这里就只记录了文件信息，不会执行下面的 copy
		if !fileInfo.Mode().IsRegular() {
			return nil
		}

		// 打开文件
		file, err := os.Open(filePath)
		if err != nil {
			return err
		}
		defer file.Close()

		// copy 文件数据到 archivingWriter
		n, err := io.Copy(t.Writer, file)
		if err != nil {
			return err
		}

		logrus.WithFields(logrus.Fields{
			"归档文件": filePath,
			"文件大小": fmt.Sprintf("%v bytes", n),
		}).Tracef("归档成功")

		return nil
	})
}
