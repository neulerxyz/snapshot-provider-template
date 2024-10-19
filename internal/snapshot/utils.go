// internal/snapshot/utils.go

package snapshot

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// StopService tries to stop a service, and forcefully kills it if it doesn't stop within the timeout period.
func StopService(serviceName string) error {
	fmt.Printf("Attempting to stop service: %s\n", serviceName)

	// Create a context with a timeout (e.g., 10 seconds)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Run the systemctl stop command with the context
	cmd := exec.CommandContext(ctx, "sudo", "systemctl", "stop", serviceName)
	err := cmd.Run()

	// If the context timeout was exceeded, the command will return an error
	if ctx.Err() == context.DeadlineExceeded {
		fmt.Printf("Service %s did not stop within the timeout. Proceeding to force kill.\n", serviceName)
	}

	if err == nil {
		fmt.Printf("Service %s stopped successfully.\n", serviceName)
		return nil
	}

	// If stopping the service failed, retrieve the PID and forcefully kill the process
	fmt.Printf("Failed to stop service %s via systemctl. Attempting to kill the process...\n", serviceName)

	pid, err := getServicePID(serviceName)
	if err != nil {
		return fmt.Errorf("failed to get PID for service %s: %v", serviceName, err)
	}

	if pid != "" {
		killCmd := exec.Command("sudo", "kill", "-9", pid)
		err := killCmd.Run()
		if err != nil {
			return fmt.Errorf("failed to kill process %s: %v", pid, err)
		}
		fmt.Printf("Process %s for service %s killed successfully.\n", pid, serviceName)
	} else {
		return fmt.Errorf("could not find PID for service %s", serviceName)
	}

	return nil
}

func getServicePID(serviceName string) (string, error) {
	var out bytes.Buffer
	cmd := exec.Command("systemctl", "show", serviceName, "--property=MainPID")
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", err
	}

	output := out.String()
	parts := strings.Split(output, "=")
	if len(parts) > 1 {
		pid := strings.TrimSpace(parts[1])
		if pid != "0" {
			return pid, nil
		}
	}
	return "", fmt.Errorf("no valid PID found for service %s", serviceName)
}

func StartService(serviceName string) error {
	cmd := exec.Command("sudo", "systemctl", "start", serviceName)
	return cmd.Run()
}

func Sleep(seconds int) {
	time.Sleep(time.Duration(seconds) * time.Second)
}
