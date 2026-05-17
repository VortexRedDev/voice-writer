package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"voice-writer/asr"
	"voice-writer/audio"
	"voice-writer/audio/file"
	"voice-writer/config"
	"voice-writer/output"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// VAD thresholds
const (
	MinSpeechRMS = 30.0 // Minimum RMS energy to consider as speech
	SilenceLimit = 10   // Number of consecutive silence chunks to detect end of speech
)

type App struct {
	ctx            context.Context
	recorder       *audio.Recorder
	asrEngine      asr.ASREngine
	outputter      *output.Outputter
	cfg            config.Config
	isRecordingKey bool
}

func NewApp() *App {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
	}

	fmt.Println("Initializing Recorder...")
	rec, err := audio.NewRecorder()
	if err != nil {
		fmt.Printf("Failed to create recorder: %v\n", err)
	} else {
		fmt.Println("Recorder Init Success")
	}

	return &App{
		recorder:  rec,
		outputter: output.NewOutputter(),
		cfg:       cfg,
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	err := copyDLLsToExecutableDir()
	if err != nil {
		fmt.Printf("Warning: Failed to copy DLLs: %v\n", err)
	}

	exePath, err := os.Executable()
	if err != nil {
		fmt.Printf("Warning: Failed to get exe path: %v\n", err)
	}
	baseDir := filepath.Dir(exePath)
	cwd, _ := os.Getwd()

	possibleModelTypes := []string{"paraformer", "sense-voice", "qwen3-asr"}
	// If a model is already selected in config, prioritize it
	if a.cfg.ModelID != "" {
		// Move the configured model to the front of the list
		for i, mt := range possibleModelTypes {
			if mt == a.cfg.ModelID {
				possibleModelTypes = append([]string{mt}, append(possibleModelTypes[:i], possibleModelTypes[i+1:]...)...)
				break
			}
		}
	}

	for _, modelType := range possibleModelTypes {
		for _, base := range []string{baseDir, cwd, filepath.Dir(cwd)} {
			modelPath := filepath.Join(base, "models", modelType, "model.int8.onnx")
			tokensPath := filepath.Join(base, "models", modelType, "tokens.txt")

			// 也检查 encoder.int8.onnx (qwen3-asr 风格)
			if modelType == "qwen3-asr" {
				modelPath = filepath.Join(base, "models", modelType, "encoder.int8.onnx")
				tokensPath = filepath.Join(base, "models", modelType, "decoder.int8.onnx")
			}

			fmt.Printf("Searching for %s model at: %s\n", modelType, modelPath)
			if _, err := os.Stat(modelPath); err == nil {
				hotwordsPath := config.GetHotwordsFilePath()
				modelConfig := asr.ModelConfig{
					Type:         modelType,
					ModelPath:    modelPath,
					TokensPath:   tokensPath,
					HotwordsPath: hotwordsPath,
					NumThreads:   4,
				}
				factory := &asr.EngineFactory{}
				engine, err := factory.NewEngine(modelConfig)
				if err != nil {
					fmt.Printf("ASR Engine Init Failed for %s: %v\n", modelType, err)
					continue
				}
				a.asrEngine = engine
				fmt.Printf("ASR Engine Init Success (%s) \n", modelType)
				go a.listenHotkey()
				return
			}
		}
	}

	fmt.Printf("ASR Error: Could not find any ASR model\n")
}

func copyDLLsToExecutableDir() error {
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}
	exeDir := filepath.Dir(exePath)
	cwd, _ := os.Getwd()

	dlls := []string{
		"onnxruntime.dll",
		"sherpa-onnx-c-api.dll",
		"sherpa-onnx-cxx-api.dll",
	}

	for _, dll := range dlls {
		targetPath := filepath.Join(exeDir, dll)
		if _, err := os.Stat(targetPath); err == nil {
			fmt.Printf("DLL '%s' already in executable directory.\n", dll)
			continue
		}

		sourcePaths := []string{
			filepath.Join(cwd, dll),
			filepath.Join(cwd, "build", "bin", dll),
			filepath.Join(cwd, "..", "build", "bin", dll),
		}

		found := false
		for _, sourcePath := range sourcePaths {
			if _, err := os.Stat(sourcePath); err == nil {
				fmt.Printf("Copying DLL '%s' from '%s' to '%s'\n", dll, sourcePath, targetPath)
				content, err := os.ReadFile(sourcePath)
				if err != nil {
					fmt.Printf("Error reading DLL '%s': %v\n", sourcePath, err)
					continue
				}
				err = os.WriteFile(targetPath, content, 0755)
				if err != nil {
					fmt.Printf("Error writing DLL '%s' to '%s': %v\n", dll, targetPath, err)
					continue
				}
				found = true
				break
			}
		}

		if !found {
			fmt.Printf("Warning: Could not find or copy DLL '%s'. ASR might fail.\n", dll)
		}
	}
	return nil
}

