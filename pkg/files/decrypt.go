package files

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/synology-cloud-sync-decrypt-tool/syndecrypt-go/pkg/core"
	"github.com/synology-cloud-sync-decrypt-tool/syndecrypt-go/pkg/util"
)

// DecryptFile 解密单个文件
func DecryptFile(inputFileName, outputFileName string, config core.DecryptConfig) error {
	// 检查输入文件是否存在
	if !util.FileExists(inputFileName) {
		return fmt.Errorf("input file does not exist: %s", inputFileName)
	}

	// 检查输出文件是否已存在
	if util.FileExists(outputFileName) {
		return fmt.Errorf("output file already exists: %s", outputFileName)
	}

	// 确保输出目录存在
	outputDir := filepath.Dir(outputFileName)
	if err := util.EnsureDir(outputDir); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	// 打开输入文件
	inputFile, err := os.Open(inputFileName)
	if err != nil {
		return fmt.Errorf("failed to open input file: %v", err)
	}
	defer inputFile.Close()

	// 创建输出文件
	outputFile, err := os.Create(outputFileName)
	if err != nil {
		return fmt.Errorf("failed to create output file: %v", err)
	}
	defer outputFile.Close()

	// 执行解密，传入文件名用于错误报告
	if err := core.DecryptStreamWithFilename(inputFile, outputFile, config, inputFileName); err != nil {
		// 如果解密失败，删除输出文件
		outputFile.Close()
		os.Remove(outputFileName)
		return fmt.Errorf("decryption failed: %v", err)
	}

	return nil
}

// DecryptFiles 解密多个文件
func DecryptFiles(inputFiles []string, outputDir string, config core.DecryptConfig) error {
	for _, inputFile := range inputFiles {
		// 生成输出文件名
		baseName := filepath.Base(inputFile)
		outputFile := filepath.Join(outputDir, baseName)

		// 如果输入文件有加密扩展名，移除它
		if ext := filepath.Ext(baseName); ext == ".cse" || ext == ".enc" {
			outputFile = filepath.Join(outputDir, baseName[:len(baseName)-len(ext)])
		}

		fmt.Printf("Decrypting %s -> %s\n", inputFile, outputFile)

		if err := DecryptFile(inputFile, outputFile, config); err != nil {
			return fmt.Errorf("failed to decrypt %s: %v", inputFile, err)
		}
	}

	return nil
}

// DecryptDirectory 递归解密目录，返回详细的统计结果
func DecryptDirectory(inputDir, outputDir string, config core.DecryptConfig) (*DecryptResults, error) {
	results := NewDecryptResults()

	err := filepath.Walk(inputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过目录
		if info.IsDir() {
			return nil
		}

		// 计算相对路径
		relPath, err := filepath.Rel(inputDir, path)
		if err != nil {
			return err
		}

		// 生成输出路径
		outputPath := filepath.Join(outputDir, relPath)

		// 如果文件有加密扩展名，移除它
		if ext := filepath.Ext(outputPath); ext == ".cse" || ext == ".enc" {
			outputPath = outputPath[:len(outputPath)-len(ext)]
		}

		// 执行解密并记录结果
		result := decryptFileWithResult(path, outputPath, config)
		results.AddResult(result)

		return nil
	})

	if err != nil {
		return results, err
	}

	// 显示结果摘要（只在控制台打印，不保存到文件）
	results.PrintSummary()

	return results, nil
}

// decryptFileWithResult 解密单个文件并返回结果
func decryptFileWithResult(inputFileName, outputFileName string, config core.DecryptConfig) DecryptResult {
	startTime := time.Now()
	result := DecryptResult{
		InputFile:  inputFileName,
		OutputFile: outputFileName,
		StartTime:  startTime,
	}

	// 检查输入文件是否存在
	if !util.FileExists(inputFileName) {
		result.Error = fmt.Sprintf("input file does not exist: %s", inputFileName)
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime).String()
		fmt.Printf("  ❌ %s - %s\n", inputFileName, result.Error)
		return result
	}

	// 检查输出文件是否已存在
	if util.FileExists(outputFileName) {
		result.Error = fmt.Sprintf("output file already exists: %s", outputFileName)
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime).String()
		fmt.Printf("  ❌ %s - %s\n", inputFileName, result.Error)
		return result
	}

	// 确保输出目录存在
	outputDir := filepath.Dir(outputFileName)
	if err := util.EnsureDir(outputDir); err != nil {
		result.Error = fmt.Sprintf("failed to create output directory: %v", err)
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime).String()
		fmt.Printf("  ❌ %s - %s\n", inputFileName, result.Error)
		return result
	}

	// 执行解密（静默执行，只输出错误信息）
	err := DecryptFile(inputFileName, outputFileName, config)
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime).String()

	if err != nil {
		result.Error = err.Error()
		fmt.Printf("  ❌ %s - %s\n", inputFileName, result.Error)
		return result
	}

	// 获取文件大小
	if info, err := os.Stat(outputFileName); err == nil {
		result.FileSize = info.Size()
	}

	result.Success = true
	// 不再输出成功信息
	return result
}

