package metronome

import (
	"fmt"
	"sync"
	"time"
)

type Metronome struct {
	BPM         int
	BeatsPerBar int
	Pattern     *Pattern
	Running     bool
	mu          sync.Mutex
	stopChan    chan struct{}
	ticker      *time.Ticker
	subscribers []chan TickEvent
	beatCount   int
	barCount    int
}

type TickEvent struct {
	Beat      int     // Номер доли в такте (1-based)
	Bar       int     // Номер такта
	Volume    float64 // Громкость (0.0-1.0)
	Sound     string  // Тип звука: accent, normal, ghost, etc
	Timestamp time.Time
}

func NewMetronome(bpm, beats int, pattern *Pattern) (*Metronome, error) {
	if bpm < 20 || bpm > 300 {
		return nil, fmt.Errorf("BPM должен быть от 20 до 300")
	}
	if beats < 1 || beats > 32 {
		return nil, fmt.Errorf("количество долей должно быть от 1 до 32")
	}

	return &Metronome{
		BPM:         bpm,
		BeatsPerBar: beats,
		Pattern:     pattern,
		Running:     false,
		stopChan:    make(chan struct{}),
		subscribers: make([]chan TickEvent, 0),
		beatCount:   0,
		barCount:    1,
	}, nil
}

func (m *Metronome) Start() error {
	m.mu.Lock()
	if m.Running {
		m.mu.Unlock()
		return fmt.Errorf("метроном уже запущен")
	}
	m.Running = true
	m.mu.Unlock()

	interval := time.Duration(float64(time.Minute) / float64(m.BPM))

	m.ticker = time.NewTicker(interval)
	m.beatCount = 0
	m.barCount = 1

	go func() {
		for {
			select {
			case <-m.ticker.C:
				m.handleTick()
			case <-m.stopChan:
				m.ticker.Stop()
				return
			}
		}
	}()

	return nil
}

func (m *Metronome) handleTick() {
	m.beatCount++
	if m.beatCount > m.BeatsPerBar {
		m.beatCount = 1
		m.barCount++
	}

	// Получаем настройки для этой доли из паттерна
	soundType, volume := m.Pattern.GetSound(m.beatCount, m.barCount)

	event := TickEvent{
		Beat:      m.beatCount,
		Bar:       m.barCount,
		Volume:    volume,
		Sound:     soundType,
		Timestamp: time.Now(),
	}

	// Отправляем звук
	m.playSound(event)

	// Уведомляем подписчиков
	m.notifySubscribers(event)

	// Визуальный индикатор в консоли
	m.printVisual(event)
}

func (m *Metronome) playSound(event TickEvent) {
	// Генерируем и проигрываем звук
	GenerateAndPlaySound(event.Sound, event.Volume, m.BPM)
}

func (m *Metronome) printVisual(event TickEvent) {
	var marker string
	switch event.Sound {
	case "accent":
		marker = "█"
	case "normal":
		marker = "▓"
	case "ghost":
		marker = "░"
	case "silent":
		marker = " "
	default:
		marker = "▒"
	}

	if event.Beat == 1 {
		fmt.Printf("\n[%03d] ", event.Bar)
	}

	fmt.Printf("%s ", marker)
}

func (m *Metronome) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.Running {
		return
	}

	m.Running = false
	close(m.stopChan)

	// Закрываем каналы подписчиков
	for _, ch := range m.subscribers {
		close(ch)
	}
	m.subscribers = nil
}

func (m *Metronome) Subscribe() <-chan TickEvent {
	ch := make(chan TickEvent, 100)
	m.mu.Lock()
	m.subscribers = append(m.subscribers, ch)
	m.mu.Unlock()
	return ch
}

func (m *Metronome) notifySubscribers(event TickEvent) {
	m.mu.Lock()
	subscribers := m.subscribers
	m.mu.Unlock()

	for _, ch := range subscribers {
		select {
		case ch <- event:
		default:
			// Пропускаем если канал полный
		}
	}
}

func (m *Metronome) SetBPM(bpm int) error {
	if bpm < 20 || bpm > 300 {
		return fmt.Errorf("BPM должен быть от 20 до 300")
	}

	m.mu.Lock()
	wasRunning := m.Running
	if wasRunning {
		m.Stop()
	}

	m.BPM = bpm

	if wasRunning {
		m.Start()
	}
	m.mu.Unlock()

	return nil
}

func (m *Metronome) SetPattern(pattern *Pattern) {
	m.mu.Lock()
	m.Pattern = pattern
	m.mu.Unlock()
}

func (m *Metronome) GetState() map[string]interface{} {
	m.mu.Lock()
	defer m.mu.Unlock()

	return map[string]interface{}{
		"bpm":           m.BPM,
		"beats_per_bar": m.BeatsPerBar,
		"running":       m.Running,
		"current_beat":  m.beatCount,
		"current_bar":   m.barCount,
		"pattern":       m.Pattern.Name,
	}
}

func (m *Metronome) Reset() {
	m.mu.Lock()
	m.beatCount = 0
	m.barCount = 1
	m.mu.Unlock()
}