func (a *App) listenHotkey() {
	if a.cfg.HotkeyRawCode == 0 {
		a.cfg.Hotkey = "f9"
		a.cfg.HotkeyRawCode = 120
		config.SaveConfig(a.cfg)
		fmt.Printf("Hotkey Listener: No hotkey configured, setting default: F9\n")
	}

	fmt.Printf("Hotkey Listener: Polling GetAsyncKeyState for raw code %d (%s)\n", a.cfg.HotkeyRawCode, a.cfg.Hotkey)

	var wasDown bool
	for {
		isDown := isKeyDown(a.cfg.HotkeyRawCode)

		if isDown && !wasDown && !a.isRecordingKey && !a.recorder.IsRecording() {
			fmt.Println("Hotkey Listener: Hotkey pressed, starting recording.")
			a.StartRecording()
		}
		if !isDown && wasDown && !a.isRecordingKey && a.recorder.IsRecording() {
			fmt.Println("Hotkey Listener: Hotkey released, stopping recording.")
			a.StopRecording()
		}

		wasDown = isDown
		time.Sleep(50 * time.Millisecond)
	}
}

// isKeyDown checks if a key is currently pressed using GetAsyncKeyState
func isKeyDown(rawCode uint16) bool {
	user32 := syscall.NewLazyDLL("user32.dll")
	proc := user32.NewProc("GetAsyncKeyState")
	ret, _, _ := proc.Call(uintptr(rawCode))
	return ret&0x8000 != 0
}

func (a *App) shutdown(ctx context.Context) {
	if a.recorder != nil {
		a.recorder.Close()
	}
	if a.asrEngine != nil {
		a.asrEngine.Close()
	}
}

func (a *App) StartRecordingHotkey() {
	fmt.Println("StartRecordingHotkey: Starting hotkey recording mode...")
	// Notify frontend to start listening for keypress
	if a.ctx != nil {
		runtime.EventsEmit(a.ctx, "hotkey-recording-started", "")
	}
}

func (a *App) RecordHotkey(keyCode uint16, keyName string) string {
	fmt.Printf("RecordHotkey: Recording key: %s (code: %d)\n", keyName, keyCode)

	// Update config
	a.cfg.Hotkey = keyName
	a.cfg.HotkeyRawCode = keyCode

	// Save config
	err := config.SaveConfig(a.cfg)
	if err != nil {
		return fmt.Sprintf("Failed to save hotkey: %v", err)
	}

	// Notify frontend
	if a.ctx != nil {
		runtime.EventsEmit(a.ctx, "hotkey-recorded", keyName)
	}

	return "Hotkey saved"
}

func (a *App) StartRecording() string {
	if a.recorder == nil {
		return "Recorder not initialized"
	}
	err := a.recorder.Start()
	if err != nil {
		return fmt.Sprintf("Error starting recording: %v", err)
	}
	// Emit status change event to frontend
	if a.ctx != nil {
		runtime.EventsEmit(a.ctx, "status-change", "recording")
	}
	return "Recording started"
}

func (a *App) StopRecording() string {
	if a.recorder == nil {
		return "Recorder not initialized"
	}
	data, err := a.recorder.Stop()
	if err != nil {
		return fmt.Sprintf("Error stopping recording: %v", err)
	}

	fmt.Printf("Stopped recording. Captured %d samples\n", len(data))

	if len(data) == 0 {
		runtime.EventsEmit(a.ctx, "status-change", "idle")
		return "No audio captured"
	}

	// Check audio energy using RMS
	rms := audio.CalculateRMS(data)
	fmt.Printf("Audio RMS energy: %.2f (threshold: %.2f)\n", rms, MinSpeechRMS)

	// If audio is too quiet (only noise), skip recognition
	if rms < MinSpeechRMS {
		fmt.Printf("Audio too quiet (noise only), skipping recognition\n")
		runtime.EventsEmit(a.ctx, "status-change", "idle")
		return ""
	}

	if a.asrEngine == nil {
		return "ASR engine not initialized"
	}
	go func() {
		runtime.EventsEmit(a.ctx, "status-change", "recognizing")
		
		// 获取最新热词
		hotwords, _ := config.LoadHotwords()
		
		text, err := a.asrEngine.Recognize(data, asr.RecognitionOptions{
			Punctuation: a.cfg.Punctuation,
			Hotwords:    hotwords,
		})
		if err != nil {
			runtime.EventsEmit(a.ctx, "recognition-error", err.Error())
			runtime.EventsEmit(a.ctx, "status-change", "idle")
			return
		}

		err = a.outputter.TypeText(text)
		if err != nil {
			runtime.EventsEmit(a.ctx, "recognition-error", fmt.Sprintf("Failed to type: %v", err))
		}

		runtime.EventsEmit(a.ctx, "recognition-result", text)
		runtime.EventsEmit(a.ctx, "status-change", "idle")
	}()

	return "Processing audio..."
}

