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

	// 下面就该开始处理数据了，这里的思路就是递归处理目录及目录下的所有文件和目录
	// 这里可以自己写个递归来处理，不过 Golang 提供了 filepath.Walk 函数，可以很方便的做这个事情
	// 直接将这个函数的处理结果返回就行，需要传给它一个源文件或目录，它就可以自己去处理
	// 我们就只需要去实现我们自己的 打包逻辑即可，不需要再去做路径相关的事情
	return filepath.WalkDir(t.Src, func(fileName string, dirEntry os.DirEntry, err error) error {
		// 因为这个闭包会返回个 error ，所以先要处理一下这个
		if err != nil {
			return err
		}
		fileInfo, _ := dirEntry.Info()
		// 这里就不需要我们自己再 os.Stat 了，它已经做好了，我们直接使用 fi 即可
		hdr, err := tar.FileInfoHeader(fileInfo, "")
		if err != nil {
			return err
		}

		// 这里需要处理下 hdr 中的 Name，因为默认文件的名字是不带路径的，
		// 打包之后所有文件就会堆在一起，这样就破坏了原本的目录结果
		// 例如： 将原本 hdr.Name 的 syslog 替换程 log/syslog
		// 这个其实也很简单，回调函数的 fileName 字段给我们返回来的就是完整路径的 log/syslog
		// strings.TrimPrefix 将 fileName 的最左侧的 / 去掉，
		// 熟悉 Linux 的都知道为什么要去掉这个
		hdr.Name = strings.TrimPrefix(fileName, string(filepath.Separator))

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
		file, err := os.Open(fileName)
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
			"归档文件": fileName,
			"文件大小": fmt.Sprintf("%v bytes", n),
		}).Tracef("归档成功")

		return nil
	})
}
