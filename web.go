package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

//go:embed static/index.html
var indexHTML []byte

// ExecutionStatus holds the current / last run state
type ExecutionStatus struct {
	Running bool     `json:"running"`
	Script  string   `json:"script"`
	Log     []string `json:"log"`
	Error   string   `json:"error,omitempty"`
	Done    bool     `json:"done"`
}

var (
	execStatus ExecutionStatus
	execMu     sync.Mutex
	globalHID  string
	globalDir  string
)

func startWebServer(port, scriptsDir, hidDev string) {
	globalHID = hidDev
	globalDir = scriptsDir

	mux := http.NewServeMux()
	mux.HandleFunc("/", handleIndex)
	mux.HandleFunc("/scripts", handleScripts)
	mux.HandleFunc("/upload", handleUpload)
	mux.HandleFunc("/layout", handleLayout)
	mux.HandleFunc("/save", handleSave)
	mux.HandleFunc("/load", handleLoad)
	mux.HandleFunc("/run", handleRun)
	mux.HandleFunc("/status", handleStatus)
	mux.HandleFunc("/delete", handleDelete)

	if err := http.ListenAndServe(":"+port, mux); err != nil {
		fmt.Printf("[!] Web server: %v\n", err)
	}
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(indexHTML)
}

func handleScripts(w http.ResponseWriter, r *http.Request) {
	entries, err := os.ReadDir(globalDir)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	var names []string
	for _, e := range entries {
		n := e.Name()
		if !e.IsDir() && (strings.HasSuffix(n, ".txt") || strings.HasSuffix(n, ".ds")) {
			names = append(names, n)
		}
	}
	if names == nil {
		names = []string{} // return [] not null
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(names)
}

func handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", 405)
		return
	}
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "parse form: "+err.Error(), 400)
		return
	}
	file, header, err := r.FormFile("script")
	if err != nil {
		http.Error(w, "no file: "+err.Error(), 400)
		return
	}
	defer file.Close()

	name := filepath.Base(header.Filename)
	// ensure safe extension
	if !strings.HasSuffix(name, ".txt") && !strings.HasSuffix(name, ".ds") {
		name += ".txt"
	}

	dst, err := os.Create(filepath.Join(globalDir, name))
	if err != nil {
		http.Error(w, "create file: "+err.Error(), 500)
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, "write file: "+err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"name": name})
}

func handleRun(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", 405)
		return
	}

	execMu.Lock()
	if execStatus.Running {
		execMu.Unlock()
		http.Error(w, "script already running", 409)
		return
	}
	execMu.Unlock()

	name := filepath.Base(r.URL.Query().Get("name"))
	if name == "" || name == "." {
		http.Error(w, "missing name", 400)
		return
	}

	content, err := os.ReadFile(filepath.Join(globalDir, name))
	if err != nil {
		http.Error(w, "script not found", 404)
		return
	}

	execMu.Lock()
	execStatus = ExecutionStatus{Running: true, Script: name, Log: []string{}}
	execMu.Unlock()

	go runScript(name, string(content))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"started": true})
}

func runScript(name, content string) {
	log := func(msg string) {
		execMu.Lock()
		execStatus.Log = append(execStatus.Log, msg)
		execMu.Unlock()
	}

	// Small grace period so the host USB device is ready
	time.Sleep(500 * time.Millisecond)

	kbd, err := NewHIDKeyboard(globalHID)
	if err != nil {
		execMu.Lock()
		execStatus.Running = false
		execStatus.Error = fmt.Sprintf("open HID device: %v", err)
		execStatus.Done = true
		execMu.Unlock()
		return
	}
	defer kbd.Close()

	execErr := ExecuteDucky(content, kbd, log)

	execMu.Lock()
	execStatus.Running = false
	execStatus.Done = true
	if execErr != nil {
		execStatus.Error = execErr.Error()
	}
	execMu.Unlock()
}

func handleStatus(w http.ResponseWriter, r *http.Request) {
	execMu.Lock()
	s := execStatus
	execMu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s)
}

func handleLayout(w http.ResponseWriter, r *http.Request) {
	type info struct {
		Code    string `json:"code"`
		Display string `json:"display"`
	}
	switch r.Method {
	case http.MethodGet:
		var available []info
		for _, code := range LayoutOrder {
			l := Layouts[code]
			available = append(available, info{Code: l.Code, Display: l.Display})
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"current": GetLayout().Code,
			"layouts": available,
		})
	case http.MethodPost:
		code := r.URL.Query().Get("code")
		if code == "" {
			var body struct {
				Code string `json:"code"`
			}
			json.NewDecoder(r.Body).Decode(&body)
			code = body.Code
		}
		if !SetLayout(code) {
			http.Error(w, "unknown layout: "+code, 400)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"code": code})
	default:
		http.Error(w, "GET or POST only", 405)
	}
}

func handleSave(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", 405)
		return
	}
	var body struct {
		Name    string `json:"name"`
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "decode: "+err.Error(), 400)
		return
	}
	name := filepath.Base(body.Name)
	if name == "" || name == "." {
		http.Error(w, "invalid name", 400)
		return
	}
	if !strings.HasSuffix(name, ".txt") && !strings.HasSuffix(name, ".ds") {
		name += ".txt"
	}
	if err := os.WriteFile(filepath.Join(globalDir, name), []byte(body.Content), 0644); err != nil {
		http.Error(w, "write: "+err.Error(), 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"name": name})
}

func handleLoad(w http.ResponseWriter, r *http.Request) {
	name := filepath.Base(r.URL.Query().Get("name"))
	if name == "" || name == "." {
		http.Error(w, "invalid name", 400)
		return
	}
	content, err := os.ReadFile(filepath.Join(globalDir, name))
	if err != nil {
		http.Error(w, "not found", 404)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"content": string(content)})
}

func handleDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "DELETE only", 405)
		return
	}
	name := filepath.Base(r.URL.Query().Get("name"))
	if name == "" || name == "." {
		http.Error(w, "missing name", 400)
		return
	}
	os.Remove(filepath.Join(globalDir, name))
	w.WriteHeader(204)
}
