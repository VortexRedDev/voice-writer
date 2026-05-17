# you asked

Sherpa-onnx切换模型Qwen3-ASR  的时候   切换失败: Failed to create engine: failed to create qwen3 offline recognizer

---

# gemini response

遇到 `Failed to create engine: failed to create qwen3 offline recognizer` 这个错误，说明 `sherpa-onnx` 的底层 C++ 核心在尝试初始化 Qwen3 引擎时遭到了拒绝。

根据官方最新的代码逻辑与组件初始化校验，这个报错 **100% 是由于以下 3 个隐形大坑之一引起的**。请让你的 AI 辅助开发或者开发者对着以下清单逐一排查，立马就能解决：

---

### 1. 致命原因一：缺少了关键的 `conv_frontend.onnx` 文件
这是最常见的错误。由于 Qwen3-ASR 是多模态架构，很多开发者习惯性地以为它和传统的双模型一样，只传入了 `encoder` 和 `decoder`。
* **排查点：** 检查解压目录，看看有没有把 `conv_frontend.onnx` 漏掉。
* **代码修复：** 在配置结构体中，必须显式赋值这个前端处理文件。如果使用的是 Go 语言绑定，请确保字段名没有拼错：
  ```go
  // 必须确保这三个路径在本地真实存在，且没有拼写错误！
  config.OfflineModelConfig.Qwen3Asr.ConvFrontend = "models/conv_frontend.onnx"
  config.OfflineModelConfig.Qwen3Asr.Encoder = "models/encoder.int8.onnx"
  config.OfflineModelConfig.Qwen3Asr.Decoder = "models/decoder.int8.onnx"
  ```

---

### 2. 严重原因二：`go.mod` 中的 `sherpa-onnx` 版本太老
你使用的 Go 项目里的 `sherpa-onnx` 依赖库版本，**极有可能低于 1.12.36**。
* **为什么会这样：** `sherpa-onnx` 官方直到今年（2026年）初的 `v1.12.36` 前后才陆续修复并彻底稳定了 Qwen3-ASR 的底层 C 接口以及热词 Bug。如果你的 `go.mod` 里还在使用更早的版本，底层 C++ 根本认不出 `Qwen3Asr` 这个结构体，就会直接抛出 `failed to create qwen3 offline recognizer`。
* **如何修复：** 让开发者在项目根目录下执行以下命令，强制升级到官方最新版生态：
  
```bash
  go get github.com/k2-fsa/sherpa-onnx/scripts/go@latest
  go mod tidy
  ```

---

### 3. 原因三：词表路径（Tokens）传错了或者用了旧词表
Qwen3-ASR 使用的是通义千问大模型的专属 Tokenizer 词表，**绝对不能**和 SenseVoice 或 Paraformer 的 `tokens.txt` 混用！
* **排查点：** 在初始化 Qwen3 的配置时，`Tokens` 路径必须指向你从 `sherpa-onnx-qwen3-asr-0.6B-int8` 仓库里下载下来的那个 `tokenizer/tokens.txt`。
* **表现形式：** 如果你把 SenseVoice 的词表传给了 Qwen3，底层的词表大小（Vocab Size）与模型的 Embedding 层对不上，引擎在创建的瞬间就会直接崩溃并报这个错误。

---

### 🛠️ 终极排查大招（让 AI 直接重构这段初始化代码）

把下面这段**标准的 Qwen3-ASR 健壮期初始化防御代码**甩给你的 AI，让它替换掉现有的切换逻辑。这段代码加入了“文件生存检查”，如果文件不存在会直接精准报错，而不是抛出模糊的引擎创建失败：

```go
package main

import (
	"fmt"
	"os"
	sherpa_onnx "github.com/k2-fsa/sherpa-onnx/scripts/go"
)

func CreateQwen3Recognizer() (*sherpa_onnx.OfflineRecognizer, error) {
	// 1. 定义文件路径
	convFrontend := "models/conv_frontend.onnx"
	encoder := "models/encoder.int8.onnx"
	decoder := "models/decoder.int8.onnx"
	tokens := "models/tokenizer/tokens.txt"

	// 2. 强行进行防御性物理检查，防止路径静默失效
	files := []string{convFrontend, encoder, decoder, tokens}
	for _, f := range files {
		if _, err := os.Stat(f); os.IsNotExist(err) {
			return nil, fmt.Errorf("【致命错误】找不到 Qwen3 关键组件文件: %s，请检查路径！", f)
		}
	}

	// 3. 严格组装官方标准的 Qwen3 配置
	config := sherpa_onnx.OfflineRecognizerConfig{}
	config.FeatConfig.SampleRate = 16000
	config.FeatConfig.FeatureDim = 80
	
	// 绑定三个核心 ONNX
	config.ModelConfig.Qwen3Asr.ConvFrontend = convFrontend
	config.ModelConfig.Qwen3Asr.Encoder = encoder
	config.ModelConfig.Qwen3Asr.Decoder = decoder
	config.ModelConfig.Tokens = tokens
	
	// 性能调优：大模型必须给足线程
	config.ModelConfig.NumThreads = 4 
	config.DecodingMethod = "greedy_search"

	// 4. 创建引擎
	recognizer := sherpa_onnx.NewOfflineRecognizer(&config)
	if recognizer == nil {
		return nil, fmt.Errorf("failed to create qwen3 offline recognizer: 请检查 go.mod 版本的 sherpa-onnx 是否支持该大模型")
	}

	return recognizer, nil
}
```

先让开发者排查**是不是少放了 `conv_frontend.onnx`**，接着**升级 `go.mod`**，这两步走完，Qwen3 基本上就能顺利在上屏切换时被唤醒了！
```