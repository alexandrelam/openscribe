// +build ignore

package main

import (
	"fmt"
	"os"
	"time"

	"github.com/alexandrelam/openscribe/internal/keyboard"
)

func main() {
	fmt.Println("Keyboard Simulation Test")
	fmt.Println("=========================")
	fmt.Println()

	// Create keyboard instance
	kb, err := keyboard.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to create keyboard: %v\n", err)
		os.Exit(1)
	}
	defer kb.Close()

	// Check permissions
	fmt.Println("Checking accessibility permissions...")
	if err := kb.CheckPermissions(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: Accessibility permissions not granted.\n\n")
		fmt.Fprintf(os.Stderr, "Please grant permissions in:\n")
		fmt.Fprintf(os.Stderr, "  System Preferences > Security & Privacy > Privacy > Accessibility\n\n")
		fmt.Fprintf(os.Stderr, "Add 'Terminal' (or your terminal app) to the list.\n\n")
		os.Exit(1)
	}

	fmt.Println("✓ Accessibility permissions granted")
	fmt.Println()
	fmt.Println("Test will type text in 5 seconds...")
	fmt.Println("Please click in a text editor (TextEdit, Notes, etc.) now!")
	fmt.Println()

	// Give user time to switch to another app
	for i := 5; i > 0; i-- {
		fmt.Printf("%d...\n", i)
		time.Sleep(1 * time.Second)
	}

	testText := "Hello from OpenScribe! This is a test of keyboard simulation."
	fmt.Printf("Typing: \"%s\"\n", testText)

	if err := kb.TypeText(testText); err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to type text: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✓ Text typed successfully!")
}
