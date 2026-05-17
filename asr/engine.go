package asr

// RecognitionOptions 识别选项
type RecognitionOptions struct {
	Punctuation bool   // 是否添加标点
	Hotwords    string // 热词（可选）
}

// ModelInfo 模型信息
type ModelInfo struct {
	Name                  string // 显示名称
	Type                  string // "paraformer" | "whisper" | "sensevoice"
	Version               string // 模型版本
	SupportsHotwords      bool   // 是否支持热词
	SupportsPunctuation   bool   // 是否支持标点
	Size                  int64  // 文件大小（字节）
	RecommendedSampleRate int    // 推荐采样率
}

// ModelConfig 模型配置
type ModelConfig struct {
	Type                  string // 模型类型
	ModelPath             string // 模型文件路径
	TokensPath            string // tokens 文件路径
	HotwordsPath          string // 热词文件路径（Paraformer 等）
	ConvFrontendPath      string // conv_frontend.onnx 路径（仅 Qwen3 需要）
	TokenizerPath         string // tokenizer 目录路径（Qwen3 需要）
	BpeVocab              string // BPE 词汇表路径（Qwen3 需要，指向 vocab.json）
	PunctuationModelPath  string // 标点模型路径
	PunctuationTokensPath string // 标点 tokens 路径
	NumThreads            int    // 推理线程数
}

// ASREngine 统一的 ASR 引擎接口
type ASREngine interface {
	// Recognize 执行语音识别，返回完整文本
	Recognize(samples []int16, opts RecognitionOptions) (string, error)

	// RecognizeDetailed 执行语音识别，返回带时间戳的详细结果
	RecognizeDetailed(samples []int16, opts RecognitionOptions) (*TimestampResult, error)

	// Reload 热重载模型配置
	Reload(config ModelConfig) error

	// Switch 切换到新的模型配置
	Switch(config ModelConfig) error

	// GetInfo 获取模型信息
	GetInfo() ModelInfo

	// GetConfig 获取当前配置
	GetConfig() ModelConfig

	// Close 释放资源
	Close()
}

// EngineBase 引擎基类，包含通用字段
type EngineBase struct {
	config ModelConfig
	info   ModelInfo
}

// GetConfig 获取当前配置
func (e *EngineBase) GetConfig() ModelConfig {
	return e.config
}

// GetInfo 获取模型信息
func (e *EngineBase) GetInfo() ModelInfo {
	return e.info
}

// TimestampResult 带时间戳的识别结果
type TimestampResult struct {
	Text       string
	Tokens     []string
	Timestamps []float32
}

// GetTimestampResult 获取带时间戳的结果（部分模型支持）
type TimestampProvider interface {
	GetTimestampResult() *TimestampResult
}

// HotwordScore 热词得分（用于调试）
type HotwordScore struct {
	Word  string
	Score float32
}
