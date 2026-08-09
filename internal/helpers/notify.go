package helpers

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/Francesco99975/casaintake/cmd/boot"
)

func Notify(topic string, message string) {
	resp, err := http.Post(fmt.Sprintf("%s/%s", boot.Environment.NTFY, topic), "text/plain",
		strings.NewReader(message))

	if err != nil {
		slog.Warn("Failed to send notification", slog.Any("error", err))
	}

	if resp != nil {
		slog.Debug("Notification sent with status code", slog.Int("status", resp.StatusCode))
		defer func() { _ = resp.Body.Close() }()
	}
}
