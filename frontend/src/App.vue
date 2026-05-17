<template>
  <div class="h-screen bg-[#0a0b0d] text-[#f3f4f6] flex overflow-hidden font-sans">
    <!-- Sidebar -->
    <aside class="w-20 glass-sidebar flex flex-col items-center py-8 z-10">
      <div class="mb-12">
        <div class="w-10 h-10 bg-blue-600 rounded-xl flex items-center justify-center shadow-lg shadow-blue-900/20">
          <svg xmlns="http://www.w3.org/2000/svg" class="w-6 h-6 text-white" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 2a3 3 0 0 0-3 3v7a3 3 0 0 0 6 0V5a3 3 0 0 0-3-3Z"/><path d="M19 10v2a7 7 0 0 1-14 0v-2"/><line x1="12" x2="12" y1="19" y2="22"/></svg>
        </div>
      </div>
      
      <nav class="flex flex-col gap-8">
        <button 
          v-for="tab in tabs" 
          :key="tab.id"
          @click="switchTab(tab.id)"
          :class="[
            'relative p-3 rounded-2xl transition-all duration-300 group',
            activeTab === tab.id ? 'text-blue-500 bg-blue-500/10' : 'text-gray-500 hover:text-gray-300 hover:bg-white/5'
          ]"
          :title="t(`nav.${tab.id}`)"
        >
          <component :is="tab.icon" class="w-6 h-6" />
          <div v-if="activeTab === tab.id" class="absolute left-0 top-1/2 -translate-y-1/2 w-1 h-6 bg-blue-500 rounded-r-full"></div>
        </button>
      </nav>

      <div class="mt-auto flex flex-col gap-4 items-center">
        <!-- Language Switcher -->
        <button 
          @click="toggleLocale"
          class="w-8 h-8 rounded-lg bg-white/5 border border-white/10 flex items-center justify-center text-[10px] font-bold text-gray-500 hover:text-gray-300 transition-colors"
          :title="locale === 'zh-CN' ? 'Switch to English' : '切换至中文'"
        >
          {{ locale === 'zh-CN' ? 'EN' : '中' }}
        </button>
        <div class="w-8 h-8 rounded-full bg-gradient-to-tr from-gray-800 to-gray-700 border border-white/10"></div>
      </div>
    </aside>

    <!-- Main Content -->
    <main class="flex-1 flex flex-col relative overflow-hidden bg-gradient-to-br from-[#0a0b0d] to-[#16181d]">
      <!-- Top Header -->
      <header class="h-16 flex items-center justify-between px-8 border-b border-white/5 bg-black/10 backdrop-blur-sm">
        <div class="flex items-center gap-4">
          <h2 class="text-lg font-medium tracking-tight">{{ t(`nav.${activeTab}`) }}</h2>
          <div class="h-4 w-[1px] bg-white/10"></div>
          <span class="text-xs text-gray-500 font-mono">{{ audioDeviceName === 'Initializing...' ? t('header.initializing') : audioDeviceName }}</span>
        </div>
        <div class="flex items-center gap-3">
          <div :class="['px-2.5 py-1 rounded-full text-[10px] font-bold uppercase tracking-wider border', statusBadgeClass]">
            {{ t(`header.${status === 'idle' ? 'ready' : status}`) }}
          </div>
        </div>
      </header>

      <!-- Scrollable Content -->
      <div class="flex-1 overflow-y-auto p-8 custom-scrollbar">
        <transition name="fade" mode="out-in">
          <div :key="activeTab" class="max-w-4xl mx-auto w-full">
            
            <!-- Home Tab -->
            <div v-if="activeTab === 'main'" class="flex flex-col items-center py-12">
              
              <!-- Horizontal Layout Container -->
              <div class="flex items-center justify-center gap-16 mb-16 w-full max-w-3xl">
                <!-- Instruction Card (Left) -->
                <div class="flex-1 text-right max-w-[280px]">
                  <p class="text-[10px] font-bold text-gray-500 uppercase tracking-[0.2em] mb-4 opacity-50">{{ t('home.holdHotkey') }}</p>
                  <div class="inline-flex flex-col items-end gap-1">
                    <div class="px-6 py-4 rounded-3xl bg-white/5 border border-white/10 shadow-2xl backdrop-blur-xl transition-all hover:bg-white/[0.08] group cursor-default">
                      <div class="flex items-baseline gap-3">
                        <span class="text-[10px] font-black text-gray-600 uppercase tracking-tighter">{{ t('home.hotkeyLabel') }}</span>
                        <span class="text-3xl font-black text-blue-500 tracking-tighter group-hover:scale-110 transition-transform inline-block origin-right">{{ appConfig.hotkey.toUpperCase() }}</span>
                      </div>
                    </div>
                    <p class="text-[10px] text-gray-600 mt-4 leading-relaxed pr-2 italic">
                      {{ locale === 'zh-CN' ? '按住预设热键开始录制语音，松开按键系统将自动识别并完成文字录入。' : 'Hold the hotkey to capture voice, release to automatically transcribe and type into the active window.' }}
                    </p>
                  </div>
                </div>

                <!-- Circular Button (Right) -->
                <div class="relative shrink-0">
                  <!-- Dynamic Glow Pulse -->
                  <div v-if="status === 'recording'" class="absolute inset-0 rounded-full bg-blue-500/20 animate-pulse-glow -m-6"></div>
                  
                  <button 
                    @mousedown="startRecording" 
                    @mouseup="stopRecording"
                    @mouseleave="stopRecording"
                    :class="[
                      'relative w-44 h-44 rounded-full flex items-center justify-center transition-all duration-700 select-none shadow-[0_0_50px_rgba(0,0,0,0.5)] border-4',
                      status === 'recording' 
                        ? 'bg-blue-600 border-blue-400 scale-110 shadow-[0_0_80px_rgba(59,130,246,0.4)]' 
                        : 'bg-white/5 hover:bg-white/10 border-white/10 hover:border-blue-500/30'
                    ]"
                  >
                    <!-- Normal Icon -->
                    <svg v-if="status !== 'recording'" xmlns="http://www.w3.org/2000/svg" class="w-16 h-16 text-white opacity-90 transition-transform group-hover:scale-110" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 2a3 3 0 0 0-3 3v7a3 3 0 0 0 6 0V5a3 3 0 0 0-3-3Z"/><path d="M19 10v2a7 7 0 0 1-14 0v-2"/><line x1="12" x2="12" y1="19" y2="22"/></svg>
                    
                    <!-- Recording Animation -->
                    <div v-else class="flex gap-2.5 items-center">
                      <div class="w-2.5 h-10 bg-white rounded-full animate-[bounce_0.8s_infinite_0ms]"></div>
                      <div class="w-2.5 h-16 bg-white rounded-full animate-[bounce_0.8s_infinite_200ms]"></div>
                      <div class="w-2.5 h-10 bg-white rounded-full animate-[bounce_0.8s_infinite_400ms]"></div>
                    </div>
                  </button>
                </div>
              </div>

              <!-- Recognition Display -->
              <div class="w-full max-w-2xl glass-card p-8 min-h-[200px] flex flex-col shadow-2xl border-t-blue-500/10 transition-all hover:bg-white/[0.04]">
                <div class="flex items-center justify-between mb-4 border-b border-white/5 pb-3">
                  <div class="flex items-center gap-2 text-[10px] font-bold text-gray-500 uppercase tracking-widest">
                    <svg xmlns="http://www.w3.org/2000/svg" class="w-3 h-3" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/></svg>
                    {{ t('home.outputLabel') }}
                  </div>
                  <div class="text-[10px] text-gray-600 font-mono">{{ currentModelId }}</div>
                </div>
                <div class="flex-1 text-lg leading-relaxed text-gray-200">
                  <p v-if="result">{{ result }}</p>
                  <p v-else class="text-gray-600 italic text-base">{{ t('home.waitRecognition') }}</p>
                </div>
              </div>
            </div>

            <!-- Audio Files Tab -->
            <div v-if="activeTab === 'files'" class="space-y-6 flex flex-col h-full max-h-[600px]">
              <div class="mb-2">
                <h3 class="text-xl font-bold text-white">{{ t('files.title') }}</h3>
                <p class="text-sm text-gray-500">{{ t('files.desc') }}</p>
              </div>

              <!-- Drop Zone -->
              <div 
                @dragover="onDragOver"
                @dragleave="onDragLeave"
                @drop="onFileDrop"
                :class="[
                  'flex-1 border-2 border-dashed rounded-3xl flex flex-col items-center justify-center transition-all duration-300 min-h-[300px]',
                  fileDropActive ? 'border-blue-500 bg-blue-500/5 scale-[0.99]' : 'border-white/10 bg-white/[0.02] hover:bg-white/[0.04]'
                ]"
              >
                <div :class="['w-16 h-16 rounded-2xl bg-blue-500/10 flex items-center justify-center mb-4 transition-transform duration-500', fileDropActive ? 'scale-110' : '']">
                  <svg xmlns="http://www.w3.org/2000/svg" class="w-8 h-8 text-blue-500" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="17 8 12 3 7 8"/><line x1="12" y1="3" x2="12" y2="15"/></svg>
                </div>
                <p class="text-lg font-medium text-gray-300">{{ t('files.dropTip') }}</p>
                <p class="text-xs text-gray-600 mt-2 font-mono uppercase tracking-widest">{{ t('files.onlyWav') }}</p>

                <!-- Processing List -->
                <div v-if="processingFiles.length > 0" class="mt-8 w-full max-w-xs space-y-2">
                  <div v-for="file in processingFiles" :key="file" class="bg-blue-600/10 border border-blue-500/20 rounded-xl px-4 py-3 flex items-center justify-between overflow-hidden">
                    <div class="flex items-center gap-3 truncate">
                      <div class="w-2 h-2 rounded-full bg-blue-500 animate-pulse"></div>
                      <span class="text-xs text-blue-400 font-medium truncate">{{ file }}</span>
                    </div>
                    <span class="text-[10px] text-blue-500 font-bold uppercase shrink-0">{{ t('files.processing') }}</span>
                  </div>
                </div>
              </div>
            </div>

            <!-- Settings Tab -->
            <div v-if="activeTab === 'settings'" class="space-y-6">
              <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
                <!-- Hotkey Card -->
                <div class="glass-card p-6">
                  <h3 class="text-sm font-semibold text-white mb-4 flex items-center gap-2">
                    <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4 text-blue-500" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="11" width="18" height="11" rx="2" ry="2"/><path d="M7 11V7a5 5 0 0 1 10 0v4"/></svg>
                    {{ t('settings.hotkeyTitle') }}
                  </h3>
                  <div class="flex items-center gap-3">
                    <div class="flex-1 bg-black/20 border border-white/10 rounded-xl px-4 py-3 text-sm font-mono text-white">
                      {{ localConfig.hotkey.toUpperCase() }}
                    </div>
                    <button 
                      @click="recordHotkey" 
                      :class="['px-6 py-3 rounded-xl text-sm font-medium transition-all shadow-lg', isRecording ? 'bg-red-500 shadow-red-900/20' : 'bg-blue-600 hover:bg-blue-500 shadow-blue-900/20']"
                    >
                      {{ isRecording ? t('settings.recording') : t('settings.change') }}
                    </button>
                  </div>
                  <p class="mt-3 text-xs text-gray-500">{{ t('settings.hotkeyDesc') }}</p>
                </div>

                <!-- Features Card -->
                <div class="glass-card p-6">
                  <h3 class="text-sm font-semibold text-white mb-4 flex items-center gap-2">
                    <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4 text-purple-500" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/></svg>
                    {{ t('settings.featuresTitle') }}
                  </h3>
                  <div class="space-y-4">
                    <label class="flex items-center justify-between group cursor-pointer">
                      <span class="text-sm text-gray-400 group-hover:text-gray-200 transition-colors">{{ t('settings.autoPunctuation') }}</span>
                      <div class="relative inline-flex items-center">
                        <input type="checkbox" v-model="localConfig.punctuation" class="sr-only peer" />
                        <div class="w-11 h-6 bg-gray-700 peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-blue-600"></div>
                      </div>
                    </label>
                  </div>
                  <p class="mt-4 text-xs text-gray-500">{{ t('settings.featuresDesc') }}</p>
                </div>
              </div>

              <!-- Hotwords Card -->
              <div class="glass-card p-6">
                <h3 class="text-sm font-semibold text-white mb-4 flex items-center gap-2">
                  <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4 text-amber-500" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/><line x1="16" y1="13" x2="8" y2="13"/><line x1="16" y1="17" x2="8" y2="17"/><polyline points="10 9 9 9 8 9"/></svg>
                  {{ t('settings.hotwordsTitle') }}
                </h3>
                <textarea 
                  v-model="hotwords" 
                  rows="6"
                  class="w-full bg-black/20 border border-white/10 rounded-xl p-4 text-white text-sm focus:ring-1 focus:ring-blue-500 focus:outline-none placeholder-gray-700 transition-all font-mono"
                  :placeholder="t('settings.hotwordsPlaceholder')"
                ></textarea>
                <div class="mt-3 flex items-center justify-between">
                  <span class="text-[10px] text-gray-600 uppercase tracking-widest font-bold">{{ t('settings.hotwordsDesc') }}</span>
                  <button @click="saveSettings" class="px-6 py-2 bg-blue-600 hover:bg-blue-500 text-white text-sm font-medium rounded-lg transition-all shadow-lg shadow-blue-900/20">
                    {{ t('settings.applyChanges') }}
                  </button>
                </div>
              </div>
            </div>

            <!-- Model Tab -->
            <div v-if="activeTab === 'model'" class="space-y-6">
              <div class="flex items-center justify-between mb-2">
                <div>
                  <h3 class="text-xl font-bold text-white">{{ t('model.title') }}</h3>
                  <p class="text-sm text-gray-500">{{ t('model.desc') }}</p>
                </div>
              </div>
              
              <div class="grid grid-cols-1 gap-4">
                <div 
                  v-for="model in availableModels" 
                  :key="model.id"
                  :class="[
                    'relative overflow-hidden group glass-card p-5 transition-all duration-300',
                    currentModelId === model.id ? 'border-blue-500/50 bg-blue-500/5' : 'hover:bg-white/[0.05]'
                  ]"
                >
                  <div class="flex items-center justify-between relative z-10">
                    <div class="flex gap-4 items-center">
                      <div :class="['w-12 h-12 rounded-2xl flex items-center justify-center', currentModelId === model.id ? 'bg-blue-500 text-white' : 'bg-white/5 text-gray-400']">
                        <svg xmlns="http://www.w3.org/2000/svg" class="w-6 h-6" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"/><polyline points="3.27 6.96 12 12.01 20.73 6.96"/><line x1="12" y1="22.08" x2="12" y2="12"/></svg>
                      </div>
                      <div>
                        <div class="flex items-center gap-2">
                          <h4 class="font-bold text-white">{{ model.name }}</h4>
                          <span v-if="currentModelId === model.id" class="px-2 py-0.5 rounded text-[10px] bg-blue-500/20 text-blue-400 font-bold uppercase tracking-widest">{{ t('model.active') }}</span>
                        </div>
                        <p class="text-xs text-gray-500 mt-1">{{ model.description }} • {{ formatSize(model.size) }}</p>
                      </div>
                    </div>
                    
                    <button 
                      v-if="currentModelId !== model.id"
                      :disabled="switchingModelId === model.id"
                      @click="switchToModel(model)"
                      :class="[
                        'px-5 py-2 rounded-xl text-xs font-bold uppercase tracking-wider transition-all',
                        switchingModelId === model.id 
                          ? 'bg-gray-800 text-gray-600' 
                          : 'bg-white/5 hover:bg-white/10 text-white border border-white/10'
                      ]"
                    >
                      {{ switchingModelId === model.id ? t('model.loading') : t('model.select') }}
                    </button>
                  </div>
                  
                  <!-- Progress bar for switching -->
                  <div v-if="switchingModelId === model.id" class="absolute bottom-0 left-0 right-0 h-1 bg-blue-500/20">
                    <div class="h-full bg-blue-500 animate-[shimmer_2s_infinite]"></div>
                  </div>
                </div>
              </div>

              <div class="mt-8 p-6 rounded-2xl bg-amber-500/5 border border-amber-500/10">
                <div class="flex gap-4">
                  <svg xmlns="http://www.w3.org/2000/svg" class="w-6 h-6 text-amber-500 flex-shrink-0" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/></svg>
                  <div>
                    <h5 class="text-sm font-bold text-amber-500">{{ t('model.noticeTitle') }}</h5>
                    <p class="text-xs text-gray-500 mt-1 leading-relaxed">{{ t('model.noticeDesc') }}</p>
                  </div>
                </div>
              </div>
            </div>

          </div>
        </transition>
      </div>

      <!-- Footer / Status Bar -->
      <footer class="h-8 flex items-center px-8 border-t border-white/5 bg-black/10 text-[10px] text-gray-600 font-mono">
        <div class="flex-1 flex gap-6">
          <span class="flex items-center gap-1.5"><div class="w-1 h-1 rounded-full bg-green-500"></div> {{ t('footer.ready') }}</span>
          <span>{{ t('footer.audioInfo') }}</span>
          <span>{{ t('footer.bufferInfo') }}</span>
        </div>
        <div>Voice Writer v1.0.0</div>
      </footer>

      <!-- Global Switch Mask -->
      <transition name="fade">
        <div v-if="switchingModelId" class="fixed inset-0 bg-black/80 backdrop-blur-xl flex items-center justify-center z-[100]">
          <div class="text-center">
            <div class="relative w-16 h-16 mx-auto mb-6">
              <div class="absolute inset-0 border-4 border-blue-500/20 rounded-full"></div>
              <div class="absolute inset-0 border-4 border-blue-500 rounded-full border-t-transparent animate-spin"></div>
            </div>
            <h3 class="text-xl font-bold text-white mb-2">{{ t('model.switchingTitle') }}</h3>
            <p class="text-gray-500 text-sm">{{ t('model.switchingDesc') }}</p>
          </div>
        </div>
      </transition>
    </main>

    <!-- Toast Notification -->
    <Transition name="fade">
      <div v-if="showToast" class="fixed bottom-12 left-1/2 -translate-x-1/2 px-6 py-3 bg-blue-600 text-white rounded-2xl shadow-2xl shadow-blue-900/40 z-[200] flex items-center gap-3 border border-white/10">
        <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3"><polyline points="20 6 9 17 4 12"/></svg>
        <span class="text-sm font-bold tracking-tight">{{ toastMessage }}</span>
      </div>
    </Transition>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, markRaw, h } from "vue";