func (a *App) GetConfig() config.Config {
	return a.cfg
}

func (a *App) SaveConfig(cfg config.Config) string {
	err := config.SaveConfig(cfg)
	if err != nil {
		return fmt.Sprintf("Failed to save config: %v", err)
	}
	a.cfg = cfg
	return "Config saved"
}

func (a *App) GetStatus() string {
	if a.recorder != nil && a.recorder.IsRecording() {
		return "recording"
	}
	return "idle"
}

func (a *App) ListAudioDevices() []string {
	return nil
}

func (a *App) GetAudioDeviceName() string {
	info := audio.GetDeviceInfo()
	if info == nil {
		return "未知设备"
	}
	name, ok := info["name"].(string)
	if !ok {
		return "未知设备"
	}
	return name
}

func (a *App) GetAudioDeviceInfo() map[string]interface{} {
	return audio.GetDeviceInfo()
}

func (a *App) GetModelInfo() map[string]interface{} {
	if a.asrEngine == nil {
		return nil
	}
	info := a.asrEngine.GetInfo()
	return map[string]interface{}{
		"name":                  info.Name,
		"type":                  info.Type,
		"supportsHotwords":      info.SupportsHotwords,
		"supportsPunctuation":   info.SupportsPunctuation,
		"version":               info.Version,
		"recommendedSampleRate": info.RecommendedSampleRate,
	}
}

func (a *App) ReloadModel() string {
	if a.asrEngine == nil {
		return "ASR engine not initialized"
	}

	// 获取当前引擎配置
	config := a.asrEngine.GetConfig()

	// 调用引擎的 Reload 方法
	if err := a.asrEngine.Reload(config); err != nil {
		return fmt.Sprintf("Failed to reload model: %v", err)
	}

	return "Model reloaded successfully"
}

func (a *App) GetHotwords() string {
	content, _ := config.LoadHotwords()
	return content
}

func (a *App) SaveHotwords(content string) string {
	err := config.SaveHotwords(content)
	if err != nil {
		return fmt.Sprintf("Failed to save hotwords: %v", err)
	}
	
	// 保存热词后自动重新加载模型以使热词生效
	if a.asrEngine != nil {
		a.ReloadModel()
	}
	
	return "Hotwords saved"
}

// ModelInfo 模型信息（用于前端）
type ModelInfoJS struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Size        int64  `json:"size"` // 文件大小（字节）
	ModelPath   string `json:"modelPath"`
	TokensPath  string `json:"tokensPath"`
	ModelDir    string `json:"modelDir"` // 模型目录
}

// formatFileSize 格式化文件大小
func formatFileSize(bytes int64) string {
	if bytes >= 1024*1024*1024 {
		return fmt.Sprintf("%.1fGB", float64(bytes)/(1024*1024*1024))
	}
	if bytes >= 1024*1024 {
		return fmt.Sprintf("%.0fMB", float64(bytes)/(1024*1024))
	}
	if bytes >= 1024 {
		return fmt.Sprintf("%.0fKB", float64(bytes)/1024)
	}
	return fmt.Sprintf("%dB", bytes)
}

// GetAvailableModels 返回固定的3个模型配置
func (a *App) GetAvailableModels() []ModelInfoJS {
	return []ModelInfoJS{
		{Id: "paraformer", Name: "Paraformer", Description: "支持热词，需外部标点模型", ModelPath: "models/paraformer/model.int8.onnx", TokensPath: "models/paraformer/tokens.txt", ModelDir: "models/paraformer"},
		{Id: "sense-voice", Name: "SenseVoice", Description: "内置标点，不支持热词", ModelPath: "models/sense-voice/model.int8.onnx", TokensPath: "models/sense-voice/tokens.txt", ModelDir: "models/sense-voice"},
		{Id: "qwen3-asr", Name: "Qwen3-ASR", Description: "内置热词和标点", ModelPath: "models/qwen3-asr/encoder.int8.onnx", TokensPath: "models/qwen3-asr/decoder.int8.onnx", ModelDir: "models/qwen3-asr"},
	}
}

