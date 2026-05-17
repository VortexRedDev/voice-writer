package asr

import "fmt"

// EngineFactory ASR 引擎工厂
// 根据模型类型创建对应的引擎实例
type EngineFactory struct{}

// NewEngine 创建引擎实例
// modelType 支持: paraformer, sense-voice, qwen3-asr (目录名)
func (f *EngineFactory) NewEngine(config ModelConfig) (ASREngine, error) {
	switch config.Type {
	case "paraformer":
		return NewParaformerEngine(config)
	case "sense-voice":
		return NewSenseVoiceEngine(config)
	case "qwen3-asr":
		return NewQwen3Engine(config)
	// 兼容旧名称
	case "sensevoice":
		config.Type = "sense-voice"
		return NewSenseVoiceEngine(config)
	case "qwen3":
		config.Type = "qwen3-asr"
		return NewQwen3Engine(config)
	default:
		return nil, fmt.Errorf("unsupported engine type: %s", config.Type)
	}
}

// SupportedEngines 返回支持的引擎类型列表
func (f *EngineFactory) SupportedEngines() []string {
	return []string{"paraformer", "sense-voice", "qwen3-asr"}
}