import { StartRecording, StopRecording, GetConfig, SaveConfig, GetHotwords, SaveHotwords, StartRecordingHotkey, RecordHotkey, GetAudioDeviceName, GetAvailableModels, SwitchToModel } from "../wailsjs/go/main/App";
import { EventsOn } from "../wailsjs/runtime/runtime";
import { useI18n } from "./i18n";

const { t, locale, setLocale } = useI18n();

// SVG Icons as components
const HomeIcon = markRaw({ render: () => h('svg', { xmlns: 'http://www.w3.org/2000/svg', viewBox: '0 0 24 24', fill: 'none', stroke: 'currentColor', 'stroke-width': '2', 'stroke-linecap': 'round', 'stroke-linejoin': 'round' }, [h('path', { d: 'm3 9 9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z' }), h('polyline', { points: '9 22 9 12 15 12 15 22' })]) });
const FilesIcon = markRaw({ render: () => h('svg', { xmlns: 'http://www.w3.org/2000/svg', viewBox: '0 0 24 24', fill: 'none', stroke: 'currentColor', 'stroke-width': '2', 'stroke-linecap': 'round', 'stroke-linejoin': 'round' }, [h('path', { d: 'M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4' }), h('polyline', { points: '17 8 12 3 7 8' }), h('line', { x1: '12', y1: '3', x2: '12', y2: '15' })]) });
const SettingsIcon = markRaw({ render: () => h('svg', { xmlns: 'http://www.w3.org/2000/svg', viewBox: '0 0 24 24', fill: 'none', stroke: 'currentColor', 'stroke-width': '2', 'stroke-linecap': 'round', 'stroke-linejoin': 'round' }, [h('path', { d: 'M12.22 2h-.44a2 2 0 0 0-2 2v.18a2 2 0 0 1-1 1.73l-.43.25a2 2 0 0 1-2 0l-.15-.08a2 2 0 0 0-2.73.73l-.22.38a2 2 0 0 0 .73 2.73l.15.1a2 2 0 0 1 1 1.72v.51a2 2 0 0 1-1 1.74l-.15.09a2 2 0 0 0-.73 2.73l.22.38a2 2 0 0 0 2.73.73l.15-.08a2 2 0 0 1 2 0l.43.25a2 2 0 0 1 1 1.73V20a2 2 0 0 0 2 2h.44a2 2 0 0 0 2-2v-.18a2 2 0 0 1 1-1.73l.43-.25a2 2 0 0 1 2 0l.15.08a2 2 0 0 0 2.73-.73l.22-.39a2 2 0 0 0-.73-2.73l-.15-.08a2 2 0 0 1-1-1.74v-.5a2 2 0 0 1 1-1.74l.15-.09a2 2 0 0 0 .73-2.73l-.22-.38a2 2 0 0 0-2.73-.73l-.15.08a2 2 0 0 1-2 0l-.43-.25a2 2 0 0 1-1-1.73V4a2 2 0 0 0-2-2z' }), h('circle', { cx: '12', cy: '12', r: '3' })]) });
const ModelIcon = markRaw({ render: () => h('svg', { xmlns: 'http://www.w3.org/2000/svg', viewBox: '0 0 24 24', fill: 'none', stroke: 'currentColor', 'stroke-width': '2', 'stroke-linecap': 'round', 'stroke-linejoin': 'round' }, [h('rect', { x: '2', y: '7', width: '20', height: '14', rx: '2', ry: '2' }), h('path', { d: 'M16 21V5a2 2 0 0 0-2-2h-4a2 2 0 0 0-2 2v16' })]) });

