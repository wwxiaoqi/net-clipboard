package main

import (
	"crypto/rand"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	// Configuration
	savePath     = "_tmp"     // Path to save notes
	listenAddr   = ":8736"    // Port to listen on
	maxNoteLen   = 64         // Maximum length of note name
	noteIdRegexp = `^[a-zA-Z0-9_-]+$` // Regex for valid note names
)

// generateRandomID creates a random note ID with 5 characters
func generateRandomID() string {
	const chars = "234579abcdefghjkmnpqrstwxyz"
	bytes := make([]byte, 5)
	rand.Read(bytes)
	
	for i := 0; i < 5; i++ {
		bytes[i] = chars[int(bytes[i])%len(chars)]
	}
	
	return string(bytes)
}

// validNoteID checks if the note ID is valid
func validNoteID(noteID string) bool {
	if noteID == "" || len(noteID) > maxNoteLen {
		return false
	}
	
	matched, _ := regexp.MatchString(noteIdRegexp, noteID)
	return matched
}

// serveNotepad handles the main functionality
func serveNotepad(w http.ResponseWriter, r *http.Request) {

	// Parse the URL path to get the note ID
	path := strings.TrimPrefix(r.URL.Path, "/")
	
	// If no note ID or invalid note ID, redirect to a new random note
	if !validNoteID(path) {
		http.Redirect(w, r, "/"+generateRandomID(), http.StatusFound)
		return
	}
	
	filePath := filepath.Join(savePath, path)
	
	// Handle POST requests (save content)
	if r.Method == http.MethodPost {
		var content string
		
		// Always read the raw body first
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Error", http.StatusInternalServerError)
			return
		}
		
		contentType := r.Header.Get("Content-Type")

		// Decode URL-encoded body to handle browser submissions properly
		if strings.Contains(contentType, "application/x-www-form-urlencoded") {

			// Parse form data - this is for browser submissions
			if err := r.ParseForm(); err == nil {

				// For browser submissions, get the "text" parameter and URL-decode it
				if textValue := r.PostFormValue("text"); textValue != "" {
					content = textValue
				} else {

					// If no "text" parameter, use the raw body
					bodyStr := string(body)
					if strings.HasPrefix(bodyStr, "text=") {
						decoded, err := url.QueryUnescape(bodyStr[5:]) // Skip "text="
						if err == nil {
							content = decoded
						} else {
							content = bodyStr
						}
					} else {
						content = bodyStr
					}
				}
			} else {
				content = string(body)
			}
		} else {
			// For curl/wget and API clients, use the raw body as-is
			content = string(body)
		}
		
		// If content is empty, delete the file
		if content == "" {
			os.Remove(filePath)
		} else {

			// Create directory if it doesn't exist
			if err := os.MkdirAll(savePath, 0755); err != nil {
				http.Error(w, "Error", http.StatusInternalServerError)
				return
			}
			
			// Write content to file
			err := os.WriteFile(filePath, []byte(content), 0644)
			if err != nil {
				http.Error(w, "Error", http.StatusInternalServerError)
				return
			}
		}
		
		// Return the content for curl and API clients
		userAgent := r.UserAgent()
		if strings.HasPrefix(userAgent, "curl") || strings.HasPrefix(userAgent, "wget") {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.Write([]byte("OK"))
		} else {
			w.WriteHeader(http.StatusOK)
		}
		return
	}
	
	// Check if raw display is requested or if user agent is curl/wget
	userAgent := r.UserAgent()
	isRaw := r.URL.Query().Get("raw") != "" || 
		strings.HasPrefix(userAgent, "curl") || 
		strings.HasPrefix(userAgent, "wget")
	
	if isRaw {
		// Check if file exists before trying to read it
		_, err := os.Stat(filePath)
		fileExists := !os.IsNotExist(err)

		if fileExists {

			// File exists, serve its content
			content, err := os.ReadFile(filePath)
			if err != nil {
				http.Error(w, "Error", http.StatusInternalServerError)
				return
			}
			
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.Write(content)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
		return
	}
	
	// Serve HTML page
	serveHTMLPage(w, r, path, filePath)
}

// cleanupFormData cleans up URL encoding from saved form data
func cleanupFormData(content string) string {

	// If content starts with "text=", remove it
	if strings.HasPrefix(content, "text=") {
		content = content[5:]
	}
	
	// Replace URL encoded characters
	content = strings.ReplaceAll(content, "%0A", "\n") // newline
	content = strings.ReplaceAll(content, "%0D", "\r") // carriage return
	content = strings.ReplaceAll(content, "%09", "\t") // tab
	content = strings.ReplaceAll(content, "%20", " ")  // space
	
	return content
}

