package disgoplus

import (
	"fmt"

	"github.com/disgoorg/disgo/discord"
)

// GetInteractionName returns a slash-command path string like "settings/show".
// The default delimiter is "/"; pass a custom one as the second argument.
func GetInteractionName(
	data discord.SlashCommandInteractionData,
	delimiter ...string,
) string {
	sep := "/"
	if len(delimiter) > 0 {
		sep = delimiter[0]
	}

	name := data.CommandName()
	if data.SubCommandGroupName != nil {
		name = fmt.Sprintf("%s%s%s", name, sep, *data.SubCommandGroupName)
	}

	if data.SubCommandName != nil {
		name = fmt.Sprintf("%s%s%s", name, sep, *data.SubCommandName)
	}

	return name
}

// ParseModalData extracts text-input values from a modal submit into a
// map[customID → value].
func ParseModalData(data discord.ModalSubmitInteractionData) map[string]string {
	result := make(map[string]string)

	for component := range data.AllComponents() {
		if ti, ok := component.(discord.TextInputComponent); ok {
			result[ti.CustomID] = ti.Value
		}
	}

	return result
}
