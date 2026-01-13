package patterns

import (
	"fmt"
	"smart-metronome/metronome"
)

var patternRegistry map[string]*metronome.Pattern

func init() {
	patternRegistry = metronome.PredefinedPatterns()
}

func LoadPattern(name string) (*metronome.Pattern, error) {
	pattern, exists := patternRegistry[name]
	if !exists {
		return nil, fmt.Errorf("паттерн '%s' не найден", name)
	}
	return pattern, nil
}

func GetAllPatterns() map[string]string {
	result := make(map[string]string)
	for name, pattern := range patternRegistry {
		result[name] = pattern.Description
	}
	return result
}

func RegisterPattern(name string, pattern *metronome.Pattern) error {
	if _, exists := patternRegistry[name]; exists {
		return fmt.Errorf("паттерн '%s' уже существует", name)
	}
	patternRegistry[name] = pattern
	return nil
}

func GetPatternNames() []string {
	names := make([]string, 0, len(patternRegistry))
	for name := range patternRegistry {
		names = append(names, name)
	}
	return names
}

func SaveCustomPattern(name, description string, beats int, patternDef []metronome.BeatDefinition) error {
	pattern := &metronome.Pattern{
		Name:        name,
		Description: description,
		Beats:       beats,
		Pattern:     patternDef,
	}
	return RegisterPattern(name, pattern)
}
