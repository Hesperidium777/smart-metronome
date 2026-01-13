package metronome

import (
	"fmt"
	"math"
	"os"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
	goaudio "github.com/go-audio/audio"
	goaudiowav "github.com/go-audio/wav"
)

// SoundGenerator управляет аудио
type SoundGenerator struct {
	sampleRate  beep.SampleRate
	initialized bool
}

var (
	soundGen   *SoundGenerator
	sampleRate = 44100
)

// Инициализация аудиосистемы
func initAudio() error {
	if soundGen == nil {
		soundGen = &SoundGenerator{
			sampleRate: beep.SampleRate(sampleRate),
		}

		// Инициализируем speaker
		err := speaker.Init(soundGen.sampleRate, soundGen.sampleRate.N(time.Second/10))
		if err != nil {
			return fmt.Errorf("ошибка инициализации аудио: %w", err)
		}

		soundGen.initialized = true
	}
	return nil
}

func GenerateAndPlaySound(soundType string, volume float64, bpm int) {
	// Инициализируем аудио если еще не инициализировано
	if soundGen == nil {
		if err := initAudio(); err != nil {
			fmt.Printf("Аудио недоступно: %v\n", err)
			return
		}
	}

	duration := time.Duration(float64(time.Minute) / float64(bpm) * 0.1)

	// Определяем частоту по типу звука
	var freq float64
	switch soundType {
	case "accent":
		freq = 880
	case "ride":
		freq = 1318.51
	case "normal":
		freq = 440
	case "ghost":
		freq = 220
		volume *= 0.3
	default:
		freq = 440
	}

	// Создаем и проигрываем звук
	streamer := createTone(freq, duration, volume)
	done := make(chan bool)
	speaker.Play(beep.Seq(streamer, beep.Callback(func() {
		done <- true
	})))
	<-done
}

// createTone создает тональный сигнал
func createTone(freq float64, duration time.Duration, volume float64) beep.Streamer {
	return beep.StreamerFunc(func(samples [][2]float64) (n int, ok bool) {
		numSamples := len(samples)
		if numSamples == 0 {
			return 0, false
		}

		// Вычисляем количество семплов для указанной длительности
		maxSamples := int(soundGen.sampleRate.D(int(duration)))
		if maxSamples <= 0 {
			return 0, false
		}

		// Ограничиваем количество семплов
		if numSamples > maxSamples {
			numSamples = maxSamples
		}

		// Генерируем синусоидальную волну
		for i := 0; i < numSamples; i++ {
			t := float64(i) / float64(soundGen.sampleRate)
			val := math.Sin(2 * math.Pi * freq * t)

			// Применяем огибающую ADSR
			envelope := adsrEnvelope(i, numSamples)
			val *= volume * envelope

			// Записываем в оба канала (стерео)
			samples[i][0] = val
			samples[i][1] = val
		}

		return numSamples, numSamples < maxSamples
	})
}

// adsrEnvelope создает ADSR огибающую
func adsrEnvelope(sample, totalSamples int) float64 {
	attack := totalSamples / 10  // 10% attack
	decay := totalSamples / 5    // 20% decay
	release := totalSamples / 10 // 10% release
	sustain := 0.7               // sustain level

	if sample < attack {
		// Attack фаза
		return float64(sample) / float64(attack)
	} else if sample < attack+decay {
		// Decay фаза
		decayPos := float64(sample-attack) / float64(decay)
		return 1.0 - (1.0-sustain)*decayPos
	} else if sample < totalSamples-release {
		// Sustain фаза
		return sustain
	} else {
		// Release фаза
		releasePos := float64(sample-(totalSamples-release)) / float64(release)
		return sustain * (1.0 - releasePos)
	}
}

