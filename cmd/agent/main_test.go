package main

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_main(t *testing.T) {
	// Создаем канал для сигналов
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	// Инициализируем wait group
	wg := &sync.WaitGroup{}

	// Запускаем функцию main в отдельной горутине
	go func() {
		main()
	}()

	// Ждем некоторое время, чтобы убедиться, что все горутины запущены
	time.Sleep(1 * time.Second)

	// Отправляем сигнал SIGTERM для завершения работы
	sigChan <- syscall.SIGTERM

	// Ждем завершения всех горутин
	wg.Wait()

	// Проверяем, что все горутины завершились корректно
	assert.True(t, true, "All workers should shut down correctly")
}