const tabs = [
  { id: 'main', label: 'Dashboard', icon: HomeIcon },
  { id: 'files', label: 'Files', icon: FilesIcon },
  { id: 'model', label: 'Engine', icon: ModelIcon },
  { id: 'settings', label: 'Settings', icon: SettingsIcon }
];

const activeTab = ref('main');

const appConfig = ref({ hotkey: 'f9', punctuation: true, hotkey_raw_code: 120, hotwords_path: '', model_id: 'paraformer' });
const localConfig = ref({ hotkey: 'f9', punctuation: true, hotkey_raw_code: 120, hotwords_path: '', model_id: 'paraformer' });
const hotwords = ref("");
const status = ref<"idle" | "recording" | "recognizing">("idle");
const result = ref("");
const isRecording = ref(false);
const showToast = ref(false);
const toastMessage = ref('');
const audioDeviceName = ref("Initializing...");
const currentModelId = ref("paraformer");
const availableModels = ref<any[]>([]);
const switchingModelId = ref("");
const processingFiles = ref<string[]>([]);
const fileDropActive = ref(false);

async function onFileDrop(e: DragEvent) {
  e.preventDefault();
  fileDropActive.value = false;
  
  const files = e.dataTransfer?.files;
  if (!files || files.length === 0) return;
  
  for (let i = 0; i < files.length; i++) {
    const file = files[i];
    // Check if it's a wav file
    if (!file.name.toLowerCase().endsWith('.wav')) {
      showToastMessage(t('files.onlyWav'));
      continue;
    }
    
    // In Wails, we need the actual path. 
    // For drag and drop in many Wails setups, you can get it from the file object
    // Note: Some browsers/Wails environments might not expose the full path directly
    // but typically for desktop apps like this it should be accessible.
    const path = (file as any).path;
    if (!path) {
      showToastMessage('Could not get file path');
      continue;
    }

    processingFiles.value.push(file.name);
    
    // Call backend
    const res = await (window as any).go.main.App.ProcessAudioFile(path);
    
    processingFiles.value = processingFiles.value.filter(f => f !== file.name);
    
    if (res === 'success') {
      showToastMessage(t('files.success'));
    } else {
      showToastMessage(t('files.error') + ': ' + res);
    }
  }
}

