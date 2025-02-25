package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"time"
)

const (
	filename       = "counter.txt"
	heartbeatFile  = "heartbeat.txt"
	lockFile       = "primary.lock"
	checkInterval  = 1 * time.Second
	updateInterval = 500 * time.Millisecond
	missedLimit    = 3 // how many missed heartbeats before taking over
)

func main() {
	// Try to become primary by creating a lock file
	isPrimary := tryBecomePrimary()

	if isPrimary {
		// This process will be the primary
		fmt.Println("Starting as PRIMARY")

		// Create a backup
		spawnBackup()

		// Run as primary
		runPrimary()
	} else {
		// This process will be the backup
		fmt.Println("Starting as BACKUP")
		runBackup()
	}
}

func tryBecomePrimary() bool {
	// First check if the heartbeat file is recent
	if primaryIsAlive() {
		return false
	}

	// Try to create the lock file
	lockFile, err := os.OpenFile(lockFile, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		// Failed to create lock file, likely because it already exists
		return false
	}

	// Write our PID to the lock file
	fmt.Fprintf(lockFile, "%d", os.Getpid())
	lockFile.Close()

	// Double check that we really are the primary now
	// This helps prevent race conditions where two processes both think they're primary
	time.Sleep(100 * time.Millisecond)

	// Create an initial heartbeat
	updateHeartbeat()

	return true
}

func primaryIsAlive() bool {
	// Check if heartbeat file exists and is recent
	info, err := os.Stat(heartbeatFile)
	if err != nil {
		// File doesn't exist or can't be accessed
		return false
	}

	// Check timestamp - if it's too old, the primary is considered dead
	timeSinceUpdate := time.Since(info.ModTime())
	return timeSinceUpdate < checkInterval*missedLimit
}

func spawnBackup() {
	// Get the current working directory
	dir, err := os.Getwd()
	if err != nil {
		fmt.Printf("Failed to get working directory: %v\n", err)
		return
	}

	// Get the absolute path to the current executable
	execPath, err := os.Executable()
	if err != nil {
		fmt.Printf("Failed to get executable path: %v\n", err)
		return
	}

	var cmd *exec.Cmd

	// Platform-specific terminal commands
	switch runtime.GOOS {
	case "darwin":
		// macOS
		cmd = exec.Command("osascript", "-e", `tell app "Terminal" to do script "cd '`+dir+`' && '`+execPath+`'"`)
	case "linux":
		// Linux - try gnome-terminal first, fallback to xterm
		cmd = exec.Command("gnome-terminal", "--", "bash", "-c", "cd '"+dir+"' && '"+execPath+"'")
		err = cmd.Run()
		if err != nil {
			// Try xterm if gnome-terminal fails
			cmd = exec.Command("xterm", "-e", "cd '"+dir+"' && '"+execPath+"'")
		}
	case "windows":
		// Windows
		cmd = exec.Command("cmd", "/C", "start", "cmd", "/k", "cd "+dir+" && "+execPath)
	default:
		fmt.Printf("Unsupported operating system: %s\n", runtime.GOOS)
		return
	}

	err = cmd.Run()
	if err != nil {
		fmt.Printf("Failed to spawn backup: %v\n", err)

		// Fallback to running with go run if this is a development environment
		sourceFile := filepath.Join(dir, "main.go")
		if _, err := os.Stat(sourceFile); err == nil {
			switch runtime.GOOS {
			case "darwin":
				cmd = exec.Command("osascript", "-e", `tell app "Terminal" to do script "cd '`+dir+`' && go run '`+sourceFile+`'"`)
			case "linux":
				cmd = exec.Command("gnome-terminal", "--", "bash", "-c", "cd '"+dir+"' && go run '"+sourceFile+"'")
			case "windows":
				cmd = exec.Command("cmd", "/C", "start", "cmd", "/k", "cd "+dir+" && go run "+sourceFile)
			}

			err = cmd.Run()
			if err != nil {
				fmt.Printf("Failed to spawn backup with go run: %v\n", err)
			}
		}
	}
}

func readCounter() int {
	data, err := os.ReadFile(filename)
	if err != nil {
		// If file doesn't exist or can't be read, start from 0
		return 0
	}
	counter, err := strconv.Atoi(string(data))
	if err != nil {
		return 0
	}
	return counter
}

func saveCounter(counter int) {
	// Save the current counter value to file
	err := os.WriteFile(filename, []byte(strconv.Itoa(counter)), 0644)
	if err != nil {
		fmt.Printf("Error saving counter: %v\n", err)
	}
}

func updateHeartbeat() {
	// Simply touch the heartbeat file to update its timestamp
	currentTime := []byte(time.Now().String())
	err := os.WriteFile(heartbeatFile, currentTime, 0644)
	if err != nil {
		fmt.Printf("Error updating heartbeat: %v\n", err)
	}
}

func runPrimary() {
	counter := readCounter()

	// Set up cleanup when primary exits
	defer func() {
		// Remove the lock file when primary exits
		os.Remove(lockFile)
	}()

	// Main loop for the primary
	for {
		counter++
		fmt.Printf("Counter: %d\n", counter)

		// Update the state - save counter to file
		saveCounter(counter)

		// Send heartbeat
		updateHeartbeat()

		// Wait before incrementing again
		time.Sleep(updateInterval)
	}
}

func runBackup() {
	missedHeartbeats := 0

	// Monitor the primary
	for {
		if !primaryIsAlive() {
			missedHeartbeats++
			fmt.Printf("Missed heartbeat: %d/%d\n", missedHeartbeats, missedLimit)

			if missedHeartbeats >= missedLimit {
				fmt.Println("PRIMARY is dead! Taking over...")

				// Remove any stale lock file
				os.Remove(lockFile)

				// Try to become the primary
				if tryBecomePrimary() {
					// Spawn a new backup
					spawnBackup()

					// Start operating as the primary
					runPrimary()
					return
				}
			}
		} else {
			// Reset counter if we detect a heartbeat
			if missedHeartbeats > 0 {
				fmt.Println("Primary is alive again!")
				missedHeartbeats = 0
			}
		}

		time.Sleep(checkInterval)
	}
}
