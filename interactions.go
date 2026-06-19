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

// ParseModalAttachments extracts file-upload attachments from a modal submit
// into a map[customID → attachments].
func ParseModalAttachments(data discord.ModalSubmitInteractionData) map[string][]discord.Attachment {
	result := make(map[string][]discord.Attachment)

	for component := range data.AllComponents() {
		fu, ok := component.(discord.FileUploadComponent)
		if !ok {
			continue
		}

		attachments := make([]discord.Attachment, 0, len(fu.Values))
		for _, id := range fu.Values {
			if a, ok := data.Resolved.Attachments[id]; ok {
				attachments = append(attachments, a)
			}
		}

		result[fu.CustomID] = attachments
	}

	return result
}

// ParseModalStringValues extracts string-select menu selections from a modal
// submit into a map[customID → selected values].
func ParseModalStringValues(data discord.ModalSubmitInteractionData) map[string][]string {
	result := make(map[string][]string)

	for component := range data.AllComponents() {
		sm, ok := component.(discord.StringSelectMenuComponent)
		if !ok {
			continue
		}

		result[sm.CustomID] = sm.Values
	}

	return result
}

// ParseModalUsers extracts user-select menu selections from a modal submit
// into a map[customID → resolved users].
func ParseModalUsers(data discord.ModalSubmitInteractionData) map[string][]discord.User {
	result := make(map[string][]discord.User)

	for component := range data.AllComponents() {
		sm, ok := component.(discord.UserSelectMenuComponent)
		if !ok {
			continue
		}

		users := make([]discord.User, 0, len(sm.Values))
		for _, id := range sm.Values {
			if u, ok := data.Resolved.Users[id]; ok {
				users = append(users, u)
			}
		}

		result[sm.CustomID] = users
	}

	return result
}

// ParseModalRoles extracts role-select menu selections from a modal submit
// into a map[customID → resolved roles].
func ParseModalRoles(data discord.ModalSubmitInteractionData) map[string][]discord.Role {
	result := make(map[string][]discord.Role)

	for component := range data.AllComponents() {
		sm, ok := component.(discord.RoleSelectMenuComponent)
		if !ok {
			continue
		}

		roles := make([]discord.Role, 0, len(sm.Values))
		for _, id := range sm.Values {
			if r, ok := data.Resolved.Roles[id]; ok {
				roles = append(roles, r)
			}
		}

		result[sm.CustomID] = roles
	}

	return result
}

// ParseModalChannels extracts channel-select menu selections from a modal
// submit into a map[customID → resolved channels].
func ParseModalChannels(data discord.ModalSubmitInteractionData) map[string][]discord.ResolvedChannel {
	result := make(map[string][]discord.ResolvedChannel)

	for component := range data.AllComponents() {
		sm, ok := component.(discord.ChannelSelectMenuComponent)
		if !ok {
			continue
		}

		channels := make([]discord.ResolvedChannel, 0, len(sm.Values))
		for _, id := range sm.Values {
			if ch, ok := data.Resolved.Channels[id]; ok {
				channels = append(channels, ch)
			}
		}

		result[sm.CustomID] = channels
	}

	return result
}

// ParseModalMentionables extracts mentionable-select menu selections from a
// modal submit into a map[customID → resolved mentionables]. Each value is
// either a discord.User or a discord.Role.
func ParseModalMentionables(data discord.ModalSubmitInteractionData) map[string][]discord.Mentionable {
	result := make(map[string][]discord.Mentionable)

	for component := range data.AllComponents() {
		sm, ok := component.(discord.MentionableSelectMenuComponent)
		if !ok {
			continue
		}

		mentionables := make([]discord.Mentionable, 0, len(sm.Values))
		for _, id := range sm.Values {
			if u, ok := data.Resolved.Users[id]; ok {
				mentionables = append(mentionables, u)
				continue
			}

			if r, ok := data.Resolved.Roles[id]; ok {
				mentionables = append(mentionables, r)
			}
		}

		result[sm.CustomID] = mentionables
	}

	return result
}

