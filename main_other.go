//go:build !windows

package main

func main() {
	// 异形/透明窗口示例仅支持 Windows，需使用 Go + C 调用 Win32 API
	// 在 Windows 下编译运行: go build -o win-desktop.exe .
	println("此示例请在 Windows 下编译运行")
}
