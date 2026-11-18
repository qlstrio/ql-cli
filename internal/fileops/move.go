package fileops

import (
	"fmt"

	"github.com/qlstrio/ql-cli/internal/util"
)

func Move(opts CommonOptions) error {
	if opts.DryRun {
		fmt.Println("===== 预览模式(不会实际移动) =====")
	}
	return Batch(opts, util.MoveFile)
}
