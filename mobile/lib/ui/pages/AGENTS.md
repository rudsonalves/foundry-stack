# AGENTS — Páginas e view models

Páginas são a entrada visual dos fluxos da aplicação.

Estrutura sugerida:

```text
pages/<feature>/<page>/<page>_page.dart
pages/<feature>/<page>/viewmodel/<page>_viewmodel.dart
pages/<feature>/<page>/models/<presentation_model>.dart
pages/<feature>/<page>/widgets/<local_widget>.dart
```

## Divisão de responsabilidades

Página:

- árvore de widgets e `BuildContext`;
- controllers, focus e form keys;
- validação simples de formulário;
- navegação, snackbar e dialog;
- listeners de comandos;
- lifecycle de widgets.

View model:

- repositories para operações simples;
- use cases para fluxos compostos;
- comandos e estado de apresentação;
- getters somente leitura.

## Comandos

- Desabilitar ações enquanto `command.isRunning`.
- Mostrar loading a partir do estado do comando.
- Exibir `command.error?.message` quando apropriado.
- Evitar executar o mesmo comando duas vezes.
- Retirar orquestração assíncrona de callbacks de botões.

## Formulários

Controllers e `GlobalKey<FormState>` pertencem ao `State` da página e devem ser
descartados em `dispose`. Regras compartilhadas ou de domínio não devem ficar
duplicadas em validators locais.

## Não fazer

- chamar API services ou Dio;
- guardar tokens;
- interpretar envelopes HTTP;
- colocar `BuildContext` no view model;
- declarar rotas globais dentro da página;
- promover todo widget local para `ui/components`.

Testar comandos do view model e renderização dos estados essenciais da página.
