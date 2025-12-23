package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"backend/internal/models"
	"backend/internal/room"
	"backend/internal/storage"
	"backend/internal/utils"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

const AI_SYSTEM_PROMPT = `
	You are helping edit a Markdown document.
	You will be given:
	1. The text before the cursor
	2. The user request
	3. The text after the cursor
	Your job is to ONLY the text that should be inserted at the cursor.
	Do NOT rewrite existing text. Do NOT restate context
	Output plain Markdown without backticks.
	You are allowed to use latex for math. For both block/inline math DO NOT put 
	a new line between $$ and the math.
	CORRECT: $$x+y$$, INCORRECT: $$\nx+y\n$$
	CORRECT: $x+y$, INCORRECT: $\nx\n$
`

const CONTEXT_SIZE = 600

var openAIClient openai.Client

func buildUserPrompt(documentContent string, rawPrompt string, cursorPosition int) string {
	cursorPosition = min(cursorPosition, len(documentContent))

	start := max(cursorPosition-CONTEXT_SIZE, 0)
	end := min(cursorPosition+CONTEXT_SIZE, len(documentContent))

	before := documentContent[start:cursorPosition]
	after := documentContent[cursorPosition:end]

	prompt := fmt.Sprintf(
		"Before cursor:\n%s\n\nUser request:\n%s\n\nAfter cursor:\n%s",
		before,
		rawPrompt,
		after,
	)

	return prompt
}

func InitOpenAIClient(apiKey string) {
	openAIClient = openai.NewClient(
		option.WithAPIKey(apiKey),
	)
}

func BroadcastAIResponse(req *models.AIRequest) {
	var prompt string

	content, err := storage.GetDocumentContent(req.RoomCode, req.DocId)
	if err != nil {
		// fall back to just the given prompt, dont add any context from the doc contents
		prompt = req.Prompt
	} else {
		prompt = buildUserPrompt(content, req.Prompt, req.CursorPosition)
	}

	chatCompletion, err := openAIClient.Chat.Completions.New(
		context.TODO(),
		openai.ChatCompletionNewParams{
			Messages: []openai.ChatCompletionMessageParamUnion{
				openai.SystemMessage(AI_SYSTEM_PROMPT),
				openai.UserMessage(prompt),
			},
			Model: openai.ChatModelGPT4oMini,
		})

	if err != nil {
		log.Printf("Could not get GPT response: %v", err)
		return
	}

	op := &models.Operation{
		Type:     "insert",
		Position: req.CursorPosition,
		Text:     chatCompletion.Choices[0].Message.Content,
	}

	currentContent, err := storage.GetDocumentContent(req.RoomCode, req.DocId)
	if err != nil {
		log.Printf("error getting document %d\n", req.DocId)
		return
	}

	newContent := utils.ApplyOperation(currentContent, op)

	err = storage.UpdateDocumentContent(req.RoomCode, newContent, req.DocId)
	if err != nil {
		log.Printf("could not update document %d\n", req.DocId)
		return
	}

	resp := models.Message{
		Type:      "operation",
		Operation: op,
		UserID:    "ai",
	}

	data, err := json.Marshal(resp)
	if err != nil {
		log.Printf("could not marshal ai response operation: %v", err)
		return
	}

	rm := room.GetRoom(req.RoomCode)
	if rm == nil {
		return
	}

	room.BroadcastToEveryone(rm, req.DocId, data)
}
