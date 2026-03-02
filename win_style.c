// win_style.c - Windows 下通过 Win32 API 设置窗口为分层窗口并设置异形区域
// 仅 Windows 编译；非 Windows 提供空实现避免链接错误
#ifdef _WIN32

#include <windows.h>

// 拖动窗口：发送 HTCAPTION 消息，模拟鼠标按在标题栏
void window_drag(void* hwnd_ptr) {
    HWND hwnd = (HWND)hwnd_ptr;
    if (hwnd) {
        ReleaseCapture();
        SendMessage(hwnd, WM_NCLBUTTONDOWN, HTCAPTION, 0);
    }
}

// 最小化窗口
void window_minimize(void* hwnd_ptr) {
    HWND hwnd = (HWND)hwnd_ptr;
    if (hwnd) {
        ShowWindow(hwnd, SW_MINIMIZE);
    }
}

// 将窗口设为分层窗口并设置整体透明度（0=全透明，255=不透明）
// 同时可设置异形区域（圆角矩形）
void apply_window_style(void* hwnd_ptr, int width, int height, unsigned char alpha, int rounded) {
    HWND hwnd = (HWND)hwnd_ptr;
    if (!hwnd) return;

    LONG ex = GetWindowLongW(hwnd, GWL_EXSTYLE);
    SetWindowLongW(hwnd, GWL_EXSTYLE, ex | WS_EX_LAYERED);
    // 仅当 alpha < 255 时才调用 SetLayeredWindowAttributes
    // 当 alpha == 255 时，不调用此函数，以免破坏 WebView2 的每像素 Alpha 通道（Per-Pixel Alpha）
    if (alpha < 255) {
        SetLayeredWindowAttributes(hwnd, 0, alpha, LWA_ALPHA);
    }

    if (rounded > 0 && width > 0 && height > 0) {
        HRGN rgn = CreateRoundRectRgn(0, 0, width + 1, height + 1, (int)rounded, (int)rounded);
        if (rgn) {
            SetWindowRgn(hwnd, rgn, TRUE);
            DeleteObject(rgn);
        }
    }
}

#else

void apply_window_style(void* hwnd_ptr, int width, int height, unsigned char alpha, int rounded) {
    (void)hwnd_ptr;
    (void)width;
    (void)height;
    (void)alpha;
    (void)rounded;
}

void window_drag(void* hwnd_ptr) { (void)hwnd_ptr; }
void window_minimize(void* hwnd_ptr) { (void)hwnd_ptr; }

#endif
