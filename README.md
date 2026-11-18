# ql-cli

一个用 Go 编写的命令行工具集，基于 [Cobra](https://github.com/spf13/cobra) 框架构建。

## 功能特性

| 命令 | 说明 |
|------|------|
| `copy` | 批量复制文件到指定目录，支持筛选文件类型、递归、保留目录结构 |
| `move` | 批量移动文件到指定目录，支持筛选文件类型、递归、保留目录结构 |
| `rename` | 批量修改文件名，支持查找替换、正则匹配、前后缀、大小写转换、清理特殊字符等 |
| `splitzip` / `unsplitzip` | 将目录按大小分割压缩为多个 zip 文件，或还原分卷 zip |
| `crypto` | 基于 [age](https://age-encryption.org/) 的文件加解密工具，支持密码短语和公钥/私钥两种模式 |
| `icongen` | 从一张高清图片自动生成全平台（iOS/Android/macOS/Windows/Linux/watchOS/Web PWA）应用图标 |
| `uuid` | UUID 生成与压缩工具 |

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

---

### `copy` — 批量复制文件

遍历指定目录，批量复制文件到目标目录。

#### 基本用法

```bash
ql-cli copy -d ./files -o ./backup -f .jpg -r -k
```

#### 参数说明

| 参数 | 说明 |
|------|------|
| `-d` / `--dir` | 源目录路径（默认 `.`） |
| `-o` / `--out` | 输出目录（默认 `.`） |
| `-f` / `--filter` | 筛选指定文件类型，可多次指定 `-f=png -f=txt`，也支持逗号分隔 `-f=png,txt` |
| `-r` / `--recursive` | 递归处理子目录 |
| `-k` / `--keep-dir` | 保留目录结构 |
| `--dry-run` | 预览模式，不实际执行复制 |

---

### `move` — 批量移动文件

遍历指定目录，批量移动文件到目标目录。

#### 基本用法

```bash
ql-cli move -d ./files -o ./backup -f .jpg -r -k
```

#### 参数说明

| 参数 | 说明 |
|------|------|
| `-d` / `--dir` | 源目录路径（默认 `.`） |
| `-o` / `--out` | 输出目录（默认 `.`） |
| `-f` / `--filter` | 筛选指定文件类型，可多次指定或逗号分隔 |
| `-r` / `--recursive` | 递归处理子目录 |
| `-k` / `--keep-dir` | 保留目录结构 |
| `--dry-run` | 预览模式，不实际执行移动 |

---

### `rename` — 批量重命名

遍历指定目录，批量修改文件名。支持查找替换、正则匹配、添加前后缀、大小写转换等多种操作。

#### 基本用法

```bash
# 查找替换
ql-cli rename -d ./files --from "IMG_" --to "Photo_"

# 正则替换
ql-cli rename -d ./files --regex --from "IMG_(\d+)" --to "Photo_$1"

# 添加前缀 + 大小写转换
ql-cli rename -d ./files --prefix "2024_" --to-lower

# 清理特殊字符 + 预览
ql-cli rename -d ./files --clean --dry-run

# 移除扩展名
ql-cli rename -d ./files --remove-ext
```

#### 参数说明

| 参数 | 说明 |
|------|------|
| `-d` / `--dir` | 要处理的目录路径（默认 `.`） |
| `-f` / `--filter` | 筛选指定文件类型 |
| `-r` / `--recursive` | 递归处理子目录 |
| `--from` | 要查找的文本 |
| `--to` | 替换为的文本 |
| `--regex` | 使用正则表达式匹配（默认开启） |
| `--prefix` | 添加文件名前缀 |
| `--suffix` | 添加文件名后缀（在扩展名前） |
| `--to-lower` | 文件名转小写 |
| `--to-upper` | 文件名转大写 |
| `--clean` | 清理特殊字符（只保留英文字母、数字、下划线、连字符、中文汉字及括号） |
| `--remove-ext` | 移除扩展名 |
| `--dry-run` | 预览模式，不实际执行重命名 |

---

### `splitzip` / `unsplitzip` — 分卷压缩与还原

将指定目录按大小分割压缩为多个 zip 文件，或还原分卷 zip。

#### 基本用法

```bash
# 压缩：将 source 目录按 10MB 分割为多个 zip
ql-cli splitzip -s /path/to/src -d /path/to/zips -p split_ -b 10

# 还原
ql-cli unsplitzip -s /path/to/zips -d /path/to/restore -p split_
```

#### 参数说明

**splitzip**

| 参数 | 说明 |
|------|------|
| `-s` / `--source` | 源文件目录 |
| `-d` / `--dir` | 输出目录（默认 `.`） |
| `-p` / `--prefix` | zip 文件前缀（默认 `split_`） |
| `-b` / `--max` | 单卷最大大小，单位 MB（默认 `2`）。若单文件超过该大小则单独压缩 |

**unsplitzip**

| 参数 | 说明 |
|------|------|
| `-s` / `--source` | 包含分割 zip 文件的目录 |
| `-d` / `--dir` | 还原输出目录（默认 `.`） |
| `-p` / `--prefix` | zip 文件前缀（需与 splitzip 一致，默认 `split_`） |

---

### `crypto` — 文件加解密

基于 [age](https://age-encryption.org/) 实现，支持两种加密方式：**密码短语** 和 **公钥/私钥**。

#### 子命令

| 子命令 | 说明 |
|--------|------|
| `keygen` | 生成 age 密钥对 |
| `encrypt` | 加密文件或目录 |
| `decrypt` | 解密文件或目录 |

#### 两种加密模式

| 模式 | 加密参数 | 解密参数 | 说明 |
|------|---------|---------|------|
| 密码短语 | `-p` | `-p` | 对称加密，加解密使用同一密码 |
| 公钥/私钥 | `-r`（公钥 `age1...`） | `-i`（私钥文件路径） | 非对称加密，公钥加密、私钥解密 |

#### 基本用法

```bash
# 生成密钥对（默认私钥保存到 ~/.age/key.txt）
ql-cli crypto keygen -o ~/.age/key.txt

# 密码短语加密单文件
ql-cli crypto encrypt -s secret.txt -p "my password" -o secret.age

# 公钥加密 + 私钥解密（非对称）
ql-cli crypto encrypt -s data.txt -r age1xxxxxxxxxxxxx -o data.age
ql-cli crypto decrypt -s data.age -i ~/.age/key.txt -o data.txt

# 批量加密目录（递归 + 仅 txt 文件 + 保留目录结构）
ql-cli crypto encrypt -s ./docs -R -f txt -p "pw" -o ./enc_out -k --suffix=.age

# 批量解密目录（递归 + 保留目录结构）
ql-cli crypto decrypt -s ./enc_out -R -i ~/.age/key.txt -o ./docs_decoded -k --suffix=.age
```

#### 参数说明

**keygen**

| 参数 | 说明 |
|------|------|
| `-o` / `--out` | 私钥输出路径（默认 `~/.age/key.txt`） |

**encrypt / decrypt 通用参数**

| 参数 | 说明 |
|------|------|
| `-s` / `--source` | 要加密/解密的文件或目录路径 |
| `-o` / `--out` | 输出路径 |
| `-R` / `--recursive` | 递归处理子目录 |
| `-f` / `--filter` | 按后缀筛选文件，可多次指定（默认全部文件） |
| `-k` / `--keep-dir` | 目录模式下保留子目录结构 |
| `--suffix` | 加密时追加 / 解密时要去除的后缀，如 `--suffix=.age` |

**encrypt 专属参数**

| 参数 | 说明 |
|------|------|
| `-r` / `--recipient` | age 公钥 `age1...`（非对称加密） |
| `-p` / `--passphrase` | 密码短语（对称加密） |

**decrypt 专属参数**

| 参数 | 说明 |
|------|------|
| `-i` / `--identity` | age 私钥文件路径（非对称解密） |
| `-p` / `--passphrase` | 密码短语（对称解密） |

> **注意：** 加密时 `-r` 和 `-p` 至少需要指定一个；解密时 `-i` 和 `-p` 至少需要指定一个。

---

### `icongen` — 全平台应用图标生成器

从一张高清图片源自动生成全平台应用图标，按平台打包为 ZIP 文件。

#### 支持的平台

| 平台 | 标识 | 输出格式 |
|------|------|---------|
| iOS / watchOS | `ios` / `watchos` | AppIcon.appiconset + Contents.json |
| Android | `android` | mipmap-*/ 目录结构 |
| macOS | `macos` | `.icns` 文件 |
| Windows | `windows` | `.ico` 文件 |
| Linux | `linux` | hicolor/ 目录 + `.desktop` |
| Web PWA | `web` | `manifest.json` + `browserconfig.xml` |

#### 基本用法

```bash
# 生成全部平台图标
ql-cli icongen -s icon.png -d ./icons -n MyApp

# 只生成指定平台
ql-cli icongen -s icon.png -d ./icons -f ios,android,web

# 仅生成 iOS + watchOS
ql-cli icongen -s icon.png -f ios,watchos
```

#### 参数说明

| 参数 | 说明 |
|------|------|
| `-s` / `--source` | 源图片路径（支持 PNG/JPG/JPEG/BMP/WEBP，建议 1024×1024 正方形） |
| `-d` / `--dir` | 输出目录（默认 `./icons`） |
| `-f` / `--formats` | 平台列表，逗号分隔（默认 `all`） |
| `-n` / `--name` | 图标/应用名称（默认 `AppIcon`） |

---

### `uuid` — UUID 工具

UUID 生成与压缩工具。压缩算法将标准 UUID（36 字符）压缩为 23 位或 22 位短字符串，可逆解压。

#### 子命令

| 子命令 | 说明 |
|--------|------|
| `gen` | 生成 UUID v4 |
| `com` | 压缩 UUID（默认 23 位，`-s` 压缩为 22 位） |
| `dec` | 解压压缩后的 UUID 回标准格式 |

#### 基本用法

```bash
# 生成 UUID v4
ql-cli uuid gen

# 压缩 UUID（默认 23 位）
ql-cli uuid com 0d6ad000-bc92-4518-8d1c-044763bfc9b4

# 压缩为 22 位
ql-cli uuid com 0d6ad000-bc92-4518-8d1c-044763bfc9b4 -s

# 解压回标准 UUID 格式
ql-cli uuid dec 0d6adAAvJJFGI0cBEdjv8m0
```

#### 参数说明

**com**

| 参数 | 说明 |
|------|------|
| UUID（位置参数） | 要压缩的标准 UUID 字符串 |
| `-s` / `--short` | 压缩为 22 位（默认 23 位） |

**dec**

| 参数 | 说明 |
|------|------|
| UUID（位置参数） | 要解压的压缩后 UUID 字符串（22 或 23 位） |

---

## 依赖库

| 分类 | 技术 |
|------|------|
| 语言 | [Go 1.27+](https://go.dev/) |
| CLI 框架 | [Cobra](https://github.com/spf13/cobra) |
| 文件加解密 | [age](https://age-encryption.org/) |
| 图像处理 | [imaging](https://github.com/disintegration/imaging) |
| 终端美化 | [fatih/color](https://github.com/fatih/color) |

## License

MIT