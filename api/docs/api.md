# Endpoints da API

Este documento apresenta o contrato HTTP da API do FoundryStack em formato legível.
A especificação formal fica em [`swagger/openapi.yaml`](./swagger/openapi.yaml).

## Estado do contrato

Cadastro, autenticação, verificação do e-mail e recuperação de senha estão
implementados. As decisões do fluxo de verificação estão registradas no backlog
encerrado
[`001-credentials.md`](../../docs/backlogs/api/closed/001-credentials.md).

## Convenções

### URL base

Em desenvolvimento local:

```text
http://localhost:8080
```

### Conteúdo

Requisições com corpo e respostas JSON usam:

```http
Content-Type: application/json
```

### Envelope de sucesso

```json
{
  "data": {}
}
```

Respostas `204 No Content` não possuem corpo.

### Envelope de erro

```json
{
  "error": {
    "code": "BAD_REQUEST",
    "message": "invalid input",
    "request_id": "request-id",
    "details": [
      {
        "field": "email",
        "message": "invalid email format"
      }
    ]
  }
}
```

`request_id` e `details` aparecem quando aplicáveis. Códigos genéricos atuais:

- `BAD_REQUEST`;
- `UNAUTHORIZED`;
- `FORBIDDEN`;
- `NOT_FOUND`;
- `CONFLICT`;
- `INTERNAL_ERROR`.

Erros específicos relevantes ao fluxo de recuperação:

- `INVALID_PASSWORD_RESET`: recuperação ausente, sintética, inválida,
  expirada, bloqueada ou já consumida;
- `RATE_LIMIT_EXCEEDED`: cooldown ou limite por e-mail/IP excedido.

### Request ID

O cliente pode enviar `X-Request-ID`. Quando ausente, a API cria um
identificador e o devolve no header da resposta.

## Resumo

| Método | Caminho | Estado | Autenticação | Finalidade |
|---|---|---|---|---|
| `POST` | `/users` | Implementado | Pública | Criar usuário |
| `POST` | `/auth/login` | Implementado | Pública | Autenticar credenciais |
| `POST` | `/auth/refresh` | Implementado | Refresh token no corpo | Renovar access token |
| `POST` | `/auth/logout` | Implementado | Refresh token no corpo | Revogar refresh token |
| `POST` | `/email-verifications` | Implementado | Pública | Enviar código de verificação |
| `POST` | `/email-verifications/confirm` | Implementado | Pública | Confirmar código e emitir prova |
| `POST` | `/password-resets` | Implementado | Pública | Solicitar recuperação de senha |
| `POST` | `/password-resets/confirm` | Implementado | Pública | Confirmar código e emitir prova |
| `POST` | `/password-resets/complete` | Implementado | Pública | Trocar senha e encerrar sessões renováveis |

## Endpoints implementados

### Criar usuário

```http
POST /users
```

Cria uma conta após a confirmação prévia do e-mail. A requisição exige a prova
temporária emitida por `POST /email-verifications/confirm`.

Requisição:

```json
{
  "name": "Ana Silva",
  "email": "ana@example.com",
  "password": "senha-segura",
  "email_verification_token": "opaque-secret"
}
```

Regras atuais:

- `name` é obrigatório;
- `email` é obrigatório, válido, normalizado e único;
- `password` deve possuir entre 8 e 72 caracteres.
- `email_verification_token` deve ser uma prova válida, ativa e vinculada ao
  mesmo e-mail.

Resposta `201 Created`:

```json
{
  "data": {
    "id": "7fd1ec52-6f58-4fe2-b482-e31c06adfba5",
    "name": "Ana Silva",
    "email": "ana@example.com"
  }
}
```

Respostas de erro:

- `400 Bad Request`: JSON inválido, campos inválidos ou
  `INVALID_EMAIL_VERIFICATION`;
- `409 Conflict`: `EMAIL_ALREADY_REGISTERED`;
- `500 Internal Server Error`: falha inesperada.

### Login

```http
POST /auth/login
```

Valida e-mail, senha e cliente, emitindo access e refresh token.

Requisição:

```json
{
  "email": "ana@example.com",
  "password": "senha-segura",
  "client_id": "foundry-stack-mobile"
}
```

`client_id` deve estar entre os valores configurados em `JWT_CLIENT_IDS`.

Resposta `200 OK`:

```json
{
  "data": {
    "access_token": "access-token",
    "refresh_token": "refresh-token",
    "token_type": "Bearer",
    "expires_in": 900
  }
}
```

Respostas de erro:

- `400 Bad Request`: JSON inválido;
- `401 Unauthorized`: credenciais inválidas ou cliente não permitido;
- `500 Internal Server Error`: falha inesperada.

Por segurança, usuário inexistente e senha incorreta recebem a mesma resposta.

### Renovar access token

```http
POST /auth/refresh
```

Emite um novo access token usando um refresh token ativo. O refresh token atual
não é rotacionado nem tem sua validade estendida.

Requisição:

```json
{
  "refresh_token": "refresh-token"
}
```

Resposta `200 OK`:

```json
{
  "data": {
    "access_token": "new-access-token",
    "token_type": "Bearer",
    "expires_in": 900
  }
}
```

Respostas de erro:

- `400 Bad Request`: JSON inválido;
- `401 Unauthorized`: refresh token ausente, inválido, expirado ou revogado;
- `500 Internal Server Error`: falha inesperada.

### Logout

```http
POST /auth/logout
```

Revoga o refresh token informado.

Requisição:

```json
{
  "refresh_token": "refresh-token"
}
```

Resposta:

```text
204 No Content
```

Repetir o logout com um token já ausente ou revogado é tratado como operação
sem efeito.

## Verificação prévia do e-mail

