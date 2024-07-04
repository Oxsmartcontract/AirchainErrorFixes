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

func checkAndInstallGo() {
    _, err := exec.LookPath("go")
    if err != nil {
        log.Println("Go not found, installing Go 1.21.6...")

        version := "1.21.6"
        arch := "amd64"
        goURL := "https://golang.org/dl/go" + version + ".linux-" + arch + ".tar.gz"
        goTar := "go" + version + ".linux-" + arch + ".tar.gz"

        err = runCommand("curl", "-O", "-L", goURL)
        if err != nil {
            log.Fatalf("failed to download Go: %v", err)
        }

        err = runCommand("tar", "-xf", goTar)
        if err != nil {
            log.Fatalf("failed to extract Go tarball: %v", err)
        }

        err = runCommand("sudo", "rm", "-rf", "/usr/local/go")
        if err != nil {
            log.Fatalf("failed to remove existing Go installation: %v", err)
        }

        err = runCommand("sudo", "mv", "-v", "go", "/usr/local")
        if err != nil {
            log.Fatalf("failed to move Go to /usr/local: %v", err)
        }

        os.Setenv("GOPATH", os.Getenv("HOME")+"/go")
        os.Setenv("PATH", os.Getenv("PATH")+":/usr/local/go/bin:"+os.Getenv("GOPATH")+"/bin")

        err = runCommand("source", "~/.bash_profile")
        if err != nil {
            log.Printf("failed to source ~/.bash_profile: %v", err)
        }

        err = runCommand("go", "version")
        if err != nil {
            log.Fatalf("failed to verify Go installation: %v", err)
        }

        log.Println("Go 1.21.6 installed successfully.")
    } else {
        log.Println("Go is already installed.")
    }
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
    checkAndInstallGo()

    ticker := time.NewTicker(40 * time.Minute)
    defer ticker.Stop()

    for {
        executeCommands()
        <-ticker.C
    }
}
