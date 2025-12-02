package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func initTemplate() error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("==========================================")
	fmt.Println("Go Project Template Initialization")
	fmt.Println("==========================================")
	fmt.Println()

	// Get current directory name as default
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting current directory: %w", err)
	}
	defaultProjectName := filepath.Base(currentDir)

	// Read go.mod to get current module path
	goModPath := "go.mod"
	currentModule, err := readModulePath(goModPath)
	if err != nil {
		return fmt.Errorf("reading go.mod: %w", err)
	}

	// Prompt for project information
	projectName := prompt(reader, fmt.Sprintf("Project name (default: %s): ", defaultProjectName), defaultProjectName)

	modulePath := prompt(reader, "Go module path (e.g., github.com/username/repo): ", "")
	for modulePath == "" {
		fmt.Println("Module path is required!")
		modulePath = prompt(reader, "Go module path (e.g., github.com/username/repo): ", "")
	}

	execName := prompt(reader, fmt.Sprintf("Executable name (default: %s): ", projectName), projectName)
	sshUser := prompt(reader, "SSH user for deployment (default: ansible): ", "ansible")
	deployTargetIP := prompt(reader, "Deployment target IP (press Enter to skip): ", "")

	// Show summary
	fmt.Println()
	fmt.Println("==========================================")
	fmt.Println("Configuration Summary:")
	fmt.Println("==========================================")
	fmt.Printf("Project name:       %s\n", projectName)
	fmt.Printf("Module path:        %s\n", modulePath)
	fmt.Printf("Executable name:    %s\n", execName)
	fmt.Printf("SSH user:           %s\n", sshUser)
	if deployTargetIP != "" {
		fmt.Printf("Deploy target IP:   %s\n", deployTargetIP)
	} else {
		fmt.Printf("Deploy target IP:   <not set>\n")
	}
	fmt.Println()

	confirm := prompt(reader, "Proceed with these settings? (y/n): ", "")
	if strings.ToLower(confirm) != "y" {
		fmt.Println("Initialization cancelled.")
		return nil
	}

	fmt.Println()
	fmt.Println("Initializing project...")

	// Update go.mod
	if err := updateGoMod(goModPath, currentModule, modulePath); err != nil {
		return fmt.Errorf("updating go.mod: %w", err)
	}
	fmt.Println("✓ Updated go.mod")

	// Update Makefile
	if err := updateMakefile("Makefile", projectName, execName, sshUser, deployTargetIP); err != nil {
		return fmt.Errorf("updating Makefile: %w", err)
	}
	fmt.Println("✓ Updated Makefile")

	// Update systemd files
	if err := updateSystemdFiles("remote", projectName); err != nil {
		fmt.Printf("Warning: %v\n", err)
	}

	fmt.Println()
	fmt.Println("==========================================")
	fmt.Println("✓ Initialization complete!")
	fmt.Println("==========================================")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  1. Review the updated files (go.mod, Makefile)")
	fmt.Println("  2. Run 'make doctor' to verify your development environment")
	fmt.Println("  3. Run 'make help' to see available commands")
	fmt.Println()
	fmt.Println("Happy coding! 🚀")

	return nil
}

func prompt(reader *bufio.Reader, message, defaultValue string) string {
	fmt.Print(message)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if input == "" {
		return defaultValue
	}
	return input
}

func readModulePath(filename string) (string, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}

	re := regexp.MustCompile(`(?m)^module\s+(.+)$`)
	matches := re.FindSubmatch(data)
	if len(matches) < 2 {
		return "", fmt.Errorf("module path not found in go.mod")
	}

	return string(matches[1]), nil
}

func updateGoMod(filename, oldModule, newModule string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	content := string(data)
	content = strings.Replace(content, "module "+oldModule, "module "+newModule, 1)

	return os.WriteFile(filename, []byte(content), 0644)
}

func updateMakefile(filename, projectName, execName, sshUser, deployTargetIP string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	content := string(data)

	// Replace PROJECT_NAME
	re := regexp.MustCompile(`PROJECT_NAME := "[^"]*"`)
	content = re.ReplaceAllString(content, fmt.Sprintf(`PROJECT_NAME := "%s"`, projectName))

	// Replace EXEC_NAME
	re = regexp.MustCompile(`EXEC_NAME := \w+`)
	content = re.ReplaceAllString(content, fmt.Sprintf(`EXEC_NAME := %s`, execName))

	// Replace SSH_USER
	re = regexp.MustCompile(`SSH_USER := \w+`)
	content = re.ReplaceAllString(content, fmt.Sprintf(`SSH_USER := %s`, sshUser))

	// Replace DEPLOY_TARGET_IP if provided
	if deployTargetIP != "" {
		re = regexp.MustCompile(`DEPLOY_TARGET_IP := [\d.]+`)
		content = re.ReplaceAllString(content, fmt.Sprintf(`DEPLOY_TARGET_IP := %s`, deployTargetIP))
	}

	return os.WriteFile(filename, []byte(content), 0644)
}

func updateSystemdFiles(dirname, projectName string) error {
	entries, err := os.ReadDir(dirname)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // remote directory doesn't exist, skip
		}
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filename := entry.Name()
		if !strings.HasSuffix(filename, ".service") && !strings.HasSuffix(filename, ".timer") {
			continue
		}

		oldPath := filepath.Join(dirname, filename)
		data, err := os.ReadFile(oldPath)
		if err != nil {
			return err
		}

		content := string(data)
		content = strings.ReplaceAll(content, "go-start", projectName)

		// Determine new filename
		var ext string
		if strings.HasSuffix(filename, ".service") {
			ext = ".service"
		} else {
			ext = ".timer"
		}
		newFilename := projectName + ext
		newPath := filepath.Join(dirname, newFilename)

		// Write to new filename
		if err := os.WriteFile(newPath, []byte(content), 0644); err != nil {
			return err
		}

		// Remove old file if the name changed
		if oldPath != newPath {
			if err := os.Remove(oldPath); err != nil {
				return err
			}
			fmt.Printf("✓ Renamed %s to %s\n", oldPath, newPath)
		} else {
			fmt.Printf("✓ Updated %s\n", newPath)
		}
	}

	return nil
}
