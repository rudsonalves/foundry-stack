# API do FoundryStack

API HTTP em Go do starter full-stack FoundryStack.

Esta base contém cadastro com verificação prévia de e-mail, recuperação de
senha, autenticação com access e refresh tokens, PostgreSQL, migrations,
middleware HTTP e encerramento gracioso.
Novos módulos de negócio podem ser adicionados sem alterar a fundação de contas
e autenticação.

## Requisitos

- Go 1.26 ou compatível;
- PostgreSQL;
- Docker e Docker Compose para o ambiente local;
- `golang-migrate` para executar migrations.

## Configuração

Na raiz do monorepo, inicialize os arquivos de ambiente com:

```bash
make env-init
```

Os arquivos são criados a partir dos exemplos correspondentes:

```text
.env.dev.example
.env.stag.example
.env.prod.example
.env.test.example
```

Os segredos JWT e de verificação de e-mail são valores Base64 que, após
decodificados, devem ter pelo menos 32 bytes. Os segredos usados para o código e
para a prova de verificação devem ser distintos. O inicializador gera os
segredos, a senha do PostgreSQL e o `APP_TOKEN` quando estiverem ausentes, sem
substituir valores já configurados.

O envio de e-mail possui dois providers:

- `memory`, permitido apenas em desenvolvimento e testes;
- `smtp`, destinado a staging e produção e configurado pelas variáveis
  `SMTP_*`.

Nos ambientes `dev` e `stag`, o código de seis dígitos também é escrito no log
da API junto ao `verification_id`, para permitir testes manuais. Esse registro
não é habilitado em `prod` nem em `test` e não inclui o endereço de e-mail.

As variáveis `EMAIL_VERIFICATION_*` controlam validade do código e da prova,
quantidade máxima de tentativas, intervalo entre reenvios, limites por e-mail e
IP e retenção dos desafios encerrados. Consulte os arquivos `.env.*.example`
para os valores esperados.

As variáveis `PASSWORD_RESET_*` configuram os segredos exclusivos, validade do
código e da prova, tentativas, cooldown, limites por e-mail e IP e retenção da
recuperação de senha. Os dois segredos devem ser distintos entre si e também
dos segredos de verificação de e-mail. `TRUSTED_PROXY_CIDRS` aceita uma lista
opcional de redes CIDR separadas por vírgula.

Os limitadores de solicitações mantidos em memória, incluindo o de recuperação
de senha, são adequados somente enquanto a API operar com uma única instância.
Antes de escalar horizontalmente, eles devem ser substituídos por um estado
compartilhado. Por padrão, o IP usado nos limites vem da conexão direta;
`X-Forwarded-For` só pode ser considerado quando o proxy de origem estiver em
uma lista explícita de redes confiáveis.

## Execução

```bash
make db-setup
make dev
```

## Verificação

```bash
make check
```

O cadastro ocorre em três chamadas públicas:

1. `POST /email-verifications` envia o código;
2. `POST /email-verifications/confirm` confirma o código e devolve uma prova
   temporária;
3. `POST /users` cria a conta usando a prova vinculada ao mesmo e-mail.

A criação da conta não inicia uma sessão. Depois do cadastro, o usuário acessa
`POST /auth/login`; a renovação e o encerramento da sessão usam,
respectivamente, `POST /auth/refresh` e `POST /auth/logout`.

A recuperação de senha usa três chamadas públicas:

1. `POST /password-resets` solicita o código sem revelar se a conta existe;
2. `POST /password-resets/confirm` confirma o código e devolve uma prova
   temporária;
3. `POST /password-resets/complete` troca a senha, consome a prova e revoga
   todos os refresh tokens, sem criar uma nova sessão.

Access tokens já emitidos permanecem válidos até o próprio TTL. Eles não podem
ser renovados após a troca porque os refresh tokens anteriores são revogados.

Consulte [`docs/api.md`](./docs/api.md) para exemplos, respostas e o contrato
da API. A especificação formal está em
[`docs/swagger/openapi.yaml`](./docs/swagger/openapi.yaml).
