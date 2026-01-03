package i18n

// en returns English translations
func en() *I18n {
	i := &I18n{}

	i.Env.SystemInitMessage = "You are a helpful assistant"

	i.Command.Help.Summary = "The following commands are currently supported:\n"
	i.Command.Help.Help = "Get command help"
	i.Command.Help.New = "Start a new conversation"
	i.Command.Help.Start = "Get your ID and start a new conversation"
	i.Command.Help.Img = "Generate an image, the complete command format is `/img image description`, for example `/img beach at moonlight`"
	i.Command.Help.Version = "Get the current version number to determine whether to update"
	i.Command.Help.Setenv = "Set user configuration, the complete command format is /setenv KEY=VALUE"
	i.Command.Help.Setenvs = "Batch set user configurations, the full format of the command is /setenvs {\"KEY1\": \"VALUE1\", \"KEY2\": \"VALUE2\"}"
	i.Command.Help.Delenv = "Delete user configuration, the complete command format is /delenv KEY"
	i.Command.Help.Clearenv = "Clear all user configuration"
	i.Command.Help.System = "View some system information"
	i.Command.Help.Redo = "Redo the last conversation, /redo with modified content or directly /redo"
	i.Command.Help.Echo = "Echo the message"
	i.Command.Help.Models = "switch chat model"

	i.Command.New.NewChatStart = "A new conversation has started"

	i.CallbackQuery.OpenModelList = "Open models list"
	i.CallbackQuery.SelectProvider = "Select a provider:"
	i.CallbackQuery.SelectModel = "Choose model:"
	i.CallbackQuery.ChangeModel = "Change model to "

	return i
}
