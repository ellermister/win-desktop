module win-desktop

go 1.21

require github.com/webview/webview_go v0.0.0-20260213002156-9a608c2ce215

// 使用本地已修改的 webview_go（支持 WebView2 透明背景）
replace github.com/webview/webview_go => ./webview_go
