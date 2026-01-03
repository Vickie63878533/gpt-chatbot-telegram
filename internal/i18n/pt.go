package i18n

// pt returns Portuguese translations
func pt() *I18n {
	i := &I18n{}

	i.Env.SystemInitMessage = "Você é um assistente útil"

	i.Command.Help.Summary = "Os seguintes comandos são suportados atualmente:\n"
	i.Command.Help.Help = "Obter ajuda sobre comandos"
	i.Command.Help.New = "Iniciar uma nova conversa"
	i.Command.Help.Start = "Obter seu ID e iniciar uma nova conversa"
	i.Command.Help.Img = "Gerar uma imagem, o formato completo do comando é `/img descrição da imagem`, por exemplo `/img praia ao luar`"
	i.Command.Help.Version = "Obter o número da versão atual para determinar se é necessário atualizar"
	i.Command.Help.Setenv = "Definir configuração do usuário, o formato completo do comando é /setenv CHAVE=VALOR"
	i.Command.Help.Setenvs = "Definir configurações do usuário em lote, o formato completo do comando é /setenvs {\"CHAVE1\": \"VALOR1\", \"CHAVE2\": \"VALOR2\"}"
	i.Command.Help.Delenv = "Excluir configuração do usuário, o formato completo do comando é /delenv CHAVE"
	i.Command.Help.Clearenv = "Limpar todas as configurações do usuário"
	i.Command.Help.System = "Ver algumas informações do sistema"
	i.Command.Help.Redo = "Refazer a última conversa, /redo com conteúdo modificado ou diretamente /redo"
	i.Command.Help.Echo = "Repetir a mensagem"
	i.Command.Help.Models = "Mudar o modelo de diálogo"

	i.Command.New.NewChatStart = "Uma nova conversa foi iniciada"

	i.CallbackQuery.OpenModelList = "Abra a lista de modelos"
	i.CallbackQuery.SelectProvider = "Escolha um fornecedor de modelos.:"
	i.CallbackQuery.SelectModel = "Escolha um modelo:"
	i.CallbackQuery.ChangeModel = "O modelo de diálogo já foi modificado para"

	return i
}
