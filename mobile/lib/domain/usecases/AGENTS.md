# AGENTS — Use cases

Use cases coordenam fluxos complexos do aplicativo.

## Criar quando

- uma ação utiliza mais de um repository;
- existe uma sequência de etapas ou decisões;
- o mesmo fluxo será reutilizado;
- o view model começaria a conter regras que não são de apresentação.

Não criar use case para simples delegação de um método de repository.

## Regras

- Receber repositories por construtor.
- Retornar `Result` ou `AsyncResult`.
- Não chamar API services, RestClient ou Dio.
- Não acessar storage diretamente.
- Não usar widgets, `BuildContext`, navegação ou feedback visual.
- Não mover para cá cache ou persistência que pertencem a repositories.

Estrutura sugerida:

```text
usecases/<feature>/<feature>_usecase.dart
usecases/<feature>/inputs/<input>.dart
```

Testes devem cobrir ordem das operações, branches e propagação de falhas.
