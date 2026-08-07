package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStoreSetGet(t *testing.T) {
	path := filepath.Join(t.TempDir(), "db.json")
	st, err := NewStore(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := st.Set("name", "张三"); err != nil {
		t.Fatal(err)
	}
	v, ok := st.Get("name")
	if !ok || v != "张三" {
		t.Fatalf("get 不符: %q %v", v, ok)
	}
}

func TestStorePersist(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "db.json")
	st, _ := NewStore(path)
	st.Set("a", "1")
	st.Set("b", "2")

	// 重新打开同一个文件，数据应还在
	st2, err := NewStore(path)
	if err != nil {
		t.Fatal(err)
	}
	if v, _ := st2.Get("a"); v != "1" {
		t.Fatalf("持久化失败，a=%q", v)
	}
	if v, _ := st2.Get("b"); v != "2" {
		t.Fatalf("持久化失败，b=%q", v)
	}
}

func TestStoreDel(t *testing.T) {
	path := filepath.Join(t.TempDir(), "db.json")
	st, _ := NewStore(path)
	st.Set("x", "9")
	if !st.Del("x") {
		t.Fatal("删除应返回 true")
	}
	if _, ok := st.Get("x"); ok {
		t.Fatal("删完还能 get 到")
	}
	if st.Del("x") {
		t.Fatal("删不存在的应返回 false")
	}
}

func TestStoreKeysSorted(t *testing.T) {
	path := filepath.Join(t.TempDir(), "db.json")
	st, _ := NewStore(path)
	st.Set("c", "1")
	st.Set("a", "1")
	st.Set("b", "1")
	keys := st.Keys()
	want := []string{"a", "b", "c"}
	for i := range want {
		if keys[i] != want[i] {
			t.Fatalf("keys 未按序: %#v", keys)
		}
	}
}

func TestStoreBadFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.json")
	// 写一个非法 JSON，打开应报错而不是静默当空库
	os.WriteFile(path, []byte("{不是json"), 0644)
	if _, err := NewStore(path); err == nil {
		t.Fatal("非法 JSON 文件应报错")
	}
}

func TestStoreHas(t *testing.T) {
	path := filepath.Join(t.TempDir(), "db.json")
	st, _ := NewStore(path)
	st.Set("x", "1")
	if !st.Has("x") {
		t.Fatal("Has 应返回 true")
	}
	if st.Has("y") {
		t.Fatal("Has 对不存在的应返回 false")
	}
}

func TestStoreUpdateMany(t *testing.T) {
	path := filepath.Join(t.TempDir(), "db.json")
	st, _ := NewStore(path)
	if err := st.UpdateMany(map[string]string{"a": "1", "b": "2"}); err != nil {
		t.Fatal(err)
	}
	// 重新打开确认持久化
	st2, _ := NewStore(path)
	if v, _ := st2.Get("a"); v != "1" {
		t.Fatalf("批量写入未持久化: %q", v)
	}
}

func TestStoreExportImport(t *testing.T) {
	path := filepath.Join(t.TempDir(), "db.json")
	st, _ := NewStore(path)
	st.Set("a", "1")
	st.Set("b", "2")
	backup := st.Export()
	if len(backup) != 2 {
		t.Fatalf("导出数量不对: %#v", backup)
	}
	// 清空再导入回去
	st.Del("a")
	st.Del("b")
	if err := st.Import(backup); err != nil {
		t.Fatal(err)
	}
	if v, _ := requireGet(st, "a"); v != "1" {
		t.Fatalf("导入后 a 不符: %q", v)
	}
}

func TestStoreRename(t *testing.T) {
	path := filepath.Join(t.TempDir(), "db.json")
	st, _ := NewStore(path)
	st.Set("old", "v")
	ok, err := st.Rename("old", "new")
	if err != nil || !ok {
		t.Fatalf("rename 应成功: ok=%v err=%v", ok, err)
	}
	if v, ok := st.Get("new"); !ok || v != "v" {
		t.Fatalf("改名后 new 应是 v，实际 %q %v", v, ok)
	}
	if _, ok := st.Get("old"); ok {
		t.Fatal("old 应已不存在")
	}
	// 改名不存在的键应返回 false
	if ok, _ := st.Rename("nope", "x"); ok {
		t.Fatal("改名不存在的键应返回 false")
	}
}

func TestStoreAppend(t *testing.T) {
	path := filepath.Join(t.TempDir(), "db.json")
	st, _ := NewStore(path)
	st.Set("log", "第一行")
	if err := st.Append("log", "第二行", "\n"); err != nil {
		t.Fatal(err)
	}
	if v, _ := st.Get("log"); v != "第一行\n第二行" {
		t.Fatalf("append 不符: %q", v)
	}
	// 给一个不存在的键追加，应该直接成为值
	if err := st.Append("fresh", "x", ","); err != nil {
		t.Fatal(err)
	}
	if v, _ := st.Get("fresh"); v != "x" {
		t.Fatalf("追加到空键不符: %q", v)
	}
}

func TestStoreClear(t *testing.T) {
	path := filepath.Join(t.TempDir(), "db.json")
	st, _ := NewStore(path)
	st.Set("a", "1")
	st.Set("b", "2")
	if err := st.Clear(); err != nil {
		t.Fatal(err)
	}
	if len(st.Keys()) != 0 {
		t.Fatalf("clear 后应无键，实际 %#v", st.Keys())
	}
	// clear 也该持久化：重开文件应为空
	st2, _ := NewStore(path)
	if len(st2.Keys()) != 0 {
		t.Fatalf("clear 未持久化，重开仍有键: %#v", st2.Keys())
	}
}

func TestStoreKeysWithPrefix(t *testing.T) {
	path := filepath.Join(t.TempDir(), "db.json")
	st, _ := NewStore(path)
	st.Set("user:1", "a")
	st.Set("user:2", "b")
	st.Set("post:1", "c")
	got := st.KeysWithPrefix("user:")
	if len(got) != 2 || got[0] != "user:1" || got[1] != "user:2" {
		t.Fatalf("前缀过滤不符: %#v", got)
	}
	// 空前缀返回全部
	if got := st.KeysWithPrefix(""); len(got) != 3 {
		t.Fatalf("空前缀应返回全部 3 个，实际 %d", len(got))
	}
}

// requireGet 是测试里的取值辅助，避免重复写 bool 判断
func requireGet(st *Store, k string) (string, bool) {
	return st.Get(k)
}
