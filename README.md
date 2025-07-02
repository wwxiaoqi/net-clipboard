<p align="center">
<img alt="Clipboard" src="favicon.png">
<br>
<em>极简网络剪贴板</em>
<br><br>

<p align="center">
<a href="README.md">中文</a> | <a href="README-EN.md">英語</a>
</p>


一款极简、无干扰的网络剪贴板。轻松共享和访问文本内容，无需登录或复杂设置。

### 特点

- **简洁界面**：干净整洁的用户界面，专注于内容
- **即时保存**：自动保存所有更改，无需手动操作
- **随机URL**：每个笔记都有唯一随机URL，方便分享
- **深色模式**：自动适应系统深色模式设置
- **响应式设计**：在任何设备上都能完美工作
- **零依赖**：纯 Go 实现，无需外部依赖
- **API友好**：支持通过 curl/wget 等工具进行纯文本访问

### 用法

#### 服务端

```bash
# 克隆仓库
git clone https://github.com/wwxiaoqi/net-clipboard
cd net-clipboard

# 运行服务器 (默认端口: 8736)
go run main.go
```

服务器默认会在 `:8736` 端口启动。

#### 客户端

##### 网页界面

只需访问服务器地址，系统会自动为您创建一个新笔记：

```
http://localhost:8736/
```

##### 使用 curl

读取笔记:
```bash
curl http://localhost:8736/noteid
```

创建/更新笔记:
```bash
echo "您的内容" | curl -d @- http://localhost:8736/noteid
```

删除笔记:
```bash
curl -d "" http://localhost:8736/noteid
```

### 配置选项

在 `main.go` 中可以修改以下配置:

```go
const (
    savePath     = "_tmp"     // 保存笔记的路径
    listenAddr   = ":8736"    // 监听端口
    maxNoteLen   = 64         // 笔记名称最大长度
    noteIdRegexp = `^[a-zA-Z0-9_-]+$` // 有效笔记名称的正则表达式
)
```