function onDragOver(e: DragEvent) {
  e.preventDefault();
  fileDropActive.value = true;
}

function onDragLeave() {
  fileDropActive.value = false;
}

// Key code mapping for hotkeys
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
  if (keyCode >= 112 && keyCode <= 123) return `f${keyCode - 111}`;
  return keyCodeToName[keyCode] || `key${keyCode}`;
}

function formatSize(bytes: number): string {
  if (!bytes || bytes === 0) return 'Unknown';
  if (bytes >= 1024 * 1024 * 1024) return (bytes / (1024 * 1024 * 1024)).toFixed(1) + 'GB';
  if (bytes >= 1024 * 1024) return Math.round(bytes / (1024 * 1024)) + 'MB';
  return Math.round(bytes / 1024) + 'KB';
}

function handleKeyDown(event: KeyboardEvent) {
  if (!isRecording.value) return;
  event.preventDefault();
  event.stopPropagation();
  const keyCode = event.keyCode;
  const keyName = getKeyName(keyCode);
  isRecording.value = false;
  localConfig.value.hotkey = keyName;
  localConfig.value.hotkey_raw_code = keyCode;
  RecordHotkey(keyCode, keyName);
  window.removeEventListener('keydown', handleKeyDown, true);
}

function toggleLocale() {
  setLocale(locale.value === 'zh-CN' ? 'en-US' : 'zh-CN');
}

