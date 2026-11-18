package uuidTool

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var CompressCmd = &cobra.Command{
	Use:   "com",
	Short: "压缩 UUID",
	Long:  "压缩 UUID 字符串，压缩后长度为 23 或 22 的 UUID 字符串，默认23位",
	RunE:  runCompress,
}

var short bool

func init() {
	CompressCmd.Flags().BoolVarP(&short, "short", "s", false, "压缩长度为22")
}

func runCompress(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		fmt.Println(color.RedString("请输入 UUID 字符串"))
		return nil
	}
	// 检查是否是 UUID 字符串。
	if !isUUID(args[0]) {
		fmt.Println(color.RedString("请输入正确的 UUID 字符串"))
		return nil
	}
	if short {
		fmt.Println(color.CyanString("压缩后长度为 22"))
	} else {
		fmt.Println(color.CyanString("压缩后长度为 23"))
	}
	compress(args[0])
	return nil
}

func isUUID(u string) bool {
	// 定义UUID v4的正则表达式
	r := regexp.MustCompile(`(?i)^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	return r.MatchString(u)
}

// compress 是压缩函数。
//
// 输入是一个 UUID 字符串，例如：
//
//	0d6ad000-bc92-4518-8d1c-044763bfc9b4
//
// 压缩步骤：
//  1. 去掉 UUID 中的 "-"，得到 32 个十六进制字符。
//  2. 保留前 5 个十六进制字符不压缩，记为 qianwu。
//  3. 剩余 27 个十六进制字符，每 3 个一组。
//     每个十六进制字符占 4 bit，3 个就是 12 bit。
//     算法把这 12 bit 重新拆成两个 6-bit 值，再分别查 Base64 字符表，
//     于是 3 个十六进制字符变成 2 个 Base64 字符。
//  4. 27 个字符正好分成 9 组，所以生成 18 个 Base64 字符。
//  5. 最终结果 = 前 5 个原始字符 + 18 个压缩字符，总长度 23。
func compress(str string) string {
	// 去掉 UUID 中的连字符。
	str = strings.ReplaceAll(str, "-", "")
	length := len(str)
	head := ""
	i := 0
	if !short {
		// 长度 23：保留前 5 个原始十六进制字符，压缩后面 18 个字符。
		head = str[0:5]
		i = 5
	} else {
		// 长度 22：保留前 2 个原始十六进制字符，压缩后面 20 个字符。
		head = str[0:2]
		i = 2
	}
	// 保留前 5 个十六进制字符，不参与压缩。

	var zi strings.Builder
	// 从第 6 个字符开始，每 3 个十六进制字符压缩成 2 个 Base64 字符。
	for i < length {
		// 每个 hex 字符表示 4 bit，取三个 4-bit 值。
		hexVal1 := hexCharValue(str[i])
		hexVal2 := hexCharValue(str[i+1])
		hexVal3 := hexCharValue(str[i+2])
		// 12 bit 拆成两个 6-bit：
		//
		// v1 = hexVal1 的全部 4 bit + hexVal2 的高 2 bit
		//    = (hexVal1 << 2) | (hexVal2 >> 2)
		//
		// v2 = hexVal2 的低 2 bit + hexVal3 的全部 4 bit
		//    = ((hexVal2 & 3) << 4) | hexVal3
		//
		// 然后分别用 Base64 字符表映射成字符。
		zi.WriteString(string(ase64KeyChars[(hexVal1<<2)|(hexVal2>>2)]))
		zi.WriteString(string(ase64KeyChars[((hexVal2&3)<<4)|hexVal3]))

		i += 3
	}

	result := head + zi.String()
	fmt.Println("压缩后：" + color.GreenString(result))
	return result
}
