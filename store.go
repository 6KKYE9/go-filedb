package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"sync"
)

// Store 是个最简单的键值库，数据落到一个 JSON 文件
// 用读写锁挡住并发读写，避免同时写文件把内容写花
type Store struct {
	mu   sync.RWMutex
	data map[string]string
	path string
}

// NewStore 打开（或新建）一个库文件
func NewStore(path string) (*Store, error) {
	s := &Store{
		data: make(map[string]string),
		path: path,
	}
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return s, nil // 文件没有就当空库
		}
		return nil, err
	}
	if len(b) > 0 {
		if err := json.Unmarshal(b, &s.data); err != nil {
			return nil, fmt.Errorf("文件内容不是合法 JSON: %w", err)
		}
	}
	return s, nil
}

func (s *Store) Set(key, val string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = val
	return s.save()
}

func (s *Store) Get(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.data[key]
	return v, ok
}

func (s *Store) Del(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.data[key]; !ok {
		return false
	}
	delete(s.data, key)
	s.save()
	return true
}

func (s *Store) Keys() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	keys := make([]string, 0, len(s.data))
	for k := range s.data {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// save 必须在持锁状态下调用
func (s *Store) save() error {
	b, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return err
	}
	// 先写临时文件再改名，避免写到一半进程挂了把库弄坏
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, b, 0644); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}
