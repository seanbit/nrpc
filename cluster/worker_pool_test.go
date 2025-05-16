package cluster

import (
	"sync"
	"testing"
	"time"
)

func TestPlayerQueueManager(t *testing.T) {
	manager := NewWorkerPool(0, 128, time.Minute, time.Minute*10) // 每个玩家独立队列
	defer manager.Stop()

	var wg sync.WaitGroup
	wg.Add(2)

	result := []string{}

	// 玩家1提交任务
	manager.Submit("player1", func() {
		time.Sleep(100 * time.Millisecond)
		result = append(result, "p1_task1")
		wg.Done()
	})
	manager.Submit("player1", func() {
		result = append(result, "p1_task2")
		wg.Done()
	})

	wg.Wait()

	// 检查任务执行顺序
	expected := []string{"p1_task1", "p1_task2"}
	if len(result) != len(expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestPlayerQueueCleanup(t *testing.T) {
	manager := NewWorkerPool(0, 128, time.Second*5, time.Second*10)
	defer manager.Stop()

	manager.Submit("player1", func() {})
	manager.Submit("player2", func() {})

	time.Sleep(15 * time.Second) // 等待超时回收

	// 确保队列被删除
	_, exists1 := manager.queues.Load("player1")
	_, exists2 := manager.queues.Load("player2")
	if exists1 || exists2 {
		t.Errorf("expected queues to be cleaned up")
	}
}

// 测试固定数量队列模式
func TestLimitedQueueMode(t *testing.T) {
	pool := NewWorkerPool(4, 10, time.Second*5, time.Second*10)
	defer pool.Stop()

	var wg sync.WaitGroup
	playerIDs := []string{"player1", "player2", "player3", "player4", "player5"}

	results := make(map[string]int)
	var mu sync.Mutex

	// 提交任务
	for _, id := range playerIDs {
		wg.Add(1)
		pool.Submit(id, func() {
			mu.Lock()
			results[id]++
			mu.Unlock()
			wg.Done()
		})
	}

	wg.Wait()

	// 验证任务被正确分配
	for _, id := range playerIDs {
		mu.Lock()
		if results[id] != 1 {
			t.Errorf("Player %s's task did not execute correctly", id)
		}
		mu.Unlock()
	}
}

// 测试每个玩家独立队列
func TestUniqueQueueMode(t *testing.T) {
	pool := NewWorkerPool(0, 10, time.Second*5, time.Second*10)
	defer pool.Stop()

	var wg sync.WaitGroup
	playerIDs := []string{"player1", "player2", "player3"}

	results := make(map[string]int)
	var mu sync.Mutex

	// 提交任务
	for _, id := range playerIDs {
		wg.Add(1)
		pool.Submit(id, func() {
			mu.Lock()
			results[id]++
			mu.Unlock()
			wg.Done()
		})
	}

	wg.Wait()

	// 验证任务执行
	for _, id := range playerIDs {
		mu.Lock()
		if results[id] != 1 {
			t.Errorf("Player %s's task did not execute correctly", id)
		}
		mu.Unlock()
	}
}
