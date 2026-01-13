package metronome

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Pattern struct {
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Beats       int              `json:"beats"`
	Pattern     []BeatDefinition `json:"pattern"`
	Cycle       int              `json:"cycle"` // Цикл повторения (в тактах)
}

type BeatDefinition struct {
	Beat    int     `json:"beat"`    // Номер доли в такте
	Sound   string  `json:"sound"`   // Тип звука
	Volume  float64 `json:"volume"`  // Громкость (0.0-1.0)
	Subdiv  int     `json:"subdiv"`  // Подразделения (триоли и т.д.)
	Accent  bool    `json:"accent"`  // Акцент
	Comment string  `json:"comment"` // Комментарий для музыканта
}

func (p *Pattern) GetSound(beat, bar int) (string, float64) {
	// Если паттерн циклический, вычисляем позицию в цикле
	if p.Cycle > 0 {
		bar = ((bar - 1) % p.Cycle) + 1
	}

	// Ищем определение для этой доли
	for _, def := range p.Pattern {
		if def.Beat == beat {
			// Проверяем, применяется ли этот паттерн к текущему такту
			// (простейшая реализация)
			return def.Sound, def.Volume
		}
	}

	// По умолчанию - обычный удар
	return "normal", 0.7
}

// LoadPatternFromFile загружает паттерн из JSON файла
func LoadPatternFromFile(filename string) (*Pattern, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения файла: %w", err)
	}

	var pattern Pattern
	if err := json.Unmarshal(data, &pattern); err != nil {
		return nil, fmt.Errorf("ошибка парсинга JSON: %w", err)
	}

	return &pattern, nil
}

// SavePatternToFile сохраняет паттерн в JSON файл
func (p *Pattern) SavePatternToFile(filename string) error {
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return fmt.Errorf("ошибка сериализации: %w", err)
	}

	// Создаем директорию если нужно
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("ошибка создания директории: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("ошибка записи файла: %w", err)
	}

	return nil
}

// PredefinedPatterns возвращает встроенные паттерны
func PredefinedPatterns() map[string]*Pattern {
	return map[string]*Pattern{
		"basic": {
			Name:        "basic",
			Description: "Базовый паттерн 4/4",
			Beats:       4,
			Pattern: []BeatDefinition{
				{Beat: 1, Sound: "accent", Volume: 1.0, Accent: true, Comment: "Сильная доля"},
				{Beat: 2, Sound: "normal", Volume: 0.7, Accent: false},
				{Beat: 3, Sound: "normal", Volume: 0.7, Accent: false},
				{Beat: 4, Sound: "normal", Volume: 0.7, Accent: false},
			},
		},
		"rock": {
			Name:        "rock",
			Description: "Рок-ритм с акцентами на малом барабане",
			Beats:       4,
			Pattern: []BeatDefinition{
				{Beat: 1, Sound: "accent", Volume: 1.0, Accent: true, Comment: "Большой барабан"},
				{Beat: 2, Sound: "normal", Volume: 0.8, Accent: true, Comment: "Малый барабан"},
				{Beat: 3, Sound: "accent", Volume: 1.0, Accent: true, Comment: "Большой барабан"},
				{Beat: 4, Sound: "normal", Volume: 0.8, Accent: true, Comment: "Малый барабан"},
			},
		},
		"jazz": {
			Name:        "jazz",
			Description: "Джазовый паттерн ride-тарелки",
			Beats:       4,
			Pattern: []BeatDefinition{
				{Beat: 1, Sound: "ride", Volume: 0.7, Comment: "Ride bell"},
				{Beat: 2, Sound: "ride", Volume: 0.5, Comment: "Ride bow"},
				{Beat: 3, Sound: "ride", Volume: 0.7, Comment: "Ride bell"},
				{Beat: 4, Sound: "ride", Volume: 0.5, Comment: "Ride bow"},
			},
		},
		"waltz": {
			Name:        "waltz",
			Description: "Вальс 3/4",
			Beats:       3,
			Pattern: []BeatDefinition{
				{Beat: 1, Sound: "accent", Volume: 1.0, Accent: true},
				{Beat: 2, Sound: "normal", Volume: 0.6},
				{Beat: 3, Sound: "normal", Volume: 0.6},
			},
		},
		"shuffle": {
			Name:        "shuffle",
			Description: "Шаффл-ритм с триольным ощущением",
			Beats:       4,
			Cycle:       2, // Двухтактный паттерн
			Pattern: []BeatDefinition{
				// Первый такт
				{Beat: 1, Sound: "accent", Volume: 1.0, Comment: "Downbeat"},
				{Beat: 2, Sound: "ghost", Volume: 0.3, Comment: "Ghost note"},
				{Beat: 3, Sound: "normal", Volume: 0.7, Comment: "Backbeat"},
				{Beat: 4, Sound: "ghost", Volume: 0.3},
				// Второй такт
				{Beat: 5, Sound: "accent", Volume: 0.9},
				{Beat: 6, Sound: "ghost", Volume: 0.3},
				{Beat: 7, Sound: "normal", Volume: 0.8},
				{Beat: 8, Sound: "ghost", Volume: 0.3},
			},
		},
		"5-4": {
			Name:        "5-4",
			Description: "Сложный размер 5/4",
			Beats:       5,
			Pattern: []BeatDefinition{
				{Beat: 1, Sound: "accent", Volume: 1.0},
				{Beat: 2, Sound: "normal", Volume: 0.6},
				{Beat: 3, Sound: "accent", Volume: 0.9},
				{Beat: 4, Sound: "normal", Volume: 0.6},
				{Beat: 5, Sound: "normal", Volume: 0.6},
			},
		},
		"7-8": {
			Name:        "7-8",
			Description: "Сложный размер 7/8 (3+2+2)",
			Beats:       7,
			Pattern: []BeatDefinition{
				{Beat: 1, Sound: "accent", Volume: 1.0},
				{Beat: 2, Sound: "normal", Volume: 0.6},
				{Beat: 3, Sound: "normal", Volume: 0.6},
				{Beat: 4, Sound: "accent", Volume: 0.8},
				{Beat: 5, Sound: "normal", Volume: 0.6},
				{Beat: 6, Sound: "accent", Volume: 0.8},
				{Beat: 7, Sound: "normal", Volume: 0.6},
			},
		},
		"poly": {
			Name:        "poly",
			Description: "Полиритмия 3:4",
			Beats:       12, // НОК(3, 4)
			Pattern: []BeatDefinition{
				// Ритм 3 поверх 4
				{Beat: 1, Sound: "accent", Volume: 1.0, Comment: "3/4 - доля 1"},
				{Beat: 4, Sound: "accent", Volume: 0.8, Comment: "3/4 - доля 2"},
				{Beat: 7, Sound: "accent", Volume: 0.8, Comment: "3/4 - доля 3"},
				{Beat: 10, Sound: "accent", Volume: 0.8, Comment: "3/4 - доля 1 следующего цикла"},
				// Ритм 4 поверх 3
				{Beat: 1, Sound: "ride", Volume: 0.6, Comment: "4/4 - доля 1"},
				{Beat: 4, Sound: "ride", Volume: 0.5, Comment: "4/4 - доля 2"},
				{Beat: 7, Sound: "ride", Volume: 0.5, Comment: "4/4 - доля 3"},
				{Beat: 10, Sound: "ride", Volume: 0.5, Comment: "4/4 - доля 4"},
			},
		},
	}
}
