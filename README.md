一个零依赖的嵌入式键值库，数据落到一个 JSON 文件。读写都带锁，写的时候先写临时文件再改名，避免写到一半崩了把库弄坏。

子命令：
  set <k> <v>       存一个键值
  get <k>            取一个键值
  del <k>            删一个键
  keys               列出所有键（按字母序）；加 -p 前缀 只列以它开头的
  has <k>            判断某个键在不在
  rename <旧> <新>   把键改名（新键已存在则覆盖）
  append <k> <内容>  往某个键的值后面追加内容（用 -sep 指定连接符，默认换行）
  clear              清空整个库
  update             从标准输入读 k=v（一行一个）批量写入
  export             把整个库导出成 JSON
  import <文件>      用一份 JSON 覆盖当前库

用法：
  go-filedb -db my.json set user 张三
  go-filedb -db my.json get user
  go-filedb -db my.json keys
  go-filedb -db my.json keys -p user:
  go-filedb -db my.json rename user name
  go-filedb -db my.json append log "新一行" -sep " | "
  go-filedb -db my.json clear

测试：
  go test
