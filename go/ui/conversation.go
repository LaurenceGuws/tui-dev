package ui

import (
	"log"

	"personal/tui-dev/go/storage"

	"github.com/rivo/tview"
)

// showConversationListView displays a list of conversation IDs.
// When a conversation is selected, onSelect is called with the chosen ID.
func showConversationListView(app *tview.Application, store *storage.Storage, logger *log.Logger, prevRoot tview.Primitive, onSelect func(convID string)) {
	list := tview.NewList()
	convs, err := store.GetConversations()
	if err != nil {
		logger.Println("Error fetching conversations:", err)
		return
	}
	for _, convID := range convs {
		list.AddItem(convID, "", 0, func(id string) func() {
			return func() {
				onSelect(id)
			}
		}(convID))
		logger.Println("Conversation list item added:", convID)
	}
	list.SetBorder(true).SetTitle("Conversations")
	list.SetDoneFunc(func() {
		app.SetRoot(prevRoot, true)
	})
	app.SetRoot(list, true)
}
