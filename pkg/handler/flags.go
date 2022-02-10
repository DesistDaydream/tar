package handler

import "github.com/spf13/pflag"

type TarHandlerFlags struct {
	ArchiveSrc  string
	ArchiveDest string
}

// AddFlag 用来为语雀用户数据设置一些值
func (flags *TarHandlerFlags) AddFlag() {
	pflag.StringVar(&flags.ArchiveSrc, "archive-src", "test_tar_dir", "待打包的路径")
	pflag.StringVar(&flags.ArchiveDest, "archive-dest", "test.tar.gz", "打包后保存的路径")
}
