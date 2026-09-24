# Arquitetura

## Visão geral

O FoundryStack separa a API e o aplicativo mobile dentro de um monorepo. Os
dois lados compartilham contratos, mas podem ser construídos e implantados de
forma independente.

```text
Flutter → API HTTP → módulos de aplicação → repositories → PostgreSQL
```

## API

A API é um monólito modular em Go:

- `cmd`: composition root, configuração, servidor, middleware e rotas;
- `internal/users`: conta, senha, verificação de e-mail e recuperação;
- `internal/auth`: login, access tokens, refresh tokens e logout;
- `internal/shared`: erros públicos, envelopes HTTP e middleware;
- `internal/bootstrap`: leitura e validação dos ambientes;
- `migrations`: evolução exclusiva do schema PostgreSQL.

Cada módulo separa, quando necessário, domínio, aplicação, infraestrutura e
transporte. Handlers não acessam o banco diretamente e repositories não
produzem respostas HTTP.

## Mobile

O aplicativo Flutter segue os fluxos:

```text
UI → ViewModel → Repository → API Service → RestClient
UI → ViewModel → UseCase → Repository → API Service → RestClient
```

- `lib/app`: composição, containers, escopos e navegação;
- `lib/core`: resultado, erros, HTTP, storage, configuração e logging;
- `lib/data`: APIs, DTOs, adapters e repositories;
- `lib/domain`: modelos estáveis e casos de uso coordenadores;
- `lib/ui`: páginas, view models, temas e componentes.

DTOs e envelopes HTTP ficam restritos a `data`. Tokens são persistidos por uma
abstração de armazenamento seguro e não chegam à UI.

## Autenticação

Access tokens são JWTs de curta duração validados por issuer, audience, client
ID, assinatura e expiração. Refresh tokens são opacos; somente o hash é salvo.
A troca de senha revoga refresh tokens ativos, mas access tokens já emitidos
continuam válidos até o próprio vencimento.

## Extensão

Adicione módulos de negócio ao lado de `users` e `auth`, com seus próprios
limites. No mobile, introduza modelos, repositories, services e páginas apenas
quando o fluxo exigir. Não transforme a base em um framework nem mova código
para `shared` antes de existir reutilização real.