async function loadConfig() {
  appConfig.value = await GetConfig();
  localConfig.value = { ...appConfig.value };
  if (appConfig.value.model_id) {
    currentModelId.value = appConfig.value.model_id;
  }
  audioDeviceName.value = await GetAudioDeviceName();
  await loadModels();
}

async function loadModels() {
  const models = await GetAvailableModels();
  if (models && models.length > 0) {
    availableModels.value = models;
  }
}

async function switchToModel(model: any) {
  const modelId = model.id || model.Id;
  switchingModelId.value = modelId;
  const res = await SwitchToModel(modelId);
  switchingModelId.value = "";
  if (res.startsWith("Switched")) {
    currentModelId.value = modelId;
    // 立即保存到配置中（后端已保存，但前端也要同步）
    localConfig.value.model_id = modelId;
    appConfig.value.model_id = modelId;
    // 同步保存到后端
    await SaveConfig(localConfig.value);
    showToastMessage(res);
  } else {
    showToastMessage(t('toast.switchFailed') + ': ' + res);
  }
}

async function loadHotwords() {
  hotwords.value = await GetHotwords();
}

function switchTab(tabId: string) {
  activeTab.value = tabId;
}

const statusBadgeClass = computed(() => {
  switch (status.value) {
    case "recording": return "bg-red-500/10 text-red-500 border-red-500/20";
    case "recognizing": return "bg-amber-500/10 text-amber-500 border-amber-500/20";
    default: return "bg-blue-500/10 text-blue-500 border-blue-500/20";
  }
});

