# AGENTS — Usuários

O módulo `users` mantém conta, perfil básico, credenciais e cadastro.

## Regras de domínio

- Nome e e-mail são obrigatórios.
- E-mail deve ser normalizado de forma consistente.
- E-mail é único no banco.
- Senha em texto puro nunca entra na entidade nem no repository.
- `PasswordHash` não é serializado em respostas.
- Hash e comparação seguem o contrato `PasswordHasher`.

## Camadas

- domain define `User`, `UserRepository`, `PasswordHasher` e erros sentinela;
- application coordena cadastro e autenticação de credenciais;
- infrastructure implementa PostgreSQL e bcrypt;
- transport decodifica requests e expõe apenas dados públicos.

## Erros

- ausência vira `ErrUserNotFound`;
- unicidade conhecida vira `ErrEmailAlreadyExists`;
- erro de banco inesperado preserva a causa e não vaza ao cliente;
- login não deve permitir enumeração por mensagens diferentes.

## Testes

Cobrir normalização, validação, hashing, e-mail duplicado, usuário ausente,
falhas de repository e ausência de `password_hash` na resposta.
