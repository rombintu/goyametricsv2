package patterns

import (
	"sync"
	"testing"
	"time"
)

func TestSemaphore(t *testing.T) {
	const maxReq = 3
	sem := NewSemaphore(maxReq)

	// Проверяем, что семафор не позволяет превысить максимальное количество запросов
	var wg sync.WaitGroup
	for i := 0; i < maxReq+1; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem.Acquire()
			defer sem.Release()
			// Имитируем работу с ресурсом
			time.Sleep(10 * time.Millisecond)
		}()
	}

	// Ждем завершения всех горутин
	wg.Wait()

	// Проверяем, что семафор работает корректно
	if len(sem.semaCh) != 0 {
		t.Errorf("Expected semaphore channel to be empty, got %d", len(sem.semaCh))
	}
}

func TestSemaphore_Concurrency(t *testing.T) {
	const maxReq = 5
	sem := NewSemaphore(maxReq)

	// Проверяем, что семафор ограничивает количество одновременно работающих горутин
	var wg sync.WaitGroup
	var count int
	var mu sync.Mutex

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem.Acquire()
			defer sem.Release()

			// Увеличиваем счетчик одновременно работающих горутин
			mu.Lock()
			count++
			if count > maxReq {
				t.Errorf("Exceeded maximum number of concurrent requests: %d", count)
			}
			mu.Unlock()

			// Имитируем работу с ресурсом
			time.Sleep(10 * time.Millisecond)

			// Уменьшаем счетчик одновременно работающих горутин
			mu.Lock()
			count--
			mu.Unlock()
		}()
	}

	// Ждем завершения всех горутин
	wg.Wait()

	// Проверяем, что семафор работает корректно
	if len(sem.semaCh) != 0 {
		t.Errorf("Expected semaphore channel to be empty, got %d", len(sem.semaCh))
	}
}

func TestSemaphore_EdgeCases(t *testing.T) {
	// Проверяем случай с максимальным количеством запросов равным 0
	sem := NewSemaphore(0)

	// Проверяем, что семафор не позволяет выполнять операции
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		sem.Acquire()
		t.Error("Expected Acquire to block indefinitely, but it did not")
	}()

	// Ждем некоторое время, чтобы убедиться, что горутина не завершилась
	time.Sleep(100 * time.Millisecond)

	// Проверяем, что семафор работает корректно
	if len(sem.semaCh) != 0 {
		t.Errorf("Expected semaphore channel to be empty, got %d", len(sem.semaCh))
	}

	// Проверяем случай с максимальным количеством запросов равным 1
	sem = NewSemaphore(1)

	// Проверяем, что семафор работает корректно
	sem.Acquire()
	sem.Release()

	// Проверяем, что семафор работает корректно
	if len(sem.semaCh) != 0 {
		t.Errorf("Expected semaphore channel to be empty, got %d", len(sem.semaCh))
	}
}
