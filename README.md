# FoundryStack

FoundryStack é um starter full-stack opinativo para aplicações com API em Go,
PostgreSQL e aplicativo Flutter. Ele entrega uma fundação funcional de contas e
autenticação, pronta para receber os módulos de negócio de um novo produto.

## Funcionalidades incluídas

- criação e gerenciamento básico de conta;
- verificação de e-mail por código;
- login com access token JWT e refresh token opaco;
- renovação e encerramento de sessão;
- recuperação e troca de senha;
- revogação de sessões renováveis após a troca de senha;
- API HTTP em Go com PostgreSQL e migrations versionadas;
- aplicativo Flutter com armazenamento seguro e refresh automático;
- ambientes de desenvolvimento, staging, produção e testes;
- testes, análise estática, builds e automações via Makefile.

## Estrutura

```text
foundry-stack/
├── api/       # API Go, PostgreSQL, migrations e OpenAPI
├── mobile/    # aplicativo Flutter
├── docs/      # arquitetura, ambientes e personalização
├── infra/     # preparação dos ambientes locais
└── Makefile   # comandos do monorepo
```

A API é um monólito modular. O mobile usa camadas explícitas de UI, domínio,
dados e infraestrutura. Consulte [a arquitetura](docs/architecture.md) para os
limites de cada camada.

## Requisitos

- Go 1.26.1 ou versão compatível;
- Flutter 3.47.5 como baseline validada, com Dart 3.12.2 ou superior compatível;
- PostgreSQL 18, diretamente ou via Docker;
- Docker e Docker Compose para o ambiente local;
- `golang-migrate` para aplicar migrations;
- Java 17 e toolchain Android compatível com a versão do Flutter;
- Xcode para builds iOS, cujo deployment target mínimo é iOS 15.

## Início rápido

Crie os arquivos locais de ambiente. O comando parte dos exemplos rastreados,
gera segredos ausentes e nunca substitui valores já configurados:

```bash
make env-init
```

Inicie o PostgreSQL, aplique as migrations e execute a API:

```bash
make db-setup
make dev
```

Em outro terminal, execute o aplicativo:

```bash
cd mobile
flutter pub get
flutter run --dart-define-from-file=dev.env
```

Em um aparelho físico, `make mobile-sync-ip` ajusta a URL de desenvolvimento
para o endereço local da máquina.

## Validação

```bash
make check
make mobile-app-env-tests
```

Com um PostgreSQL de testes configurado:

```bash
make -C api test-integration ENV=test
```

Builds de distribuição podem ser verificados com:

```bash
cd mobile
flutter build appbundle --release --dart-define-from-file=prod.env
flutter build ios --release --no-codesign --dart-define-from-file=prod.env
```

## Configuração e personalização

- [Ambientes, banco e execução](docs/environments.md)
- [Arquitetura da API e do mobile](docs/architecture.md)
- [Renomeação, marca e identificadores](docs/customization.md)
- [Contrato HTTP](api/docs/api.md)
- [OpenAPI](api/docs/swagger/openapi.yaml)
- [Guia do aplicativo](mobile/README.md)

Os arquivos `.env.*.example` são seguros para versionamento. Arquivos `.env`
locais, credenciais, tokens e configurações da máquina são ignorados pelo Git.

## Licenças

O código do FoundryStack é distribuído sob a [licença MIT](LICENSE). As fontes
empacotadas mantêm suas licenças SIL Open Font License nos respectivos
diretórios em `mobile/assets/fonts`.

Contribuições seguem o [guia de contribuição](CONTRIBUTING.md). Vulnerabilidades
devem ser reportadas conforme a [política de segurança](SECURITY.md).
