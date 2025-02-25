package ui

import (
	"fmt"
	"log"
	"strings"

	"personal/tui-dev/go/ollama"
	"personal/tui-dev/go/storage"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// RunUI sets up and runs the TUI chat application.
func RunUI(client *ollama.Client, store *storage.Storage, logger *log.Logger) {
	// activeConversationID holds the current conversation. Default is "default".
	activeConversationID := "default"
	// conversationHistory holds the current session's messages.
	var conversationHistory []ollama.ChatMessage

	app := tview.NewApplication()

	// Set up the chat history view.
	chatView := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWrap(true)
	chatView.SetBorder(true).SetTitle("Chat History")

	// Set up the input field.
	inputField := tview.NewInputField().
		SetLabel("You: ").
		SetFieldWidth(0)
	inputField.SetBorder(true)

	// Layout using Flex.
	layout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(chatView, 0, 1, false).
		AddItem(inputField, 3, 0, true)

	// Global key binding:
	// F2 opens the CRUD view for the active conversation.
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		logger.Println("Global key captured:", event.Key(), event.Rune(), event.Modifiers())

		if event.Key() == tcell.KeyCtrlE {
			showCRUDView(app, store, logger, layout, activeConversationID)
			return nil
		}
		// Ctrl+L opens the conversation list view.
		if event.Key() == tcell.KeyCtrlH {
			showConversationListView(app, store, logger, layout, func(convID string) {
				// When a conversation is selected, load its messages.
				msgs, err := store.GetMessagesForConversation(convID)
				if err != nil {
					logger.Println("Error loading conversation:", err)
					return
				}
				var newHistory []ollama.ChatMessage
				chatView.Clear()
				for _, m := range msgs {
					newHistory = append(newHistory, ollama.ChatMessage{
						Role:    m.Role,
						Content: m.Content,
					})
					bubble := createBubble(m.Role, m.Content, "[blue]")
					fmt.Fprintf(chatView, "%s\n\n", bubble)
				}
				conversationHistory = newHistory
				activeConversationID = convID
				logger.Println("Loaded conversation", convID, "into current session")
				app.SetRoot(layout, true)
			})
			return nil
		}
		return event
	})

	// Also add a key capture for the input field.
	inputField.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		logger.Println("InputField key captured:", event.Key(), event.Rune(), event.Modifiers())
		if event.Key() == tcell.KeyF2 {
			logger.Println("Opening CRUD view for current conversation (Triggered by F2 in InputField)")
			showCRUDView(app, store, logger, layout, activeConversationID)
			return nil
		}
		return event
	})

	// When Enter is pressed in the input field, send the message.
	inputField.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			message := inputField.GetText()
			if message == "" {
				return
			}
			inputField.SetText("")
			userBubble := createBubble("You", message, "[green]")
			fmt.Fprintf(chatView, "%s\n\n", userBubble)
			logger.Println("User message added to UI:", message)

			conversationHistory = append(conversationHistory, ollama.ChatMessage{
				Role:    "user",
				Content: message,
			})
			logger.Println("Conversation history updated (user):", conversationHistory)
			if err := store.AddMessage(activeConversationID, "user", message); err != nil {
				logger.Println("DB Error on AddMessage (user):", err)
			} else {
				logger.Println("DB: Added user message")
			}

			go func() {
				chatReq := ollama.ChatRequest{
					Model:    "llama3.2",
					Messages: conversationHistory,
					Stream:   true,
				}
				resp, err := client.GenerateChatCompletion(chatReq)
				var apiResponse string
				if err != nil {
					apiResponse = fmt.Sprintf("Error: %v", err)
					logger.Println("API Error:", err)
				} else {
					apiResponse = resp.Message.Content
					logger.Println("API Response received:", apiResponse)
				}
				if strings.TrimSpace(apiResponse) == "" {
					logger.Println("Warning: API returned blank response for message:", message)
				}

				conversationHistory = append(conversationHistory, ollama.ChatMessage{
					Role:    "assistant",
					Content: apiResponse,
				})
				logger.Println("Conversation history updated (assistant):", conversationHistory)
				if err := store.AddMessage(activeConversationID, "assistant", apiResponse); err != nil {
					logger.Println("DB Error on AddMessage (assistant):", err)
				} else {
					logger.Println("DB: Added assistant message")
				}

				app.QueueUpdateDraw(func() {
					assistantBubble := createBubble(chatReq.Model, apiResponse, "[yellow]")
					fmt.Fprintf(chatView, "%s\n\n", assistantBubble)
				})
			}()
		}
	})

	if err := app.SetRoot(layout, true).EnableMouse(true).Run(); err != nil {
		logger.Println("Application error:", err)
		panic(err)
	}
}
