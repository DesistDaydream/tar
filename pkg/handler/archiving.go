package handler

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
)

func Archiving(src, dst, extension string) (err error) {
	// 创建文件
	fileDescriptor, err := os.Create(dst)
	if err != nil {
		return
	}
	defer fileDescriptor.Close()

	// TODO: 改成接口
	var archivingWriter *tar.Writer
	// var archivingWriter *zip.Writer

	switch extension {
	// TODO: 写个接口，分成两个
	case "zip":
		// 通过 fw 来创建 zip.Write
		zw := zip.NewWriter(fileDescriptor)
		defer func() {
			// 检测一下是否成功关闭
			if err := zw.Close(); err != nil {
				log.Fatalln(err)
			}
		}()

		return Zip(zw, dst)
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

func Targz(archivingWriter *tar.Writer, src string) error {
	// 下面就该开始处理数据了，这里的思路就是递归处理目录及目录下的所有文件和目录
	// 这里可以自己写个递归来处理，不过 Golang 提供了 filepath.Walk 函数，可以很方便的做这个事情
	// 直接将这个函数的处理结果返回就行，需要传给它一个源文件或目录，它就可以自己去处理
	// 我们就只需要去实现我们自己的 打包逻辑即可，不需要再去做路径相关的事情
	return filepath.Walk(src, func(fileName string, fi os.FileInfo, err error) error {
		// 因为这个闭包会返回个 error ，所以先要处理一下这个
		if err != nil {
			return err
		}

		// 这里就不需要我们自己再 os.Stat 了，它已经做好了，我们直接使用 fi 即可
		hdr, err := tar.FileInfoHeader(fi, "")
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
		if err := archivingWriter.WriteHeader(hdr); err != nil {
			return err
		}

		// 判断下文件是否是标准文件，如果不是就不处理了，
		// 如： 目录，这里就只记录了文件信息，不会执行下面的 copy
		if !fi.Mode().IsRegular() {
			return nil
		}

		// 打开文件
		file, err := os.Open(fileName)
		if err != nil {
			return err
		}
		defer file.Close()

		// copy 文件数据到 archivingWriter
		n, err := io.Copy(archivingWriter, file)
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

func Zip(zw *zip.Writer, src string) error {
	// 下面来将文件写入 zw ，因为有可能会有很多个目录及文件，所以递归处理
	return filepath.Walk(src, func(path string, fi os.FileInfo, errBack error) (err error) {
		if errBack != nil {
			return errBack
		}

		// 通过文件信息，创建 zip 的文件信息
		fh, err := zip.FileInfoHeader(fi)
		if err != nil {
			return
		}

		// 替换文件信息中的文件名
		fh.Name = strings.TrimPrefix(path, string(filepath.Separator))

		// 这步开始没有加，会发现解压的时候说它不是个目录
		if fi.IsDir() {
			fh.Name += "/"
		}

		// 写入文件信息，并返回一个 Write 结构
		w, err := zw.CreateHeader(fh)
		if err != nil {
			return
		}

		// 检测，如果不是标准文件就只写入头信息，不写入文件数据到 w
		// 如目录，也没有数据需要写
		if !fh.Mode().IsRegular() {
			return nil
		}

		// 打开要压缩的文件
		fr, err := os.Open(path)
		if err != nil {
			return
		}
		defer fr.Close()

		// 将打开的文件 Copy 到 w
		n, err := io.Copy(w, fr)
		if err != nil {
			return
		}
		// 输出压缩的内容
		fmt.Printf("成功压缩文件： %s, 共写入了 %d 个字符的数据\n", path, n)

		return nil
	})
}
