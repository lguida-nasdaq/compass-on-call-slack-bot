package slackmsg

import (
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/metriodev/pompiers/internal/domain"
	"github.com/slack-go/slack"
)

func ToSlackMessage(s domain.CurrentOnCallSchedule) ([]byte, error) {
	var elements []slack.RichTextElement
	for _, schedule := range s.Schedules {
		elements = append(elements, scheduleToBlock(schedule))
	}

	message := slack.NewBlockMessage(
		slack.NewRichTextBlock(
			"rich_text",
			slack.NewRichTextList(slack.RTEListBullet, 0, elements...),
		),
	)

	// response := map[string]interface{}{
	// 	"response_type": slack.ResponseTypeEphemeral,
	// 	"blocks":        message.Blocks,
	// }

	// message.Type = slack.ResponseTypeEphemeral

	payload, err := json.Marshal(&message)
	if err != nil {
		slog.Error("Error marshalling slack message", "error", err)
		return nil, fmt.Errorf("failed to marshal slack message: %w", err)
	}

	return payload, nil
}

func scheduleToBlock(schedule domain.Schedule) *slack.RichTextSection {
	return slack.NewRichTextSection(
		slack.NewRichTextSectionTextElement(
			fmt.Sprintf("%s: ", schedule.Name),
			&slack.RichTextSectionTextStyle{Bold: true},
		),
		slack.NewRichTextSectionTextElement(
			schedule.OnCallUsers,
			&slack.RichTextSectionTextStyle{},
		),
	)
}
