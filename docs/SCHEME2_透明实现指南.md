# 方案二：基于 HTML 背景透明的实现指南

## 和方案一的区别

| 方案一（当前） | 方案二（目标） |
|----------------|----------------|
| 用 `SetWindowRgn` 把窗口裁成异形 | 窗口仍可异形，但**不依赖裁切** |
| WebView 内容区是**不透明**的（白底） | WebView **背景透明**，HTML 里 `background: transparent` 会透出桌面 |
| 无法实现“只有部分区域有内容、其余透明” | 可实现**真正基于 HTML/CSS 的透明/异形** |

所以：**只有方案二才能实现“HTML 背景透明 + 自定义形状”的透明窗口**。

---

## 方案二容易实现吗？

**结论：可行，但不算“开箱即用”，需要改 webview 的 C/C++ 层并重新编译。**

- **难度**：中等。改动点集中、代码量小（约 20～40 行），但需要：
  1. 能拿到并修改 webview 的源码（fork 或 vendor）
  2. 在 WebView2 **创建完成的回调**里多写几行 COM 调用
  3. 确保编译时链到 WebView2 SDK（一般 webview_go 已带）

- **主要步骤**：
  1. 定位：在 `webview_go` 的 **libs/webview/include/webview.h** 里找到 `CreateCoreWebView2Controller` 的**完成回调**（即收到 `ICoreWebView2Controller*` 的地方）。
  2. 在该回调里拿到 controller 后，增加：
     - `QueryInterface` 取 `ICoreWebView2Controller2`
     - 调用 `put_DefaultBackgroundColor`，传入 **Alpha = 0**（完全透明）。
  3. 窗口样式：创建窗口时加上 `WS_EX_LAYERED`（或创建后用 `SetWindowLong` 加），这样透明背景才能正确显示。

下面按“要改什么、改在哪、怎么改”给出具体做法。

---

## 具体要改的代码位置

webview_go 的 Windows 实现都在 **libs/webview/include/webview.h** 里（单头文件，约 3600 行）。需要改两处：

### 1. 找到 WebView2 创建完成的回调

在 `webview.h` 里搜索（或按行号附近找）：

- `CreateCoreWebView2Controller`
- 或完成回调函数里接收 `ICoreWebView2Controller *controller` / `*result` 的地方

通常形如（伪代码）：

```c
// 回调：CreateCoreWebView2Controller 完成时被调用
HRESULT Invoke(HRESULT errorCode, ICoreWebView2Controller *controller) {
  if (FAILED(errorCode) || !controller) return errorCode;
  // ... 现有代码：保存 controller、设置 Bounds 等 ...
}
```

就在这个 `Invoke` 里、在**使用 controller 做别的事之前**，加上下面“插入的代码”。

### 2. 在回调里插入：设置透明背景

在拿到 `controller` 且未释放的同一回调里加入（C++）：

```c
// ----- 插入开始：让 WebView2 背景透明，便于 HTML 透明/异形窗口 -----
{
  ICoreWebView2Controller2 *controller2 = NULL;
  HRESULT hr = controller->lpVtbl->QueryInterface(controller,
      &IID_ICoreWebView2Controller2, (void**)&controller2);
  if (SUCCEEDED(hr) && controller2) {
    COREWEBVIEW2_COLOR color = { 0, 0, 0, 0 };  // A=0 表示完全透明
    controller2->lpVtbl->put_DefaultBackgroundColor(controller2, color);
    controller2->lpVtbl->Release(controller2);
  }
}
// ----- 插入结束 -----
```

注意：

- 若用的是 C++（`webview.h` 里可能是 C++），接口调用可能是 `controller->QueryInterface(...)` 和 `controller2->put_DefaultBackgroundColor(...)`，逻辑相同。
- `COREWEBVIEW2_COLOR` 和 `IID_ICoreWebView2Controller2` 来自 WebView2 SDK，一般 webview 已包含对应头文件；若编译报错，在包含 WebView2 头文件后再试。

### 3. 窗口加上分层样式（WS_EX_LAYERED）

这样系统才会按透明通道合成窗口。

- 若 `webview.h` 里用 `CreateWindowEx` 创建窗口：在 `dwExStyle` 里加上 `WS_EX_LAYERED`（0x00080000）。
- 若窗口已先创建，可在创建后立刻用 `SetWindowLong(hwnd, GWL_EXSTYLE, exStyle | WS_EX_LAYERED)` 补上。

你当前项目里的 **win_style.c** 已经在用 `WS_EX_LAYERED`，若采用方案二，webview 内部也建议加上，避免未调用你 C 代码时窗口仍不透明。

---

## 如何获得并修改 webview 源码

1. **克隆带 submodule 的 webview_go**  
   ```bash
   git clone --recursive https://github.com/webview/webview_go.git
   cd webview_go
   ```
2. **改 libs/webview/include/webview.h**  
   按上面两处修改（完成回调里加透明背景 + 窗口加 WS_EX_LAYERED）。
3. **用本仓库引用你的 fork**  
   在你自己的项目里：
   ```bash
   go mod edit -replace=github.com/webview/webview_go=../path/to/your/webview_go
   ```
   然后 `go build` 会用到你改过的 webview。

---

## 方案二实现后的效果

- HTML 里 `body { background: transparent; }`（或某块 `background: transparent`）会**真正透出桌面**。
- 可继续用方案一的 **SetWindowRgn** 做异形裁切，或只做透明、不裁切，由你决定。
- 二者结合：**方案二负责“透明”**，**方案一负责“异形”**，即可实现基于 HTML 背景透明的自定义透明/异形窗口。

---

## 小结

- **方案一**：只能“硬裁切”成异形，**不能**实现基于 HTML 背景的透明。
- **方案二**：在 webview 的 C/C++ 里给 WebView2 设 `DefaultBackgroundColor` 透明 + 窗口设 `WS_EX_LAYERED`，**能**实现基于 HTML 背景透明的自定义透明窗口；实现难度中等，改动集中、代码量小，需要改依赖并重新编译。

按本文在 **CreateCoreWebView2Controller 完成回调**里加上上述几行，即可在现有方案一的基础上获得真正的“HTML 背景透明”能力。
