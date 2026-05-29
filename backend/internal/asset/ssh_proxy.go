package asset

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
	"golang.org/x/crypto/ssh"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type sshConnectRequest struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
}

func handleSSH(w http.ResponseWriter, r *http.Request) {
	// Try to get asset credentials from DB
	assetIDStr := r.URL.Query().Get("asset_id")
	if assetIDStr != "" {
		id, err := strconv.ParseUint(assetIDStr, 10, 64)
		if err == nil {
			a, _ := repo.GetAsset(r.Context(), id)
			if a != nil && a.SSHUser != "" {
				port := 22
				if p, err := strconv.Atoi(a.SSHPort); err == nil && p > 0 {
					port = p
				}
				connectSSH(w, r, sshConnectRequest{
					Host:     a.IP,
					Port:     port,
					User:     a.SSHUser,
					Password: a.SSHPassword,
				})
				return
			}
		}
	}

	// No stored credentials - upgrade and wait for JSON message
	connectSSHWithCreds(w, r)
}

func connectSSHWithCreds(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("websocket upgrade", "error", err)
		return
	}
	defer conn.Close()

	_, msg, err := conn.ReadMessage()
	if err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte("Failed to read credentials"))
		return
	}

	var req sshConnectRequest
	if err := json.Unmarshal(msg, &req); err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte("Invalid credentials format"))
		return
	}

	if req.Host == "" || req.User == "" {
		conn.WriteMessage(websocket.TextMessage, []byte("Host and user are required"))
		return
	}
	if req.Port == 0 {
		req.Port = 22
	}

	sshProxy(conn, req)
}

func connectSSH(w http.ResponseWriter, r *http.Request, req sshConnectRequest) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("websocket upgrade", "error", err)
		return
	}
	defer conn.Close()

	sshProxy(conn, req)
}

func sshProxy(ws *websocket.Conn, req sshConnectRequest) {
	config := &ssh.ClientConfig{
		User:            req.User,
		Auth:            []ssh.AuthMethod{ssh.Password(req.Password)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	addr := req.Host + ":" + strconv.Itoa(req.Port)
	sshClient, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		ws.WriteMessage(websocket.TextMessage, []byte("\r\n\x1b[31mSSH connection failed: "+err.Error()+"\x1b[0m\r\n"))
		return
	}
	defer sshClient.Close()

	session, err := sshClient.NewSession()
	if err != nil {
		ws.WriteMessage(websocket.TextMessage, []byte("\r\n\x1b[31mSSH session failed: "+err.Error()+"\x1b[0m\r\n"))
		return
	}
	defer session.Close()

	stdinPipe, _ := session.StdinPipe()
	stdoutPipe, _ := session.StdoutPipe()
	stderrPipe, _ := session.StderrPipe()

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}
	if err := session.RequestPty("xterm-256color", 40, 120, modes); err != nil {
		ws.WriteMessage(websocket.TextMessage, []byte("\r\n\x1b[31mPTY request failed\x1b[0m\r\n"))
		return
	}

	if err := session.Shell(); err != nil {
		ws.WriteMessage(websocket.TextMessage, []byte("\r\n\x1b[31mShell start failed\x1b[0m\r\n"))
		return
	}

	ws.WriteMessage(websocket.TextMessage, []byte("\r\n\x1b[32mConnected to "+req.Host+"\x1b[0m\r\n"))

	done := make(chan struct{})

	// WebSocket -> SSH stdin
	go func() {
		defer close(done)
		for {
			msgType, data, err := ws.ReadMessage()
			if err != nil {
				return
			}
			if msgType == websocket.TextMessage {
				var cmd map[string]any
				if json.Unmarshal(data, &cmd) == nil {
					if resize, ok := cmd["resize"]; ok {
						if r, ok := resize.(map[string]any); ok {
							if rows, ok := r["rows"].(float64); ok {
								if cols, ok := r["cols"].(float64); ok {
									session.WindowChange(int(rows), int(cols))
								}
							}
						}
						continue
					}
				}
			}
			stdinPipe.Write(data)
		}
	}()

	// SSH stdout/stderr -> WebSocket
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := stdoutPipe.Read(buf)
			if err != nil {
				return
			}
			ws.WriteMessage(websocket.BinaryMessage, buf[:n])
		}
	}()

	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := stderrPipe.Read(buf)
			if err != nil {
				return
			}
			ws.WriteMessage(websocket.BinaryMessage, buf[:n])
		}
	}()

	// Keep alive until session ends
	session.Wait()
	io.Copy(io.Discard, stdoutPipe)
	io.Copy(io.Discard, stderrPipe)
	ws.WriteMessage(websocket.TextMessage, []byte("\r\n\x1b[31mConnection closed\x1b[0m\r\n"))
}
