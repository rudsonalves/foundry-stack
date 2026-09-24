# AGENTS — Data

`data` contém integrações externas e a orquestração dos dados do aplicativo.

## Responsabilidades

- expor repositories para view models e use cases;
- implementar repositories;
- construir e interpretar chamadas da API;
- manter DTOs próximos aos endpoints;
- coordenar armazenamento e cache;
- devolver `AsyncResult<T>`.

## Dependências

`data` pode depender de `core`, `domain`, DTOs e serviços da própria camada.
Não pode depender de widgets, páginas, view models, `BuildContext` ou
navegação.

## Separação

- `services/apis`: transporte, endpoints, envelopes e DTOs;
- `repositories`: operações orientadas ao aplicativo, cache e persistência.

API services usam `RestClient`, nunca Dio diretamente. Repositories usam API
services, nunca paths HTTP.

## Modelos e DTOs

DTO representa um contrato de transporte. Modelo de domínio representa um
conceito estável do aplicativo. Mesmo quando possuem os mesmos campos, DTO e
modelo permanecem tipos distintos: a coincidência estrutural não transforma o
contrato externo em modelo da aplicação.

Mapas JSON, nomes `snake_case`, status HTTP e envelopes não chegam à UI.
DTOs também não atravessam a interface pública do API service. O service faz o
mapping por meio de um adapter próprio da integração.

## Testes

Cobrir requests, parsing, mapping, cache e coordenação. Usar fakes; nenhum teste
unitário realiza chamadas reais de rede.
