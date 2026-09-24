# AGENTS — Repositories

Repositories são a fronteira de dados consumida por view models e use cases.

## Regras

- Nomear métodos por operações do aplicativo, não por verbos HTTP.
- Definir contratos focados e implementações `*RepositoryImpl`.
- Receber APIs e armazenamento pelo construtor.
- Retornar `Result` ou `AsyncResult`.
- Usar `Unit` quando não houver retorno significativo.
- Esconder detalhes de endpoints, DTOs internos e persistência.
- Manter cache pequeno, explícito e com estratégia de invalidação.
- Não manter controllers, estado visual ou navegação.

Estrutura sugerida:

```text
repositories/<feature>/<feature>_repository.dart
repositories/<feature>/<feature>_repository_impl.dart
```

## Estado

Um repository pode manter estado como usuário atual, anúncios em cache ou última
sincronização quando isso pertence ao comportamento dos dados. Expor somente
getters ou streams necessários.

## Testes

Cobrir coordenação de APIs, propagação de falhas, persistência, cache,
invalidação e streams. APIs e storage devem ser fakes.
