package asr

import (
	"fmt"
	"os"

	sherpa "github.com/k2-fsa/sherpa-onnx-go/sherpa_onnx"
)

// Qwen3Engine 基于 sherpa-onnx Qwen3-ASR 的引擎实现
type Qwen3Engine struct {
	EngineBase
	recognizer *sherpa.OfflineRecognizer
}

// NewQwen3Engine 创建 Qwen3 引擎
func NewQwen3Engine(config ModelConfig) (*Qwen3Engine, error) {
	// 获取文件大小（统计所有模型文件）
	var fileSize int64
	if info, err := os.Stat(config.ModelPath); err == nil {
		fileSize = info.Size()
	}

	engine := &Qwen3Engine{
		EngineBase: EngineBase{
			config: config,
			info: ModelInfo{
				Name:                  "Qwen3-ASR",
				Type:                  "qwen3",
				Version:               "int8",
				SupportsHotwords:      true, // Qwen3 支持热词
				SupportsPunctuation:   true, // Qwen3 内置标点
				Size:                  fileSize,
				RecommendedSampleRate: 16000,
			},
		},
	}

	// Qwen3 使用 Qwen3-ASR 模型结构
	// 读取热词文件内容
	// var hotwordsContent string
	// if config.HotwordsPath != "" {
	// 	if data, err := os.ReadFile(config.HotwordsPath); err == nil {
	// 		hotwordsContent = string(data)
	// 	}
	// }

	configPtr := &sherpa.OfflineRecognizerConfig{
		FeatConfig: sherpa.FeatureConfig{
			SampleRate: 16000,
			FeatureDim: 80,
		},
		ModelConfig: sherpa.OfflineModelConfig{
			Qwen3ASR: sherpa.OfflineQwen3ASRModelConfig{
				Encoder:      config.ModelPath,        // encoder.int8.onnx
				Decoder:      config.TokensPath,       // decoder.int8.onnx
				ConvFrontend: config.ConvFrontendPath, // conv_frontend.onnx
				Tokenizer:    config.TokenizerPath,    // tokenizer 目录
			},
			Tokens:       "", // Qwen3 不需要单独的 tokens 文件
			NumThreads:   config.NumThreads,
			Debug:        0,
			ModelType:    "qwen3",
			ModelingUnit: "cjkchar+bpe",   // Qwen3 需要指定建模单元
			BpeVocab:     config.BpeVocab, // vocab.json 路径
		},
		DecodingMethod: "modified_beam_search", // 使用热词时必须使用 modified_beam_search
		HotwordsFile:   config.HotwordsPath,
		HotwordsScore:  10.0,
	}

	recognizer := sherpa.NewOfflineRecognizer(configPtr)
	if recognizer == nil {
		return nil, fmt.Errorf("failed to create qwen3 offline recognizer")
	}

	engine.recognizer = recognizer
	return engine, nil
}

// Recognize 执行语音识别
func (e *Qwen3Engine) Recognize(samples []int16, opts RecognitionOptions) (string, error) {
	result, err := e.recognizeInternal(samples, opts)
	if err != nil {
		return "", err
	}
	if result == nil {
		return "", nil
	}

	// Qwen3 内置标点，直接返回
	return result.Text, nil
}

// RecognizeDetailed 执行语音识别，返回带时间戳的详细结果
func (e *Qwen3Engine) RecognizeDetailed(samples []int16, opts RecognitionOptions) (*TimestampResult, error) {
	result, err := e.recognizeInternal(samples, opts)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, nil
	}

	return &TimestampResult{
		Text:       result.Text,
		Tokens:     result.Tokens,
		Timestamps: result.Timestamps,
	}, nil
}

func (e *Qwen3Engine) recognizeInternal(samples []int16, opts RecognitionOptions) (*sherpa.OfflineRecognizerResult, error) {
	if e.recognizer == nil {
		return nil, fmt.Errorf("recognizer not initialized")
	}

	// Convert int16 to float32 samples (range [-1, 1])
	floatSamples := make([]float32, len(samples))
	for i, s := range samples {
		floatSamples[i] = float32(s) / 32768.0
	}

	stream := sherpa.NewOfflineStream(e.recognizer)
	if stream == nil {
		return nil, fmt.Errorf("failed to create offline stream")
	}
	defer sherpa.DeleteOfflineStream(stream)

	stream.AcceptWaveform(16000, floatSamples)
	e.recognizer.Decode(stream)

	return stream.GetResult(), nil
}

// Reload 热重载模型配置
func (e *Qwen3Engine) Reload(config ModelConfig) error {
	// 关闭旧引擎
	if e.recognizer != nil {
		sherpa.DeleteOfflineRecognizer(e.recognizer)
		e.recognizer = nil
	}

	// 重新初始化
	engine, err := NewQwen3Engine(config)
	if err != nil {
		return err
	}

	e.config = engine.config
	e.info = engine.info
	e.recognizer = engine.recognizer
	return nil
}

// Switch 切换到新的模型配置
func (e *Qwen3Engine) Switch(config ModelConfig) error {
	return e.Reload(config)
}

// Close 释放资源
func (e *Qwen3Engine) Close() {
	if e.recognizer != nil {
		sherpa.DeleteOfflineRecognizer(e.recognizer)
		e.recognizer = nil
	}
}
