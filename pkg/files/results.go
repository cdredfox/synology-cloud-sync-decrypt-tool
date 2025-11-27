package files

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// DecryptResult 记录单个文件的解密结果
type DecryptResult struct {
	InputFile    string    `json:"input_file"`
	OutputFile   string    `json:"output_file"`
	Success      bool      `json:"success"`
	Error        string    `json:"error,omitempty"`
	FileSize     int64     `json:"file_size"`
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`
	Duration     string    `json:"duration"`
	// 目录处理时的统计信息
	FileCount    int       `json:"file_count,omitempty"`
	SuccessCount int       `json:"success_count,omitempty"`
	FailedCount  int       `json:"failed_count,omitempty"`
}

// DecryptResults 记录批量解密的结果
type DecryptResults struct {
	mu            sync.Mutex
	Results       []DecryptResult `json:"results"`
	TotalFiles    int             `json:"total_files"`
	SuccessCount  int             `json:"success_count"`
	FailedCount   int             `json:"failed_count"`
	StartTime     time.Time       `json:"start_time"`
	EndTime       time.Time       `json:"end_time"`
	TotalDuration string          `json:"total_duration"`
}

// NewDecryptResults 创建新的结果记录器
func NewDecryptResults() *DecryptResults {
	return &DecryptResults{
		Results:   make([]DecryptResult, 0),
		StartTime: time.Now(),
	}
}

// AddResult 添加一个解密结果
func (dr *DecryptResults) AddResult(result DecryptResult) {
	dr.mu.Lock()
	defer dr.mu.Unlock()

	dr.Results = append(dr.Results, result)
	if result.Success {
		dr.SuccessCount++
	} else {
		dr.FailedCount++
	}
	dr.TotalFiles++
}

// Finish 完成结果记录
func (dr *DecryptResults) Finish() {
	dr.mu.Lock()
	defer dr.mu.Unlock()

	dr.EndTime = time.Now()
	duration := dr.EndTime.Sub(dr.StartTime)
	dr.TotalDuration = duration.String()
}

// PrintSummary 打印结果摘要
func (dr *DecryptResults) PrintSummary() {
	// 确保总耗时已计算
	dr.Finish()

	// 总是显示基本的统计信息和总耗时
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("解密完成报告")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("总文件数: %d\n", dr.TotalFiles)
	fmt.Printf("成功: %d\n", dr.SuccessCount)
	fmt.Printf("失败: %d\n", dr.FailedCount)
	fmt.Printf("总耗时: %s\n", dr.TotalDuration)
	fmt.Println(strings.Repeat("=", 60))

	// 只有在有失败时才显示失败文件列表
	if dr.FailedCount > 0 {
		fmt.Println("\n失败文件列表:")
		for _, result := range dr.Results {
			if !result.Success {
				fmt.Printf("  ❌ %s - %s\n", result.InputFile, result.Error)
			}
		}
	}
	// 不显示成功文件列表（保持静默）
}


// SaveReport 保存详细报告到文件
func (dr *DecryptResults) SaveReport(outputDir string) error {
	dr.Finish()

	reportFile := filepath.Join(outputDir, "decryption_report.txt")
	file, err := os.Create(reportFile)
	if err != nil {
		return fmt.Errorf("failed to create report file: %v", err)
	}
	defer file.Close()

	// 写入报告标题
	fmt.Fprintf(file, "Synology Cloud Sync 解密报告\n")
	fmt.Fprintf(file, "生成时间: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Fprintf(file, "%s\n\n", strings.Repeat("=", 60))

	// 写入摘要
	fmt.Fprintf(file, "摘要统计:\n")
	fmt.Fprintf(file, "  总文件数: %d\n", dr.TotalFiles)
	fmt.Fprintf(file, "  成功: %d\n", dr.SuccessCount)
	fmt.Fprintf(file, "  失败: %d\n", dr.FailedCount)
	fmt.Fprintf(file, "  总耗时: %s\n\n", dr.TotalDuration)

	// 写入失败文件
	if dr.FailedCount > 0 {
		fmt.Fprintf(file, "失败文件:\n")
		for _, result := range dr.Results {
			if !result.Success {
				fmt.Fprintf(file, "  ❌ %s\n", result.InputFile)
				fmt.Fprintf(file, "     错误: %s\n", result.Error)
				fmt.Fprintf(file, "     时间: %s\n\n", result.Duration)
			}
		}
	}

	// 写入成功文件
	if dr.SuccessCount > 0 {
		fmt.Fprintf(file, "成功文件:\n")
		for _, result := range dr.Results {
			if result.Success {
				fmt.Fprintf(file, "  ✅ %s\n", result.InputFile)
				fmt.Fprintf(file, "     输出: %s\n", result.OutputFile)
				fmt.Fprintf(file, "     大小: %d 字节\n", result.FileSize)
				fmt.Fprintf(file, "     时间: %s\n\n", result.Duration)
			}
		}
	}

	fmt.Printf("\n详细报告已保存到: %s\n", reportFile)
	return nil
}

// PrintProgress 打印进度信息
func (dr *DecryptResults) PrintProgress(currentFile string, current, total int) {
	percentage := float64(current) * 100.0 / float64(total)
	fmt.Printf("\r进度: %.1f%% (%d/%d) - 当前文件: %s", percentage, current, total, filepath.Base(currentFile))
}

// GetSuccessRate 获取成功率
func (dr *DecryptResults) GetSuccessRate() float64 {
	if dr.TotalFiles == 0 {
		return 0.0
	}
	return float64(dr.SuccessCount) * 100.0 / float64(dr.TotalFiles)
}