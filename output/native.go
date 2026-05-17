package output

import (
	"fmt"
	"syscall"
	"unsafe"
)

var (
	user32             = syscall.NewLazyDLL("user32.dll")
	procSendInput      = user32.NewProc("SendInput")
	procGetKeyNameText = user32.NewProc("GetKeyNameTextW")
	procMapVirtualKey  = user32.NewProc("MapVirtualKeyW")
)

const (
	InputKeyboard    = 1
	KeyeventfUnicode = 0x0004
	KeyeventfKeyup   = 0x0002
)

type input struct {
	Type uint32
	Ki   keybdinput
}

type keybdinput struct {
	WVk         uint16
	WScan       uint16
	DwFlags     uint32
	Time        uint32
	DwExtraInfo uintptr
}

// SendUnicode 输入字符
func SendUnicode(text string) {
	for _, r := range text {
		ki := keybdinput{
			WScan:   uint16(r),
			DwFlags: KeyeventfUnicode,
		}

		in := input{
			Type: InputKeyboard,
			Ki:   ki,
		}

		// 按下
		procSendInput.Call(1, uintptr(unsafe.Pointer(&in)), unsafe.Sizeof(in))

		// 释放
		ki.DwFlags |= KeyeventfKeyup
		in.Ki = ki
		procSendInput.Call(1, uintptr(unsafe.Pointer(&in)), unsafe.Sizeof(in))
	}
}

// GetKeyName returns the human-readable name of a key from its raw code.
func GetKeyName(rawCode uint16) string {
	// 1. 将虚拟键码转换为扫描码 (MAPVK_VK_TO_VSC = 0)
	scanCode, _, _ := procMapVirtualKey.Call(uintptr(rawCode), 0)
	
	// 2. 构造 lParam (扫描码左移 16 位，并设置扩展键位)
	lParam := uint32(scanCode) << 16
	
	buf := make([]uint16, 256)
	ret, _, _ := procGetKeyNameText.Call(uintptr(lParam), uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
	
	if ret > 0 {
		return syscall.UTF16ToString(buf[:ret])
	}
	return fmt.Sprintf("Key_%d", rawCode)
}
