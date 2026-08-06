package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
)

// 用法：go-filedb -db x.json set name 张三
// 子命令：set / get / del / keys / has / update / export / import
func main() {
	dbPath := flag.String("db", "filedb.json", "库文件路径")
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "子命令：set <k> <v> / get <k> / del <k> / keys / has <k> / update（从标准输入读 k=v）/ export / import <文件>")
		os.Exit(2)
	}

	st, err := NewStore(*dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "打开库失败: %v\n", err)
		os.Exit(1)
	}

	switch args[0] {
	case "set":
		if len(args) < 3 {
			fmt.Fprintln(os.Stderr, "set 需要 key 和 value")
			os.Exit(2)
		}
		if err := st.Set(args[1], args[2]); err != nil {
			fmt.Fprintf(os.Stderr, "写入失败: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("ok")
	case "get":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "get 需要 key")
			os.Exit(2)
		}
		if v, ok := st.Get(args[1]); ok {
			fmt.Println(v)
		} else {
			fmt.Printf("（没有 %s）\n", args[1])
			os.Exit(1)
		}
	case "del":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "del 需要 key")
			os.Exit(2)
		}
		if st.Del(args[1]) {
			fmt.Println("已删除")
		} else {
			fmt.Println("本来就没有")
		}
	case "keys":
		for _, k := range st.Keys() {
			fmt.Println(k)
		}
	case "has":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "has 需要 key")
			os.Exit(2)
		}
		if st.Has(args[1]) {
			fmt.Println("有")
		} else {
			fmt.Println("没有")
			os.Exit(1)
		}
	case "update":
		// 从标准输入读 k=v 一行一个，批量写入
		pairs, perr := readPairs(os.Stdin)
		if perr != nil {
			fmt.Fprintf(os.Stderr, "读取失败: %v\n", perr)
			os.Exit(1)
		}
		if err := st.UpdateMany(pairs); err != nil {
			fmt.Fprintf(os.Stderr, "写入失败: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("已更新 %d 条\n", len(pairs))
	case "export":
		// 把整个库导出成 JSON 文本
		b, err := json.MarshalIndent(st.Export(), "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "序列化失败: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(b))
	case "import":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "import 需要文件路径")
			os.Exit(2)
		}
		b, rerr := os.ReadFile(args[1])
		if rerr != nil {
			fmt.Fprintf(os.Stderr, "读文件失败: %v\n", rerr)
			os.Exit(1)
		}
		var pairs map[string]string
		if err := json.Unmarshal(b, &pairs); err != nil {
			fmt.Fprintf(os.Stderr, "JSON 解析失败: %v\n", err)
			os.Exit(1)
		}
		if err := st.Import(pairs); err != nil {
			fmt.Fprintf(os.Stderr, "导入失败: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("已导入 %d 条\n", len(pairs))
	default:
		fmt.Fprintf(os.Stderr, "不认识子命令: %s\n", args[0])
		os.Exit(2)
	}
}

// readPairs 从 reader 读 k=v 行，跳过空行和没有等号的行
func readPairs(r *os.File) (map[string]string, error) {
	pairs := map[string]string{}
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || !strings.Contains(line, "=") {
			continue
		}
		i := strings.Index(line, "=")
		pairs[line[:i]] = line[i+1:]
	}
	return pairs, sc.Err()
}
