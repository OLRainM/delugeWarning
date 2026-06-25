package async

import (
	"log"
	"sync"
)

// Job 是一个异步任务（TTS 合成、推送、COS 操作等）。
type Job func()

// Queue 是进程内 worker pool + channel 队列，慢操作异步执行，避免阻塞同步上报路径。
type Queue struct {
	jobs    chan Job
	wg      sync.WaitGroup
	once    sync.Once
	closed  bool
	mu      sync.Mutex
}

func New(workers, buffer int) *Queue {
	if workers <= 0 {
		workers = 4
	}
	if buffer <= 0 {
		buffer = 1024
	}
	q := &Queue{jobs: make(chan Job, buffer)}
	for i := 0; i < workers; i++ {
		q.wg.Add(1)
		go q.worker()
	}
	return q
}

func (q *Queue) worker() {
	defer q.wg.Done()
	for job := range q.jobs {
		func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("[async] 任务 panic 已恢复: %v", r)
				}
			}()
			job()
		}()
	}
}

// Submit 提交任务。队列已满时不阻塞调用方，直接降级为同步执行，保证不丢任务。
func (q *Queue) Submit(job Job) {
	q.mu.Lock()
	closed := q.closed
	q.mu.Unlock()
	if closed {
		return
	}
	select {
	case q.jobs <- job:
	default:
		log.Printf("[async] 队列已满，降级同步执行")
		job()
	}
}

// Close 优雅关闭，等待在途任务完成。
func (q *Queue) Close() {
	q.once.Do(func() {
		q.mu.Lock()
		q.closed = true
		q.mu.Unlock()
		close(q.jobs)
		q.wg.Wait()
	})
}
