# Go + C 混编实现 webview 异形/透明窗口

## 思路概览

- **可以**通过 Golang + 部分 C 混编来改“窗口”行为，实现异形窗口（圆角、不规则形状）和分层透明。
- 分为两种程度：
  1. **只改窗口（不改 webview 源码）**：用 CGO 在创建窗口后调 Win32 API（本示例做法）。
  2. **改 webview 的 C/C++ 层**：在 WebView2 创建时设置 `DefaultBackgroundColor` 透明，才能做到“页面背景真正透明”。

---

## 本示例（方案 A：只改窗口）

- **`win_style.c`**：C 里对传入的 `HWND` 做：
  - `SetWindowLong(..., WS_EX_LAYERED)`：窗口设为分层窗口
  - `SetLayeredWindowAttributes(..., LWA_ALPHA)`：整体透明度（可选）
  - `CreateRoundRectRgn` + `SetWindowRgn`：圆角矩形异形窗口
- **`main.go`**：用 webview 创建窗口后，通过 `Dispatch` 在主线程里拿到 `Window()` 的 HWND，调用 C 的 `apply_window_style`。

效果：窗口是圆角矩形；若只做“整窗半透明”可改 C 里 alpha。  
限制：**WebView 内容区域默认仍是白底**，因为 webview 的 C 层没有把 WebView2 的 `DefaultBackgroundColor` 设为透明。

---

## 若要“页面背景也透明”（方案 B：改 webview C++ 层）

需要动 webview 库里创建 WebView2 的那段 C++ 代码（webview_go 的 `libs` 里，或上游 [webview/webview](https://github.com/webview/webview) 的 Windows 实现）：

1. **在 WebView2 创建完成的回调里**拿到 `ICoreWebView2Controller`，再：
   - `QueryInterface` 取 `ICoreWebView2Controller2`
   - 调用 `put_DefaultBackgroundColor`，传入 `COREWEBVIEW2_COLOR{A=0}`（完全透明）
2. **创建窗口时**在 `CreateWindowEx` 里加上 `WS_EX_LAYERED`，或创建后用 `SetWindowLong` 加上，这样分层+透明才能一起生效。

这样 Go 侧不用改，只要用你改过的 webview 库（例如 go mod replace 指向本地 fork）重新编译即可。

---

## 编译与运行（Windows）

```bash
cd e:\projects\win-desktop
go mod tidy
go build -ldflags="-H windowsgui" -o win-desktop.exe .
.\win-desktop.exe
```

若 `go mod tidy` 因网络失败，可先手动执行一次 `go get github.com/webview/webview_go` 再试。

---

## 异形窗口扩展

在 `win_style.c` 里可以换成或叠加其他区域 API，例如：

- `CreateEllipticRgn`：椭圆/圆形窗口
- `CreatePolygonRgn`：多边形
- `CombineRgn`：组合多个区域

均通过 `SetWindowRgn(hwnd, rgn, TRUE)` 应用到窗口。
