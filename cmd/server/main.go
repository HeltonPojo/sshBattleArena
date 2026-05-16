package main

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/bubbletea"

	"github.com/muesli/termenv"

	"github.com/HeltoPojo/sshBattleArena/internal/game"
	"github.com/HeltoPojo/sshBattleArena/internal/server"
	"github.com/HeltoPojo/sshBattleArena/internal/tui"
)

const (
	host        = "0.0.0.0"
	port        = 2222
	hostKeyFile = ".ssh_host_key"
)

func main() {
	if err := ensureHostKey(hostKeyFile); err != nil {
		log.Fatalf("host key: %v", err)
	}

	reg := server.NewRegistry()
	broadcast := server.NewBroadcaster(reg)
	gl := game.NewGameLoop(broadcast)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go gl.Run(ctx)

	s, err := wish.NewServer(
		wish.WithAddress(fmt.Sprintf("%s:%d", host, port)),
		wish.WithHostKeyPath(hostKeyFile),
		wish.WithMiddleware(
			bubbletea.MiddlewareWithProgramHandler(
				func(s ssh.Session) *tea.Program {
					sessionID := s.Context().SessionID()
					gl.AddPlayer(sessionID)

					model := tui.NewModel(sessionID, gl.InputCh())
					opts := append(bubbletea.MakeOptions(s), tea.WithAltScreen())
					p := tea.NewProgram(model, opts...)

					reg.SetProgram(sessionID, p)

					go func() {
						<-s.Context().Done()
						gl.RemovePlayer(sessionID)
						reg.Remove(sessionID)
						gl.ResetIfEmpty()
					}()

					return p
				}, termenv.ANSI256,
			),
			// Connection limiter — runs BEFORE bubbletea (middleware order is reversed in Wish).
			func(next ssh.Handler) ssh.Handler {
				return func(s ssh.Session) {
					sessionID := s.Context().SessionID()
					if !reg.TryRegister(sessionID) {
						fmt.Fprintln(s, "Arena is full! Try again later.")
						return
					}
					next(s)
				}
			},
		),
	)
	if err != nil {
		log.Fatalf("could not create server: %v", err)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)

	log.Printf("SSH server listening on %s:%d", host, port)
	go func() {
		if err := s.ListenAndServe(); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-done
	log.Println("shutting down...")
	cancel()
	if err := s.Shutdown(context.Background()); err != nil {
		log.Fatalf("shutdown error: %v", err)
	}
}

// ensureHostKey generates an ed25519 host key if one does not already exist.
func ensureHostKey(path string) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	}

	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return fmt.Errorf("generate key: %w", err)
	}

	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return fmt.Errorf("marshal key: %w", err)
	}

	pemBlock := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privBytes,
	})

	if err := os.WriteFile(path, pemBlock, 0600); err != nil {
		return fmt.Errorf("write key: %w", err)
	}

	log.Printf("generated new host key: %s", path)
	return nil
}
