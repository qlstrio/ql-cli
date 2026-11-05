package cmd

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

var splitzipCmd = &cobra.Command{
	Use:   "splitzip",
	Short: "将文件分割压缩为多个zip文件",
	Long:  "将文件分割压缩为多个zip文件，保留文件结构。",
	RunE:  runSplitZip,
}
var (
	zipfolder    string
	outputfolder string
	zprefix      string
	maxSize      int64
)

func init() {
	splitzipCmd.Flags().StringVarP(&zipfolder, "source", "s", "", "源文件目录")
	splitzipCmd.Flags().StringVarP(&outputfolder, "dir", "d", ".", "输出目录")
	splitzipCmd.Flags().StringVarP(&zprefix, "prefix", "p", "split_", "zip文件前缀")
	splitzipCmd.Flags().Int64VarP(&maxSize, "max", "b", 2, "最大文件大小，单位MB，若单文件大小超过最大文件大小，单独压缩")
	rootCmd.AddCommand(splitzipCmd)
}

func runSplitZip(cmd *cobra.Command, args []string) error {
	var err error
	zipfolder, err = getAbsDirPath(zipfolder)
	if err != nil {
		return err
	}
	outputfolder, err = getAbsDirPath(outputfolder)
	if err != nil {
		return err
	}
	DoZip(zipfolder, outputfolder, maxSize*1024*1024, zprefix)
	return nil
}

type FileSize struct {
	Size int64
	Path string
}

func DoZip(sourceFilePath string, outputPath string, maxSize int64, prefix string) {
	var fileSize []FileSize
	filepath.Walk(sourceFilePath, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		f := FileSize{
			Size: info.Size(),
			Path: path,
		}
		fileSize = append(fileSize, f)
		return nil
	})
	totoalFileSize := len(fileSize)
	ret, _ := groupBySize(fileSize, maxSize)
	currFileSize := 0
	for index, stringInfo := range ret {
		currFileSize += len(stringInfo)
		buildCmdZip(stringInfo, fmt.Sprintf("%v%v.zip", prefix, index), outputPath, sourceFilePath)
	}
	fmt.Println(totoalFileSize, currFileSize)
}

type ByLargest []FileSize

func (f ByLargest) Len() int           { return len(f) }
func (f ByLargest) Swap(i, j int)      { f[i], f[j] = f[j], f[i] }
func (f ByLargest) Less(i, j int) bool { return f[i].Size > f[j].Size }

func groupBySize(files []FileSize, maxSize int64) ([][]string, error) {
	var ret [][]string

	sort.Sort(ByLargest(files))
	for i, e := range files {
		if e.Size == -1 {
			continue
		}
		if e.Size > maxSize {
			// 单个文件大小超过最大文件大小，单独压缩
			ret = append(ret, []string{e.Path})
			continue
		}
		files[i].Size = -1

		group := []string{e.Path}
		d := maxSize - e.Size

		j := i + 1
		s := e.Size
		for ; j < len(files); j++ {
			if files[j].Size == -1 {
				continue
			}
			if files[j].Size <= d {
				s += files[j].Size
				d = maxSize - s
				group = append(group, files[j].Path)
				files[j].Size = -1
			}
			if s > maxSize {
				break
			}
		}
		ret = append(ret, group)
	}
	return ret, nil
}

func buildCmdZip(oneZipInfo []string, zipName string, outputPath string, sourceFilePath string) {
	cmdStr := fmt.Sprintf("cd %v && zip -r %v/%v %v", sourceFilePath, outputPath, zipName, buildFileArr(oneZipInfo, sourceFilePath))
	_, err := exec_shell(cmdStr)
	if err != nil {
		fmt.Println(err)
	}
}

func buildFileArr(oneZipInfo []string, sourceFilePath string) string {
	cmdStr := ""
	for _, strInfo := range oneZipInfo {
		cmdStr += strings.Replace(strInfo, sourceFilePath+"/", "", -1) + " "
	}
	return cmdStr
}

// 阻塞式的执行外部shell命令的函数,等待执行完毕并返回标准输出
func exec_shell(s string) (string, error) {
	// 函数返回一个*Cmd，用于使用给出的参数执行name指定的程序
	cmd := exec.Command("/bin/bash", "-c", s)
	// 读取io.Writer类型的cmd.Stdout，再通过bytes.Buffer(缓冲byte类型的缓冲器)将byte类型转化为string类型(out.String():这是bytes类型提供的接口)
	var out bytes.Buffer
	cmd.Stdout = &out
	// Run执行c包含的命令，并阻塞直到完成。  这里stdout被取出，cmd.Wait()无法正确获取stdin,stdout,stderr，则阻塞在那了
	err := cmd.Run()

	return out.String(), err
}
