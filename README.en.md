# Voice Writer

Offline desktop voice-to-text input tool — hold a hotkey to speak, release to recognize and type.

**Windows first · Single binary, zero dependencies · Fully offline**

---

## Features

- **Push-to-Talk**: Hold F9 to record, release to auto-recognize and type via simulated keyboard
- **Three ASR engines**: Paraformer, SenseVoice, Qwen3-ASR — switch at runtime without restart
- **Hotword support**: Paraformer / Qwen3-ASR support custom hotwords via a text file
- **Built-in punctuation**: SenseVoice and Qwen3-ASR output punctuation natively; Paraformer can use an external punctuation model
- **System tray**: Runs in background with a tray menu for quick settings

## Requirements

- Windows 10/11 64-bit
- [WebView2 Runtime](https://developer.microsoft.com/microsoft-edge/webview2/) (included in Win11)
- Microphone

## Model Setup

Download ONNX models and place them under `models/` next to the executable:

| Model | Download |
|-------|----------|
| Paraformer | [sherpa-onnx-paraformer-zh-2023-09-14](https://huggingface.co/csukuangfj/sherpa-onnx-paraformer-zh-2023-09-14) |
| Paraformer (hotword) | [sherpa-onnx-paraformer-v2-zh-2023-09-14](https://huggingface.co/csukuangfj/sherpa-onnx-paraformer-v2-zh-2023-09-14) |
| SenseVoice | [sherpa-onnx-sense-voice-zh-en-ja-ko-yue-2024-07-17](https://huggingface.co/csukuangfj/sherpa-onnx-sense-voice-zh-en-ja-ko-yue-2024-07-17) |
| Qwen3-ASR | [sherpa-onnx-qwen3-asr-0.6B-int8](https://huggingface.co/csukuangfj/sherpa-onnx-qwen3-asr-0.6B-int8) |
| Punctuation (Paraformer optional) | [sherpa-onnx-punct-ct-transformer-zh-en-vocab727k-2024-04-14](https://huggingface.co/csukuangfj/sherpa-onnx-punct-ct-transformer-zh-en-vocab727k-2024-04-14) |

Expected directory layout (next to the executable):

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
└── punctuation/                   # Paraformer only (optional)
    ├── model.int8.onnx            # ~50MB
    └── tokens.json
```

See [DEVELOPMENT_PLAN.md](DEVELOPMENT_PLAN.md#模型下载) for more details.

## Build

```bash
# Install Wails CLI
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# Development (hot reload)
wails dev

# Production build
wails build
```

## Tech Stack

| Module | Choice |
|--------|--------|
| Desktop framework | [Wails v2](https://wails.io/) (Go + Vue 3) |
| Frontend | Vue 3 + TypeScript + Tailwind CSS |
| ASR engine | [sherpa-onnx](https://github.com/k2-fsa/sherpa-onnx) v1.13.2 |
| Audio capture | [malgo](https://github.com/gen2brain/malgo) |
| Keyboard emulation | [robotgo](https://github.com/go-vgo/robotgo) |

## Architecture

```
Hold hotkey → Capture PCM via microphone → Release → sherpa-onnx offline recognition
         → Punctuation processing → robotgo keyboard type → Target window
```

No network, no Python, single binary distribution.

## Related Projects

- [CapsWriter-Offline](https://github.com/HaujetZhao/CapsWriter-Offline) — feature-rich voice input tool written in Python
- [sherpa-onnx](https://github.com/k2-fsa/sherpa-onnx) — cross-platform ONNX inference framework
