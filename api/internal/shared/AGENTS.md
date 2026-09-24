# AGENTS — Infraestrutura compartilhada

`shared` contém código transversal utilizado por mais de um módulo.

## Critério de entrada

Mover código para `shared` somente quando houver reutilização real e semântica
comum. Não usar como pasta de conveniência.

## Erros

- `AppError` representa significado público e envolve a causa interna.
- Códigos de erro são estáveis para clientes.
- Usar construtores coerentes para bad request, unauthorized, forbidden, not
  found e conflict.
- Erro inesperado vira `INTERNAL_ERROR` sem detalhes internos.
- Validação pode incluir violações de campo.

## HTTP

- `AppHandler` centraliza retorno de erros.
- Respostas usam os envelopes `data` e `error`.
- Serializar antes de escrever status quando a serialização puder falhar.
- Definir `Content-Type: application/json`.
- Middleware deve chamar o próximo handler exatamente uma vez.
- A ordem do chain é parte do comportamento e deve ser testada.

## Middleware

- request ID deve ser propagado;
- CORS aceita somente origens configuradas;
- logging não inclui segredos;
- recovery impede que panic encerre o servidor;
- middleware compartilhado não contém regra de um módulo.

## Testes

Cobrir status, headers, JSON, ordem de middleware, panic, preflight, request ID
e falhas de escrita quando praticável.
