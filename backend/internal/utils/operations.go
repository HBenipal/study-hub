package utils

import (
	"log"

	"backend/internal/models"
)

// applies operation (insert/delete) to given content
func ApplyOperation(content string, operation *models.Operation) string {
	switch operation.Type {
	case "insert":
		if operation.Position > len(content) {
			operation.Position = len(content)
		}

		return content[:operation.Position] + operation.Text + content[operation.Position:]
	case "delete":
		if operation.Position > len(content) {
			operation.Position = len(content)
		}

		end := min(operation.Position+operation.Length, len(content))

		return content[:operation.Position] + content[end:]
	default:
		log.Printf("operation not supported: %s\n", operation.Type)
	}

	return content
}
