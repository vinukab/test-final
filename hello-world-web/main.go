package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const indexHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Hello World</title>
<style>
html, body { margin: 0; padding: 0; height: 100%; }
body {
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
  font-size: 1rem;
  line-height: 1.5;
  color: #1a1a1a;
  background: #f7f7f8;
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  padding: 1rem;
  box-sizing: border-box;
}
main {
  text-align: center;
  max-width: 40rem;
  width: 100%;
}
h1 {
  font-size: clamp(2rem, 6vw, 4rem);
  margin: 0 0 0.5rem 0;
}
p { font-size: 1rem; margin: 0; color: #555; }
</style>
</head>
<body>
<main>
<h1>Hello, World!</h1>
<p>Welcome to the test-final web application.</p>
</main>
</body>
</html>`

const notFoundHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Page Not Found</title>
<style>
html, body { margin: 0; padding: 0; height: 100%; }
body {
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
  font-size: 1rem;
  line-height: 1.5;
  color: #1a1a1a;
  background: #f7f7f8;
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  padding: 1rem;
  box-sizing: border-box;
}
main { text-align: center; max-width: 40rem; width: 100%; }
h1 { font-size: clamp(1.75rem, 5vw, 3rem); margin: 0 0 0.5rem 0; }
p { font-size: 1rem; margin: 0 0 1rem 0; color: #555; }
a { color: #0366d6; text-decoration: none; }
a:hover { text-decoration: underline; }
</style>
</head>
<body>
<main>
<h1>404 — Page Not Found</h1>
<p>The page you requested could not be found.</p>
<p><a href="/">Return to home</a></p>
</main>
</body>
</html>`

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Content-Security-Policy", "default-src 'self'; style-src 'unsafe-inline'")
		h.Set("Referrer-Policy", "no-referrer")
		h.Set("Server", "")
		next.ServeHTTP(w, r)
	})
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)
	_, _ = w.Write([]byte(notFoundHTML))
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		notFoundHandler(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(indexHTML))
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", rootHandler)
	mux.HandleFunc("/health", healthHandler)

	server := &http.Server{
		Addr:         ":9090",
		Handler:      securityHeaders(mux),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	idleConnsClosed := make(chan struct{})
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			log.Printf("shutdown error: %v", err)
		}
		close(idleConnsClosed)
	}()

	log.Println("listening on :9090")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
	<-idleConnsClosed
}