// GenerateWAV создает WAV файл с метрономом
func (m *Metronome) GenerateWAV(filename string, durationSeconds int) error {
	fmt.Printf("Генерация WAV файла: %s (%d секунд)...\n", filename, durationSeconds)

	// Создаем WAV файл
	out, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("ошибка создания файла: %w", err)
	}
	defer out.Close()

	// Параметры аудио
	totalSamples := sampleRate * durationSeconds
	intervalSamples := int(float64(sampleRate) * 60.0 / float64(m.BPM))

	// Создаем буфер для аудиоданных
	buf := &goaudio.IntBuffer{
		Data: make([]int, totalSamples),
		Format: &goaudio.Format{
			SampleRate:  sampleRate,
			NumChannels: 2,
		},
		SourceBitDepth: 16,
	}

	beatCount := 0
	barCount := 1

	for i := 0; i < totalSamples; i++ {
		// Проверяем, нужно ли воспроизвести удар
		if i%intervalSamples == 0 {
			beatCount++
			if beatCount > m.BeatsPerBar {
				beatCount = 1
				barCount++
			}

			soundType, volume := m.Pattern.GetSound(beatCount, barCount)
			m.addSoundToBuffer(buf, i, intervalSamples, soundType, volume)
		}
	}

	// Кодируем в WAV (используем псевдоним goaudiowav)
	enc := goaudiowav.NewEncoder(out, sampleRate, 16, 2, 1)
	if err := enc.Write(buf); err != nil {
		return fmt.Errorf("ошибка записи WAV: %w", err)
	}

	if err := enc.Close(); err != nil {
		return fmt.Errorf("ошибка закрытия encoder: %w", err)
	}

	return nil
}

func (m *Metronome) addSoundToBuffer(buf *goaudio.IntBuffer, startIdx, durationSamples int, soundType string, volume float64) {
	var freq float64
	switch soundType {
	case "accent":
		freq = 880
	case "ride":
		freq = 1318.51
	case "normal":
		freq = 440
	case "ghost":
		freq = 220
		volume *= 0.3
	default:
		freq = 440
	}

	// Длительность звука - 10% от интервала
	soundSamples := durationSamples / 10

	for i := 0; i < soundSamples && startIdx+i < len(buf.Data); i++ {
		t := float64(i) / float64(buf.Format.SampleRate)
		sample := math.Sin(2 * math.Pi * freq * t)

		// ADSR огибающая
		envelope := m.adsrEnvelope(i, soundSamples)

		// Применяем volume и огибающую
		sample *= volume * envelope

		// Конвертируем в 16-bit и добавляем в оба канала
		val := int(sample * 32767)
		idx := startIdx + i
		if idx < len(buf.Data) {
			buf.Data[idx] = val
		}
	}
}

// adsrEnvelope создает ADSR огибающую для метронома
func (m *Metronome) adsrEnvelope(sample, totalSamples int) float64 {
	attack := totalSamples / 10  // 10% attack
	decay := totalSamples / 5    // 20% decay
	release := totalSamples / 10 // 10% release
	sustain := 0.7               // sustain level

	if sample < attack {
		// Attack фаза
		return float64(sample) / float64(attack)
	} else if sample < attack+decay {
		// Decay фаза
		decayPos := float64(sample-attack) / float64(decay)
		return 1.0 - (1.0-sustain)*decayPos
	} else if sample < totalSamples-release {
		// Sustain фаза
		return sustain
	} else {
		// Release фаза
		releasePos := float64(sample-(totalSamples-release)) / float64(release)
		return sustain * (1.0 - releasePos)
	}
}

// Простая альтернатива без сложных зависимостей
type SimpleAudio struct{}

func (a *SimpleAudio) PlayBeep(isAccent bool) {
	if isAccent {
		fmt.Print("█ ") // Сильный удар
		// Системный beep для Windows/Linux
		fmt.Print("\a")
	} else {
		fmt.Print("▓ ") // Обычный удар
	}
}

// Альтернативная упрощенная версия GenerateWAV
func (m *Metronome) GenerateSimpleWAV(filename string, durationSeconds int) error {
	fmt.Printf("Создание файла %s...\n", filename)

	// Создаем простой текстовый файл с паттерном вместо WAV
	out, err := os.Create(filename + ".txt")
	if err != nil {
		return err
	}
	defer out.Close()

	// Записываем информацию о паттерне
	content := fmt.Sprintf("Metronome Pattern\nBPM: %d\nTime Signature: %d/4\nPattern: %s\nDuration: %d seconds\n\n",
		m.BPM, m.BeatsPerBar, m.Pattern.Name, durationSeconds)

	_, err = out.WriteString(content)
	return err
}
