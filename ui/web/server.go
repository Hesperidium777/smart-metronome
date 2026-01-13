package metronome

import (
	"fmt"
	"os/exec"
	"runtime"
)

// SoundPlayer - простой проигрыватель звуков
type SoundPlayer struct {
	useBeep bool
}

func NewSoundPlayer() *SoundPlayer {
	return &SoundPlayer{
		useBeep: checkBeepAvailable(),
	}
}

func checkBeepAvailable() bool {
	// Проверяем, доступна ли утилита beep в системе
	if runtime.GOOS == "windows" {
		return false
	}
	cmd := exec.Command("which", "beep")
	return cmd.Run() == nil
}

func GenerateAndPlaySound(soundType string, volume float64, bpm int) {
	// Простая реализация - выводим в консоль и/или используем системные звуки
	switch soundType {
	case "accent":
		fmt.Print("█ ")
		playSystemBeep(880, volume)
	case "normal":
		fmt.Print("▓ ")
		playSystemBeep(440, volume)
	case "ghost":
		fmt.Print("░ ")
		playSystemBeep(220, volume*0.3)
	case "ride":
		fmt.Print("◉ ")
		playSystemBeep(1318, volume*0.7)
	default:
		fmt.Print("▒ ")
		playSystemBeep(440, volume)
	}
}

func playSystemBeep(freq float64, volume float64) {
	// Простой способ для Windows
	if runtime.GOOS == "windows" {
		// Для Windows можно использовать PowerShell
		cmd := exec.Command("powershell", "[console]::beep(800, 100)")
		cmd.Run()
		return
	}

	// Для Linux/Mac используем системную утилиту beep если доступна
	cmd := exec.Command("beep", "-f", fmt.Sprintf("%.0f", freq), "-l", "100")
	cmd.Run()
}

// PlaySimpleBeep - простая реализация бипа для всех платформ
func PlaySimpleBeep() {
	// Просто выводим символ и воспроизводим через команду
	if runtime.GOOS == "windows" {
		// Windows
		exec.Command("cmd", "/c", "echo", "\a").Run()
	} else {
		// Linux/Mac
		fmt.Print("\a")
	}
}
