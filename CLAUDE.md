# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概述
这是一个 Synology Cloud Sync 解密工具的 Go 语言实现，用于解密 Synology NAS 设备通过 Cloud Sync 功能加密的文件。

## 核心命令

### 构建项目
```bash
# 基本构建
go build -o syndecrypt cmd/syndecrypt/main.go

# 跨平台构建
GOOS=linux GOARCH=amd64 go build -o syndecrypt-linux-amd64 cmd/syndecrypt/main.go
GOOS=darwin GOARCH=amd64 go build -o syndecrypt-darwin-amd64 cmd/syndecrypt/main.go
GOOS=windows GOARCH=amd64 go build -o syndecrypt-windows-amd64.exe cmd/syndecrypt/main.go
```

### 运行测试
```bash
go test ./...
```

### 依赖管理
```bash
go mod download
go mod tidy
```

## 架构设计

### 模块结构
- `cmd/syndecrypt/` - 命令行入口，使用 docopt 进行参数解析
- `pkg/core/` - 核心解密算法实现
  - `crypto.go` - OpenSSL KDF、AES-256-CBC、RSA-OAEP 算法
  - `decrypt.go` - 流式解密处理逻辑
  - `stream.go` - 数据流管理
- `pkg/files/` - 文件处理逻辑
  - `decrypt.go` - 文件解密调度
  - `results.go` - 结果统计和报告
- `pkg/util/` - 工具函数
  - `lz4.go` - LZ4 解压处理（依赖系统 lz4 命令）
  - `file.go` - 文件操作工具

### 核心算法
1. **密钥派生**: OpenSSL KDF (EVP_BytesToKey) - 基于密码生成加密密钥
2. **对称加密**: AES-256-CBC with PKCS7 padding
3. **非对称加密**: RSA-OAEP for 私钥解密场景
4. **压缩**: LZ4 解压（通过系统命令调用）
5. **验证**: MD5 摘要验证解密结果

### 处理流程
1. 解析加密文件格式，提取元数据（盐值、IV、会话密钥等）
2. 根据解密模式（密码或 RSA 私钥）派生或解密会话密钥
3. 使用会话密钥进行 AES-256-CBC 解密
4. 对解密结果进行 LZ4 解压
5. 验证 MD5 摘要，确保数据完整性

## 开发注意事项

### 外部依赖
- 系统必须安装 `lz4` 命令行工具，可通过 `which lz4` 验证

### 文件格式支持
支持以下加密文件扩展名：`.cse`, `.enc`, `.cloudsync`, `.csenc`

### 错误处理
- 所有解密操作都有详细的错误信息和统计报告
- 使用 `pkg/files/results.go` 中的 `ResultStats` 跟踪处理结果
- 单文件失败不会影响批量处理中的其他文件

### 性能优化
- 流式处理大文件，内存占用低
- 支持批量文件和目录递归处理
- 单可执行文件部署，无运行时依赖