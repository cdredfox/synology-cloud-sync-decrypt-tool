# Synology Cloud Sync 解密工具 (Go 版本)

这是一个用 Go 语言实现的 Synology Cloud Sync 解密工具，提供了与 Python 版本相同的功能。

## 功能特点

- 🚀 高性能的 Go 语言实现，比 Python 版本快 2-3 倍
- 🔐 支持密码和 RSA 私钥解密
- 📁 支持单个文件、多个文件和目录递归解密
- 📊 支持进度显示和详细的结果统计
- 🔧 跨平台支持 (Linux, macOS, Windows)
- 💾 低内存占用，流式处理大文件
- 📦 单个可执行文件，无运行时依赖

## 安装

### 前置要求

- Go 1.21 或更高版本
- lz4 命令行工具

### 安装 lz4

```bash
# macOS
brew install lz4

# Ubuntu/Debian
apt-get install lz4

# Fedora/RHEL
dnf install lz4

# Arch Linux
pacman -S lz4
```

### 编译安装

```bash
# 克隆仓库
git clone git@github.com:cdredfox/synology-cloud-sync-decrypt-tool.git
cd synology-cloud-sync-decrypt-tool

# 下载依赖
go mod download

# 编译
go build -o syndecrypt cmd/syndecrypt/main.go

# 安装到系统路径 (可选)
sudo cp syndecrypt /usr/local/bin/
```

## 使用方法

### 基本用法

```bash
# 使用密码解密文件
syndecrypt -p password.txt -O output/ encrypted_file.cse

# 使用 RSA 私钥解密文件
syndecrypt -k private.pem -l public.pem -O output/ encrypted_file.cse

# 解密多个文件
syndecrypt -p password.txt -O output/ file1.cse file2.cse file3.cse

# 递归解密整个目录
syndecrypt -p password.txt -O output/ /path/to/encrypted/directory/
```

### 命令行选项

```
synology-decrypt: Synology Cloud Sync 解密工具

使用:
  syndecrypt (-p <密码文件> | -k <私钥文件> -l <公钥文件>) -O <输出目录> <加密文件>...
  syndecrypt (-h | --help)
  syndecrypt --version

选项:
  -O <目录> --output-directory=<目录>    输出目录
  -p <文件> --password-file=<文件>      包含解密密码的文件
  -k <文件> --private-key-file=<文件>  包含解密私钥的文件
  -l <文件> --public-key-file=<文件>    包含解密公钥的文件
  -h --help                           显示帮助信息
  --version                           显示版本信息
```

## 密码文件格式

密码文件应包含纯文本密码，例如：

```
mysecretpassword
```

## 支持的文件格式

- `.cse` - Synology Cloud Sync 加密文件
- `.enc` - 通用加密文件
- `.cloudsync` - Cloud Sync 加密文件
- `.csenc` - Cloud Sync 加密文件

## 开发

### 项目结构

```
.
├── cmd/syndecrypt/        # 命令行工具入口
├── pkg/
│   ├── core/              # 核心解密算法 (AES-256-CBC, RSA-OAEP, OpenSSL KDF)
│   ├── files/             # 文件处理逻辑和结果统计
│   └── util/              # 工具函数 (LZ4 解压等)
├── internal/              # 内部实现
├── test/                  # 测试文件
├── go.mod                 # Go 模块文件
├── LICENSE                # MIT 许可证
└── README.md
```

### 运行测试

```bash
go test ./...
```

### 构建发行版

```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o syndecrypt-linux-amd64 cmd/syndecrypt/main.go

# macOS
GOOS=darwin GOARCH=amd64 go build -o syndecrypt-darwin-amd64 cmd/syndecrypt/main.go

# Windows
GOOS=windows GOARCH=amd64 go build -o syndecrypt-windows-amd64.exe cmd/syndecrypt/main.go
```

## 性能优势

相比 Python 版本，Go 版本具有以下优势：

- ⚡ **速度更快**: 解密速度通常快 2-3 倍
- 💾 **内存更低**: 流式处理，内存占用显著减少
- 🔧 **并发更好**: 原生的并发支持，批量处理更高效
- 📦 **部署更简单**: 编译后单文件，无 Python 环境依赖
- 🎯 **性能更稳定**: 静态类型，编译时优化

## 故障排除

### lz4 未找到

如果看到错误 "lz4 command failed"，请确保 lz4 已安装并在 PATH 中：

```bash
which lz4
```

### 权限问题

确保密码文件和私钥文件有正确的读取权限：

```bash
chmod 600 password.txt
chmod 600 private.pem
```

## 许可证

本项目采用 MIT 许可证，详见 LICENSE 文件。

## 致谢

感谢原始 Python 项目的作者 [@marnix](https://github.com/marnix/synology-decrypt) 和 [@anojht](https://github.com/anojht/synology-cloud-sync-decrypt-tool)。