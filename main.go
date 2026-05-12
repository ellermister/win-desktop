//go:build windows

package main

/*
#cgo windows CFLAGS: -D_WIN32
#cgo windows LDFLAGS: -lgdi32
extern void apply_window_style(void* hwnd_ptr, int width, int height, unsigned char alpha, int rounded);
extern void window_drag(void* hwnd_ptr);
extern void window_minimize(void* hwnd_ptr);
extern void window_focus(void* hwnd_ptr);
extern void window_position_near(void* hwnd_ptr, void* parent_ptr);
extern void window_set_enabled(void* hwnd_ptr, int enabled);
*/
import "C"
import (
	"os"
	"runtime"
	"sync"
	"unsafe"

	webview "github.com/webview/webview_go"
)

var (
	aboutMu     sync.Mutex
	aboutOpened bool
	aboutHWND   unsafe.Pointer

	confirmMu     sync.Mutex
	confirmOpened bool
	confirmHWND   unsafe.Pointer
)

func init() {
	// 必须在任何 WebView2 初始化之前设置，否则会一直白屏
	_ = os.Setenv("WEBVIEW2_DEFAULT_BACKGROUND_COLOR", "0x00000000")
}

func main() {
	w := webview.New(true)
	defer w.Destroy()

	w.SetTitle("异形/透明窗口示例")
	// 窗口边距, 用于阴影，避免阴影被裁剪显示为直角
	marginWidth := 12
	w.SetSize(460+marginWidth*2, 360+marginWidth*2, webview.HintFixed)

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

	w.Bind("open_about", func() {
		openAboutWindow(w.Window())
	})

	w.Bind("open_confirm", func() {
		openConfirmWindow(w.Window())
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
      margin: 12px; /* 给阴影留出空间 */
      width: calc(100% - 24px);
      height: calc(100% - 24px);
      box-shadow: 0 0 12px rgba(0,0,0,0.95); /* 四周均匀、随圆角轮廓扩散的阴影 */
      background: rgba(40, 44, 52, 0.95); /* 深色半透明背景 */
      border-radius: 12px; /* 圆角 */
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
      padding: 26px 28px 30px;
      display: flex;
      flex-direction: column;
      justify-content: center;
      align-items: center;
      text-align: center;
      gap: 10px;
    }

    h2 { margin: 0; font-weight: 300; }
    p { margin: 0; font-size: 14px; color: #ccc; line-height: 1.6; }

    .button-grid {
      width: 100%;
      display: grid;
      grid-template-columns: repeat(2, minmax(0, 1fr));
      gap: 10px;
      margin-top: 12px;
    }
    
    button.action-btn {
      min-height: 36px;
      padding: 8px 14px;
      background: #61dafb;
      border: none;
      border-radius: 4px;
      color: #282c34;
      font-weight: bold;
      cursor: pointer;
      transition: background 0.2s;
    }
    button.action-btn:hover { background: #4fa8d1; }

    .secondary-btn {
      background: rgba(255,255,255,0.16) !important;
      color: #fff !important;
    }

    .counter-status {
      min-height: 20px;
      color: #8be9fd;
      font-size: 13px;
    }

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
      <div class="counter-status" id="counterStatus">Click Me 计数：0</div>
      <div class="button-grid">
        <button class="action-btn" onclick="alert('Hello from Go!')">原始 Alert</button>
        <button class="action-btn" id="counterButton" onclick="incrementCounter()">Click Me +1</button>
        <button class="action-btn secondary-btn" onclick="open_about()">关于</button>
        <button class="action-btn secondary-btn" onclick="open_confirm()">确认型窗口</button>
      </div>
    </div>
  </div>
  <script>
    let clickCount = 0;
    function incrementCounter() {
      clickCount += 1;
      document.getElementById('counterButton').textContent = 'Click Me：' + clickCount;
      document.getElementById('counterStatus').textContent = 'Click Me 计数：' + clickCount;
    }
  </script>
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

func openAboutWindow(parent unsafe.Pointer) {
	aboutMu.Lock()
	if aboutOpened {
		hwnd := aboutHWND
		aboutMu.Unlock()
		if hwnd != nil {
			C.window_focus(unsafe.Pointer(hwnd))
		}
		return
	}
	aboutOpened = true
	aboutMu.Unlock()

	go func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()

		about := webview.New(true)
		defer about.Destroy()
		defer func() {
			aboutMu.Lock()
			aboutOpened = false
			aboutHWND = nil
			aboutMu.Unlock()
		}()

		about.SetTitle("关于")
		marginWidth := 12
		about.SetSize(360+marginWidth*2, 240+marginWidth*2, webview.HintFixed)
		hwnd := about.Window()
		aboutMu.Lock()
		aboutHWND = hwnd
		aboutMu.Unlock()
		if hwnd != nil {
			C.window_position_near(unsafe.Pointer(hwnd), unsafe.Pointer(parent))
		}

		about.Bind("about_drag", func() {
			hwnd := about.Window()
			if hwnd != nil {
				C.window_drag(unsafe.Pointer(hwnd))
			}
		})

		about.Bind("about_minimize", func() {
			hwnd := about.Window()
			if hwnd != nil {
				C.window_minimize(unsafe.Pointer(hwnd))
			}
		})

		about.Bind("about_close", func() {
			about.Terminate()
		})

		about.Bind("about_interact", func() string {
			return "来自 Go 的二级窗口交互成功"
		})

		about.SetHtml(`<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <style>
    body, html {
      margin: 0; padding: 0;
      width: 100%; height: 100%;
      background: transparent;
      overflow: hidden;
      font-family: "Segoe UI", sans-serif;
    }

    .app-container {
      margin: 12px;
      width: calc(100% - 24px);
      height: calc(100% - 24px);
      box-shadow: 0 0 12px rgba(0,0,0,0.95);
      background: rgba(32, 36, 44, 0.96);
      border-radius: 12px;
      display: flex;
      flex-direction: column;
      border: 1px solid rgba(255,255,255,0.12);
      color: #fff;
    }

    .title-bar {
      height: 32px;
      display: flex;
      justify-content: space-between;
      align-items: center;
      padding: 0 10px;
      background: rgba(255,255,255,0.06);
      border-top-left-radius: 12px;
      border-top-right-radius: 12px;
      user-select: none;
    }

    .title-text {
      font-size: 13px;
      flex-grow: 1;
      display: flex;
      align-items: center;
      height: 100%;
      cursor: default;
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
    }

    .btn-close { background: #ff5f56; }
    .btn-min { background: #ffbd2e; }

    .content {
      flex: 1;
      padding: 22px;
      display: flex;
      flex-direction: column;
      justify-content: center;
      align-items: center;
      text-align: center;
      gap: 10px;
    }

    h2 { margin: 0; font-weight: 300; }
    p { margin: 0; font-size: 14px; color: #ccc; line-height: 1.6; }

    .action-btn {
      margin-top: 8px;
      padding: 8px 18px;
      background: #61dafb;
      border: none;
      border-radius: 4px;
      color: #282c34;
      font-weight: bold;
      cursor: pointer;
    }

    .status {
      min-height: 20px;
      color: #8be9fd;
      font-size: 13px;
    }
  </style>
</head>
<body>
  <div class="app-container">
    <div class="title-bar">
      <div class="title-text" onmousedown="about_drag()">关于 win-desktop</div>
      <div class="window-controls">
        <button class="control-btn btn-min" onclick="about_minimize()" title="Minimize"></button>
        <button class="control-btn btn-close" onclick="about_close()" title="Close"></button>
      </div>
    </div>
    <div class="content">
      <h2>win-desktop</h2>
      <p>这是一个由主窗口打开的二级透明窗口。</p>
      <p>它拥有独立的 WebView、窗口拖动、最小化、关闭和 JS/Go 交互。</p>
      <button class="action-btn" onclick="testInteraction()">测试交互</button>
      <div id="status" class="status"></div>
    </div>
  </div>
  <script>
    async function testInteraction() {
      const message = await about_interact();
      document.getElementById('status').textContent = message;
    }
  </script>
</body>
</html>`)

		about.Run()
	}()
}

func openConfirmWindow(parent unsafe.Pointer) {
	confirmMu.Lock()
	if confirmOpened {
		hwnd := confirmHWND
		confirmMu.Unlock()
		if hwnd != nil {
			C.window_focus(unsafe.Pointer(hwnd))
		}
		return
	}
	confirmOpened = true
	confirmMu.Unlock()

	if parent != nil {
		C.window_set_enabled(unsafe.Pointer(parent), 0)
	}

	go func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
		defer func() {
			confirmMu.Lock()
			confirmOpened = false
			confirmHWND = nil
			confirmMu.Unlock()

			if parent != nil {
				C.window_set_enabled(unsafe.Pointer(parent), 1)
				C.window_focus(unsafe.Pointer(parent))
			}
		}()

		confirm := webview.New(true)
		defer confirm.Destroy()

		confirm.SetTitle("确认操作")
		marginWidth := 12
		confirm.SetSize(380+marginWidth*2, 230+marginWidth*2, webview.HintFixed)

		confirmMu.Lock()
		hwnd := confirm.Window()
		confirmHWND = hwnd
		confirmMu.Unlock()
		if hwnd != nil {
			C.window_position_near(unsafe.Pointer(hwnd), unsafe.Pointer(parent))
			C.window_focus(unsafe.Pointer(hwnd))
		}

		confirm.Bind("confirm_drag", func() {
			hwnd := confirm.Window()
			if hwnd != nil {
				C.window_drag(unsafe.Pointer(hwnd))
			}
		})

		confirm.Bind("confirm_ok", func() {
			confirm.Terminate()
		})

		confirm.Bind("confirm_cancel", func() {
			confirm.Terminate()
		})

		confirm.Bind("confirm_close", func() {
			confirm.Terminate()
		})

		confirm.SetHtml(`<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <style>
    body, html {
      margin: 0; padding: 0;
      width: 100%; height: 100%;
      background: transparent;
      overflow: hidden;
      font-family: "Segoe UI", sans-serif;
    }

    .app-container {
      margin: 12px;
      width: calc(100% - 24px);
      height: calc(100% - 24px);
      box-shadow: 0 0 12px rgba(0,0,0,0.95);
      background: rgba(44, 38, 36, 0.97);
      border-radius: 12px;
      display: flex;
      flex-direction: column;
      border: 1px solid rgba(255,255,255,0.12);
      color: #fff;
    }

    .title-bar {
      height: 32px;
      display: flex;
      justify-content: space-between;
      align-items: center;
      padding: 0 10px;
      background: rgba(255,255,255,0.06);
      border-top-left-radius: 12px;
      border-top-right-radius: 12px;
      user-select: none;
      cursor: default;
    }

    .dialog-title-text {
      display: inline-flex;
      align-items: center;
      height: 100%;
      padding-right: 16px;
      cursor: default;
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
    }

    .btn-close { background: #ff5f56; }

    .content {
      flex: 1;
      padding: 22px;
      display: flex;
      flex-direction: column;
      justify-content: center;
      align-items: center;
      text-align: center;
      gap: 12px;
    }

    h2 { margin: 0; font-weight: 300; }
    p { margin: 0; font-size: 14px; color: #ddd; line-height: 1.6; }

    .actions {
      display: flex;
      gap: 12px;
      margin-top: 8px;
    }

    button {
      padding: 8px 20px;
      border: none;
      border-radius: 4px;
      font-weight: bold;
      cursor: pointer;
    }

    .ok { background: #61dafb; color: #282c34; }
    .cancel { background: rgba(255,255,255,0.16); color: #fff; }
  </style>
</head>
<body>
  <div class="app-container">
    <div class="title-bar">
      <div class="dialog-title-text" onmousedown="confirm_drag()">确认操作</div>
      <div class="window-controls">
        <button class="control-btn btn-close" onclick="confirm_close()" title="Close"></button>
      </div>
    </div>
    <div class="content">
      <h2>确认型窗口</h2>
      <p>这是一个模态对话框。关闭它之前，主窗口会被禁用，无法继续交互。</p>
      <div class="actions">
        <button class="ok" onclick="confirm_ok()">确认</button>
        <button class="cancel" onclick="confirm_cancel()">取消/关闭</button>
      </div>
    </div>
  </div>
</body>
</html>`)

		confirm.Run()
	}()
}
