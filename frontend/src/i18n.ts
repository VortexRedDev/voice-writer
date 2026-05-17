import { ref, computed } from 'vue';

export type Locale = 'zh-CN' | 'en-US';

const messages = {
  'zh-CN': {
    nav: {
      main: '主面板',
      files: '音频处理',
      settings: '设置',
      model: '模型引擎'
    },
    files: {
      title: '音频文件识别',
      desc: '将音频文件拖放到此处进行识别。结果将以 .txt 和 .srt 格式保存在文件同目录下。',
      dropTip: '将音频文件 (WAV) 拖到此处',
      processing: '正在处理...',
      success: '识别完成，已保存 .txt 和 .srt 文件至同目录',
      error: '处理失败',
      formatLabel: '输出格式',
      onlyWav: '目前仅支持 16-bit PCM WAV 格式'
    },
    header: {
      initializing: '正在初始化...',
      systemReady: '系统就绪',
      recording: '正在录音',
      processing: '正在处理',
      ready: '就绪'
    },
    home: {
      outputLabel: '识别结果输出',
      waitRecognition: '等待语音输入...',
      holdHotkey: '按住快捷键开始录音',
      hotkeyLabel: '快捷键'
    },
    settings: {
      hotkeyTitle: '唤醒快捷键',
      change: '修改',
      recording: '按下按键...',
      hotkeyDesc: '用于在任何窗口触发语音输入的全局快捷键。',
      featuresTitle: '识别功能',
      autoPunctuation: '自动标点符号',
      featuresDesc: '使用 AI 模型自动为文本添加逗号和句号。',
      hotwordsTitle: '自定义热词',
      hotwordsPlaceholder: '例如：\n人工智能 2.0\n语音助手 1.5',
      hotwordsDesc: '每行一个词组 • 词组 权重',
      applyChanges: '应用修改'
    },
    model: {
      title: '可用模型',
      desc: '选择用于语音识别的 ASR 引擎',
      active: '正在使用',
      loading: '加载中',
      select: '选择',
      noticeTitle: '模型说明',
      noticeDesc: '不同模型各有侧重。Qwen3-ASR 支持热词和内置标点，SenseVoice 对通用语音识别极其精准。请确保模型文件已放置在 models 目录下。',
      switchingTitle: '正在切换引擎',
      switchingDesc: '请稍候，正在将 ONNX 模型加载到内存中...'
    },
    toast: {
      saved: '设置已保存',
      switchFailed: '切换失败'
    },
    footer: {
      ready: '系统就绪',
      audioInfo: '音频: 16kHz 单声道',
      bufferInfo: '缓冲区: 已重采样'
    }
  },
  'en-US': {
    nav: {
      main: 'Dashboard',
      files: 'Audio Files',
      settings: 'Settings',
      model: 'Engine'
    },
    files: {
      title: 'Audio Recognition',
      desc: 'Drag and drop audio files here for recognition. Results will be saved as .txt and .srt in the same directory.',
      dropTip: 'Drop WAV files here',
      processing: 'Processing...',
      success: 'Done! Saved .txt and .srt files to folder',
      error: 'Processing failed',
      formatLabel: 'Output Format',
      onlyWav: 'Only 16-bit PCM WAV is supported'
    },
    header: {
      initializing: 'Initializing...',
      systemReady: 'System Ready',
      recording: 'Recording',
      processing: 'Processing',
      ready: 'Ready'
    },
    home: {
      outputLabel: 'Recognition Output',
      waitRecognition: 'Wait for recognition...',
      holdHotkey: 'Hold hotkey to start recording',
      hotkeyLabel: 'Hotkey'
    },
    settings: {
      hotkeyTitle: 'Activation Hotkey',
      change: 'Change',
      recording: 'Recording...',
      hotkeyDesc: 'Global hotkey to trigger voice input from anywhere.',
      featuresTitle: 'Recognition Features',
      autoPunctuation: 'Auto Punctuation',
      featuresDesc: 'Optimize text with commas and periods using AI models.',
      hotwordsTitle: 'Custom Hotwords',
      hotwordsPlaceholder: 'Example:\nOpenAI 2.0\nGemini 1.5',
      hotwordsDesc: 'One phrase per line • Word Weight',
      applyChanges: 'Apply Changes'
    },
    model: {
      title: 'Available Models',
      desc: 'Select an ASR engine for recognition',
      active: 'Active',
      loading: 'Loading',
      select: 'Select',
      noticeTitle: 'Model Notice',
      noticeDesc: 'Each model has different strengths. Qwen3-ASR supports hotwords and built-in punctuation, while SenseVoice is extremely accurate for general speech.',
      switchingTitle: 'Switching Engine',
      switchingDesc: 'Please wait while the ONNX model is being loaded into memory...'
    },
    toast: {
      saved: 'Settings Saved',
      switchFailed: 'Switch Failed'
    },
    footer: {
      ready: 'System Ready',
      audioInfo: 'Audio: 16kHz Mono',
      bufferInfo: 'Buffer: Resampled'
    }
  }
};

const currentLocale = ref<Locale>('zh-CN');

export function useI18n() {
  const t = (path: string) => {
    const keys = path.split('.');
    let result: any = messages[currentLocale.value];
    for (const key of keys) {
      if (result[key]) {
        result = result[key];
      } else {
        return path;
      }
    }
    return result;
  };

  const setLocale = (locale: Locale) => {
    currentLocale.value = locale;
  };

  return {
    t,
    locale: computed(() => currentLocale.value),
    setLocale
  };
}
