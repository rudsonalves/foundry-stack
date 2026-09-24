# Aplicativo mobile do FoundryStack

Aplicativo Flutter do starter full-stack FoundryStack.

## Estado atual

O projeto possui a base de autenticação implementada: cadastro, login,
restauração e encerramento de sessão, armazenamento seguro de tokens, refresh
automático e composição das dependências. A Splash já restaura a sessão ao
iniciar: abre a Home quando o refresh é válido, abre a Login quando não existe
sessão e oferece nova tentativa em falhas transitórias. Login e criação de
usuário já estão disponíveis como base para a evolução da interface.

## Arquitetura

A estrutura segue uma arquitetura explícita em camadas:

```text
UI → ViewModel → Repository → API/Service → RestClient
```

A pasta `lib/core` contém a infraestrutura compartilhada, incluindo HTTP,
resultados, armazenamento seguro e política de tokens. A camada `data` contém
o serviço e o repository de autenticação. Repository e interceptor compartilham
o mesmo `AuthTokenStore`, sem expor tokens ou chaves de storage à UI.
`lib/app` reúne a composição dos containers, o router e o widget raiz, pois essa
camada conhece e conecta as demais partes da aplicação.

Documentação:

- `docs/ARCHITECTURE.md`: responsabilidades, dependências e fluxos;
- `docs/Starting.md`: preparação e execução local;
- `AGENTS.md`: regras gerais para construção de código;
- arquivos `AGENTS.md` internos: regras específicas de cada camada.

## Execução

```bash
flutter pub get
flutter run \
  --dart-define=BASE_URL=http://10.0.2.2:8080 \
  --dart-define=APP_MODE=dev \
  --dart-define=AUTH_CLIENT_ID=foundry-stack-mobile
```

`AUTH_CLIENT_ID` identifica o tipo de cliente e não é segredo. O valor padrão é
`foundry-stack-mobile`; a API continua responsável por validar se ele está autorizado.

Também é possível usar os arquivos de ambiente locais. Na raiz do monorepo,
gere e sincronize esses arquivos com as configurações da API:

```bash
make env-init
```

Depois execute:

```bash
flutter run --dart-define-from-file=dev.env
```

`APP_ACCESS_TOKEN` recebe o mesmo valor de `APP_TOKEN` do ambiente equivalente
da API. A configuração fica pronta para o controle de acesso da aplicação,
mesmo que esse token ainda não seja enviado nas requisições atuais.

## Sessão

- o login só termina com sucesso depois que access e refresh token são salvos;
- a restauração usa o refresh token e diferencia sessão restaurada, sessão
  ausente e falha transitória;
- o interceptor renova o access token e repete uma requisição protegida uma
  única vez;
- `401` no refresh limpa a sessão; timeout, conexão indisponível e `5xx`
  preservam as credenciais;
- o logout sempre tenta limpar a sessão local, mesmo quando a revogação remota
  falha.

Quando o logout ocorre sem conexão, o refresh token pode continuar válido no
servidor até expirar. O dispositivo deixa de considerá-lo autenticado, mas a
revogação remota não é garantida nesse cenário.

Para validar os testes de configuração do `AppEnv` com os `.env` de teste:

```bash
make mobile-app-env-tests
```

## Testes e análise

```bash
flutter analyze
flutter test
```

O contrato inicial utilizado pelo aplicativo está documentado em
`../api/docs/swagger/openapi.yaml`.
