package main

import (
    "bufio"
    "bytes"
    "fmt"
    "log"
    "net/http"
    "net/url"
    "os"
    "os/exec"
    "regexp"
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

func runCommand(name string, args ...string) ([]byte, error) {
    cmd := exec.Command(name, args...)
    var out bytes.Buffer
    cmd.Stdout = &out
    cmd.Stderr = &out
    err := cmd.Run()
    return out.Bytes(), err
}

func executeCommands() ([]byte, error) {
    commands := [][]string{
        {"sudo", "systemctl", "stop", "stationd"},
        {"cd", "tracks"},
        {"go", "run", "cmd/main.go", "rollback"},
        {"go", "run", "cmd/main.go", "rollback"},
        {"go", "run", "cmd/main.go", "rollback"},
        {"sudo", "systemctl", "restart", "stationd"},
        {"sudo", "journalctl", "-u", "stationd", "-f", "--no-hostname", "-o", "cat"},
    }

    var finalOutput bytes.Buffer

    for _, cmd := range commands {
        if len(cmd) > 1 && cmd[0] == "cd" {
            if err := os.Chdir(cmd[1]); err != nil {
                return nil, fmt.Errorf("failed to change directory: %v", err)
            }
        } else {
            output, err := runCommand(cmd[0], cmd[1:]...)
            finalOutput.Write(output)
            if err != nil {
                return finalOutput.Bytes(), fmt.Errorf("failed to execute command %s: %v", cmd[0], err)
            }
        }
    }
    return finalOutput.Bytes(), nil
}

func parsePodNumber(logData []byte) string {
    re := regexp.MustCompile(`Pod Number= (\d+)`)
    match := re.FindSubmatch(logData)
    if match != nil {
        return string(match[1])
    }
    return "N/A"
}

func main() {
    botToken := getInput("Enter your Telegram bot token: ")
    chatID := getInput("Enter your Telegram chat ID: ")

    loc, err := time.LoadLocation("Asia/Tehran")
    if err != nil {
        log.Fatalf("Failed to load location: %v", err)
    }

    commandTicker := time.NewTicker(20 * time.Minute)
    logTicker := time.NewTicker(10 * time.Minute)
    defer commandTicker.Stop()
    defer logTicker.Stop()

    var lastCommandOutput []byte
    var lastCommandError error
    startTime := time.Now().In(loc).Format(time.RFC1123)

    for {
        select {
        case <-commandTicker.C:
            lastCommandOutput, lastCommandError = executeCommands()
        case <-logTicker.C:
            currentTime := time.Now().In(loc).Format(time.RFC1123)
            if lastCommandError != nil {
                errMsg := fmt.Sprintf("Time: %s\nStatus: Error\nDetails: %v", currentTime, lastCommandError)
                log.Println(errMsg)
                sendTelegramMessage(botToken, chatID, errMsg)
            } else {
                podNumber := parsePodNumber(lastCommandOutput)
                successMsg := fmt.Sprintf("Time: %s\nStatus: Success\nDetails: All commands executed successfully since %s.\nPod Number: %s", currentTime, startTime, podNumber)
                log.Println(successMsg)
                sendTelegramMessage(botToken, chatID, successMsg)
            }
        }
    }
}
