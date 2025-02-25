package ollama

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

// Client is used to interact with the Ollama API.
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
	Logger     *log.Logger
}

// NewClient returns a new Ollama API client with logging enabled.
func NewClient(baseURL string) *Client {
	logFile, err := os.OpenFile("log.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	logger := log.New(logFile, "OLLAMA: ", log.LstdFlags)
	logger.Println("Ollama client initialized. API Base URL:", baseURL)
	return &Client{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		Logger: logger,
	}
}

// ChatMessage represents a single message in a chat conversation.
type ChatMessage struct {
	Role      string      `json:"role"`
	Content   string      `json:"content"`
	Images    []string    `json:"images,omitempty"`
	ToolCalls interface{} `json:"tool_calls,omitempty"`
}

// ChatRequest is the payload for the /api/chat endpoint.
type ChatRequest struct {
	Model     string                 `json:"model"`
	Messages  []ChatMessage          `json:"messages"`
	Options   map[string]interface{} `json:"options,omitempty"`
	Format    interface{}            `json:"format,omitempty"`
	Stream    bool                   `json:"stream,omitempty"`
	KeepAlive int64                  `json:"keep_alive,omitempty"`
}

// ChatResponse is a simplified response from the /api/chat endpoint.
type ChatResponse struct {
	Model     string      `json:"model"`
	CreatedAt string      `json:"created_at"`
	Message   ChatMessage `json:"message"`
	Done      bool        `json:"done"`
}

// GenerateChatCompletion sends a chat request to the Ollama API and
// handles a streaming response by accumulating the message content.
func (c *Client) GenerateChatCompletion(reqData ChatRequest) (*ChatResponse, error) {
	url := c.BaseURL + "/api/chat"

	// Ensure options exist and set default max_tokens.
	if reqData.Options == nil {
		reqData.Options = make(map[string]interface{})
	}
	reqData.Options["max_tokens"] = 256
	reqData.Stream = true // enable streaming

	// Log the full request payload.
	requestJSON, _ := json.MarshalIndent(reqData, "", "  ")
	c.Logger.Println("Sending chat request to Ollama API:", string(requestJSON))

	body, err := json.Marshal(reqData)
	if err != nil {
		c.Logger.Println("Error marshaling chat request:", err)
		return nil, err
	}

	resp, err := c.HTTPClient.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		c.Logger.Println("HTTP chat request failed:", err)
		return nil, err
	}
	defer resp.Body.Close()

	c.Logger.Println("Ollama API chat response status:", resp.Status)

	var fullContent string
	decoder := json.NewDecoder(resp.Body)
	for {
		var partial ChatResponse
		if err := decoder.Decode(&partial); err != nil {
			if err == io.EOF {
				break
			}
			c.Logger.Println("Error decoding chat streaming response:", err)
			return nil, err
		}
		c.Logger.Println("Received chat chunk:", partial.Message.Content)
		fullContent += partial.Message.Content
		if partial.Done {
			break
		}
	}

	if fullContent == "" {
		errMsg := "received empty chat response from Ollama API"
		c.Logger.Println("WARNING:", errMsg)
		return nil, fmt.Errorf(errMsg)
	}

	finalResponse := &ChatResponse{
		Model:     reqData.Model,
		CreatedAt: time.Now().Format(time.RFC3339),
		Message: ChatMessage{
			Role:    "assistant",
			Content: fullContent,
		},
		Done: true,
	}
	c.Logger.Println("Final aggregated chat response:", finalResponse.Message.Content)
	return finalResponse, nil
}
