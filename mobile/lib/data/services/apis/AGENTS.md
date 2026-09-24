# AGENTS — API services

API services são a camada do aplicativo imediatamente acima do `RestClient`.

## Responsabilidades

- conhecer paths, métodos, headers, query e body;
- serializar requests;
- interpretar o envelope `{data, error}`;
- usar adapters para converter modelos da aplicação em DTOs de request;
- usar adapters para converter DTOs de response em modelos da aplicação;
- transformar falhas de parsing em `AppError`;
- retornar `AsyncResult<T>`.

Estrutura sugerida:

```text
services/apis/<feature>/<operation>_api.dart
services/apis/<feature>/dtos/<operation>_request_dto.dart
services/apis/<feature>/dtos/<operation>_response_dto.dart
```

## Regras

- Usar `RestClientRequest` em todas as chamadas.
- Não importar Dio.
- Não acessar storage diretamente.
- Não manter sessão ou cache.
- Não depender de repositories, UI ou navegação.
- Não expor DTOs, mapas JSON ou envelopes em sua interface pública.
- Manter DTOs imutáveis, com `toMap` e `fromMap` quando aplicável.
- Validar campos obrigatórios durante o parsing.
- Manter adapters de mapping próximos ao service, inclusive para conversões
  triviais, preservando uma fronteira uniforme.

## Testes

Usar um `RestClient` fake e verificar path, método, query, headers, body,
parsing, erros do envelope e responses incompletas.
