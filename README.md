# Voice Writer

离线语音输入法桌面软件 — 按住快捷键说话，松开即识别并键入文字。

**Windows 优先 · 单进程零依赖 · 完全离线**

---

## 功能

- **Push-to-Talk**：按住 F9 录音，松开自动识别并模拟键盘键入
- **三种 ASR 引擎**：支持 Paraformer、SenseVoice、Qwen3-ASR，运行时热切换
- **热词支持**：Paraformer / Qwen3-ASR 可通过热词文件提升特定词识别率
- **内置标点**：SenseVoice 和 Qwen3-ASR 内置标点，Paraformer 可加载外部标点模型
- **系统托盘**：后台常驻，托盘菜单快速设置

## 系统要求

- Windows 10/11 64 位
- [WebView2 Runtime](https://developer.microsoft.com/microsoft-edge/webview2/)（通常 Win11 自带）
- 麦克风

## 下载模型

首次使用需要下载 ONNX 模型文件，放入可执行文件同目录下的 `models/` 文件夹：

| 模型 | 下载地址 |
|------|---------|
| Paraformer | [sherpa-onnx-paraformer-zh-2023-09-14](https://huggingface.co/csukuangfj/sherpa-onnx-paraformer-zh-2023-09-14) |
| Paraformer (热词版) | [sherpa-onnx-paraformer-v2-zh-2023-09-14](https://huggingface.co/csukuangfj/sherpa-onnx-paraformer-v2-zh-2023-09-14) |
| SenseVoice | [sherpa-onnx-sense-voice-zh-en-ja-ko-yue-2024-07-17](https://huggingface.co/csukuangfj/sherpa-onnx-sense-voice-zh-en-ja-ko-yue-2024-07-17) |
| Qwen3-ASR | [sherpa-onnx-qwen3-asr-0.6B-int8](https://huggingface.co/csukuangfj/sherpa-onnx-qwen3-asr-0.6B-int8) |
| 标点模型 (Paraformer 可选) | [sherpa-onnx-punct-ct-transformer-zh-en-vocab727k-2024-04-14](https://huggingface.co/csukuangfj/sherpa-onnx-punct-ct-transformer-zh-en-vocab727k-2024-04-14) |

下载后目录结构（与可执行文件同目录）：

```
models/
├── paraformer/
│   ├── model.int8.onnx            # ~200MB
│   └── tokens.txt
├── sense-voice/
│   ├── model.int8.onnx            # ~80MB
│   └── tokens.txt
├── qwen3-asr/
│   ├── encoder.int8.onnx          # ~1.7GB
│   ├── decoder.int8.onnx          # ~700MB
│   ├── conv_frontend.onnx         # ~200MB
│   └── tokenizer/
│			├── merges.txt
│			├── tokenizer_config.json
│       └── vocab.json
└── punctuation/                   # Paraformer 专用（可选）
    ├── model.int8.onnx            # ~50MB
    └── tokens.json
```

详细模型说明见 [DEVELOPMENT_PLAN.md](./DEVELOPMENT_PLAN.md#模型下载)。

## 编译

```bash
# 安装 Wails CLI
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# 开发模式（热重载）
wails dev

# 生产构建
wails build
```

## 技术栈

| 模块 | 选型 |
|------|------|
| 桌面框架 | [Wails v2](https://wails.io/)（Go + Vue 3） |
| 前端 | Vue 3 + TypeScript + Tailwind CSS |
| ASR 引擎 | [sherpa-onnx](https://github.com/k2-fsa/sherpa-onnx) v1.13.2 |
| 音频采集 | [malgo](https://github.com/gen2brain/malgo) |
| 键盘模拟 | [robotgo](https://github.com/go-vgo/robotgo) |

## 架构

```
按住热键 → 麦克风采集 PCM → 松键 → sherpa-onnx 离线识别
         → 标点处理 → robotgo 模拟键盘键入 → 目标窗口
```

无网络、无 Python、单二进制分发。

## 相关项目

- [CapsWriter-Offline](https://github.com/HaujetZhao/CapsWriter-Offline) — 功能更丰富的语音输入法，Python 实现
- [sherpa-onnx](https://github.com/k2-fsa/sherpa-onnx) — 跨平台 ONNX 推理框架
