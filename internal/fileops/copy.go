package fileops

import (
	"fmt"

	"github.com/qlstrio/ql-cli/internal/util"
)

func Copy(opts CommonOptions) error {
	if opts.DryRun {
		fmt.Println("===== 预览模式(不会实际复制) =====")
	}
	return Batch(opts, util.CopyFile)
}
