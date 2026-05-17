<template>
  <div v-if="show" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
    <div class="bg-gray-900 p-6 rounded-lg shadow-xl w-96 border border-gray-700 text-gray-100">
      <h2 class="text-xl font-bold mb-4">设置</h2>
      <!-- 快捷键配置 -->
      <div class="mb-4">
        <label class="block text-sm text-gray-400 mb-1">快捷键</label>
        <div class="flex gap-2">
          <input 
            v-model="localConfig.hotkey" 
            readonly
            class="flex-1 bg-gray-800 border border-gray-700 rounded p-2 text-white text-sm"
            :placeholder="isRecording ? '请按下按键...' : 'F9'"
          />
          <button 
            @click="recordHotkey" 
            :class="['px-3 py-1 rounded text-xs', isRecording ? 'bg-red-600' : 'bg-gray-700']"
          >
            {{ isRecording ? '录制中...' : '修改' }}
          </button>
        </div>
      </div>


      <!-- 热词配置 -->
      <div class="mb-4">
        <label class="block text-sm text-gray-400 mb-1">自定义热词 (每行一个词，空格后接权重，例如: 人工智能 2.0)</label>
        <textarea 
          v-model="hotwords" 
          rows="4"
          class="w-full bg-gray-800 border border-gray-700 rounded p-2 text-white text-sm"
          placeholder="例如：人工智能 2.0"
        ></textarea>
      </div>

      <div class="mb-4">
        <label class="flex items-center text-sm text-gray-300">
          <input type="checkbox" v-model="localConfig.punctuation" class="mr-2" />
          自动标点优化
        </label>
      </div>

      <div class="flex justify-end gap-2">
        <button @click="$emit('close')" class="px-4 py-2 bg-gray-700 rounded hover:bg-gray-600">取消</button>
        <button @click="save" class="px-4 py-2 bg-blue-600 rounded hover:bg-blue-700">保存</button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted, onUnmounted } from 'vue';
import { SaveConfig, GetHotwords, SaveHotwords, StartRecordingHotkey, RecordHotkey } from '../../wailsjs/go/main/App';
import { EventsOn } from '../../wailsjs/runtime/runtime';

const props = defineProps<{ show: boolean, config: any }>();
const emit = defineEmits(['close', 'saved']);
const localConfig = ref({ ...props.config });
const hotwords = ref("");
const isRecording = ref(false);

// Key code to name mapping
const keyCodeToName: Record<number, string> = {
  112: 'f1', 113: 'f2', 114: 'f3', 115: 'f4', 116: 'f5', 117: 'f6',
  118: 'f7', 119: 'f8', 120: 'f9', 121: 'f10', 122: 'f11', 123: 'f12',
  65: 'a', 66: 'b', 67: 'c', 68: 'd', 69: 'e', 70: 'f',
  71: 'g', 72: 'h', 73: 'i', 74: 'j', 75: 'k', 76: 'l',
  77: 'm', 78: 'n', 79: 'o', 80: 'p', 81: 'q', 82: 'r',
  83: 's', 84: 't', 85: 'u', 86: 'v', 87: 'w', 88: 'x',
  89: 'y', 90: 'z',
  48: '0', 49: '1', 50: '2', 51: '3', 52: '4', 53: '5',
  54: '6', 55: '7', 56: '8', 57: '9',
};

function getKeyName(keyCode: number): string {
  // Check if it's a function key
  if (keyCode >= 112 && keyCode <= 123) {
    return `f${keyCode - 111}`;
  }
  // Check if it's a letter or number
  if (keyCodeToName[keyCode]) {
    return keyCodeToName[keyCode];
  }
  // Return unknown
  return `key${keyCode}`;
}

function handleKeyDown(event: KeyboardEvent) {
  if (!isRecording.value) return;
  
  event.preventDefault();
  event.stopPropagation();
  
  const keyCode = event.keyCode;
  const keyName = getKeyName(keyCode);
  
  console.log("Recording hotkey:", keyName, "code:", keyCode);
  
  // Stop recording mode
  isRecording.value = false;
  localConfig.value.hotkey = keyName;
  localConfig.value.hotkey_raw_code = keyCode;
  
  // Call backend to save
  RecordHotkey(keyCode, keyName);
}

watch(() => props.config, (newVal) => {
  localConfig.value = { ...newVal };
});

watch(() => props.show, (show) => {
  if (show && isRecording.value) {
    window.addEventListener('keydown', handleKeyDown, true);
  } else {
    window.removeEventListener('keydown', handleKeyDown, true);
  }
});

onMounted(async () => {
  hotwords.value = await GetHotwords();
  EventsOn("hotkey-recording-started", () => {
    // Already handled by v-if in template
  });
});

onUnmounted(() => {
  window.removeEventListener('keydown', handleKeyDown, true);
});

async function recordHotkey() {
  isRecording.value = true;
  window.addEventListener('keydown', handleKeyDown, true);
  await StartRecordingHotkey();
}

async function save() {
  await SaveConfig(localConfig.value);
  await SaveHotwords(hotwords.value);
  emit('saved');
  emit('close');
}
</script>
