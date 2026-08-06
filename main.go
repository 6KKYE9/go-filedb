package main

import (
	"flag"
	"fmt"
	"os"
)

// 用法：go-filedb -db x.json set name 张三
// 子命令：set / get / del / keys
func main() {
	dbPath := flag.String("db", "filedb.json", "库文件路径")
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "子命令：set <k> <v> / get <k> / del <k> / keys")
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
	default:
		fmt.Fprintf(os.Stderr, "不认识子命令: %s\n", args[0])
		os.Exit(2)
	}
}
