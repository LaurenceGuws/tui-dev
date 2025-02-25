package main

import (
	"log"
	"os"

	"personal/tui-dev/go/ollama"  // Adjust this to match your module name
	"personal/tui-dev/go/storage" // Adjust the import path accordingly
	"personal/tui-dev/go/ui"      // Adjust the import path accordingly
)

func main() {
	// Create a logfile.
	logFile, err := os.OpenFile("log.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	defer logFile.Close()
	logger := log.New(logFile, "", log.LstdFlags)
	logger.Println("Application started.")

	// Create a new Ollama client.
	client := ollama.NewClient("http://localhost:11434")

	// Initialize SQLite storage for chat history.
	store, err := storage.NewStorage("chat_history.db", logger)
	if err != nil {
		logger.Println("Error initializing storage:", err)
		panic(err)
	}

	// Run the TUI.
	ui.RunUI(client, store, logger)
}
