package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"smart-metronome/metronome"
	"smart-metronome/patterns"
	"smart-metronome/ui/cli"
)

var (
	bpm       int
	beats     int
	pattern   string
	output    string
	visualize bool
	tap       bool
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "metronome",
		Short: "Умный метроном с паттернами",
		Long: `Продвинутый метроном для музыкантов с поддержкой сложных ритмических паттернов,
полиритмий и визуализацией.`,
	}

	// Команда запуска метронома
	var startCmd = &cobra.Command{
		Use:   "start",
		Short: "Запустить метроном",
		Run:   runMetronome,
	}

	startCmd.Flags().IntVarP(&bpm, "bpm", "b", 120, "Темп (удары в минуту)")
	startCmd.Flags().IntVarP(&beats, "beats", "c", 4, "Количество долей в такте")
	startCmd.Flags().StringVarP(&pattern, "pattern", "p", "basic", "Ритмический паттерн")
	startCmd.Flags().StringVarP(&output, "output", "o", "speaker", "Выход: speaker, wav, или both")
	startCmd.Flags().BoolVarP(&visualize, "visualize", "v", false, "Включить визуализацию")

	// Команда для режима тапа
	var tapCmd = &cobra.Command{
		Use:   "tap",
		Short: "Режим определения темпа по тапу",
		Run:   runTapMode,
	}

	// Команда для просмотра паттернов
	var patternsCmd = &cobra.Command{
		Use:   "patterns",
		Short: "Показать доступные паттерны",
		Run:   showPatterns,
	}

	// Команда для генерации WAV файла
	var generateCmd = &cobra.Command{
		Use:   "generate [output.wav]",
		Short: "Сгенерировать WAV файл с паттерном",
		Args:  cobra.ExactArgs(1),
		Run:   generateWAV,
	}

	generateCmd.Flags().IntVarP(&bpm, "bpm", "b", 120, "Темп (удары в минуту)")
	generateCmd.Flags().IntVarP(&beats, "beats", "c", 4, "Количество долей в такте")
	generateCmd.Flags().StringVarP(&pattern, "pattern", "p", "basic", "Ритмический паттерн")

	// Команда для запуска веб-интерфейса
	var webCmd = &cobra.Command{
		Use:   "web",
		Short: "Запустить веб-интерфейс",
		Run:   runWebInterface,
	}

	webCmd.Flags().IntVarP(&bpm, "bpm", "b", 120, "Темп по умолчанию")
	webCmd.Flags().StringVarP(&pattern, "pattern", "p", "basic", "Паттерн по умолчанию")

	rootCmd.AddCommand(startCmd, tapCmd, patternsCmd, generateCmd, webCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runMetronome(cmd *cobra.Command, args []string) {
	// Загружаем паттерн
	pat, err := patterns.LoadPattern(pattern)
	if err != nil {
		log.Fatalf("Ошибка загрузки паттерна: %v", err)
	}

	// Создаем метроном
	metro, err := metronome.NewMetronome(bpm, beats, pat)
	if err != nil {
		log.Fatalf("Ошибка создания метронома: %v", err)
	}

	fmt.Printf("🎵 Метроном запущен\n")
	fmt.Printf("   Темп: %d BPM\n", bpm)
	fmt.Printf("   Такт: %d/4\n", beats)
	fmt.Printf("   Паттерн: %s\n", pattern)
	fmt.Printf("   Нажмите Ctrl+C для остановки\n\n")

	// Запускаем CLI интерфейс если нужно
	if visualize {
		go cli.RunVisualization(metro)
	}

	// Запускаем метроном
	if output == "wav" || output == "both" {
		filename := fmt.Sprintf("metronome_%dbpm_%s.wav", bpm, pattern)
		if err := metro.GenerateWAV(filename, 60); err != nil {
			log.Printf("Ошибка генерации WAV: %v", err)
		} else {
			fmt.Printf("Файл сохранен: %s\n", filename)
		}
	}

	if output == "speaker" || output == "both" {
		if err := metro.Start(); err != nil {
			log.Fatalf("Ошибка запуска: %v", err)
		}

		// Ожидаем сигнала завершения
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		metro.Stop()
		fmt.Println("\nМетроном остановлен")
	}
}

func runTapMode(cmd *cobra.Command, args []string) {
	fmt.Println("🎵 Режим тапа")
	fmt.Println("Нажимайте пробел в ритме для определения BPM")
	fmt.Println("Нажмите Enter для выхода")

	tapTempo := cli.NewTapTempo()
	if err := tapTempo.Run(); err != nil {
		log.Fatalf("Ошибка: %v", err)
	}
}

func showPatterns(cmd *cobra.Command, args []string) {
	fmt.Println("📋 Доступные ритмические паттерны:")
	fmt.Println(string(cli.RepeatChar("=", 50)))

	allPatterns := patterns.GetAllPatterns()
	for name, desc := range allPatterns {
		fmt.Printf("• %-15s - %s\n", name, desc)
	}

	fmt.Println("\nПример использования:")
	fmt.Println("  metronome start -b 120 -p rock -v")
}

func generateWAV(cmd *cobra.Command, args []string) {
	filename := args[0]

	pat, err := patterns.LoadPattern(pattern)
	if err != nil {
		log.Fatalf("Ошибка загрузки паттерна: %v", err)
	}

	metro, err := metronome.NewMetronome(bpm, beats, pat)
	if err != nil {
		log.Fatalf("Ошибка создания метронома: %v", err)
	}

	// Генерируем 60 секунд аудио
	if err := metro.GenerateWAV(filename, 60); err != nil {
		log.Fatalf("Ошибка генерации WAV: %v", err)
	}

	fmt.Printf("✅ Файл успешно создан: %s\n", filename)
	fmt.Printf("   Длительность: 60 секунд\n")
	fmt.Printf("   Темп: %d BPM\n", bpm)
	fmt.Printf("   Паттерн: %s\n", pattern)
}

func runWebInterface(cmd *cobra.Command, args []string) {
	fmt.Printf("🌐 Веб-интерфейс запускается на http://localhost:8080\n")
	fmt.Println("Нажмите Ctrl+C для остановки")

}