// LoadPasswordFromFile 从文件加载密码
func LoadPasswordFromFile(passwordFile string) ([]byte, error) {
	return util.ReadBinaryFile(passwordFile)
}

// LoadPrivateKeyFromFile 从文件加载私钥
func LoadPrivateKeyFromFile(privateKeyFile string) ([]byte, error) {
	return util.ReadBinaryFile(privateKeyFile)
}

// LoadPublicKeyFromFile 从文件加载公钥
func LoadPublicKeyFromFile(publicKeyFile string) ([]byte, error) {
	return util.ReadBinaryFile(publicKeyFile)
}

// CopyFile 复制文件（用于测试）
func CopyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

// GetFileSize 获取文件大小
func GetFileSize(filename string) (int64, error) {
	info, err := os.Stat(filename)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

// IsEncryptedFile 检查是否是加密文件
func IsEncryptedFile(filename string) bool {
	ext := filepath.Ext(filename)
	return ext == ".cse" || ext == ".enc" || ext == ".cloudsync"
}

// GenerateOutputFilename 生成输出文件名
func GenerateOutputFilename(inputFile string) string {
	baseName := filepath.Base(inputFile)
	ext := filepath.Ext(baseName)

	// 移除加密扩展名
	if ext == ".cse" || ext == ".enc" || ext == ".cloudsync" {
		return baseName[:len(baseName)-len(ext)]
	}

	// 添加解密后缀
	return baseName + ".decrypted"
}

// ValidateConfig 验证解密配置
func ValidateConfig(config core.DecryptConfig) error {
	if config.Password == nil && config.PrivateKey == nil {
		return errors.New("either password or private key must be provided")
	}

	if config.Password != nil && config.PrivateKey != nil {
		return errors.New("cannot provide both password and private key")
	}

	return nil
}

// ProgressCallback 进度回调函数
type ProgressCallback func(current, total int64)

// DecryptFileWithProgress 带进度回调的解密
func DecryptFileWithProgress(inputFileName, outputFileName string, config core.DecryptConfig, callback ProgressCallback) error {
	// 获取文件大小
	fileSize, err := GetFileSize(inputFileName)
	if err != nil {
		return err
	}

	// 打开输入文件
	inputFile, err := os.Open(inputFileName)
	if err != nil {
		return err
	}
	defer inputFile.Close()

	// 创建进度跟踪读取器
	progressReader := &progressReader{
		reader:   inputFile,
		total:    fileSize,
		callback: callback,
	}

	// 创建输出文件
	outputFile, err := os.Create(outputFileName)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	// 执行解密
	if err := core.DecryptStream(progressReader, outputFile, config); err != nil {
		outputFile.Close()
		os.Remove(outputFileName)
		return err
	}

	return nil
}

// progressReader 跟踪读取进度
type progressReader struct {
	reader   io.Reader
	current  int64
	total    int64
	callback ProgressCallback
}

func (pr *progressReader) Read(p []byte) (int, error) {
	n, err := pr.reader.Read(p)
	pr.current += int64(n)
	if pr.callback != nil {
		pr.callback(pr.current, pr.total)
	}
	return n, err
}

// BatchDecryptOptions 批量解密选项
type BatchDecryptOptions struct {
	InputDir     string
	OutputDir    string
	Recursive    bool
	FilePattern  string
	Config       core.DecryptConfig
	ProgressFunc ProgressCallback
}

// BatchDecrypt 批量解密文件
func BatchDecrypt(options BatchDecryptOptions) error {
	results := NewDecryptResults()

	if options.Recursive {
		dirResults, err := DecryptDirectory(options.InputDir, options.OutputDir, options.Config)
		if err != nil {
			return err
		}
		// 使用目录解密的结果更新统计
		results.TotalFiles = dirResults.TotalFiles
		results.SuccessCount = dirResults.SuccessCount
		results.FailedCount = dirResults.FailedCount
		return nil
	}

	// 非递归模式：解密指定目录下的文件
	files, err := filepath.Glob(filepath.Join(options.InputDir, options.FilePattern))
	if err != nil {
		return err
	}

	fmt.Printf("找到 %d 个匹配文件\n", len(files))

	for i, file := range files {
		if !IsEncryptedFile(file) {
			continue
		}

		outputFile := filepath.Join(options.OutputDir, GenerateOutputFilename(file))

		// 显示进度
		if options.ProgressFunc != nil {
			options.ProgressFunc(int64(i), int64(len(files)))
		}

		// 执行解密并记录结果
		result := decryptFileWithResult(file, outputFile, options.Config)
		results.AddResult(result)
	}

	// 显示结果摘要（只在控制台打印，不保存到文件）
	results.PrintSummary()

	return nil
}