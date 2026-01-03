package i18n

// zhHans returns Simplified Chinese translations
func zhHans() *I18n {
	i := &I18n{}

	i.Env.SystemInitMessage = "你是一个得力的助手"

	i.Command.Help.Summary = "当前支持以下命令:\n"
	i.Command.Help.Help = "获取命令帮助"
	i.Command.Help.New = "发起新的对话"
	i.Command.Help.Start = "获取你的ID, 并发起新的对话"
	i.Command.Help.Img = "生成一张图片, 命令完整格式为 `/img 图片描述`, 例如`/img 月光下的沙滩`"
	i.Command.Help.Version = "获取当前版本号, 判断是否需要更新"
	i.Command.Help.Setenv = "设置用户配置，命令完整格式为 /setenv KEY=VALUE"
	i.Command.Help.Setenvs = "批量设置用户配置, 命令完整格式为 /setenvs {\"KEY1\": \"VALUE1\", \"KEY2\": \"VALUE2\"}"
	i.Command.Help.Delenv = "删除用户配置，命令完整格式为 /delenv KEY"
	i.Command.Help.Clearenv = "清除所有用户配置"
	i.Command.Help.System = "查看当前一些系统信息"
	i.Command.Help.Redo = "重做上一次的对话, /redo 加修改过的内容 或者 直接 /redo"
	i.Command.Help.Echo = "回显消息"
	i.Command.Help.Models = "切换对话模型"

	i.Command.New.NewChatStart = "新的对话已经开始"

	i.CallbackQuery.OpenModelList = "打开模型列表"
	i.CallbackQuery.SelectProvider = "选择一个模型提供商:"
	i.CallbackQuery.SelectModel = "选择一个模型:"
	i.CallbackQuery.ChangeModel = "对话模型已修改至"

	return i
}
