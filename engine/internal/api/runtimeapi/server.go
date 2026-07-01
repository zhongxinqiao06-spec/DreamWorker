package runtimeapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/zhongxinqiao06-spec/DreamWorker/engine/internal/app/workspace"
	"github.com/zhongxinqiao06-spec/DreamWorker/engine/internal/contracts/generated"
)

type ReadyMessage struct {
	OK            bool   `json:"ok"`
	Event         string `json:"event"`
	BaseURL       string `json:"baseUrl"`
	EngineVersion string `json:"engineVersion"`
	TraceID       string `json:"trace_id"`
}

type ErrorResponse = generated.DreamWorkerError

func NewMux(token string) *http.ServeMux {
	return NewMuxWithStore(token, workspace.NewStore(workspace.WithTraceID(NewTraceID)))
}

func NewMuxWithStore(token string, store *workspace.Store) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/runtime/ping", RuntimePingHandler(token))
	RegisterWorkspaceRoutes(mux, token, store)
	return mux
}

func RuntimePingHandler(token string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "仅支持 GET /runtime/ping。", "请使用 GET 方法重新调用 runtime.ping。")
			return
		}

		if !hasValidToken(r, token) {
			writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "缺少有效的本地引擎访问令牌。", "请通过 Electron Main 持有的本地令牌重新调用。")
			return
		}

		writeJSON(w, http.StatusOK, Ping(""))
	}
}

func Serve(ctx context.Context, token string, readyWriter io.Writer) error {
	return ServeWithStore(ctx, token, readyWriter, workspace.NewStore(workspace.WithTraceID(NewTraceID)))
}

func ServeWithStore(ctx context.Context, token string, readyWriter io.Writer, store *workspace.Store) error {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("listen local engine: %w", err)
	}

	server := &http.Server{
		Handler:           NewMuxWithStore(token, store),
		ReadHeaderTimeout: 5 * time.Second,
	}

	ready := ReadyMessage{
		OK:            true,
		Event:         "engine.ready",
		BaseURL:       "http://" + listener.Addr().String(),
		EngineVersion: EngineVersion,
		TraceID:       NewTraceID(),
	}
	if err := json.NewEncoder(readyWriter).Encode(ready); err != nil {
		_ = listener.Close()
		return fmt.Errorf("write ready message: %w", err)
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- server.Serve(listener)
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutdown local engine: %w", err)
		}
		return nil
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return fmt.Errorf("serve local engine: %w", err)
	}
}

func hasValidToken(r *http.Request, token string) bool {
	return token == "" || r.Header.Get("Authorization") == "Bearer "+token
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, code string, message string, userAction string) {
	writeJSON(w, status, ErrorResponse{
		Code:        code,
		Message:     message,
		Recoverable: status < http.StatusInternalServerError,
		UserAction:  userAction,
		TraceID:     NewTraceID(),
	})
}
