# ql-cli

一个用 Go 编写的命令行工具集，基于 [Cobra](https://github.com/spf13/cobra) 框架构建。

## 功能特性

- **copy** — 批量复制文件到指定目录，支持筛选指定文件类型、递归，保留文件结构
- **move** — 批量移动文件到指定目录,支持筛选指定文件类型、递归，保留文件结构
- **rename** — 批量修改文件名,支持查找替换、正则匹配、添加前后缀、大小写转换、清理特殊字符、扩展名过滤等
- **splitzip** — 将指定目录下的文件分割压缩为多个zip文件

## 安装

### 从源码构建

```bash
git clone https://github.com/qlstrio/ql-cli.git
cd ql-cli
go build -o ql-cli
```

### go install 安装

```bash
go install github.com/qlstrio/ql-cli@latest
```

## 命令列表

### `copy` — 批量复制文件
遍历指定目录，批量复制文件到目标目录。支持筛选指定文件类型、递归，保留文件结构。

#### 基本用法
```bash
ql copy -d ./files --filter ".jpg" --dst "./backup"
```

### `move` — 批量移动文件
遍历指定目录，批量移动文件指定类型文件到目标目录。如果目标目录不存在，会自动创建。

#### 基本用法
```bash
ql move -d ./files --filter ".jpg" --dst "./backup"
```

### `rename` — 批量重命名
遍历指定目录，批量修改文件名。支持查找替换、正则匹配、添加前后缀、大小写转换等操作。

#### 基本用法
```bash
ql rename -d ./files --from "IMG_" --to "Photo_"
```
### `splitzip` — 压缩文件
遍历指定目录，将所有文件分割压缩为多个zip文件。

#### 基本用法

```bash
ql splitzip -s ./files -d ./backup
```

## 技术栈

- [Go 1.26+](https://go.dev/)
- [Cobra](https://github.com/spf13/cobra) — CLI 框架

## License

MIT
