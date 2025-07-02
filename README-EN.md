<p align="center">
<img alt="Clipboard" src="favicon.png">
<br>
<em>Minimalist Web Clipboard</em>
<br><br>

<p align="center">
<a href="README-EN.md">英語</a> | <a href="README.md">中文</a>
</p>


A minimalist, distraction-free web clipboard. Share and access text content easily without logins or complicated setup.

### Features

- **Clean Interface**: Neat and tidy user interface focused on content
- **Auto-save**: Automatically saves all changes without manual intervention
- **Random URLs**: Each note has a unique random URL for easy sharing
- **Dark Mode**: Automatically adapts to system dark mode settings
- **Responsive Design**: Works perfectly on any device
- **Zero Dependencies**: Pure Go implementation with no external dependencies
- **API Friendly**: Supports plain text access via curl/wget and other tools

### Usage

#### Server

```bash
# Clone the repository
git clone https://github.com/wwxiaoqi/net-clipboard
cd net-clipboard

# Run the server (default port: 8736)
go run main.go
```

The server will start on port `:8736` by default.

#### Client

##### Web Interface

Simply visit the server address, and a new note will be automatically created for you:

```
http://localhost:8736/
```

##### Using curl

Read a note:
```bash
curl http://localhost:8736/noteid
```

Create/update a note:
```bash
echo "Your content" | curl -d @- http://localhost:8736/noteid
```

Delete a note:
```bash
curl -d "" http://localhost:8736/noteid
```

### Configuration Options

The following configurations can be modified in `main.go`:

```go
const (
    savePath     = "_tmp"     // Path to save notes
    listenAddr   = ":8736"    // Port to listen on
    maxNoteLen   = 64         // Maximum length of note name
    noteIdRegexp = `^[a-zA-Z0-9_-]+$` // Regex for valid note names
)
```
