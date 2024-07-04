package main

import (
    "log"
    "os"
    "os/exec"
    "time"
)

func runCommand(name string, args ...string) error {
    cmd := exec.Command(name, args...)
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    return cmd.Run()
}

func executeCommands() {
    commands := [][]string{
        {"sudo", "systemctl", "stop", "stationd"},
        {"cd", "tracks"},
        {"go", "run", "cmd/main.go", "rollback"},
        {"go", "run", "cmd/main.go", "rollback"},
        {"go", "run", "cmd/main.go", "rollback"},
        {"sudo", "systemctl", "restart", "stationd"},
        {"sudo", "journalctl", "-u", "stationd", "-f", "--no-hostname", "-o", "cat"},
    }

    for _, cmd := range commands {
        if len(cmd) > 1 && cmd[0] == "cd" {
            if err := os.Chdir(cmd[1]); err != nil {
                log.Fatalf("failed to change directory: %v", err)
            }
        } else {
            if err := runCommand(cmd[0], cmd[1:]...); err != nil {
                log.Printf("failed to execute command: %v", err)
            }
        }
    }
}

func main() {
    ticker := time.NewTicker(40 * time.Minute)
    defer ticker.Stop()

    for {
        executeCommands()
        <-ticker.C
    }
}
