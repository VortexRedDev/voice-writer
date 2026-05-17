# Voice Writer — 开发计划

> 语音输入法桌面软件 | Windows 优先 | 单进程零依赖
> 
> *最后更新：2026-05-15*

---

## 进度

| 阶段 | 状态 | 产出 |
|------|------|------|
| 阶段 1 — 项目脚手架 | ✅ 完成 | Wails v2.12 + Vue 3 + Tailwind + voice-writer.exe |
| 阶段 2 — 音频采集链路 | ✅ 完成 | malgo 封装 |
| 阶段 3 — ASR 推理接入 | ✅ 完成 | sherpa-onnx 推理 |
| 阶段 4 — 键盘输出 | ✅ 完成 | robotgo 热键 + 键入 |
| 阶段 5 — UI 集成 | ✅ 完成 | 系统托盘 + 设置页 |

**当前阶段**：阶段 6 — 模型选择与热切换

**环境**：Go 1.26.0 · Node v20.18.1 · npm 10.8.2 · Windows

**待安装**：WebView2 Evergreen Runtime（Wails 窗口运行前提）  
　　　　　 https://developer.microsoft.com/microsoft-edge/webview2/

---

## 技术选型

| 模块 | 选型 | 说明 |
|------|------|------|
| 桌面框架 | Wails v2 | Go 后端 + Web 前端 |
| 前端 | Vue 3 + TypeScript + Tailwind CSS | 状态展示 / 设置页 |
| 音频采集 | malgo (`github.com/gen2brain/malgo`) | miniaudio Go binding，跨平台 |
| ASR 推理 | sherpa-onnx (`k2-fsa/sherpa-onnx`) | 内嵌 Paraformer ONNX，单进程 |
| 系统交互 | robotgo (`github.com/go-vgo/robotgo`) | 全局热键 + 键盘模拟 |
| 交互模式 | push-to-talk（按住说话，松开识别） | 非流式，链路短 |

---

## 架构总览

```
用户按下热键 → malgo 采集 PCM → []int16 buffer (持续累积)
用户松开热键 → 停止采集 → 下采样 16kHz → sherpa-onnx 推理
sherpa-onnx 返回文本 → robotgo 模拟键盘逐字输入 → 目标窗口
```

无网络、无 Python、无中间队列。单 Go 二进制 + 前端静态资源。

---

## 项目结构

```
voice-writer/
├── main.go                 # Wails 入口
├── app.go                  # App 结构体与生命周期
├── audio/
│   └── capture.go          # malgo 封装 (Start/Stop/Buffer)
├── asr/
│   └── recognizer.go       # sherpa-onnx 封装
├── output/
│   └── keyboard.go         # robotgo 键盘模拟 + 热键注册
├── frontend/               # Vue 3 前端
│   ├── src/
│   │   ├── App.vue
│   │   ├── components/
│   │   │   ├── StatusBar.vue        # 录音/识别状态指示
│   │   │   ├── ResultPanel.vue      # 识别结果列表
│   │   │   └── SettingsDialog.vue   # 设置面板
│   │   ├── stores/
│   │   │   └── app.ts               # Pinia 状态管理
│   │   └── assets/
│   │       └── main.css             # Tailwind 入口
│   ├── index.html
│   ├── package.json
│   ├── vite.config.ts
│   └── tailwind.config.ts
├── models/                  # ONNX 模型文件 (gitignore)
├── build/
│   └── app.ico
├── wails.json
├── go.mod
└── go.sum
```

---

## 模型下载

