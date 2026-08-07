package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
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

// Has 判断某个键在不在
func (s *Store) Has(key string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.data[key]
	return ok
}

// UpdateMany 批量写入一组键值，全部成功才返回
// 中途任一写入失败则保留内存里已写的部分并返回错误
func (s *Store) UpdateMany(pairs map[string]string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for k, v := range pairs {
		s.data[k] = v
	}
	return s.save()
}

// Export 返回当前全部键值的一份拷贝，方便备份
func (s *Store) Export() map[string]string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make(map[string]string, len(s.data))
	for k, v := range s.data {
		out[k] = v
	}
	return out
}

// Import 用一份外部数据覆盖当前库（先清空再写入），用于恢复备份
func (s *Store) Import(pairs map[string]string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data = make(map[string]string, len(pairs))
	for k, v := range pairs {
		s.data[k] = v
	}
	return s.save()
}

// Rename 把旧键改成新键，新键已存在则覆盖；旧键不存在返回 false
func (s *Store) Rename(oldKey, newKey string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.data[oldKey]; !ok {
		return false, nil
	}
	v := s.data[oldKey]
	delete(s.data, oldKey)
	s.data[newKey] = v
	return true, s.save()
}

// Append 往某个键的值后面追加内容（用 sep 连接，sep 为空就直接拼）
func (s *Store) Append(key, chunk, sep string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if v, ok := s.data[key]; ok && v != "" {
		s.data[key] = v + sep + chunk
	} else {
		s.data[key] = chunk
	}
	return s.save()
}

// Clear 清空整个库
func (s *Store) Clear() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data = make(map[string]string)
	return s.save()
}

// KeysWithPrefix 返回以 prefix 开头的键（按字母序）
func (s *Store) KeysWithPrefix(prefix string) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []string
	for k := range s.data {
		if strings.HasPrefix(k, prefix) {
			out = append(out, k)
		}
	}
	sort.Strings(out)
	return out
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
	// Windows 上 os.Rename 不会覆盖已有文件
	os.Remove(s.path)
	return os.Rename(tmp, s.path)
}
