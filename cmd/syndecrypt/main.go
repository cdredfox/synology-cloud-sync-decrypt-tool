package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/docopt/docopt-go"
	"github.com/synology-cloud-sync-decrypt-tool/syndecrypt-go/pkg/core"
	"github.com/synology-cloud-sync-decrypt-tool/syndecrypt-go/pkg/files"
	"github.com/synology-cloud-sync-decrypt-tool/syndecrypt-go/pkg/util"
)

const version = "1.0.0"

const usage = `Synology Cloud Sync Decryption Tool

Usage:
  syndecrypt (-p <password-file> | -k <private-key-file> -l <public-key-file>) -O <output-directory> <encrypted-file>...
  syndecrypt (-h | --help)
  syndecrypt --version

Options:
  -O <directory> --output-directory=<directory>  Output directory
  -p <file> --password-file=<file>            File containing decryption password
  -k <file> --private-key-file=<file>        File containing decryption private key
  -l <file> --public-key-file=<file>        File containing decryption public key
  -h --help                              Show this help message
  --version                              Show version

Examples:
  # Decrypt with password
  syndecrypt -p password.txt -O output/ encrypted_file.cse

  # Decrypt with private key
  syndecrypt -k private.pem -l public.pem -O output/ file1.cse file2.cse

  # Recursive directory decryption
  syndecrypt -p password.txt -O output/ /path/to/encrypted/dir/

More information:
  https://github.com/anojht/synology-cloud-sync-decrypt-tool
`

func main() {
	args, err := docopt.ParseDoc(usage)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse arguments: %v\n", err)
		os.Exit(1)
	}

	if args["--version"].(bool) {
		fmt.Printf("synology-decrypt version %s\n", version)
		os.Exit(0)
	}

	// 解析参数
	outputDir := args["--output-directory"].(string)

	// 获取加密文件列表
	var encryptedFiles []string
	if files, ok := args["<encrypted-file>"].([]string); ok {
		encryptedFiles = files
	} else if file, ok := args["<encrypted-file>"].(string); ok {
		encryptedFiles = []string{file}
	}

	// 创建解密配置
	var config core.DecryptConfig

	// 检查密码文件
	if passwordFile, ok := args["--password-file"].(string); ok && passwordFile != "" {
		password, err := util.ReadBinaryFile(passwordFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to read password file: %v\n", err)
			os.Exit(1)
		}
		config.Password = password
	}

	// 检查私钥文件
	if privateKeyFile, ok := args["--private-key-file"].(string); ok && privateKeyFile != "" {
		privateKey, err := util.ReadBinaryFile(privateKeyFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to read private key file: %v\n", err)
			os.Exit(1)
		}
		config.PrivateKey = privateKey
	}

	// 检查公钥文件
	if publicKeyFile, ok := args["--public-key-file"].(string); ok && publicKeyFile != "" {
		publicKey, err := util.ReadBinaryFile(publicKeyFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to read public key file: %v\n", err)
			os.Exit(1)
		}
		config.PublicKey = publicKey
	}

	// 验证配置
	if err := files.ValidateConfig(config); err != nil {
		fmt.Fprintf(os.Stderr, "Configuration validation failed: %v\n", err)
		os.Exit(1)
	}

	// 确保输出目录存在
	if err := util.EnsureDir(outputDir); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create output directory: %v\n", err)
		os.Exit(1)
	}

	// 处理每个加密文件
	results := files.NewDecryptResults()

	for _, encryptedFile := range encryptedFiles {
		result := processFileWithResult(encryptedFile, outputDir, config)
		// 如果是目录，直接使用目录内的详细统计结果
		if result.FileCount > 0 {
			results.TotalFiles += result.FileCount
			results.SuccessCount += result.SuccessCount
			results.FailedCount += result.FailedCount
		} else {
			// 如果是单个文件，使用普通统计
			results.AddResult(result)
		}
	}

	// 显示结果摘要（只在控制台打印，不保存到文件）
	results.PrintSummary()
}

// processFileWithResult 处理单个文件或目录并返回结果
func processFileWithResult(inputPath, outputDir string, config core.DecryptConfig) files.DecryptResult {
	startTime := time.Now()
	result := files.DecryptResult{
		InputFile:  inputPath,
		StartTime:  startTime,
	}

	info, err := os.Stat(inputPath)
	if err != nil {
		result.Error = fmt.Sprintf("cannot access file: %v", err)
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime).String()
		fmt.Printf("  ❌ %s - %s\n", inputPath, result.Error)
		return result
	}

	if info.IsDir() {
		// 如果是目录，递归处理并获取详细统计
		dirResults, err := files.DecryptDirectory(inputPath, outputDir, config)
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime).String()

		if err != nil {
			result.Error = err.Error()
			fmt.Printf("  ❌ 目录 %s - %s\n", inputPath, result.Error)
			return result
		}

		// 目录处理成功，但我们需要返回目录内的详细统计结果
		// 而不是把整个目录当作一个文件
		result.Success = true
		result.OutputFile = outputDir
		result.FileCount = dirResults.TotalFiles
		result.SuccessCount = dirResults.SuccessCount
		result.FailedCount = dirResults.FailedCount

		// 静默处理成功的目录解密，不输出成功信息
		return result
	}

	// 如果是单个文件
	outputFile := generateOutputFileName(inputPath, outputDir)
	result.OutputFile = outputFile

	err = files.DecryptFile(inputPath, outputFile, config)
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime).String()

	if err != nil {
		result.Error = err.Error()
		fmt.Printf("  ❌ %s - %s\n", inputPath, result.Error)
		return result
	}

	// 获取文件大小
	if info, err := os.Stat(outputFile); err == nil {
		result.FileSize = info.Size()
	}

	result.Success = true
	// 静默处理成功的文件解密，不输出成功信息
	return result
}

// generateOutputFileName 生成输出文件名
func generateOutputFileName(inputFile, outputDir string) string {
	baseName := filepath.Base(inputFile)
	ext := filepath.Ext(baseName)

	// 如果是加密文件扩展名，移除它
	if isEncryptedExtension(ext) {
		baseName = baseName[:len(baseName)-len(ext)]
	}

	return filepath.Join(outputDir, baseName)
}

// isEncryptedExtension 检查是否是加密文件扩展名
func isEncryptedExtension(ext string) bool {
	encryptedExts := []string{".cse", ".enc", ".cloudsync", ".csenc"}
	ext = strings.ToLower(ext)
	for _, encryptedExt := range encryptedExts {
		if ext == encryptedExt {
			return true
		}
	}
	return false
}

// 显示进度信息
func showProgress(current, total int64) {
	percentage := float64(current) * 100.0 / float64(total)
	fmt.Printf("\r进度: %.1f%% (%d/%d bytes)", percentage, current, total)
}

// 错误处理辅助函数
func handleError(message string, err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", message, err)
		os.Exit(1)
	}
}