# Golang + WebView2 透明异形窗口示例 (Transparent Shaped Window)

这是一个使用 Go 语言结合 WebView2 (Edge) 实现的 Windows 桌面应用示例。它展示了如何创建一个**背景完全透明**、**无边框**、**自定义形状（圆角+阴影）**的现代化窗口。

与传统的 `SetWindowRgn` 裁切不同，本项目利用了 WebView2 的 DirectComposition 透明特性，使得窗口形状完全由 HTML/CSS `border-radius` 和 `box-shadow` 决定，实现了真正的抗锯齿圆角和完美的窗口阴影。

## ✨ 特性

- **真·透明背景**：HTML 中的 `background: transparent` 直接透视到桌面。
- **自定义形状**：使用 CSS `border-radius` 实现圆角。
- **窗口阴影**：使用 CSS `box-shadow` 实现原生级的窗口阴影。
- **无边框交互**：自定义标题栏，支持拖拽移动、最小化、关闭。
- **Golang 后端**：使用 Go 处理系统调用和业务逻辑。

## 🚀 如何运行

### 前置要求

- Go 1.16+
- Windows 10/11
- WebView2 Runtime (通常 Windows 10/11 已内置)

提示：

```shellscript
# 使用 w64devkit 时如报错 "cannot parse _cgo_.o as ELF, Mach-O, PE"，
# 所以必须 使用 MSYS2，并且安装 mingw-w64 和配置golang 环境变量
pacman -S mingw-w64-ucrt-x86_64-gcc

export PATH="/c/Program Files/Go/bin:$PATH"

# 在 MSYS2 UCRT64 终端里执行此脚本编译 Windows 程序。
# 必须指定 GOOS=windows，否则会报 "build constraints exclude all Go files"。
```

### 编译与运行

```bash
# 1. 克隆项目
git clone https://github.com/ellermister/win-desktop.git
cd win-desktop

# 2. 编译 (隐藏控制台窗口)
GOOS=windows go build -ldflags="-H windowsgui" -o win-desktop.exe .

# 3. 运行
./win-desktop.exe

# 4. 再次编译时清理缓存
go clean -cache
GOOS=windows go build -ldflags="-H windowsgui" -o win-desktop.exe .
```

---

## 🛠️ 核心实现原理 (How it Works)

实现透明异形窗口的核心在于：**让 WebView2 控件背景透明** 以及 **配置正确的 Windows 窗口样式**。默认情况下，WebView2 会渲染白色或黑色背景，覆盖 HTML 的透明部分。

### 1. 修改 `webview_go` (C++ 层)

标准版的 `webview_go` 库不支持透明背景。我们需要修改其核心头文件 `webview.h`。

**关键修改点 A：在 WebView2 控制器创建后设置背景色为透明**
找到 `CreateCoreWebView2Controller` 的回调函数（通常在 `Invoke` 方法中），在获取到 `controller` 后立即设置：

```cpp
// 引入 ICoreWebView2Controller2 接口
ICoreWebView2Controller2 *controller2 = nullptr;
if (SUCCEEDED(controller->QueryInterface(IID_ICoreWebView2Controller2, (void **)&controller2))) {
    // 设置背景颜色为全透明 (Alpha = 0)
    COREWEBVIEW2_COLOR color = { 0, 0, 0, 0 };
    controller2->put_DefaultBackgroundColor(color);
    controller2->Release();
}
```

**关键修改点 B：移除 `SetLayeredWindowAttributes`**
在创建窗口时，我们需要开启 `WS_EX_LAYERED` 样式以支持透明，但**不能**调用 `SetLayeredWindowAttributes(hwnd, 0, 255, LWA_ALPHA)`。

- **原因**：当 Alpha 设为 255 时，该 API 会强制窗口变为不透明，导致 WebView2 的透明像素变黑。
- **做法**：在 `webview.h` 中找到 `CreateWindowEx` 后的 `SetLayeredWindowAttributes` 调用并将其注释掉。

**关键修改点 C：窗口样式配置 (Window Styles)**
在创建窗口 (`CreateWindowExW`) 时：

