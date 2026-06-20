package main

import (
	"embed"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

//go:embed static/*
var staticFS embed.FS

func main() {
	port := flag.Int("port", 8374, "服务端口")
	dbPath := flag.String("db", "tkd.db", "SQLite数据库文件路径")
	flag.Parse()

	if err := InitDB(*dbPath); err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
		os.Exit(1)
	}
	log.Printf("数据库已加载: %s", *dbPath)

	mux := http.NewServeMux()

	mux.HandleFunc("/api/meta", HandleMeta)
	mux.HandleFunc("/api/students", HandleStudents)
	mux.HandleFunc("/api/students/", HandleStudentByID)
	mux.HandleFunc("/api/students/red-above", HandleRedBeltStudents)
	mux.HandleFunc("/api/training", HandleTrainingRecords)
	mux.HandleFunc("/api/training/monthly-duration", HandleMonthlyTrainingDuration)
	mux.HandleFunc("/api/training/monthly-type-count", HandleMonthlyTrainingTypeCount)
	mux.HandleFunc("/api/exams", HandleExamRecords)

	mux.HandleFunc("/", serveStatic)

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("跆拳道馆管理系统启动于 http://localhost%s", addr)
	if err := http.ListenAndServe(addr, corsMiddleware(mux)); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func serveStatic(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if path == "/" || path == "" {
		path = "/index.html"
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	filePath := "static" + path
	data, err := staticFS.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if strings.HasSuffix(filePath, ".html") {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
	} else if strings.HasSuffix(filePath, ".css") {
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
	} else if strings.HasSuffix(filePath, ".js") {
		w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	} else if strings.HasSuffix(filePath, ".json") {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
	}
	w.Write(data)
}
