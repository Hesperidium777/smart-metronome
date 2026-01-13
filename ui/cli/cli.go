package cli

import (
	"fmt"
	"math"
	"strings"
	"time"

	"smart-metronome/metronome"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func RepeatChar(char string, count int) string {
	return strings.Repeat(char, count)
}

func RunVisualization(metro *metronome.Metronome) {
	app := tview.NewApplication()

	// Создаем интерфейс
	beatDisplay := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetChangedFunc(func() {
			app.Draw()
		})

	infoDisplay := tview.NewTextView().
		SetTextAlign(tview.AlignCenter).
		SetDynamicColors(true)

	grid := tview.NewGrid().
		SetRows(3, 0, 3).
		SetColumns(0).
		SetBorders(true)

	grid.AddItem(infoDisplay, 0, 0, 1, 1, 0, 0, false)
	grid.AddItem(beatDisplay, 1, 0, 1, 1, 0, 0, false)

	// Обновляем информацию
	updateInfo := func() {
		state := metro.GetState()
		var status string
		if running, ok := state["running"].(bool); ok && running {
			status = "[green]▶ Воспроизведение[white]"
		} else {
			status = "[red]⏸ Остановлено[white]"
		}

		info := fmt.Sprintf("[yellow]BPM: %v | Такт: %v/4 | Паттерн: %v | %s",
			state["bpm"], state["beats_per_bar"], state["pattern"], status)
		infoDisplay.SetText(info)
	}

	// Подписываемся на события метронома
	events := metro.Subscribe()
	go func() {
		for event := range events {
			app.QueueUpdateDraw(func() {
				// Отображаем текущую долю
				beatText := ""
				beatsPerBar := metro.BeatsPerBar
				for i := 1; i <= beatsPerBar; i++ {
					if i == event.Beat {
						// Текущая доля - выделяем
						beatText += fmt.Sprintf(`["%d"][red]●[""] `, i)
					} else {
						beatText += fmt.Sprintf("[gray]○ ")
					}
				}
				beatDisplay.SetText(beatText)

				// Обновляем информацию
				updateInfo()
			})
		}
	}()

	// Обработка клавиш
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEscape:
			app.Stop()
			metro.Stop()
			return nil
		case tcell.KeyCtrlC:
			app.Stop()
			metro.Stop()
			return nil
		case tcell.KeyPause:
			if metro.Running {
				metro.Stop()
			} else {
				metro.Start()
			}
			return nil
		default:
			// Обработка символов
			switch event.Rune() {
			case '+', '=':
				bpm := metro.BPM + 5
				if bpm <= 300 {
					metro.SetBPM(bpm)
				}
				return nil
			case '-', '_':
				bpm := metro.BPM - 5
				if bpm >= 20 {
					metro.SetBPM(bpm)
				}
				return nil
			case 'r', 'R':
				metro.Reset()
				return nil
			case 'q', 'Q':
				app.Stop()
				metro.Stop()
				return nil
			}
		}
		return event
	})

	// Запускаем интерфейс
	if err := app.SetRoot(grid, true).SetFocus(grid).Run(); err != nil {
		panic(err)
	}
}

type TapTempo struct {
	taps    []time.Time
	lastTap time.Time
	minTaps int
	maxTaps int
	timeout time.Duration
}

func NewTapTempo() *TapTempo {
	return &TapTempo{
		taps:    make([]time.Time, 0),
		minTaps: 2,
		maxTaps: 8,
		timeout: 2 * time.Second,
	}
}

func (t *TapTempo) Run() error {
	fmt.Println("\nНажимайте пробел в ритме...")

	// Используем tview для чтения клавиш без Enter
	app := tview.NewApplication()
	textView := tview.NewTextView().
		SetTextAlign(tview.AlignCenter).
		SetDynamicColors(true)

	go func() {
		for {
			time.Sleep(100 * time.Millisecond)
			app.QueueUpdateDraw(func() {
				t.updateDisplay(textView)
			})
		}
	}()

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEscape, tcell.KeyEnter:
			app.Stop()
			return nil
		case tcell.KeyPause:
			t.registerTap()
			return nil
		default:
			switch event.Rune() {
			case 'c', 'C':
				t.clear()
				return nil
			case 'q', 'Q':
				app.Stop()
				return nil
			}
		}
		return event
	})

	textView.SetText(t.getInitialText())
	return app.SetRoot(textView, true).Run()
}

