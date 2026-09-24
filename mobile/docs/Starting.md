# Iniciando o aplicativo mobile

## Requisitos

- Flutter SDK configurado;
- dispositivo ou emulador Android/iOS;
- API FoundryStack acessível pelo dispositivo.

## Dependências

Dentro de `mobile/`:

```bash
flutter pub get
```

## Executar

Na raiz do monorepo, gere os ambientes da API e do aplicativo:

```bash
make env-init
```

Para um dispositivo físico, sincronize a URL com o IP da máquina:

```bash
make mobile-sync-ip
```

O aplicativo exige uma URL válida quando o cliente HTTP for inicializado:

```bash
flutter run \
  --dart-define=BASE_URL=http://10.0.2.2:8080 \
  --dart-define=APP_MODE=dev \
  --dart-define=AUTH_CLIENT_ID=foundry-stack-mobile
```

`AUTH_CLIENT_ID` usa `foundry-stack-mobile` como padrão e não é segredo. A API valida
se o identificador recebido pertence aos clientes autorizados.

No emulador Android, `10.0.2.2` aponta para a máquina host. Em um dispositivo
físico, use um endereço da rede local que seja acessível pelo aparelho.

Para executar os testes de `AppEnv` com as fixtures de `dart-define` em
`mobile/test/env/`, use o target dedicado:

```bash
make mobile-app-env-tests
```

As fixtures cobrem os casos de `AUTH_CLIENT_ID=""`, `AUTH_CLIENT_ID="   "`
e `AUTH_CLIENT_ID="   hhhhhh   "`, além do novo `EXPECTED_AUTH_CLIENT_ID`
usado pelo teste.

## Qualidade

```bash
flutter analyze
flutter test
```

Também é possível executar a verificação conjunta a partir da raiz do monorepo:

```bash
make mobile-check
```

## Antes de implementar

Leia:

1. `mobile/AGENTS.md`;
2. `mobile/docs/ARCHITECTURE.md`;
3. o `AGENTS.md` mais próximo da pasta que será alterada.