1. 移除 `WS_OVERLAPPEDWINDOW`，改为 `**WS_POPUP`**，以移除所有系统装饰（标题栏、边框）。
2. 添加 `**WS_EX_NOREDIRECTIONBITMAP`**，这是 DirectComposition 透明的关键。
3. 必须处理 `**WM_NCHITTEST`** 消息：如果 `DefWindowProc` 返回 `HTTRANSPARENT`（点击了透明区域），强制返回 `HTCLIENT`，否则鼠标点击会穿透窗口直接传给桌面。

```cpp
case WM_NCHITTEST: {
  LRESULT hit = DefWindowProcW(hwnd, msg, wp, lp);
  if (hit == HTTRANSPARENT) {
    return HTCLIENT; // 强制捕获透明区域的鼠标事件
  }
  return hit;
}
```

### 2. 窗口初始化 (Go)

在 `main.go` 中，建议使用 `webview.HintFixed` 来初始化窗口大小。

- **原因**：`HintNone` (默认) 会给窗口添加 `WS_THICKFRAME` (可调大小边框)。在透明模式下，这个不可见的系统边框可能会导致窗口边缘出现不需要的白色或黑色轮廓。
- **做法**：
  ```go
  w.SetSize(400, 300, webview.HintFixed)
  ```
  注意：这也意味着失去了拖拽边缘改变窗口大小的功能，需自行通过 HTML/JS 实现 Resizer。

### 3. 窗口样式与拖拽 (Go + C)

在 `win_style.c` 中，我们通过 Win32 API 辅助实现无边框窗口的交互。

- **拖拽窗口**：由于没有原生标题栏，我们需要拦截 HTML 元素的鼠标事件，在 Go 中调用 C 函数发送 Windows 消息：
  ```c
  // C 语言实现
  void window_drag(void* hwnd_ptr) {
      HWND hwnd = (HWND)hwnd_ptr;
      ReleaseCapture();
      SendMessage(hwnd, WM_NCLBUTTONDOWN, HTCAPTION, 0); // 模拟点击标题栏
  }
  ```
- **Go 绑定**：
  ```go
  w.Bind("window_drag", func() {
      C.window_drag(unsafe.Pointer(w.Window()))
  })
  ```
- **HTML 调用**：
  ```html
  <div class="title-bar" onmousedown="window_drag()">My App</div>
  ```

### 3. HTML/CSS 视觉设计

窗口的形状完全由 CSS 决定。

```css
body {
    background: transparent; /* 关键：透出桌面 */
}
.app-container {
    background: rgba(40, 44, 52, 0.95); /* 窗口实际背景 */
    border-radius: 12px; /* 圆角 */
    box-shadow: 0 4px 20px rgba(0,0,0,0.5); /* 阴影 */
    margin: 10px; /* 留出阴影空间 */
}
```

---

## ⚠️ 常见问题 (Troubleshooting)

**Q: 窗口背景是黑色的，不透明？**
A: 这通常是因为 `SetLayeredWindowAttributes` 被错误调用了。请检查 `webview.h` 或你的 C 代码中是否调用了 `SetLayeredWindowAttributes(..., 255, LWA_ALPHA)`。必须移除该调用，让 WebView2 接管 Alpha 通道。

**Q: 拖拽时窗口闪烁或卡顿？**
A: 使用 `SendMessage(..., HTCAPTION, ...)` 是最流畅的原生拖拽方式。避免使用 JavaScript 计算坐标来 `SetWindowPos`，那会导致严重的性能问题。

**Q: 只有 Windows 11 能用吗？**
A: 该方案在 Windows 10 (1809+) 和 Windows 11 上均可运行。Windows 7 不支持 WebView2 的透明背景特性。

---

## 📝 目录结构

- `main.go`: 主程序，包含 HTML 内容和 Go-JS 绑定。
- `win_style.c`: Windows API 辅助函数 (拖拽、最小化)。
- `webview_go/`: 本地修改版的 webview 库 (包含修改后的 `webview.h`)。
- `go.mod`: 指向本地 `webview_go` 的 replace 指令。

## License

MIT