package ui

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"personal/tui-dev/go/storage"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// showCRUDView displays a list of messages (CRUD operations) for the active conversation.
// It accepts prevRoot (the main chat layout) and activeConversationID to load only those messages.
func showCRUDView(app *tview.Application, store *storage.Storage, logger *log.Logger, prevRoot tview.Primitive, activeConversationID string) {
	logger.Println("Opening CRUD view for conversation:", activeConversationID)
	list := tview.NewList()
	msgs, err := store.GetMessagesForConversation(activeConversationID)
	if err != nil {
		logger.Println("Error fetching messages for CRUD view:", err)
		return
	}
	for _, msg := range msgs {
		title := fmt.Sprintf("%d [%s]", msg.ID, msg.Role)
		list.AddItem(title, msg.Content, 0, nil)
		logger.Println("CRUD view item added:", title, msg.Content)
	}
	list.SetBorder(true).SetTitle("Messages (CRUD) - Conversation: " + activeConversationID)
	list.SetDoneFunc(func() {
		logger.Println("Exiting CRUD view; restoring previous layout")
		app.SetRoot(prevRoot, true)
	})
	// Delete on Enter.
	list.SetSelectedFunc(func(index int, mainText string, secondaryText string, shortcut rune) {
		itemIDStr := strings.Split(mainText, " ")[0]
		id, err := strconv.ParseInt(itemIDStr, 10, 64)
		if err == nil {
			if err := store.DeleteMessage(id); err != nil {
				logger.Println("Error deleting message with ID", id, ":", err)
			} else {
				logger.Println("Deleted message with ID:", id)
				showCRUDView(app, store, logger, prevRoot, activeConversationID)
			}
		} else {
			logger.Println("Error parsing item ID:", itemIDStr, err)
		}
	})
	// Inside the CRUD view, if a key event with rune value 5 is detected, trigger editing.
	list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyRune && event.Rune() == 5 {
			index := list.GetCurrentItem()
			mainText, secondaryText := list.GetItemText(index)
			itemIDStr := strings.Split(mainText, " ")[0]
			id, err := strconv.ParseInt(itemIDStr, 10, 64)
			if err != nil {
				logger.Println("Error parsing item ID for edit:", itemIDStr, err)
				return event
			}
			logger.Println("Editing message with ID:", id)
			// For this minimal change, simply log that edit was triggered.
			// (Insert your edit logic here if desired.)
			logger.Println("Edit triggered for message ID:", id, "with current text:", secondaryText)
			return nil
		}
		return event
	})
	app.SetRoot(list, true)
}
