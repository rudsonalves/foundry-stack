# AGENTS — Core

`core` contém infraestrutura transversal utilizada pelas demais camadas.

## Responsabilidades

- resultado e erros compartilhados;
- comandos assíncronos;
- configuração global;
- abstração HTTP e adapters de infraestrutura;
- logging e utilitários realmente compartilhados.

## Regras

- Não depender de `data`, `domain` ou `ui`.
- Não incluir regras de anúncios ou de uma feature.
- Esconder Dio dentro de `services/client_http/dio`.
- Expor integrações por contratos pequenos.
- Não deixar exceções cruas atravessarem camadas.
- Não registrar tokens, credenciais ou corpos sensíveis em logs.
- Manter `Result`, `AppError` e `Command` genéricos.
- Adicionar testes ao alterar contratos ou adapters.

## HTTP

O contrato público é `RestClient`. API services constroem
`RestClientRequest` e recebem `Result<RestClientResponse>`.

Novos métodos HTTP devem entrar primeiro no contrato. Parsing de endpoints não
pertence ao `core`.

## Configuração

Configuração de execução fica em `resources/app_env.dart`. Valores obrigatórios
devem falhar cedo quando inválidos. Segredos não devem ser definidos como
constantes no código.

## Não fazer

- importar páginas, repositories ou DTOs;
- adicionar helpers usados por apenas uma feature;
- navegar, mostrar snackbar ou usar `BuildContext`;
- criar uma segunda abstração de resultado ou HTTP.
