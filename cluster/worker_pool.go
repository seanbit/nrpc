package cluster

import (
	"hash/fnv"
	"strconv"
	"sync"
	"time"
)

// WorkerQueue 代表玩家的任务队列
type WorkerQueue struct {
	taskChan   chan func()
	lastActive time.Time // 记录最近活跃时间
	mu         sync.Mutex
	wg         sync.WaitGroup
}

// NewWorkerQueue 创建新的 WorkerQueue
func NewWorkerQueue(bufferSize int) *WorkerQueue {
	pq := &WorkerQueue{
		taskChan:   make(chan func(), bufferSize),
		lastActive: time.Now(),
	}
	pq.start()
	return pq
}

// start 启动任务处理 goroutine
func (pq *WorkerQueue) start() {
	go func() {
		for task := range pq.taskChan {
			task()
			pq.mu.Lock()
			pq.lastActive = time.Now() // 任务执行时更新活跃时间
			pq.mu.Unlock()
			pq.wg.Done()
		}
	}()
}

// Submit 提交任务到队列
func (pq *WorkerQueue) Submit(task func()) {
	pq.mu.Lock()
	pq.lastActive = time.Now() // 更新活跃时间
	pq.mu.Unlock()
	pq.wg.Add(1)
	pq.taskChan <- task
}

// Stop 停止队列，等待任务完成
func (pq *WorkerQueue) Stop() {
	pq.wg.Wait()
	close(pq.taskChan)
}

// WorkerPool 统一管理玩家队列
type WorkerPool struct {
	queues        sync.Map // 存储玩家的 PlayerQueue
	queueNum      int      // 队列数量（0 表示每个玩家一个队列）
	queueSize     int
	quit          chan bool // 用于停止清理 goroutine
	clearInterval time.Duration
	clearTimePass time.Duration
}

// NewWorkerPool 创建 PlayerQueueManager
func NewWorkerPool(queueNum, queueSize int, clearInterval, clearTimePass time.Duration) *WorkerPool {
	manager := &WorkerPool{
		queueNum:      queueNum,
		queueSize:     queueSize,
		quit:          make(chan bool),
		clearInterval: clearInterval,
		clearTimePass: clearTimePass,
	}
	return manager
}

func (m *WorkerPool) Start() {
	if m.clearInterval > 0 && m.clearTimePass > 0 {
		go m.cleanupOldQueues() // 启动后台清理 goroutine
	}
}

// getQueue 获取或创建玩家的任务队列
func (m *WorkerPool) getQueue(uid string) *WorkerQueue {
	var key string
	if m.queueNum > 0 {
		index := m.hash(uid) % m.queueNum
		key = strconv.Itoa(index) // 保证 key 类型一致，避免 sync.Map 误判不同 key
	} else {
		key = uid
	}
	if val, ok := m.queues.Load(key); ok {
		return val.(*WorkerQueue)
	}
	queue := NewWorkerQueue(m.queueSize)
	m.queues.Store(key, queue)
	return queue
}

// Submit 提交任务到玩家的队列
func (m *WorkerPool) Submit(shareID string, task func()) {
	queue := m.getQueue(shareID)
	queue.Submit(task)
}

// Stop 停止队列管理器
func (m *WorkerPool) Stop() {
	close(m.quit)
	m.queues.Range(func(key, value interface{}) bool {
		queue := value.(*WorkerQueue)
		queue.Stop()
		return true
	})
}

// cleanupOldQueues 定期清理长时间未使用的队列
func (m *WorkerPool) cleanupOldQueues() {
	ticker := time.NewTicker(m.clearInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			now := time.Now()
			m.queues.Range(func(key, value interface{}) bool {
				queue := value.(*WorkerQueue)
				queue.mu.Lock()
				if now.Sub(queue.lastActive) > m.clearTimePass { // 超过未使用则释放
					queue.Stop()
					m.queues.Delete(key)
				}
				queue.mu.Unlock()
				return true
			})
		case <-m.quit:
			return
		}
	}
}

// 计算 FNV-1a Hash 并取模
func (m *WorkerPool) hash(uid string) int {
	h := fnv.New32a()
	h.Write([]byte(uid))
	return int(h.Sum32())
}