async function startRecording() {
  if (status.value !== "idle") return;
  const msg = await StartRecording();
  if (msg === "Recording started") {
    status.value = "recording";
    result.value = "";
  }
}

async function stopRecording() {
  if (status.value !== "recording") return;
  status.value = "recognizing";
  await StopRecording();
}

async function recordHotkey() {
  isRecording.value = true;
  window.addEventListener('keydown', handleKeyDown, true);
  await StartRecordingHotkey();
}

function showToastMessage(msg: string) {
  toastMessage.value = msg;
  showToast.value = true;
  setTimeout(() => showToast.value = false, 3000);
}

async function saveSettings() {
  await SaveConfig(localConfig.value);
  await SaveHotwords(hotwords.value);
  appConfig.value = { ...localConfig.value };
  showToastMessage(t('toast.saved'));
}

onMounted(() => {
  loadConfig();
  loadHotwords();
  EventsOn("status-change", (s: any) => status.value = s);
  EventsOn("recognition-result", (t: string) => result.value = t);
  EventsOn("recognition-error", (e: string) => {
    showToastMessage(e);
    status.value = "idle";
  });
});
</script>

<style>
@keyframes shimmer {
  0% { transform: translateX(-100%); }
  100% { transform: translateX(100%); }
}

.custom-scrollbar::-webkit-scrollbar {
  width: 4px;
}
.custom-scrollbar::-webkit-scrollbar-track {
  background: transparent;
}
.custom-scrollbar::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.05);
  border-radius: 10px;
}
.custom-scrollbar::-webkit-scrollbar-thumb:hover {
  background: rgba(255, 255, 255, 0.1);
}
</style>
