package asr

import (
	"fmt"
	"os"
	"voice-writer/asr/strategy"

	sherpa "github.com/k2-fsa/sherpa-onnx-go/sherpa_onnx"
)

// ParaformerEngine 基于 sherpa-onnx Paraformer 的引擎实现
type ParaformerEngine struct {
	EngineBase
	recognizer *sherpa.OfflineRecognizer
	punct      *strategy.OfflinePunctStrategy
}

// NewParaformerEngine 创建 Paraformer 引擎
func NewParaformerEngine(config ModelConfig) (*ParaformerEngine, error) {
	// 获取文件大小
	var fileSize int64
	if info, err := os.Stat(config.ModelPath); err == nil {
		fileSize = info.Size()
	}

	engine := &ParaformerEngine{
		EngineBase: EngineBase{
			config: config,
			info: ModelInfo{
				Name:                  "Paraformer",
				Type:                  "paraformer",
				Version:               "int8",
				SupportsHotwords:      false, // Paraformer 目前仅支持 greedy_search，不支持热词
				SupportsPunctuation:   true,
				Size:                  fileSize,
				RecommendedSampleRate: 16000,
			},
		},
	}

	// 初始化标点模型
	if config.PunctuationModelPath != "" {
		engine.punct = strategy.NewOfflinePunctStrategy(config.PunctuationModelPath, config.PunctuationTokensPath)
	}

	// 创建 recognizer
	configPtr := &sherpa.OfflineRecognizerConfig{
		FeatConfig: sherpa.FeatureConfig{
			SampleRate: 16000,
			FeatureDim: 80,
		},
		ModelConfig: sherpa.OfflineModelConfig{
			Paraformer: sherpa.OfflineParaformerModelConfig{
				Model: config.ModelPath,
			},
			Tokens:       config.TokensPath,
			NumThreads:   config.NumThreads,
			Debug:        0,
			ModelType:    "paraformer",
			ModelingUnit: "cjkchar", // 中文单字建模
		},
		DecodingMethod: "greedy_search",
	}

	recognizer := sherpa.NewOfflineRecognizer(configPtr)
	if recognizer == nil {
		return nil, fmt.Errorf("failed to create offline recognizer")
	}

	engine.recognizer = recognizer
	return engine, nil
}

// Recognize 执行语音识别
func (e *ParaformerEngine) Recognize(samples []int16, opts RecognitionOptions) (string, error) {
	result, err := e.recognizeInternal(samples, opts)
	if err != nil {
		return "", err
	}
	if result == nil {
		return "", nil
	}

	text := result.Text

	// 如果开启了标点优化
	if opts.Punctuation {
		if e.punct != nil && e.punct.IsAvailable() {
			// 使用专门的标点模型
			punctuatedText, err := e.punct.Apply(text)
			if err == nil {
				return punctuatedText, nil
			}
		}
		// 如果标点模型不可用，回退到基于时间戳的后处理
		return postProcessPunctuation(result), nil
	}

	return text, nil
}

// RecognizeDetailed 执行语音识别，返回带时间戳的详细结果
func (e *ParaformerEngine) RecognizeDetailed(samples []int16, opts RecognitionOptions) (*TimestampResult, error) {
	result, err := e.recognizeInternal(samples, opts)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, nil
	}

	text := result.Text
	// 如果开启了标点优化，应用到返回的 Text 中
	if opts.Punctuation {
		if e.punct != nil && e.punct.IsAvailable() {
			punctuatedText, err := e.punct.Apply(text)
			if err == nil {
				text = punctuatedText
			}
		} else {
			text = postProcessPunctuation(result)
		}
	}

	return &TimestampResult{
		Text:       text,
		Tokens:     result.Tokens,
		Timestamps: result.Timestamps,
	}, nil
}

func (e *ParaformerEngine) recognizeInternal(samples []int16, opts RecognitionOptions) (*sherpa.OfflineRecognizerResult, error) {
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
func (e *ParaformerEngine) Reload(config ModelConfig) error {
	// 关闭旧引擎
	if e.recognizer != nil {
		sherpa.DeleteOfflineRecognizer(e.recognizer)
		e.recognizer = nil
	}

	// 重新初始化
	engine, err := NewParaformerEngine(config)
	if err != nil {
		return err
	}

	e.config = engine.config
	e.info = engine.info
	e.recognizer = engine.recognizer
	return nil
}

// Switch 切换到新的模型配置
func (e *ParaformerEngine) Switch(config ModelConfig) error {
	return e.Reload(config)
}

// Close 释放资源
func (e *ParaformerEngine) Close() {
	if e.recognizer != nil {
		sherpa.DeleteOfflineRecognizer(e.recognizer)
		e.recognizer = nil
	}
}
