# Gemini Project Context: voice-writer

`voice-writer` is a desktop-based voice input method application designed for Windows. It aims to provide a single-process, zero-dependency, and offline speech-to-text experience.

## Project Overview

-   **Type:** Wails Application (Go Backend + Vue 3 Frontend)
-   **Core Mission:** Offline voice-to-text with keyboard emulation.
-   **Key Technologies:**
    -   **Backend:** Go 1.22+
    -   **Framework:** [Wails v2](https://wails.io/)
    -   **Audio Capture:** `malgo` (miniaudio binding)
    -   **ASR Engine:** `sherpa-onnx` (using Paraformer ONNX models)
    -   **System Interaction:** `robotgo` (for global hotkeys and keyboard simulation)
    -   **Frontend:** Vue 3, TypeScript, Tailwind CSS, Vite

## Architecture (Planned)

The application follows a "push-to-talk" model:
1.  **Capture:** User holds a hotkey -> `malgo` captures PCM data into a buffer.
2.  **Process:** User releases hotkey -> PCM data is resampled to 16kHz.
3.  **Recognize:** `sherpa-onnx` performs offline inference.
4.  **Output:** `robotgo` simulates keyboard events to type the text into the active window.

## Building and Running

### Development
To run in development mode with hot reloading:
```bash
wails dev
```

### Build
To build a production executable:
```bash
wails build
```

### Frontend Setup
If you need to work directly on the frontend:
```bash
cd frontend
npm install
npm run dev
```

## Directory Structure

-   `main.go`: Application entry point and Wails configuration.
-   `app.go`: Main application struct, lifecycle hooks (`startup`, `shutdown`), and methods bound to the frontend.
-   `audio/`: (Planned) Audio capture logic using `malgo`.
-   `asr/`: (Planned) ASR inference logic using `sherpa-onnx`.
-   `output/`: (Planned) Keyboard emulation and hotkey management using `robotgo`.
-   `frontend/`: The Vue 3 frontend project.
    -   `src/App.vue`: Main UI component.
    -   `src/stores/`: State management (Pinia).
-   `models/`: Directory for storing ONNX model files (ignored by git).
-   `build/`: Build assets (icons, Windows manifest).
-   `DEVELOPMENT_PLAN.md`: Detailed project roadmap and design decisions.

## Development Conventions

-   **State Management:** Use Pinia in the frontend for global state.
-   **Communication:** Use Wails Events for Go-to-Frontend notifications (e.g., status changes, recognition results).
-   **Concurrency:** Audio capture runs in C threads via `malgo`; use mutexes or channels to safely transfer data to Go.
-   **Error Handling:** Prioritize robust error handling for hardware interactions (microphone, keyboard).

## Current Status (as of 2026-05-13)

-   [x] Phase 1: Project Scaffold (Wails + Vue + Tailwind)
-   [x] Phase 2: Audio Capture (malgo)
-   [x] Phase 3: ASR Integration (sherpa-onnx)
-   [x] Phase 4: Keyboard Output (robotgo)
-   [x] Phase 5: UI Integration & Tray
