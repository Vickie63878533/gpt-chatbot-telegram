package i18n

// zhHant returns Traditional Chinese translations
func zhHant() *I18n {
	i := &I18n{}

	i.Env.SystemInitMessage = "你是一個得力的助手"

	i.Command.Help.Summary = "當前支持的命令如下：\n"
	i.Command.Help.Help = "獲取命令幫助"
	i.Command.Help.New = "開始一個新對話"
	i.Command.Help.Start = "獲取您的ID並開始一個新對話"
	i.Command.Help.Img = "生成圖片，完整命令格式為`/img 圖片描述`，例如`/img 海灘月光`"
	i.Command.Help.Version = "獲取當前版本號確認是否需要更新"
	i.Command.Help.Setenv = "設置用戶配置，完整命令格式為/setenv KEY=VALUE"
	i.Command.Help.Setenvs = "批量設置用户配置, 命令完整格式為 /setenvs {\"KEY1\": \"VALUE1\", \"KEY2\": \"VALUE2\"}"
	i.Command.Help.Delenv = "刪除用戶配置，完整命令格式為/delenv KEY"
	i.Command.Help.Clearenv = "清除所有用戶配置"
	i.Command.Help.System = "查看一些系統信息"
	i.Command.Help.Redo = "重做上一次的對話 /redo 加修改過的內容 或者 直接 /redo"
	i.Command.Help.Echo = "回显消息"
	i.Command.Help.Models = "切換對話模式"

	i.Command.New.NewChatStart = "開始一個新對話"

	i.CallbackQuery.OpenModelList = "打開模型清單"
	i.CallbackQuery.SelectProvider = "選擇一個模型供應商:"
	i.CallbackQuery.SelectModel = "選擇一個模型:"
	i.CallbackQuery.ChangeModel = "對話模型已經修改至"

	return i
}