func (t *TapTempo) registerTap() {
	now := time.Now()

	// Удаляем старые тапы
	t.cleanOldTaps(now)

	// Добавляем новый тап
	t.taps = append(t.taps, now)
	t.lastTap = now

	// Ограничиваем количество тапов
	if len(t.taps) > t.maxTaps {
		t.taps = t.taps[len(t.taps)-t.maxTaps:]
	}
}

func (t *TapTempo) cleanOldTaps(now time.Time) {
	cutoff := now.Add(-t.timeout)
	validTaps := make([]time.Time, 0)
	for _, tap := range t.taps {
		if tap.After(cutoff) {
			validTaps = append(validTaps, tap)
		}
	}
	t.taps = validTaps
}

func (t *TapTempo) calculateBPM() (int, float64) {
	if len(t.taps) < t.minTaps {
		return 0, 0
	}

	// Вычисляем средний интервал между тапами
	var totalInterval time.Duration
	for i := 1; i < len(t.taps); i++ {
		totalInterval += t.taps[i].Sub(t.taps[i-1])
	}

	avgInterval := totalInterval / time.Duration(len(t.taps)-1)
	bpm := int(time.Minute / avgInterval)

	// Вычисляем стабильность (коэффициент вариации)
	var sumSqDiff float64
	for i := 1; i < len(t.taps); i++ {
		diff := float64(t.taps[i].Sub(t.taps[i-1]) - avgInterval)
		sumSqDiff += diff * diff
	}

	stdDev := time.Duration(math.Sqrt(sumSqDiff / float64(len(t.taps)-1)))
	stability := 100 * (1 - float64(stdDev)/float64(avgInterval))

	return bpm, stability
}

func (t *TapTempo) updateDisplay(textView *tview.TextView) {
	bpm, stability := t.calculateBPM()

	var status string
	if len(t.taps) == 0 {
		status = "[yellow]Ожидание тапов...[-]\nНажмите [green]ПРОБЕЛ[-] в ритме"
	} else if len(t.taps) < t.minTaps {
		status = fmt.Sprintf("[yellow]Тапов: %d/%d[-]\nПродолжайте...", len(t.taps), t.minTaps)
	} else {
		stabilityColor := "green"
		if stability < 80 {
			stabilityColor = "yellow"
		}
		if stability < 60 {
			stabilityColor = "red"
		}

		status = fmt.Sprintf("[white]BPM: [green]%d[-]\n", bpm) +
			fmt.Sprintf("Стабильность: [%s]%.1f%%[-]\n", stabilityColor, stability) +
			fmt.Sprintf("Тапов: %d\n", len(t.taps)) +
			"[gray]ESC/Enter - выход, C - очистить[-]"
	}

	// Добавляем визуализацию ритма
	if len(t.taps) > 1 {
		status += "\n\n"
		maxBars := 20
		for i := 0; i < maxBars; i++ {
			if i < len(t.taps) {
				status += "[green]█[-]"
			} else {
				status += "[gray]░[-]"
			}
		}
	}

	textView.SetText(status)
}

func (t *TapTempo) clear() {
	t.taps = make([]time.Time, 0)
}

func (t *TapTempo) getInitialText() string {
	return `[yellow]═══════════════════════════════════
          ТАП-ТЕМПО
═══════════════════════════════════[-]

Нажимайте [green]ПРОБЕЛ[-] в ритме музыки
Система определит BPM автоматически

[gray]ESC/Enter - выход
C - очистить тапы[-]`
}

// SimpleVisualization - простая визуализация в консоли
func SimpleVisualization(metro *metronome.Metronome) {
	fmt.Println("🎵 Простая визуализация метронома")
	fmt.Println("Нажмите Ctrl+C для выхода")

	events := metro.Subscribe()

	for event := range events {
		var symbol string
		switch event.Sound {
		case "accent":
			symbol = "█"
		case "normal":
			symbol = "▓"
		case "ghost":
			symbol = "░"
		case "ride":
			symbol = "◉"
		default:
			symbol = "▒"
		}

		if event.Beat == 1 {
			fmt.Printf("\n[%03d] ", event.Bar)
		}
		fmt.Printf("%s ", symbol)
	}
}