### Solicitar ou reenviar código

```http
POST /email-verifications
```

Requisição:

```json
{
  "email": "ana@example.com"
}
```

Resposta `202 Accepted`:

```json
{
  "data": {
    "verification_id": "7fd1ec52-6f58-4fe2-b482-e31c06adfba5",
    "code_expires_at": "2026-09-10T15:15:00Z",
    "resend_available_at": "2026-09-10T15:01:00Z"
  }
}
```

Comportamento:

- normaliza o e-mail;
- invalida desafio e prova anteriores antes de emitir outro código;
- não cria usuário;
- responde `409 Conflict` com `EMAIL_ALREADY_REGISTERED` quando já houver conta;
- responde erro temporário se o envio falhar;
- aplica cooldown e limites por e-mail e IP;
- responde `429 Too Many Requests` quando um limite for excedido.

### Confirmar código

```http
POST /email-verifications/confirm
```

Requisição:

```json
{
  "verification_id": "7fd1ec52-6f58-4fe2-b482-e31c06adfba5",
  "code": "012345"
}
```

Resposta `200 OK`:

```json
{
  "data": {
    "email_verification_token": "opaque-secret",
    "expires_at": "2026-09-11T15:00:00Z"
  }
}
```

Comportamento:

- trata o código como string para preservar zeros à esquerda;
- código válido emite uma prova temporária ligada ao e-mail;
- repetir uma confirmação válida rotaciona a prova anterior;
- código incorreto incrementa o contador de tentativas;
- a quinta falha invalida o desafio;
- código ou desafio inválido, expirado, invalidado ou bloqueado responde
  `INVALID_EMAIL_VERIFICATION` sem diferenciar a causa.

### Criar usuário após a verificação

`POST /users` exige:

```json
{
  "name": "Ana Silva",
  "email": "ana@example.com",
  "password": "senha-segura",
  "email_verification_token": "opaque-secret"
}
```

Resposta `201 Created`:

```json
{
  "data": {
    "id": "7fd1ec52-6f58-4fe2-b482-e31c06adfba5",
    "name": "Ana Silva",
    "email": "ana@example.com"
  }
}
```

Regras:

- a prova deve estar válida, ativa e vinculada ao mesmo e-mail;
- usuário e consumo da prova são gravados atomicamente;
- prova ausente, inválida, expirada, consumida ou divergente responde
  `INVALID_EMAIL_VERIFICATION`;
- e-mail já cadastrado responde `EMAIL_ALREADY_REGISTERED`;
- uma repetição após sucesso também responde `EMAIL_ALREADY_REGISTERED`;
- a resposta não contém access nem refresh token;
- após o cadastro, o usuário autentica por `POST /auth/login`.

## Tokens de autenticação

Valores padrão atuais:

- access token: 15 minutos (`JWT_ACCESS_TTL_SECONDS=900`);
- refresh token: 48 horas (`JWT_REFRESH_TTL_SECONDS=172800`).

O access token é JWT. O refresh token é opaco e apenas seu hash é persistido.
Endpoints protegidos futuros usarão:

```http
Authorization: Bearer <access_token>
```

Nunca registrar access token, refresh token, senha, código de verificação ou
`email_verification_token`.

## Recuperação de senha

### Solicitar recuperação

```http
POST /password-resets
```

Requisição:

```json
{
  "email": "ana@example.com"
}
```

Resposta `202 Accepted`:

```json
{
  "data": {
    "password_reset_id": "7fd1ec52-6f58-4fe2-b482-e31c06adfba5",
    "code_expires_at": "2026-09-23T15:15:00Z",
    "resend_available_at": "2026-09-23T15:01:00Z"
  }
}
```

A resposta é deliberadamente igual para contas existentes e inexistentes. Para
uma conta inexistente, o identificador e os prazos são sintéticos e nenhum
e-mail é enviado. Cooldown e limites por e-mail e IP são aplicados nos dois
casos; excesso responde `429 Too Many Requests`.

O limitador atual reside em memória e pressupõe uma única instância da API.
Headers de encaminhamento só são considerados quando a conexão vem de uma rede
configurada em `TRUSTED_PROXY_CIDRS`.

### Confirmar código

```http
POST /password-resets/confirm
```

Requisição:

```json
{
  "password_reset_id": "7fd1ec52-6f58-4fe2-b482-e31c06adfba5",
  "code": "012345"
}
```

Resposta `200 OK`:

```json
{
  "data": {
    "password_reset_token": "opaque-secret",
    "expires_at": "2026-09-23T15:30:00Z"
  }
}
```

O código possui seis dígitos e é tratado como string. Código incorreto aumenta
o contador; a quinta falha invalida o desafio. Repetir uma confirmação válida
rotaciona a prova anterior. Todas as rejeições semânticas usam
`INVALID_PASSWORD_RESET` sem revelar a causa.

### Trocar senha

```http
POST /password-resets/complete
```

Requisição:

```json
{
  "password_reset_id": "7fd1ec52-6f58-4fe2-b482-e31c06adfba5",
  "password_reset_token": "opaque-secret",
  "new_password": "nova-senha-segura"
}
```

Resposta:

```text
204 No Content
```

A senha deve possuir de 8 a 72 caracteres. Em uma única transação, a API troca
a senha, consome a prova e revoga todos os refresh tokens do usuário. A chamada
não emite tokens nem cria sessão; o usuário deve fazer login com a nova senha.
Repetir a conclusão com a mesma prova responde `INVALID_PASSWORD_RESET`.

Access tokens já emitidos são JWTs autônomos e permanecem válidos até o próprio
TTL. Como os refresh tokens anteriores são revogados, eles não podem ser usados
para renovar a sessão.
