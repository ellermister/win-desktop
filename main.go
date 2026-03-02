//go:build windows

package main

/*
#cgo windows CFLAGS: -D_WIN32
#cgo windows LDFLAGS: -lgdi32
extern void apply_window_style(void* hwnd_ptr, int width, int height, unsigned char alpha, int rounded);
extern void window_drag(void* hwnd_ptr);
extern void window_minimize(void* hwnd_ptr);
*/
import "C"
import (
	"os"
	"unsafe"

	webview "github.com/webview/webview_go"
)

func init() {
	// 必须在任何 WebView2 初始化之前设置，否则会一直白屏
	_ = os.Setenv("WEBVIEW2_DEFAULT_BACKGROUND_COLOR", "0x00000000")
}

func main() {
	w := webview.New(true)
	defer w.Destroy()

	w.SetTitle("异形/透明窗口示例")
	w.SetSize(400, 300, webview.HintFixed)

	// 绑定 Go 函数供 JS 调用
	w.Bind("window_drag", func() {
		hwnd := w.Window()
		if hwnd != nil {
			C.window_drag(unsafe.Pointer(hwnd))
		}
	})

	w.Bind("window_minimize", func() {
		hwnd := w.Window()
		if hwnd != nil {
			C.window_minimize(unsafe.Pointer(hwnd))
		}
	})

	w.Bind("window_close", func() {
		w.Terminate()
	})

	// 页面背景透明（若底层 WebView 也支持透明，会透出桌面）
	// HTML+CSS 实现圆角和阴影，模拟窗口边框
	html := `<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <style>
    body, html {
      margin: 0; padding: 0;
      width: 100%; height: 100%;
      background: transparent; /* 关键：背景完全透明 */
      overflow: hidden; /* 防止滚动条出现 */
      font-family: "Segoe UI", sans-serif;
    }
    
    /* 主容器：模拟窗口本体 */
    .app-container {
      margin: 10px; /* 给阴影留出空间 */
      width: calc(100% - 20px);
      height: calc(100% - 20px);
      background: rgba(40, 44, 52, 0.95); /* 深色半透明背景 */
      border-radius: 12px; /* 圆角 */
      box-shadow: 0 4px 20px rgba(0,0,0,0.5); /* 窗口阴影 */
      display: flex;
      flex-direction: column;
      border: 1px solid rgba(255,255,255,0.1); /* 微弱的边框高光 */
      color: #fff;
    }

    /* 标题栏 */
    .title-bar {
      height: 32px;
      display: flex;
      justify-content: space-between;
      align-items: center;
      padding: 0 10px;
      background: rgba(255,255,255,0.05);
      border-top-left-radius: 12px;
      border-top-right-radius: 12px;
      user-select: none;
      /* 拖动区域 */
      cursor: default;
    }

    .title-text {
      font-size: 13px;
      flex-grow: 1; /* 占据剩余空间 */
      display: flex;
      align-items: center;
      height: 100%;
    }

    .window-controls {
      display: flex;
      gap: 8px;
    }

    .control-btn {
      width: 12px;
      height: 12px;
      border-radius: 50%;
      border: none;
      cursor: pointer;
      padding: 0;
      transition: transform 0.1s;
    }
    
    .btn-close { background: #ff5f56; }
    .btn-min { background: #ffbd2e; }
    .btn-max { background: #27c93f; } /* 仅展示，未绑定功能 */

    .control-btn:hover { filter: brightness(1.2); }
    .control-btn:active { transform: scale(0.9); }

    /* 内容区域 */
    .content {
      flex: 1;
      padding: 20px;
      display: flex;
      flex-direction: column;
      justify-content: center;
      align-items: center;
      text-align: center;
    }

    h2 { margin: 0 0 10px 0; font-weight: 300; }
    p { font-size: 14px; color: #ccc; line-height: 1.6; }
    
    button.action-btn {
      margin-top: 20px;
      padding: 8px 20px;
      background: #61dafb;
      border: none;
      border-radius: 4px;
      color: #282c34;
      font-weight: bold;
      cursor: pointer;
      transition: background 0.2s;
    }
    button.action-btn:hover { background: #4fa8d1; }

  </style>
</head>
<body>
  <div class="app-container">
    <div class="title-bar">
      <!-- 拖动区域：绑定到文字区域 -->
      <div class="title-text" onmousedown="window_drag()">My Transparent App</div>
      <div class="window-controls">
        <button class="control-btn btn-min" onclick="window_minimize()" title="Minimize"></button>
        <button class="control-btn btn-max" title="Maximize"></button>
        <button class="control-btn btn-close" onclick="window_close()" title="Close"></button>
      </div>
    </div>
    <div class="content">
      <h2>Hello, Transparency!</h2>
      <p>This is a custom-shaped window rendered with HTML & CSS.</p>
      <p>Background is transparent, shadows are real.</p>
      <button class="action-btn" onclick="alert('Hello from Go!')">Click Me</button>
    </div>
  </div>
</body>
</html>`
	w.SetHtml(html)

	/*
		// 在主线程（消息循环）中执行：给窗口加 WS_EX_LAYERED 并设置圆角区域
		w.Dispatch(func() {
			hwnd := w.Window()
			if hwnd != nil {
				C.apply_window_style(unsafe.Pointer(hwnd), 400, 300, 255, 0)
				// 参数：HWND, 宽, 高, 整体透明度(255=不透明), 圆角半径(0=矩形)
			}
		})
	*/

	w.Run()
}
