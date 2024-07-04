package main

import (
    "bufio"
    "fmt"
    "log"
    "net/http"
    "net/url"
    "os"
    "os/exec"
    "time"
)

func getInput(prompt string) string {
    fmt.Print(prompt)
    scanner := bufio.NewScanner(os.Stdin)
    scanner.Scan()
    return scanner.Text()
}

func sendTelegramMessage(botToken, chatID, message string) error {
    apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)
    values := url.Values{
        "chat_id": {chatID},
        "text":    {message},
    }

    resp, err := http.PostForm(apiURL, values)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("failed to send message: %s", resp.Status)
    }
    return nil
}

func runCommand(name string, args ...string) error {
    cmd := exec.Command(name, args...)
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    return cmd.Run()
}

func executeCommands() error {
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
                return fmt.Errorf("failed to change directory: %v", err)
            }
        } else {
            if err := runCommand(cmd[0], cmd[1:]...); err != nil {
                return fmt.Errorf("failed to execute command %s: %v", cmd[0], err)
            }
        }
    }
    return nil
}

func main() {
    botToken := getInput("Enter your Telegram bot token: ")
    chatID := getInput("Enter your Telegram chat ID: ")

    ticker := time.NewTicker(40 * time.Minute)
    defer ticker.Stop()

    for {
        err := executeCommands()
        if err != nil {
            errMsg := fmt.Sprintf("Error running commands: %v", err)
            log.Println(errMsg)
            sendTelegramMessage(botToken, chatID, errMsg)
        } else {
            successMsg := "All commands executed successfully!"
            log.Println(successMsg)
            sendTelegramMessage(botToken, chatID, successMsg)
        }
        <-ticker.C
    }
}
