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

// 将窗口恢复、置前并获取焦点
void window_focus(void* hwnd_ptr) {
    HWND hwnd = (HWND)hwnd_ptr;
    if (hwnd) {
        if (IsIconic(hwnd)) {
            ShowWindow(hwnd, SW_RESTORE);
        }
        BringWindowToTop(hwnd);
        SetForegroundWindow(hwnd);
        SetFocus(hwnd);
    }
}

// 将子窗口放在父窗口附近，模拟常见桌面应用弹窗位置
void window_position_near(void* hwnd_ptr, void* parent_ptr) {
    HWND hwnd = (HWND)hwnd_ptr;
    HWND parent = (HWND)parent_ptr;
    if (!hwnd || !parent) return;

    RECT parent_rect;
    RECT child_rect;
    if (!GetWindowRect(parent, &parent_rect) || !GetWindowRect(hwnd, &child_rect)) return;

    int child_width = child_rect.right - child_rect.left;
    int child_height = child_rect.bottom - child_rect.top;
    int x = parent_rect.left + 48;
    int y = parent_rect.top + 48;

    SetWindowPos(hwnd, HWND_TOP, x, y, child_width, child_height, SWP_SHOWWINDOW);
}

// 启用或禁用窗口，用于实现模态对话框
void window_set_enabled(void* hwnd_ptr, int enabled) {
    HWND hwnd = (HWND)hwnd_ptr;
    if (hwnd) {
        EnableWindow(hwnd, enabled ? TRUE : FALSE);
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
void window_focus(void* hwnd_ptr) { (void)hwnd_ptr; }
void window_position_near(void* hwnd_ptr, void* parent_ptr) { (void)hwnd_ptr; (void)parent_ptr; }
void window_set_enabled(void* hwnd_ptr, int enabled) { (void)hwnd_ptr; (void)enabled; }

#endif
