package asr

import (
	"fmt"
	"os"
	"regexp"
	"voice-writer/asr/strategy"

	sherpa "github.com/k2-fsa/sherpa-onnx-go/sherpa_onnx"
)

// SenseVoiceEngine 基于 sherpa-onnx SenseVoice 的引擎实现
type SenseVoiceEngine struct {
	EngineBase
	recognizer *sherpa.OfflineRecognizer
	punct      *strategy.OfflinePunctStrategy
}

// NewSenseVoiceEngine 创建 SenseVoice 引擎
func NewSenseVoiceEngine(config ModelConfig) (*SenseVoiceEngine, error) {
	// 获取文件大小
	var fileSize int64
	if info, err := os.Stat(config.ModelPath); err == nil {
		fileSize = info.Size()
	}

	engine := &SenseVoiceEngine{
		EngineBase: EngineBase{
			config: config,
			info: ModelInfo{
				Name:                  "SenseVoice",
				Type:                  "sensevoice",
				Version:               "int8",
				SupportsHotwords:      false, // SenseVoice 不支持热词
				SupportsPunctuation:   true,  // SenseVoice 内置标点
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
			SenseVoice: sherpa.OfflineSenseVoiceModelConfig{
				Model:    config.ModelPath,
				Language: "", // auto detect
			},
			Tokens:     config.TokensPath,
			NumThreads: config.NumThreads,
			Debug:      0,
		},
		DecodingMethod: "greedy_search",
	}

	recognizer := sherpa.NewOfflineRecognizer(configPtr)
	if recognizer == nil {
		return nil, fmt.Errorf("failed to create sensevoice offline recognizer")
	}

	engine.recognizer = recognizer
	return engine, nil
}

// Recognize 执行语音识别
func (e *SenseVoiceEngine) Recognize(samples []int16, opts RecognitionOptions) (string, error) {
	result, err := e.recognizeInternal(samples, opts)
	if err != nil {
		return "", err
	}
	if result == nil {
		return "", nil
	}

	text := result.Text

	// SenseVoice 的输出通常包含标签如 <|zh|><|HAPPY|> 等，需要清理
	text = cleanSenseVoiceTags(text)

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
func (e *SenseVoiceEngine) RecognizeDetailed(samples []int16, opts RecognitionOptions) (*TimestampResult, error) {
	result, err := e.recognizeInternal(samples, opts)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, nil
	}

	text := cleanSenseVoiceTags(result.Text)
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

func (e *SenseVoiceEngine) recognizeInternal(samples []int16, opts RecognitionOptions) (*sherpa.OfflineRecognizerResult, error) {
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

// cleanSenseVoiceTags 移除 SenseVoice 特有的标签（如 <|zh|>, <|HAPPY|> 等）
func cleanSenseVoiceTags(text string) string {
	re := regexp.MustCompile(`<\|.*?\|>`)
	return re.ReplaceAllString(text, "")
}

// Reload 热重载模型配置
func (e *SenseVoiceEngine) Reload(config ModelConfig) error {
	// 关闭旧引擎
	if e.recognizer != nil {
		sherpa.DeleteOfflineRecognizer(e.recognizer)
		e.recognizer = nil
	}

	// 重新初始化
	engine, err := NewSenseVoiceEngine(config)
	if err != nil {
		return err
	}

	e.config = engine.config
	e.info = engine.info
	e.recognizer = engine.recognizer
	return nil
}

// Switch 切换到新的模型配置
func (e *SenseVoiceEngine) Switch(config ModelConfig) error {
	return e.Reload(config)
}

// Close 释放资源
func (e *SenseVoiceEngine) Close() {
	if e.recognizer != nil {
		sherpa.DeleteOfflineRecognizer(e.recognizer)
		e.recognizer = nil
	}
}
