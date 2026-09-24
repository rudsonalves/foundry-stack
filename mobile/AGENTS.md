# AGENTS — Mobile

Este arquivo orienta alterações em todo o aplicativo Flutter do FoundryStack.

## Objetivo

O aplicativo mobile demonstra a fundação de conta e autenticação oferecida pelo
starter. Funcionalidades específicas devem ser adicionadas em módulos próprios,
sem acoplar regras de produto à infraestrutura compartilhada.

## Arquitetura

Fluxo simples:

```text
UI → ViewModel → Repository → API/Service → RestClient → Dio
```

Fluxo com coordenação de múltiplas operações:

```text
UI → ViewModel → UseCase → Repository → API/Service → RestClient → Dio
```

Camadas:

- `lib/app`: composition root, containers, navegação e widget raiz;
- `lib/core`: infraestrutura transversal;
- `lib/data`: APIs, DTOs e repositories;
- `lib/domain`: conceitos estáveis e casos de uso;
- `lib/ui`: páginas, view models e componentes visuais;
- `test`: testes que acompanham a estrutura de `lib`.

## Regras globais

- A UI não chama APIs, Dio ou armazenamento diretamente.
- View models expõem estado e `Command`; não recebem `BuildContext`.
- Repositories escondem detalhes de transporte e persistência.
- API services constroem requisições e interpretam respostas; adapters próximos
  aos services traduzem entre modelos da aplicação e DTOs da API.
- DTOs da API não são expostos a repositories, use cases, view models ou UI.
- Adapters de mapping permanecem junto à integração da API e tornam explícita
  essa fronteira mesmo quando a conversão é trivial.
- Use cases existem apenas quando há coordenação relevante.
- Falhas esperadas atravessam camadas como `Result` e `AppError`.
- Exceções cruas não devem chegar à UI.
- Dependências são recebidas por construtor.
- Não introduzir bibliotecas ou abstrações antes de um caso de uso real.
- Reutilizar a infraestrutura de `core` antes de criar alternativas.

## Dependências entre camadas

- `ui` pode depender de `domain`, contratos de `data` e `core`;
- `data` pode depender de `domain` e `core`;
- `domain/common` deve permanecer independente de Flutter e infraestrutura;
- `core` não depende de `data`, `domain` ou `ui`.

## Estado e trabalho assíncrono

- Ações iniciadas pela UI devem usar `Command0` ou `Command1`.
- Páginas observam loading, sucesso e falha através do comando.
- Evitar chamadas assíncronas diretamente em callbacks quando um view model
  representa o fluxo.
- Repositories podem manter cache pequeno quando ele representa estado de dados,
  nunca estado visual.

## Testes

- Adicionar testes ao mudar `core`, parsing, APIs, repositories, use cases ou
  comportamento de view models.
- Usar fakes para `RestClient`, API services e armazenamento.
- Testes unitários não fazem chamadas reais de rede.
- Páginas relevantes devem cobrir loading, sucesso, vazio e erro.

## Antes de criar código

Consultar o guia mais próximo:

- `lib/core/AGENTS.md`;
- `lib/data/AGENTS.md`;
- `lib/data/repositories/AGENTS.md`;
- `lib/data/services/apis/AGENTS.md`;
- `lib/domain/AGENTS.md`;
- `lib/domain/usecases/AGENTS.md`;
- `lib/ui/AGENTS.md`;
- `lib/ui/components/AGENTS.md`;
- `lib/ui/pages/AGENTS.md`.

Consulte também `docs/ARCHITECTURE.md`.