// SwitchToModel 切换到指定模型
func (a *App) SwitchToModel(modelId string) string {
	if a.asrEngine == nil {
		return "ASR engine not initialized"
	}

	// 获取项目根目录
	cwd, _ := os.Getwd()
	hotwordsPath := config.GetHotwordsFilePath()
	punctModelPath := filepath.Join(cwd, "models/punctuation/model.int8.onnx")
	punctTokensPath := filepath.Join(cwd, "models/punctuation/tokens.json")

	// 根据 modelId 确定模型配置
	var cfg asr.ModelConfig

	switch modelId {
	case "paraformer":
		cfg = asr.ModelConfig{
			Type:                  "paraformer",
			ModelPath:             filepath.Join(cwd, "models/paraformer/model.int8.onnx"),
			TokensPath:            filepath.Join(cwd, "models/paraformer/tokens.txt"),
			HotwordsPath:          hotwordsPath,
			PunctuationModelPath:  punctModelPath,
			PunctuationTokensPath: punctTokensPath,
			NumThreads:            4,
		}
	case "sense-voice":
		cfg = asr.ModelConfig{
			Type:                  "sense-voice",
			ModelPath:             filepath.Join(cwd, "models/sense-voice/model.int8.onnx"),
			TokensPath:            filepath.Join(cwd, "models/sense-voice/tokens.txt"),
			PunctuationModelPath:  punctModelPath,
			PunctuationTokensPath: punctTokensPath,
			NumThreads:            4,
		}
	case "qwen3-asr":
		cfg = asr.ModelConfig{
			Type:                  "qwen3-asr",
			ModelPath:             filepath.Join(cwd, "models/qwen3-asr/encoder.int8.onnx"),
			TokensPath:            filepath.Join(cwd, "models/qwen3-asr/decoder.int8.onnx"),
			ConvFrontendPath:      filepath.Join(cwd, "models/qwen3-asr/conv_frontend.onnx"),
			TokenizerPath:         filepath.Join(cwd, "models/qwen3-asr/tokenizer"),
			BpeVocab:              filepath.Join(cwd, "models/qwen3-asr/tokenizer/vocab.json"),
			HotwordsPath:          hotwordsPath,
			PunctuationModelPath:  punctModelPath,
			PunctuationTokensPath: punctTokensPath,
			NumThreads:            4,
		}
	default:
		return "Unknown model type: " + modelId
	}

	// 获取当前引擎类型
	currentType := a.asrEngine.GetConfig().Type

	// 如果类型改变，需要创建新的引擎实例
	if currentType != cfg.Type {
		// 关闭旧引擎
		a.asrEngine.Close()

		// 使用工厂创建新引擎
		factory := &asr.EngineFactory{}
		engine, err := factory.NewEngine(cfg)
		if err != nil {
			return fmt.Sprintf("Failed to create engine: %v", err)
		}

		a.asrEngine = engine
		return "Switched to model: " + modelId
	}

	// 同类型引擎，只需切换配置
	if err := a.asrEngine.Switch(cfg); err != nil {
		return fmt.Sprintf("Failed to switch model: %v", err)
	}

	a.cfg.ModelID = modelId
	_ = config.SaveConfig(a.cfg)

	return "Switched to model: " + modelId
}

// ProcessAudioFile 处理音频文件识别
func (a *App) ProcessAudioFile(filePath string) string {
	if a.asrEngine == nil {
		return "ASR engine not initialized"
	}

	// 1. 解码音频文件
	samples, _, err := file.DecodeWavToS16(filePath)
	if err != nil {
		return fmt.Sprintf("Failed to decode audio: %v", err)
	}

	// 2. 获取最新热词
	hotwords, _ := config.LoadHotwords()

	// 3. 执行识别
	detailedResult, err := a.asrEngine.RecognizeDetailed(samples, asr.RecognitionOptions{
		Punctuation: a.cfg.Punctuation,
		Hotwords:    hotwords,
	})
	if err != nil {
		return fmt.Sprintf("Recognition failed: %v", err)
	}

	if detailedResult == nil {
		return "No speech detected"
	}

	// 4. 保存结果到同目录
	ext := filepath.Ext(filePath)
	basePath := filePath[:len(filePath)-len(ext)]

	// 保存 .txt
	err = os.WriteFile(basePath+".txt", []byte(detailedResult.Text), 0644)
	if err != nil {
		return fmt.Sprintf("Failed to save .txt result: %v", err)
	}

	// 保存 .srt
	srtContent := asr.ToSRT(detailedResult)
	if srtContent != "" {
		err = os.WriteFile(basePath+".srt", []byte(srtContent), 0644)
		if err != nil {
			return fmt.Sprintf("Failed to save .srt result: %v", err)
		}
	}

	return "success"
}