所有 ONNX 模型从 Hugging Face 的 [csukuangfj/sherpa-onnx](https://huggingface.co/csukuangfj) 下载。解压到 `models/` 目录下，目录名与代码中 `modelType` 对应。

### 模型清单

| 模型类型 | 代码目录名 | 模型仓库 | 所需文件 | 备注 |
|---------|-----------|---------|---------|------|
| Paraformer | `paraformer/` | [sherpa-onnx-paraformer-zh-2023-09-14](https://huggingface.co/csukuangfj/sherpa-onnx-paraformer-zh-2023-09-14) | `model.int8.onnx` + `tokens.txt` | 基础模型，无热词 |
| Paraformer (热词版) | `paraformer/` | [sherpa-onnx-paraformer-v2-zh-2023-09-14](https://huggingface.co/csukuangfj/sherpa-onnx-paraformer-v2-zh-2023-09-14) | `model.int8.onnx` + `tokens.txt` + `context_graph.onnx` | 支持热词 |
| SenseVoice | `sense-voice/` | [sherpa-onnx-sense-voice-zh-en-ja-ko-yue-2024-07-17](https://huggingface.co/csukuangfj/sherpa-onnx-sense-voice-zh-en-ja-ko-yue-2024-07-17) | `model.int8.onnx` + `tokens.txt` | 内置标点，多语言，不支持热词 |
| Qwen3-ASR | `qwen3-asr/` | [sherpa-onnx-qwen3-asr-0.6B-int8](https://huggingface.co/csukuangfj/sherpa-onnx-qwen3-asr-0.6B-int8) | `encoder.int8.onnx` + `decoder.int8.onnx` + `conv_frontend.onnx` + `tokenizer/tokens.txt` | 内置标点 + 热词 |
| 标点模型 | `punctuation/` | [sherpa-onnx-punct-ct-transformer-zh-en-vocab727k-2024-04-14](https://huggingface.co/csukuangfj/sherpa-onnx-punct-ct-transformer-zh-en-vocab727k-2024-04-14) | `model.int8.onnx` + `tokens.json` | Paraformer 专用，其他模型内置标点则不需要 |

### 下载后目录结构

```
models/
├── paraformer/
│   ├── model.int8.onnx            # ~200MB
│   ├── tokens.txt
│   └── (可选) context_graph.onnx   # 热词版独有，~1MB
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
└── punctuation/                   # 外部标点模型（Paraformer 专用）
    ├── model.int8.onnx            # ~50MB
    └── tokens.json
```

> **注意**：Qwen3-ASR 下载后需要将 `tokenizer/tokens.txt` 单独放置，或在 `app.go` 中额外指定 Tokens 路径。

---

## 设计决策：音频缓冲区方案

### 约束条件

`malgo` 的音频回调跑在 **C 线程**上，不是 Go goroutine。回调在录音期间持续触发，每次携带一段 PCM 数据（如 40ms 片段）。核心要求：

- 回调不能阻塞（否则丢帧）
- 采集期间只累积数据，不做 ASR
- 按键松开后，完整 buffer 一次性送入识别引擎

### 当前方案：Mutex + Slice（v0.1 push-to-talk）

```go
type Recorder struct {
    mu     sync.Mutex
    buffer []int16
}

// malgo 回调中（C 线程）—— 微秒级操作，不会阻塞
func (r *Recorder) onData(samples []int16) {
    r.mu.Lock()
    r.buffer = append(r.buffer, samples...)
    r.mu.Unlock()
}

// 松开热键时返回完整 PCM
func (r *Recorder) Stop() []int16 {
    r.mu.Lock()
    defer r.mu.Unlock()
    result := r.buffer
    r.buffer = nil
    return result
}

// 录音结束后统一处理
func Process(input []int16, srcRate, dstRate int) []int16 {
    return resample(input, srcRate, dstRate)
}
```

**选择理由**：
- 一个锁、一个 slice，复杂度极低
- `Lock/Unlock` 操作微秒级，C 线程不阻塞
- push-to-talk 模式下不需要"一边采集一边识别"，管道化没有收益
- 处理逻辑（重采样等）集中在 `Process` 函数，与采集解耦

### 演进方案：Channel 管道（v0.2 流式识别）

当系统升级为流式 ASR（边说边出字）时，用 channel 替换 Mutex + Buffer：

```go
rawCh  := make(chan []int16, 64)   // malgo 回调 → 重采样 goroutine
procCh := make(chan []int16, 32)   // 重采样后 → ASR goroutine

// malgo 回调：必须非阻塞发送
func onData(samples []int16) {
    select {
    case rawCh <- samples:
    default:
        // channel 满时丢弃（或使用 ring buffer）
    }
}

// goroutine 1: 预处理
go func() {
    for chunk := range rawCh {
        procCh <- resample(chunk, srcRate, 16000)
    }
}()

// goroutine 2: 流式送入 ASR
go func() {
    for chunk := range procCh {
        rec.AcceptWaveform(16000, chunk)
        // 实时获取部分结果
        text := rec.GetResult()
        emit(text)
    }
}()
```

**演进成本**：录音侧 `onData` 改为非阻塞 channel send，其余保持不变。`Process` 函数直接对接到 goroutine 管道。

### 两种方案对照

| 维度 | v0.1 Mutex + Slice | v0.2 Channel 管道 |
|------|-------------------|-------------------|
| 模式 | push-to-talk | 流式 |
| C 线程安全 | Lock/Unlock，无阻塞风险 | 非阻塞 send，channel 满时降级 |
| 处理时机 | 录音结束后一次性 | 采集的同时持续处理 |
| 适用场景 | 按住说、松开发送 | 实时出字 |
| 复杂度 | 低 | 中（goroutine 管理、容量调优） |

---

## 分阶段计划

### 阶段 1 — 项目脚手架 ✅

**目标**：可编译的空 Wails 项目，三方依赖就位。

| # | 任务 | 详情 | 状态 |
|---|------|------|------|
| 1.1 | Wails CLI 安装 | v2.12.0，需设 GOPROXY=goproxy.cn | ✅ |
| 1.2 | 项目结构创建 | main.go / app.go / wails.json / Vue 前端 (13 文件) | ✅ |
| 1.3 | Go 依赖引入 | go mod tidy → wails v2.12.0 | ✅ |
| 1.4 | 前端构建 | npm install + npm run build → dist/ 生成 | ✅ |
| 1.5 | 编译验证 | go build → voice-writer.exe (6.3MB) | ✅ |

> 踩坑：echo 写入文件会污染引号/尖括号，后续阶段用 write_file / edit_file。

### 阶段 2 — 音频采集链路 ✅

**目标**：热键控制录音，PCM 数据正确采集到内存 buffer。

| # | 任务 | 详情 |
|---|------|------|
| 2.1 | malgo 设备枚举与初始化 | 枚举输入设备，选择默认麦克风 |
| 2.2 | 开始/停止控制 | 热键按下→Start，松开→Stop |
| 2.3 | PCM buffer 管理 | 回调中 append 到 `[]int16`，加锁保护 |
| 2.4 | 采样率转换 | 设备原生采样率 → 16kHz |
| 2.5 | 验证测试 | 录制音频写入 WAV 文件，确认数据正确 |

### 阶段 3 — ASR 推理接入 ✅

**目标**：加载 Paraformer 模型，PCM 输入 → 文本输出。

| # | 任务 | 详情 |
|---|------|------|
| 3.1 | 模型下载 | 从 sherpa-onnx release 下载 Paraformer ONNX |
| 3.2 | Recognizer 封装 | 模型加载、`AcceptWaveform`、`GetResult` |
| 3.3 | 端到端测试 | WAV 文件 → 识别 → 打印文本 |
| 3.4 | 识别 API 封装 | `Recognize(samples []int16) (string, error)` |

### 阶段 4 — 键盘输出 ✅

**目标**：全局热键 → 录音 → 识别 → 键入，完整链路打通。

| # | 任务 | 详情 |
|---|------|------|
| 4.1 | 全局热键注册 | robotgo 注册 `Ctrl+Shift+V`（默认） |
| 4.2 | 键盘模拟键入 | 逐字符输出中文文本 |
| 4.3 | 完整链路联调 | 按住→说话→松开→识别→键入 |
| 4.4 | 边界处理 | 空录音保护、识别超时、焦点窗口切换 |

### 阶段 5 — UI 集成 ✅

**目标**：可用的桌面应用，托盘 + 前端界面完整。

| # | 任务 | 详情 |
|---|------|------|
| 5.1 | 系统托盘 | 托盘图标 + 右键菜单（设置/退出） |
| 5.2 | 前后端通信 | Wails Bind：Go 状态 ↔ Vue 响应式 |
| 5.3 | 状态页 | 空闲/录音中/识别中 状态指示 |
| 5.4 | 设置页 | 热键自定义、模型选择、灵敏度调节 |
| 5.5 | 打包 | `wails build` 生成 Windows 安装包 |

---

## 核心 API 设计

### audio/capture.go

```go
type Recorder struct { ... }

func NewRecorder() (*Recorder, error)
func (r *Recorder) Start() error           // 开始采集
func (r *Recorder) Stop() ([]int16, error) // 停止并返回 PCM
func (r *Recorder) IsRecording() bool
```

### asr/recognizer.go

```go
type Engine struct { ... }

func NewEngine(modelPath string) (*Engine, error)
func (e *Engine) Recognize(samples []int16) (string, error)
func (e *Engine) Close()
```

### Wails 绑定 (app.go)

```go
type App struct { ... }

// 暴露给前端
func (a *App) GetStatus() Status           // 当前状态
func (a *App) GetHistory() []HistoryItem   // 历史记录
func (a *App) UpdateSettings(s Settings)   // 更新设置

// 事件回调（Go → 前端）
func (a *App) OnStatusChange(status Status)
func (a *App) OnResult(text string)
```

---

## 风险与缓解

| 风险 | 影响 | 缓解 |
|------|------|------|
| CGO 三连编译链配置 | 高 | Wails 官方文档 + MinGW64 环境；备选：各自封装纯 Go fallback |
| sherpa-onnx 模型体积 (~200MB) | 中 | 首次启动异步加载 + 进度提示；后续版本支持模型预下载 |
| robotgo TypeStr 中文兼容 | 中 | 实测 Unicode 输入；备选：剪贴板粘贴方案 |
| 采样率转换精度 | 低 | 线性插值即可；备选：集成轻量 DSP 库 |
| 非流式延迟（松键后等待识别） | 低 | MVP 可接受；后续版本可选流式 ASR 扩展 |

---

## 版本规划

- **v0.1 MVP**：阶段 1-5 全部完成，按住说话→松开发送基本可用
- **v0.2**：流式识别（实时出字）+ 标点优化
- **v0.3**：热词/自定义词典 + 多语言支持
- **v1.0**：跨平台（macOS / Linux）+ 安装包分发

---

## 阶段 6 — 模型选择与热切换

**目标**：支持用户选择不同 ASR 模型，实现运行时热切换，支持热词功能。

### 6.1 ASR 引擎接口抽象

**目标**：建立统一的引擎接口，支持多种模型实现。

```go
// asr/engine.go

// ASREngine 统一接口
type ASREngine interface {
    Recognize(samples []int16, opts RecognitionOptions) (string, error)
    Reload(config ModelConfig) error  // 热重载
    GetInfo() ModelInfo
    Close()
}

// 识别选项
type RecognitionOptions struct {
    Punctuation bool   // 是否添加标点
    Hotwords    string // 热词（可选）
}

// 模型信息
type ModelInfo struct {
    Name                 string   // 显示名称
    Type                 string   // "paraformer" | "whisper" | "sensevoice"
    Version              string
    SupportsHotwords     bool     // 是否支持热词
    SupportsPunctuation  bool     // 是否支持标点
    Size                 int64    // MB
    RecommendedSampleRate int
}

// 模型配置
type ModelConfig struct {
    Type        string  // 模型类型
    ModelPath   string  // 模型文件路径
    TokensPath  string  // tokens 文件路径
    HotwordsPath string // 热词文件路径
    NumThreads  int     // 线程数
}
```

| # | 任务 | 状态 |
|---|------|------|
| 6.1.1 | 创建 `asr/engine.go` 定义引擎接口 | 待开始 |
| 6.1.2 | 重构 `asr/recognizer.go` 实现引擎接口 | 待开始 |
| 6.1.3 | 添加引擎注册机制 | 待开始 |

### 6.2 模型信息标签页

**目标**：新增"模型"标签页，显示当前模型信息并支持切换。

**UI 设计**：

```
模型标签页
├── 当前模型信息卡片
│   ├── 模型名称
│   ├── 模型类型
│   ├── 是否支持热词
│   └── 模型大小
├── 模型选择器
│   ├── Paraformer（当前）
│   ├── Whisper（待支持）
│   └── SenseVoice（待支持）
├── 模型路径配置
│   ├── 模型文件路径
│   └── Tokens 文件路径
├── 下载模型（未来功能）
│   └── 从预设 URL 下载
└── 重载模型按钮
```

| # | 任务 | 状态 |
|---|------|------|
| 6.2.1 | 创建模型标签页 UI | 待开始 |
| 6.2.2 | 显示当前模型信息 | 待开始 |
| 6.2.3 | 添加模型选择下拉框 | 待开始 |
| 6.2.4 | 添加模型路径配置 | 待开始 |
| 6.2.5 | 添加重载模型按钮 | 待开始 |

### 6.3 模型热切换机制

**目标**：运行时切换 ASR 引擎，不中断应用。

```go
func (a *App) SwitchModel(config ModelConfig) error {
    // 1. 创建新引擎
    newEngine, err := asr.NewEngine(config)
    if err != nil {
        return err
    }
    
    // 2. 原子替换
    oldEngine := a.asrEngine
    a.asrEngine = newEngine
    
    // 3. 关闭旧引擎
    if oldEngine != nil {
        oldEngine.Close()
    }
    
    // 4. 保存配置
    a.cfg.ModelConfig = config
    config.SaveConfig(a.cfg)
    
    return nil
}
```

| # | 任务 | 状态 |
|---|------|------|
| 6.3.1 | 实现 `SwitchModel` 方法 | 待开始 |
| 6.3.2 | 添加模型配置持久化 | 待开始 |
| 6.3.3 | 添加切换中状态提示 | 待开始 |

### 6.4 热词支持（Paraformer-v2）

**目标**：使用支持热词的模型启用热词功能。

**备选模型**：
- Paraformer-v2：`sherpa-onnx-paraformer-v2-zh-2023-09-14`
- 下载：https://huggingface.co/csukuangfj/sherpa-onnx-paraformer-v2-zh-2023-09-14
- 特性：支持 `greedy_search` + 热词

| # | 任务 | 状态 |
|---|------|------|
| 6.4.1 | 测试 Paraformer-v2 热词功能 | 待开始 |
| 6.4.2 | 启用热词配置传递 | 待开始 |
| 6.4.3 | 测试热词识别效果 | 待开始 |

### 6.5 Whisper 模型支持（未来）

**目标**：支持多语言识别。

| # | 任务 | 状态 |
|---|------|------|
| 6.5.1 | 实现 Whisper 引擎 | 待开始 |
| 6.5.2 | 添加 Whisper 模型配置 | 待开始 |
| 6.5.3 | 测试多语言识别 | 待开始 |

### 6.6 多模型架构设计（工厂 + 策略模式）

**目标**：支持 Paraformer、SenseVoice、Qwen3-ASR 三种模型，每种模型的热词和标点处理策略不同。

#### 6.6.1 模型能力矩阵

| 模型 | 热词支持 | 标点支持 | 标点实现方式 | 热词实现方式 |
|------|---------|---------|-------------|-------------|
| Paraformer | ✅ 需要 `context_graph.onnx` | ❌ 需要专用标点模型 | 外部 `OfflinePunctuation` | `context_graph.onnx` |
| SenseVoice | ❌ 不支持 | ✅ 内置 | 内置标点 | 不支持 |
| Qwen3-ASR | ✅ 支持 | ✅ 内置 | 内置标点 | 内置热词 |

#### 6.6.2 架构类图

```
┌─────────────────────────────────────────────────────────────────┐
│                          工厂模式                                │
└─────────────────────────────────────────────────────────────────┘

                         ┌──────────────────┐
                         │   EngineFactory   │
                         └────────┬─────────┘
                                  │CreateEngine(type)
                                  │
          ┌───────────────────────┼───────────────────────┐
          │                       │                       │
          ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│ ParaformerEngine│    │SenseVoiceEngine │    │  Qwen3Engine    │
└────────┬────────┘    └────────┬────────┘    └────────┬────────┘
         │                       │                       │
         │ Implements            │ Implements            │ Implements
         ▼                       ▼                       ▼
┌─────────────────────────────────────────────────────────────────┐
│                     ASREngine 接口                               │
├─────────────────────────────────────────────────────────────────┤
│ + Recognize(samples, opts) → (text, error)                      │
│ + GetInfo() → ModelInfo                                         │
│ + GetConfig() → ModelConfig                                     │
│ + Switch(config) → error                                         │
│ + Close()                                                        │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                       策略模式                                    │
└─────────────────────────────────────────────────────────────────┘

         ┌──────────────────┐        ┌──────────────────┐
         │ PunctStrategy    │        │  HotwordStrategy │
         └────────┬─────────┘        └────────┬─────────┘
                  │                            │
       ┌─────────┴─────────┐        ┌─────────┴─────────┐
       │                   │        │                   │
       ▼                   ▼        ▼                   ▼
┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│  Builtin    │    │  Offline    │    │ ContextGraph│    │   None      │
│  Punct      │    │  Punct      │    │  Strategy   │    │  Strategy   │
│  Strategy   │    │  Strategy   │    │             │    │             │
└─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘
  (SenseVoice,      (Paraformer      (Paraformer,     (SenseVoice)
   Qwen3内置)        专用标点模型)      Qwen3内置)


┌─────────────────────────────────────────────────────────────────┐
│                       完整处理流程                               │
└─────────────────────────────────────────────────────────────────┘

用户说话 ──► ASR 识别 ──► 标点处理 ──► 键盘输出
                              │
          ┌───────────────────┼───────────────────┐
          │                   │                   │
          ▼                   ▼                   ▼
   内置标点 (SenseVoice)   专用标点模型      跳过标点
   Qwen3-ASR              (Paraformer)        处理
```

#### 6.6.3 目录结构

```
asr/
├── engine.go           # ASREngine 接口、RecognitionOptions、ModelInfo
├── factory.go          # EngineFactory 工厂
│
├── paraformer/
│   └── engine.go       # ParaformerEngine 实现
│                       # - HotwordStrategy: ContextGraphStrategy
│                       # - PunctStrategy: OfflinePunctStrategy
│
├── sensevoice/
│   └── engine.go       # SenseVoiceEngine 实现
│                       # - HotwordStrategy: NoneStrategy
│                       # - PunctStrategy: BuiltinPunctStrategy
│
├── qwen3/
│   └── engine.go        # Qwen3Engine 实现
│                       # - HotwordStrategy: BuiltinHotwordStrategy
│                       # - PunctStrategy: BuiltinPunctStrategy
│
├── punct/
│   └── offline.go      # OfflinePunctuation 策略实现
│
└── strategy/
    ├── hotword.go      # HotwordStrategy 接口
    └── punct.go        # PunctStrategy 接口
```

#### 6.6.4 核心接口定义

```go
// asr/engine.go

// ASREngine 统一接口
type ASREngine interface {
    Recognize(samples []int16, opts RecognitionOptions) (string, error)
    Reload(config ModelConfig) error
    Switch(config ModelConfig) error
    GetInfo() ModelInfo
    GetConfig() ModelConfig
    Close()
}

// RecognitionOptions 识别选项
type RecognitionOptions struct {
    Punctuation bool   // 是否启用标点
    Hotwords     string // 热词
}

// ModelInfo 模型信息
type ModelInfo struct {
    Name                  string
    Type                  string   // "paraformer" | "sensevoice" | "qwen3"
    Version               string
    SupportsHotwords      bool
    SupportsPunctuation   bool
    NeedsContextGraph     bool    // Paraformer 需要 context_graph.onnx
    NeedsExternalPunct    bool    // Paraformer 需要外部标点模型
}

// ModelConfig 模型配置
type ModelConfig struct {
    Type           string  // 模型类型
    ModelPath      string  // 模型文件路径
    TokensPath     string  // tokens 文件路径
    ContextGraphPath string // 热词图路径 (可选)
    PunctModelPath string  // 标点模型路径 (可选)
    NumThreads     int
}
```

#### 6.6.5 策略接口

```go
// asr/strategy/punct.go

// PunctStrategy 标点策略接口
type PunctStrategy interface {
    Apply(text string) (string, error)
    IsAvailable() bool
}

// BuiltinPunctStrategy 内置标点 (SenseVoice, Qwen3)
type BuiltinPunctStrategy struct {}

func (s *BuiltinPunctStrategy) Apply(text string) (string, error) {
    // 这些模型内置标点，不需要额外处理
    return text, nil
}

// OfflinePunctStrategy 外部标点模型 (Paraformer)
type OfflinePunctStrategy struct {
    modelPath string
}
```

```go
// asr/strategy/hotword.go

// HotwordStrategy 热词策略接口
type HotwordStrategy interface {
    SetHotwords(hotwords string) error
    IsAvailable() bool
}

// ContextGraphStrategy 上下文图热词 (Paraformer, Qwen3)
type ContextGraphStrategy struct {
    contextGraphPath string
}

// NoneStrategy 不支持热词 (SenseVoice)
type NoneStrategy struct {}
```

#### 6.6.6 工厂模式实现

```go
// asr/factory.go

type EngineFactory struct{}

func (f *EngineFactory) CreateEngine(config ModelConfig) (ASREngine, error) {
    switch config.Type {
    case "paraformer":
        return NewParaformerEngine(config)
    case "sensevoice":
        return NewSenseVoiceEngine(config)
    case "qwen3":
        return NewQwen3Engine(config)
    default:
        return nil, fmt.Errorf("unsupported engine type: %s", config.Type)
    }
}
```

#### 6.6.7 任务列表

| # | 任务 | 状态 |
|---|------|------|
| 6.6.1 | 创建策略接口 (PunctStrategy, HotwordStrategy) | 待开始 |
| 6.6.2 | 实现 OfflinePunctStrategy (外部标点模型) | 待开始 |
| 6.6.3 | 实现 ContextGraphStrategy (热词图) | 待开始 |
| 6.6.4 | 实现 SenseVoiceEngine | 待开始 |
| 6.6.5 | 实现 Qwen3Engine | 待开始 |
| 6.6.6 | 实现 EngineFactory 工厂 | 待开始 |
| 6.6.7 | 重构 ParaformerEngine 使用策略模式 | 待开始 |
| 6.6.8 | 更新前端模型选择 UI | 待开始 |
| 6.6.9 | 集成测试三种模型 | 待开始 |
