# Ambientes, banco e execução

## Ambientes

A API reconhece `dev`, `stag`, `prod` e `test`. Os exemplos rastreados ficam em
`api/.env.<ambiente>.example`; `make env-init` cria os arquivos locais sem
sobrescrever valores existentes.

O mobile usa `dev.env`, `stag.env` e `prod.env`, gerados localmente. Exemplos
sem segredos ficam ao lado deles com o sufixo `.example`.

Produção e staging exigem SMTP. Desenvolvimento e testes podem usar o provider
em memória. Nunca reutilize os segredos de JWT, verificação de e-mail e
recuperação de senha entre finalidades ou ambientes.

## PostgreSQL e migrations

O ambiente local usa PostgreSQL 18 por padrão. O schema é criado somente pelas
migrations em `api/migrations`:

```bash
make db-up
make migrate-up
```

Para criar uma migration:

```bash
make -C api migrate-create MIGRATION_NAME=add_example
```

Teste sempre a sequência completa em banco vazio e o rollback correspondente.
`db-reset` remove dados e é bloqueado para `ENV=prod`.

## API

```bash
make dev
make stag
make prod
```

O endereço padrão é `http://localhost:8080`; Swagger UI pode ser iniciado com
`make -C api swagger-up`.

## Mobile

```bash
cd mobile
flutter run --dart-define-from-file=dev.env
```

Em emulador Android, `10.0.2.2` aponta para o host. Para aparelho físico, use
`make mobile-sync-ip`. URLs de staging e produção são placeholders e devem ser
substituídas antes de um build real.
