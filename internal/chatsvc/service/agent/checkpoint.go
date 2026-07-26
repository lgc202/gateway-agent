package agent

import (
	"context"
	"sync"
)

// checkpointStore 暂存 Eino Runner 生成的运行状态，供 Agent 在中断后取出。
type checkpointStore struct {
	mu   sync.Mutex
	data map[string][]byte
}

// newCheckpointStore 创建当前 Agent 实例独享的 Checkpoint Store。
func newCheckpointStore() *checkpointStore {
	return &checkpointStore{data: make(map[string][]byte)}
}

// Get 读取指定运行状态。
func (s *checkpointStore) Get(_ context.Context, id string) ([]byte, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	state, exists := s.data[id]
	return state, exists, nil
}

// Set 保存指定运行状态。
func (s *checkpointStore) Set(_ context.Context, id string, state []byte) error {
	s.restore(id, state)
	return nil
}

// restore 将持久化的运行状态装载到当前 Agent 的内存 Store。
func (s *checkpointStore) restore(id string, state []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data[id] = state
}
