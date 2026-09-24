# AGENTS — Domain

`domain` contém conceitos estáveis da aplicação e casos de uso.

## Dependências

Modelos de `domain/common` devem usar apenas Dart e outros tipos de domínio.
Não podem depender de Flutter, Dio, navegação, storage ou envelopes HTTP.

Use cases podem depender de contratos de repositories e de `Result`, mas
continuam independentes de widgets e transporte.

## Modelos

Preferir classes Dart pequenas, campos finais e construtores explícitos.
Organizar por contexto:

```text
domain/common/catalog/models/item.dart
domain/common/orders/models/order.dart
domain/common/profile/models/profile.dart
```

Usar modelos de domínio quando representam significado estável, combinam
fontes ou isolam o aplicativo de um contrato externo. DTOs de endpoint
permanecem em `data`, ainda que momentaneamente possuam os mesmos campos do
modelo da aplicação.

## Regras

- Não colocar request DTO em domínio.
- Não interpretar envelopes HTTP.
- Não armazenar estado visual.
- Não criar entidades apenas por simetria arquitetural.
- Cobrir parsing, invariantes e enums quando existirem.
