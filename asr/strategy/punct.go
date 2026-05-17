package strategy

import (
	"fmt"

	sherpa "github.com/k2-fsa/sherpa-onnx-go/sherpa_onnx"
)

// PunctStrategy 标点策略接口
// 不同模型处理标点的方式不同：
// - 内置标点：模型输出已带标点，不需要额外处理
// - 外部标点：需要调用专门的标点模型
type PunctStrategy interface {
	// Apply 对文本应用标点
	Apply(text string) (string, error)

	// IsAvailable 策略是否可用
	IsAvailable() bool
}

// BuiltinPunctStrategy 内置标点策略（用于 SenseVoice、Qwen3 等内置标点的模型）
type BuiltinPunctStrategy struct{}

func (s *BuiltinPunctStrategy) Apply(text string) (string, error) {
	// 内置标点，文本已经带标点，直接返回
	return text, nil
}

func (s *BuiltinPunctStrategy) IsAvailable() bool {
	return true
}

// NoPunctStrategy 无标点策略（用于不需要标点的场景）
type NoPunctStrategy struct{}

func (s *NoPunctStrategy) Apply(text string) (string, error) {
	// 移除文本中的标点符号
	result := make([]rune, 0, len(text))
	for _, r := range text {
		if isChinesePunct(r) || isEnglishPunct(r) {
			continue
		}
		result = append(result, r)
	}
	return string(result), nil
}

func (s *NoPunctStrategy) IsAvailable() bool {
	return true
}

func isChinesePunct(r rune) bool {
	return r == '。' || r == '，' || r == '？' || r == '！' ||
		r == '、' || r == '；' || r == '：' || r == '"' ||
		r == '\'' || r == '(' || r == ')' || r == '[' ||
		r == ']' || r == '《' || r == '》'
}

func isEnglishPunct(r rune) bool {
	return r == '.' || r == ',' || r == '?' || r == '!' ||
		r == ';' || r == ':' || r == '\'' || r == '(' ||
		r == ')' || r == '[' || r == ']' || r == '-' || r == '"'
}

// OfflinePunctStrategy 外部标点模型策略（用于 Paraformer 等需要专用标点模型的引擎）
// 使用 sherpa-onnx 的 OfflinePunctuation 模型
type OfflinePunctStrategy struct {
	modelPath  string
	tokensPath string
	punctModel *sherpa.OfflinePunctuation
}

// NewOfflinePunctStrategy 创建外部标点策略
func NewOfflinePunctStrategy(modelPath, tokensPath string) *OfflinePunctStrategy {
	s := &OfflinePunctStrategy{
		modelPath:  modelPath,
		tokensPath: tokensPath,
	}

	// 初始化标点模型 - sherpa-onnx 新版本使用 CtTransformer 字段
	config := sherpa.OfflinePunctuationConfig{
		Model: sherpa.OfflinePunctuationModelConfig{
			CtTransformer: modelPath,
		},
	}

	punctModel := sherpa.NewOfflinePunctuation(&config)
	if punctModel != nil {
		s.punctModel = punctModel
	}

	return s
}

// IsAvailable 检查标点模型是否可用
func (s *OfflinePunctStrategy) IsAvailable() bool {
	return s.punctModel != nil
}

// Apply 对文本应用标点
func (s *OfflinePunctStrategy) Apply(text string) (string, error) {
	if !s.IsAvailable() {
		return text, fmt.Errorf("punctuation model not available")
	}

	// 调用 sherpa-onnx 的 OfflinePunctuation 模型
	result := s.punctModel.AddPunct(text)
	if result == "" {
		return text, nil
	}

	return result, nil
}
