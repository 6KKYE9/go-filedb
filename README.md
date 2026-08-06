一个零依赖的嵌入式键值库，数据落到一个 JSON 文件。读写都带锁，写的时候先写临时文件再改名，避免写到一半崩了把库弄坏。

子命令：
  set <k> <v>   存一个键值
  get <k>      取一个键值
  del <k>      删一个键
  keys         列出所有键（按字母序）

用法：
  go-filedb -db my.json set user 张三
  go-filedb -db my.json get user
  go-filedb -db my.json keys