package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

const banner = `
  _____ _ _     _____                                                    
 |  ___(_) | __|  ___|__  _ __ ___ __ _ _ __   ___ ___  _ __   __ _ _ __  
 | |_  | | |/ /| |_ / _ \| '__/ __/ _' | '_ \ / __/ _ \| '_ \ / _' | '__| 
 |  _| | |   < |  _| (_) | | | (_| (_| | | | | (_| (_) | |_) | (_| | |   
 |_|   |_|_|\_\|_|  \___/|_|  \___\__,_|_| |_|\___\___/| .__/ \__,_|_|   
                                                         |_|              
  Concurrent Static File Server v1.0
`

func main() {
	// ── CLI Flags ──────────────────────────────────────────────
	port := flag.Int("port", 8080, "TCP port to listen on (e.g., 8080, 3000)")
	root := flag.String("root", "./public", "Root directory to serve files from")
	host := flag.String("host", "0.0.0.0", "Host address to bind to")
	showVersion := flag.Bool("version", false, "Show version information")

	flag.Usage = func() {
		fmt.Fprint(os.Stderr, banner)
		fmt.Fprintln(os.Stderr, "Usage:")
		fmt.Fprintln(os.Stderr, "  file-server [options]")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Examples:")
		fmt.Fprintln(os.Stderr, "  file-server -port 3000 -root ./public")
		fmt.Fprintln(os.Stderr, "  file-server -port 8080 -root /var/www/html")
		fmt.Fprintln(os.Stderr, "  file-server -host 127.0.0.1 -port 443 -root ./assets")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Options:")
		flag.PrintDefaults()
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Verify gzip compression:")
		fmt.Fprintln(os.Stderr, "  curl -I -H \"Accept-Encoding: gzip\" http://localhost:<port>/file.html")
		fmt.Fprintln(os.Stderr, "  # Look for \"Content-Encoding: gzip\" in the response headers")
	}

	flag.Parse()

	// ── Handle version flag ────────────────────────────────────
	if *showVersion {
		fmt.Println("file-server v1.0.0")
		fmt.Println("Concurrent Static File Server — Go Infrastructure Project")
		os.Exit(0)
	}

	// ── Validate port range ────────────────────────────────────
	if *port < 1 || *port > 65535 {
		fmt.Fprintf(os.Stderr, "Error: port must be between 1 and 65535, got %d\n", *port)
		os.Exit(1)
	}

	// ── Resolve absolute path for root ─────────────────────────
	absRoot, err := resolveRoot(*root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: cannot resolve root directory %q: %v\n", *root, err)
		os.Exit(1)
	}

	// ── Build server configuration ─────────────────────────────
	cfg := DefaultConfig()
	cfg.Host = *host
	cfg.Port = *port
	cfg.Root = absRoot

	// ── Print startup banner ───────────────────────────────────
	fmt.Fprint(os.Stderr, banner)
	log.Printf("📂 Root directory: %s", cfg.Root)
	log.Printf("🌐 Address:        %s:%d", cfg.Host, cfg.Port)

	// ── Create and start server ────────────────────────────────
	srv, err := NewServer(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if err := srv.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}

// resolveRoot returns the absolute path for the given root directory,
// validating that it exists and is a directory.
func resolveRoot(root string) (string, error) {
	info, err := os.Stat(root)
	if err != nil {
		return "", fmt.Errorf("cannot access directory: %w", err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("%q is not a directory", root)
	}

	absPath, err := os.Getwd()
	if err != nil {
		return root, nil // fallback to relative path
	}

	// If root is already absolute, return as-is
	if len(root) > 0 && (root[0] == '/' || (len(root) > 1 && root[1] == ':')) {
		return root, nil
	}

	return absPath + string(os.PathSeparator) + root, nil
}