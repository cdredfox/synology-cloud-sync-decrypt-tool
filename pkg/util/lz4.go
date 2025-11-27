package util

import (
	"fmt"
	"io"
	"os/exec"
	"sync"
)

type Lz4Decompressor struct {
	cmd       *exec.Cmd
	stdin     io.WriteCloser
	stdout    io.ReadCloser
	handler   func([]byte)
	filename  string
	mu        sync.Mutex
	isClosed  bool
}

func NewLz4Decompressor(decompressedChunkHandler func([]byte)) (*Lz4Decompressor, error) {
	return NewLz4DecompressorWithFilename(decompressedChunkHandler, "")
}

func NewLz4DecompressorWithFilename(decompressedChunkHandler func([]byte), filename string) (*Lz4Decompressor, error) {
	cmd := exec.Command("lz4", "-d")

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdin pipe: %v", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		stdin.Close()
		return nil, fmt.Errorf("failed to create stdout pipe: %v", err)
	}

	if err := cmd.Start(); err != nil {
		stdin.Close()
		stdout.Close()
		return nil, fmt.Errorf("failed to start lz4: %v", err)
	}

	decomp := &Lz4Decompressor{
		cmd:      cmd,
		stdin:    stdin,
		stdout:   stdout,
		handler:  decompressedChunkHandler,
		filename: filename,
	}

	// 启动 goroutine 读取解压后的数据
	go decomp.readOutput()

	return decomp, nil
}

func (l *Lz4Decompressor) readOutput() {
	buffer := make([]byte, 64*1024) // 64KB buffer
	for {
		n, err := l.stdout.Read(buffer)
		if err != nil {
			if err != io.EOF {
				// 检查是否是关闭时的正常错误，避免误报
				l.mu.Lock()
				isClosed := l.isClosed
				l.mu.Unlock()

				// 只有在非关闭状态下才报告错误，避免竞争条件的误报
				if !isClosed {
					if l.filename != "" {
						fmt.Printf("❌ 文件 %s - lz4解压错误: %v\n", l.filename, err)
					} else {
						fmt.Printf("❌ lz4解压错误: %v\n", err)
					}
				}
			}
			return
		}
		if n > 0 {
			l.handler(buffer[:n])
		}
	}
}

func (l *Lz4Decompressor) Write(data []byte) error {
	_, err := l.stdin.Write(data)
	return err
}

func (l *Lz4Decompressor) Close() error {
	// 设置关闭状态，避免竞争条件的误报
	l.mu.Lock()
	l.isClosed = true
	l.mu.Unlock()

	// 关闭输入管道
	if err := l.stdin.Close(); err != nil {
		return fmt.Errorf("failed to close stdin: %v", err)
	}

	// 等待命令完成
	if err := l.cmd.Wait(); err != nil {
		return fmt.Errorf("lz4 command failed: %v", err)
	}

	return nil
}

// Base64Decode 解码 base64 字符串
func Base64Decode(data string) ([]byte, error) {
	// 实现 base64 解码
	return nil, fmt.Errorf("not implemented")
}