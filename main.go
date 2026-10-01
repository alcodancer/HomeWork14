// main.go (обновленный)
package main

import (
	"chess/internal/model/player"
	"chess/internal/repository"
	"chess/internal/service"
	"chess/internal/logger"
	"context"
	"fmt"
	"bufio"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func main() {
	// Создаем контекст с отменой для graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Создаем канал для сигналов ОС
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Создаем репозиторий с указанием директории для данных
	repo := repository.NewRepository("data")

	// Создаем сервис
	gameService := service.NewGameService(repo)

	// Создаем логгер
	logger := logger.NewLogger(repo)
	logger.Start()
	defer logger.Stop()

	// Выводим информацию о загруженных данных
	players := repo.GetPlayers()
	games := repo.GetGames()
	moves := repo.GetMoves()
	boards := repo.GetBoards()
	
	if len(players) > 0 || len(games) > 0 || len(moves) > 0 || len(boards) > 0 {
		fmt.Println("\n=== ВОССТАНОВЛЕННЫЕ ДАННЫЕ ===")
		fmt.Printf("Игроков: %d\n", len(players))
		fmt.Printf("Игр: %d\n", len(games))
		fmt.Printf("Ходов: %d\n", len(moves))
		fmt.Printf("Досок: %d\n", len(boards))
		fmt.Println()
	}

	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("=== ШАХМАТЫ ===")
	fmt.Println("Выберите режим:")
	fmt.Println("1. Одна доска (ручной режим)")
	fmt.Println("2. Много досок (авто-симуляция)")
	fmt.Print("Ваш выбор: ")

	scanner.Scan()
	mode := strings.TrimSpace(scanner.Text())

	// Запускаем горутину для обработки сигналов
	go func() {
		select {
		case sig := <-sigChan:
			fmt.Printf("\n[Main] Получен сигнал: %v\n", sig)
			fmt.Println("[Main] Сохранение данных перед завершением...")
			
			// Сохраняем все данные
			if err := repo.SaveAll(); err != nil {
				fmt.Printf("Ошибка сохранения данных: %v\n", err)
			}
			
			// Останавливаем сервис
			gameService.Stop()
			
			// Отменяем контекст
			cancel()
			
			// Даем время на завершение
			time.Sleep(1 * time.Second)
			
			fmt.Println("[Main] Приложение завершено")
			os.Exit(0)
		case <-ctx.Done():
			return
		}
	}()

	switch mode {
	case "1":
		runSingleBoard(gameService, repo, scanner, ctx)
	case "2":
		runMultiBoard(gameService, repo, ctx)
	default:
		fmt.Println("Неверный выбор")
	}

	// Сохраняем данные перед завершением
	fmt.Println("\n[Main] Сохранение данных...")
	if err := repo.SaveAll(); err != nil {
		fmt.Printf("Ошибка сохранения данных: %v\n", err)
	}

	// Ожидаем завершения
	<-ctx.Done()
	fmt.Println("[Main] Завершение работы")
}

func runSingleBoard(gameService *service.GameService, repo repository.Repository, scanner *bufio.Scanner, ctx context.Context) {
	player1 := askPlayerName("первого", 1)
	player2 := askPlayerName("второго", 2)
	boardSize := askBoardSize()

	gameID := gameService.InitGame(boardSize, player1, player2)
	state := gameService.GetGameState(gameID)
	if state == nil {
		fmt.Println("Ошибка создания игры")
		return
	}

	displayBoard(state)

	for !state.Game.IsFinished() {
		select {
		case <-ctx.Done():
			fmt.Println("\n[Main] Получен сигнал завершения в режиме одной доски")
			return
		default:
			fmt.Printf("\nХод игрока %d: %s\n", state.CurrentPlayer+1,
				[]*player.Player{state.Player1, state.Player2}[state.CurrentPlayer].GetName())
			fmt.Print("Введите ход (например: A1 B2) или команду: ")

			if !scanner.Scan() {
				return
			}

			input := strings.TrimSpace(scanner.Text())
			if input == "" {
				continue
			}

			parts := strings.Fields(input)
			if len(parts) == 0 {
				continue
			}

			switch strings.ToLower(parts[0]) {
			case "сдался":
				state.mu.Lock()
				state.Game.SetFinished(true)
				state.Game.SetWinner(2 - state.CurrentPlayer)
				state.mu.Unlock()
				displayBoard(state)
				fmt.Printf("Игрок %d сдался!\n", state.CurrentPlayer+1)
				// Сохраняем состояние после завершения игры
				repo.SaveAll()

			case "выход":
				fmt.Println("Выход из игры...")
				gameService.Stop()
				repo.SaveAll()
				return

			default:
				if len(parts) != 2 {
					fmt.Println("Неверный формат. Используйте: A1 B2")
					continue
				}

				from := parts[0]
				to := parts[1]

				fromRow, fromCol, err1 := parseCell(from)
				toRow, toCol, err2 := parseCell(to)

				if err1 != nil || err2 != nil {
					fmt.Println("Неверный формат клетки. Используйте: A1 B2")
					continue
				}

				fromRow--
				fromCol--
				toRow--
				toCol--

				success := gameService.MakeMove(gameID, fromRow, fromCol, toRow, toCol)
				if success {
					displayBoard(state)
					// Сохраняем состояние после каждого хода
					repo.SaveAll()
				} else {
					fmt.Println("Ход не удался. Попробуйте снова.")
				}
			}
		}
	}

	fmt.Println("\n=== ИГРА ЗАВЕРШЕНА ===")
	if winner := state.Game.GetWinner(); winner > 0 {
		fmt.Printf("Победил игрок %d: %s\n", winner,
			[]*player.Player{state.Player1, state.Player2}[winner-1].GetName())
	} else {
		fmt.Println("Ничья!")
	}
	fmt.Printf("Всего ходов: %d\n", state.Game.GetCurrentTurn())
}

func runMultiBoard(gameService *service.GameService, repo repository.Repository, ctx context.Context) {
	fmt.Print("Сколько досок создать? ")
	var numBoards int
	fmt.Scan(&numBoards)

	if numBoards < 1 {
		numBoards = 1
	}
	if numBoards > 10 {
		numBoards = 10
	}

	fmt.Print("Какой размер доски? ")
	boardSize := askBoardSize()

	fmt.Printf("\nЗапуск симуляции на %d досках размером %dx%d...\n", numBoards, boardSize, boardSize)
	fmt.Println("Нажмите Enter для выхода или Ctrl+C для завершения...")

	gameService.StartMultiBoardSimulation(numBoards, boardSize)

	// Ждем нажатия Enter или сигнала завершения
	done := make(chan bool)
	go func() {
		bufio.NewScanner(os.Stdin).Scan()
		done <- true
	}()

	select {
	case <-done:
		fmt.Println("\n[Main] Остановка симуляции по запросу пользователя")
		gameService.Stop()
		repo.SaveAll()
	case <-ctx.Done():
		fmt.Println("\n[Main] Остановка симуляции по сигналу ОС")
		gameService.Stop()
		repo.SaveAll()
	}
}

func askPlayerName(order string, id int) *player.Player {
	var name string
	fmt.Printf("Как зовут %s игрока? ", order)
	fmt.Scan(&name)
	return player.NewPlayer(name, id)
}

func askBoardSize() int {
	var size int
	fmt.Print("Какое будет размер игрового поля? ")
	fmt.Scan(&size)
	if size < 4 {
		size = 4
	}
	if size > 20 {
		size = 20
	}
	return size
}

func displayBoard(state *service.GameState) {
	fmt.Print("\033[H\033[2J")
	fmt.Println(state.Board.String())
}

func parseCell(cell string) (int, int, error) {
	if len(cell) < 2 {
		return 0, 0, fmt.Errorf("неверный формат")
	}

	col := int(cell[0] - 'A')
	if col < 0 || col > 25 {
		return 0, 0, fmt.Errorf("неверная колонка")
	}

	row, err := strconv.Atoi(cell[1:])
	if err != nil {
		return 0, 0, err
	}

	return row, col, nil
}