// serveHTMLPage serves the HTML notepad page
func serveHTMLPage(w http.ResponseWriter, r *http.Request, noteID, filePath string) {

	// Check if the file exists and read its content
	var content string
	if fileData, err := os.ReadFile(filePath); err == nil {
		// Clean up any potentially URL-encoded content
		content = cleanupFormData(string(fileData))
	}
	
	// Set security headers to prevent XSS attacks
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	
	// Serve the HTML page with template, properly escaping all dynamic content
	// Escape content for safe HTML insertion
	escapedContent := html.EscapeString(content)
	escapedNoteID := html.EscapeString(noteID)
	
	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Clipboard - %s</title>
    <link rel="icon" href="/favicon.ico" sizes="any">
    <link rel="icon" href="/favicon.svg" type="image/svg+xml">
    <style>
        :root {
            --primary-color: #4a6fa5;
            --secondary-color: #6383b5;
            --bg-color: #f8f9fa;
            --text-color: #212529;
            --border-color: #dfe3e8;
            --input-bg: #ffffff;
            --shadow-color: rgba(0, 0, 0, 0.1);
            --header-height: 60px;
            --footer-height: 40px;
            --transition-speed: 0.3s;
        }
        
        @media (prefers-color-scheme: dark) {
            :root {
                --primary-color: #4a7cb5;
                --secondary-color: #5a8bc5;
                --bg-color: #1a1e24;
                --text-color: #e4e6eb;
                --border-color: #3a3f48;
                --input-bg: #2a2e35;
                --shadow-color: rgba(0, 0, 0, 0.3);
            }
        }
        
        * {
            box-sizing: border-box;
            margin: 0;
            padding: 0;
        }
        
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, 'Open Sans', 'Helvetica Neue', sans-serif;
            background-color: var(--bg-color);
            color: var(--text-color);
            line-height: 1.6;
            transition: background-color var(--transition-speed), color var(--transition-speed);
            height: 100vh;
            overflow: hidden;
        }
        
        .app-container {
            display: flex;
            flex-direction: column;
            height: 100vh;
        }
        
        header {
            height: var(--header-height);
            background-color: var(--primary-color);
            color: white;
            display: flex;
            align-items: center;
            padding: 0 20px;
            box-shadow: 0 2px 4px var(--shadow-color);
            z-index: 10;
        }
        
        .header-content {
            display: flex;
            justify-content: space-between;
            align-items: center;
            width: 100%%;
        }
        
        .note-id {
            font-size: 1.2rem;
            font-weight: 600;
        }
        
        .header-actions {
            display: flex;
            gap: 10px;
        }
        
        .action-btn {
            background: var(--secondary-color);
            color: white;
            border: none;
            border-radius: 4px;
            padding: 8px 12px;
            cursor: pointer;
            font-size: 14px;
            transition: background-color 0.2s;
        }
        
        .action-btn:hover {
            background-color: var(--primary-color);
        }
        
        .content-area {
            flex: 1;
            overflow: hidden;
            padding: 20px;
            display: flex;
            flex-direction: column;
        }
        
        #content {
            flex: 1;
            width: 100%%;
            padding: 15px;
            border: 1px solid var(--border-color);
            border-radius: 6px;
            background-color: var(--input-bg);
            color: var(--text-color);
            font-size: 16px;
            line-height: 1.6;
            resize: none;
            outline: none;
            box-shadow: 0 1px 3px var(--shadow-color);
            transition: border-color 0.2s, background-color var(--transition-speed), color var(--transition-speed);
        }
        
        #content:focus {
            border-color: var(--primary-color);
            box-shadow: 0 0 0 2px rgba(74, 111, 165, 0.2);
        }
        
        footer {
            height: var(--footer-height);
            background-color: var(--input-bg);
            border-top: 1px solid var(--border-color);
            display: flex;
            align-items: center;
            justify-content: space-between;
            padding: 0 20px;
            font-size: 14px;
            color: #888;
        }
        
        .status {
            display: flex;
            align-items: center;
            gap: 5px;
        }
        
        .status-icon {
            width: 8px;
            height: 8px;
            border-radius: 50%%;
            background-color: #4caf50;
        }
        
        #printable {
            display: none;
            white-space: pre-wrap;
            word-break: break-word;
        }
        
        @media (max-width: 768px) {
            .header-content {
                flex-direction: column;
                align-items: flex-start;
                gap: 10px;
                padding: 10px 0;
            }
            
            header {
                height: auto;
                padding: 0 15px;
            }
            
            .content-area {
                padding: 15px;
            }
            
            #content {
                font-size: 16px;
                padding: 12px;
            }
            
            .header-actions {
                width: 100%%;
                justify-content: flex-end;
            }
            
            .action-btn {
                padding: 6px 10px;
                font-size: 13px;
            }
            
            footer {
                padding: 0 15px;
            }
        }
        
        @media print {
            .app-container {
                display: none;
            }
            
            #printable {
                display: block;
                padding: 20px;
                font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, 'Open Sans', 'Helvetica Neue', sans-serif;
                color: #000;
                background: #fff;
            }
        }
    </style>
