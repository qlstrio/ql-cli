package uuidTool

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var DecompressCmd = &cobra.Command{
	Use:   "dec",
	Short: "解压 UUID",
	Long:  "解压 UUID 字符串，支持解压长度为 23 或 22 的 UUID 字符串",
	RunE:  runDecompress,
}

func runDecompress(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		fmt.Println(color.RedString("请输入 UUID 字符串"))
		return nil
	}
	u := args[0]
	if len(u) != 23 && len(u) != 22 {
		fmt.Println(color.RedString("请输入正确的 UUID 字符串"))
		return nil
	}
	decompress(u)
	return nil
}

// decompress 是解压函数。
//
// 它支持两种压缩后的长度：
//
//  1. 长度 23：
//     前 5 个字符是原始十六进制字符，后面 18 个字符是压缩数据。
//     18 个 Base64 字符每 2 个还原成 3 个十六进制字符，
//     共还原 27 个十六进制字符，加上前 5 个，正好 32 个。
//
//  2. 长度 22：
//     前 2 个字符是原始十六进制字符，后面 20 个字符是压缩数据。
//     20 个 Base64 字符每 2 个还原成 3 个十六进制字符，
//     共还原 30 个十六进制字符，加上前 2 个，正好 32 个。
//
// 最后把 32 个十六进制字符按 UUID 的 8-4-4-4-12 格式拼接。
func decompress(TypeIDStr string) {
	// AsciiTo64 是反查表：Base64 字符 -> 6-bit 数值。
	var AsciiTo64 [128]int

	for i := range 64 {
		AsciiTo64[ase64KeyChars[i]] = i
	}

	if len(TypeIDStr) == 23 {
		// 长度 23：保留前 5 个原始十六进制字符，解压后面 18 个字符。
		hexChars := make([]byte, 0, 27)
		// 每 2 个 Base64 字符还原成 3 个十六进制字符。
		for i := 5; i < 23; i += 2 {
			lhs := AsciiTo64[TypeIDStr[i]]
			rhs := AsciiTo64[TypeIDStr[i+1]]
			// 逆运算：
			//
			// lhs 是 6 bit：
			//   hex1 = lhs 的高 4 bit = lhs >> 2
			//   hex2 的高 2 bit = lhs 的低 2 bit
			//
			// rhs 是 6 bit：
			//   hex2 的低 2 bit = rhs 的高 2 bit = rhs >> 4
			//   hex3 = rhs 的低 4 bit = rhs & 0xF
			//
			// 所以：
			//   hex1 = lhs >> 2
			//   hex2 = ((lhs & 3) << 2) | (rhs >> 4)
			//   hex3 = rhs & 0xF
			hexChars = append(hexChars, hexByte(lhs>>2))
			hexChars = append(hexChars, hexByte(((lhs&3)<<2)|(rhs>>4)))
			hexChars = append(hexChars, hexByte(rhs&0xF))
		}
		// 前 5 个字符 + 解压出的 27 个字符 = 32 个十六进制字符。
		TypeIDStr = TypeIDStr[0:5] + string(hexChars)
	} else if len(TypeIDStr) == 22 {
		// 长度 22：保留前 2 个原始十六进制字符，解压后面 20 个字符。
		hexChars := make([]byte, 0, 30)

		for i := 2; i < 22; i += 2 {
			lhs := AsciiTo64[TypeIDStr[i]]
			rhs := AsciiTo64[TypeIDStr[i+1]]

			hexChars = append(hexChars, hexByte(lhs>>2))
			hexChars = append(hexChars, hexByte(((lhs&3)<<2)|(rhs>>4)))
			hexChars = append(hexChars, hexByte(rhs&0xF))
		}
		// 前 2 个字符 + 解压出的 30 个字符 = 32 个十六进制字符。
		TypeIDStr = TypeIDStr[0:2] + string(hexChars)
	}
	// 按 UUID 格式 8-4-4-4-12 重新加上连字符。
	result := strings.Join([]string{
		TypeIDStr[0:8],
		TypeIDStr[8:12],
		TypeIDStr[12:16],
		TypeIDStr[16:20],
		TypeIDStr[20:],
	}, "-")

	fmt.Println("解压后：" + color.GreenString(result))
}
