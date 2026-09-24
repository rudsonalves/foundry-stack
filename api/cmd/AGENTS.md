# AGENTS — Composição e servidor

`cmd` é o composition root da API. Ele conecta configuração, banco,
repositories, serviços, handlers, middleware e servidor HTTP.

## Responsabilidades

- carregar e validar configuração;
- abrir e encerrar recursos;
- construir dependências na ordem correta;
- registrar rotas públicas e protegidas;
- montar middleware HTTP;
- iniciar e encerrar o servidor.

## Regras

- Não colocar regra de negócio em `cmd`.
- Não executar queries diretamente.
- Manter construção explícita; evitar service locator global.
- Agrupar inicialização por infraestrutura, repositories, domínio, aplicação,
  transport e middleware.
- Falhar cedo para configuração ou dependência inválida.
- Rotas protegidas usam `registerProtected` e `RequireUser`.
- Não aplicar autenticação implicitamente a um grupo que também contenha rotas
  públicas.
- Alterações de rota devem atualizar testes, `docs/api.md` e OpenAPI.
- Configurar timeouts e encerramento de forma explícita.

## Testes

- validar registro de método e path;
- verificar que rotas protegidas rejeitam requisições sem token;
- garantir que rotas públicas não recebem challenge Bearer;
- testar configuração do servidor e middleware sem iniciar rede real.