</head>
<body>
    <div class="app-container">
        <header>
            <div class="header-content">
                <div class="note-id">Clipboard: %s</div>
                <div class="header-actions">
                    <button class="action-btn" id="new-note">New Note</button>
                    <button class="action-btn" id="copy-url">Copy URL</button>
                </div>
            </div>
        </header>
        
        <div class="content-area">
            <textarea id="content" placeholder="Start writing...">%s</textarea>
        </div>
        
        <footer>
            <div class="status">
                <div class="status-icon"></div>
                <span id="save-status">Changes saved</span>
            </div>
            <div id="char-count">0 characters</div>
        </footer>
    </div>
    
    <pre id="printable"></pre>
    
    <script>
        // Initialize variables
        const textarea = document.getElementById('content');
        const printable = document.getElementById('printable');
        const saveStatus = document.getElementById('save-status');
        const charCount = document.getElementById('char-count');
        const statusIcon = document.querySelector('.status-icon');
        let content = textarea.value;
        let saveTimeout;

        // Update character count initially
        updateCharCount();
        
        // Initialize the printable contents with the initial value of the textarea
        printable.appendChild(document.createTextNode(content));
        
        // Set focus to the textarea
        textarea.focus();
        
        // Setup event listeners
        document.getElementById('new-note').addEventListener('click', () => {
            window.location.href = '/' + Math.random().toString(36).substring(2, 7);
        });
        
        document.getElementById('copy-url').addEventListener('click', () => {
            navigator.clipboard.writeText(window.location.href)
                .then(() => {
                    alert('URL已复制到剪贴板');
                })
                .catch(err => {
                    alert('复制失败: ' + err);
                });
        });
        
        textarea.addEventListener('input', () => {
            updateSaveStatus('Saving...', '#ffc107');
            updateCharCount();
            clearTimeout(saveTimeout);
            saveTimeout = setTimeout(uploadContent, 500);
        });
        
        // Update character count function
        function updateCharCount() {
            charCount.textContent = textarea.value.length + ' characters';
        }
        
        // Update save status function
        function updateSaveStatus(message, color) {
            saveStatus.textContent = message;
            statusIcon.style.backgroundColor = color;
        }
        
        // Function to upload content
        function uploadContent() {
            if (content !== textarea.value) {
                const temp = textarea.value;
                const request = new XMLHttpRequest();
                
                updateSaveStatus('Saving...', '#ffc107');
                
                request.open('POST', window.location.href, true);
                request.setRequestHeader('Content-Type', 'text/plain; charset=UTF-8');
                
                request.onload = function() {
                    if (request.readyState === 4) {
                        // Update content variable if request was successful
                        content = temp;
                        updateSaveStatus('Changes saved', '#4caf50');
                        
                        // Check again after delay
                        setTimeout(() => {
                            if (content === textarea.value) {
                                // If content hasn't changed, start monitoring again
                                setTimeout(checkForChanges, 1000);
                            } else {
                                // If content has changed, upload immediately
                                uploadContent();
                            }
                        }, 500);
                    }
                }
                
                request.onerror = function() {
                    updateSaveStatus('Error saving', '#f44336');
                    // Try again after 2 seconds
                    setTimeout(uploadContent, 2000);
                }
                
                // Send the content directly
                request.send(temp);
                
                // Update the printable contents
                printable.textContent = temp;
            } else {
                // If content hasn't changed, start monitoring
                checkForChanges();
            }
        }
        
        // Function to check for changes
        function checkForChanges() {
            if (content !== textarea.value) {
                uploadContent();
            } else {
                setTimeout(checkForChanges, 1000);
            }
        }
        
        // Start the content upload process
        uploadContent();
    </script>
</body>
</html>`, escapedNoteID, escapedNoteID, escapedContent)
	
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}

func main() {
	// Create the save directory if it doesn't exist
	if err := os.MkdirAll(savePath, 0755); err != nil {
		log.Fatalf("Failed to create save directory: %v", err)
	}
	
	// Configure static file serving for favicon
	http.Handle("/favicon.ico", http.FileServer(http.Dir(".")))
	http.Handle("/favicon.svg", http.FileServer(http.Dir(".")))
	
	// Set up the main handler
	http.HandleFunc("/", serveNotepad)
	
	// Start the server
	log.Printf("Starting server: %s", listenAddr)
	if err := http.ListenAndServe(listenAddr, nil); err != nil {
		log.Fatalf("Server error: %v", err)
	}
} 