package strategy

// HotwordStrategy 热词策略接口
// 不同模型处理热词的方式不同：
// - 内置热词：模型支持直接传入热词参数
// - 上下文图：需要通过 context_graph.onnx 实现
// - 不支持：模型不支持热词功能
type HotwordStrategy interface {
	// SetHotwords 设置热词
	SetHotwords(hotwords string) error

	// IsAvailable 策略是否可用
	IsAvailable() bool
}

// NoneHotwordStrategy 不支持热词的策略（用于 SenseVoice 等）
type NoneHotwordStrategy struct{}

func (s *NoneHotwordStrategy) SetHotwords(hotwords string) error {
	// 不支持热词，静默忽略
	return nil
}

func (s *NoneHotwordStrategy) IsAvailable() bool {
	return false
}

// BuiltinHotwordStrategy 内置热词策略（用于 Qwen3 等支持直接热词参数的模型）
type BuiltinHotwordStrategy struct {
	hotwords string
}

func (s *BuiltinHotwordStrategy) SetHotwords(hotwords string) error {
	s.hotwords = hotwords
	return nil
}

func (s *BuiltinHotwordStrategy) IsAvailable() bool {
	return true
}

func (s *BuiltinHotwordStrategy) GetHotwords() string {
	return s.hotwords
}

// ContextGraphHotwordStrategy 上下文图热词策略（用于 Paraformer 等）
// 需要 context_graph.onnx 文件
type ContextGraphHotwordStrategy struct {
	contextGraphPath string
	hotwords         string
}

func (s *ContextGraphHotwordStrategy) SetHotwords(hotwords string) error {
	s.hotwords = hotwords
	// 注意：实际的 context_graph 需要通过专门的工具生成
	// 这里只是保存热词配置，实际生效需要重新构建 context_graph
	return nil
}

func (s *ContextGraphHotwordStrategy) IsAvailable() bool {
	// 检查 context_graph 文件是否存在
	// if contextGraphPath != "" && file exists {
	//     return true
	// }
	return s.contextGraphPath != ""
}

func (s *ContextGraphHotwordStrategy) GetContextGraphPath() string {
	return s.contextGraphPath
}

func (s *ContextGraphHotwordStrategy) GetHotwords() string {
	return s.hotwords
}
