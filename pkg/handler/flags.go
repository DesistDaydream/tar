package handler

import "github.com/spf13/pflag"

type TarHandlerFlags struct {
	ArchiveSrc  string
	ArchiveDest string
	Extension   string
}

// AddFlag 用来为语雀用户数据设置一些值
func (flags *TarHandlerFlags) AddFlag() {
	pflag.StringVar(&flags.ArchiveSrc, "archive-src", "test_tar_dir", "待打包的目录，当前目录的相对路径")
	pflag.StringVar(&flags.ArchiveDest, "archive-dest", "tmp", "打包后保存的目录，当前目录的相对路径")
	pflag.StringVar(&flags.Extension, "extension", "tar.gz", "归档包文件的扩展名")
}
