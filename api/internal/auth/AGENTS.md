# AGENTS — Autenticação

O módulo `auth` autentica credenciais, emite access tokens e administra refresh
tokens. Ele não decide permissões sobre recursos do domínio da aplicação.

## Responsabilidades

- login e emissão de credenciais;
- parsing e validação de access token;
- persistência, expiração e revogação de refresh token;
- renovação e logout;
- identidade autenticada no contexto HTTP;
- validação de clientes permitidos.

## Regras de segurança

- Access token é JWT de curta duração.
- Refresh token é opaco; somente seu hash é persistido.
- Comparações e geração de tokens usam primitivas criptográficas adequadas.
- Não registrar tokens ou secrets.
- Validar issuer, audience, client ID, expiração e assinatura.
- `RequireUser` rejeita header ausente, malformado, duplicado ou com scheme
  diferente de Bearer.
- `401` deve incluir `WWW-Authenticate: Bearer`.
- Logout deve ser idempotente somente se o contrato declarar isso.
- Rotação e detecção de reutilização devem usar transação e testes de
  concorrência quando forem implementadas.

## Limites

- Senha e usuário pertencem a `users`; auth usa seus contratos.
- Permissões de negócio pertencem aos módulos de domínio que forem adicionados.
- Transport não persiste tokens diretamente.
- Infrastructure não produz resposta HTTP.

## Testes

Cobrir credenciais inválidas sem revelar qual campo falhou, claims inválidas,
expiração, cliente não permitido, geração e hashing, refresh revogado e
middleware Bearer.
