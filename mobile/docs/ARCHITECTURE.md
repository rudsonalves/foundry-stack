# Arquitetura do aplicativo mobile

## Visão geral

O aplicativo Flutter do FoundryStack adota uma arquitetura em camadas. O objetivo é
manter a interface simples, isolar integrações e tornar os fluxos da aplicação
previsíveis e testáveis.

## Princípios

- separação clara de responsabilidades;
- dependências recebidas por construtor;
- erros explícitos com `Result` e `AppError`;
- infraestrutura escondida atrás de contratos;
- regras de negócio independentes de Flutter;
- API services responsáveis pelo fluxo HTTP e parsing;
- adapters responsáveis pela tradução entre modelos e DTOs;
- DTOs restritos à integração com a API;
- repositories responsáveis pela visão de dados do aplicativo;
- use cases apenas para fluxos com coordenação real;
- evolução incremental, sem camadas ou abstrações vazias desnecessárias.

## Estrutura

```text
lib/
├── app/
│   ├── dependencies/
│   │   └── bindings/
│   ├── routing/
│   └── app_widget.dart
├── core/
│   ├── resources/
│   ├── result/
│   └── services/
├── data/
│   ├── repositories/
│   └── services/apis/
│       └── <feature>/
│           ├── adapters/
│           └── dtos/
├── domain/
│   ├── common/
│   └── usecases/
├── ui/
│   ├── components/
│   └── pages/
└── main.dart
```

Os testes devem acompanhar essa organização:

```text
test/
├── app/
├── core/
├── data/
├── domain/
└── ui/
```

## Fluxos

Uma operação simples, como carregar os anúncios disponíveis:

```text
Page
  → ViewModel
    → ListingRepository
      → ListingService
        → ListingApiAdapter
          → DTO/HTTP
            → RestClient
              → Dio
```

Uma operação composta, como publicar um anúncio com fotos e localização:

```text
Page
  → ViewModel
    → PublishListingUseCase
      → Repositories
        → APIs/Services
          → Adapters
            → DTO/HTTP
              → RestClient
```

Não criar um use case para cada método de repository. Ele se justifica quando
coordena múltiplos repositories, possui várias etapas ou representa um fluxo
reutilizável.

## Core

`core` contém infraestrutura compartilhada:

- `Result`, `AppError`, `Unit` e `Command`;
- configuração por `dart-define`;
- contrato HTTP e implementação Dio;
- `AuthTokenStore` e sua implementação sobre armazenamento seguro;
- interceptor com refresh automático e retry controlado;
- logging de desenvolvimento;
- composição das dependências transversais.

`core` não conhece anúncios, usuários, páginas ou regras de domínio.

## Containers e ciclos de vida

Uma explicação passo a passo da implementação está em
[`dependency_containers.md`](./articles/dependency_containers.md).

A composição de dependências separa três containers:

```text
RootContainer (composition root; não é AutoInjector)
├── CoreContainer (AutoInjector)
├── AppContainer (AutoInjector)
└── createOnboardingScope() → OnboardingScope (AutoInjector temporário)
```

`CoreContainer` mantém as instâncias fundamentais e compartilhadas, como
configuração, logging, `RestClient`, `LocalSecureStorage` e `AuthTokenStore`.
Este último permanece no core porque é usado pelo `AuthInterceptor`; a aplicação
operacional e o interceptor devem observar a mesma fonte de sessão.

`AppContainer` e cada `OnboardingScope` possuem registros e ciclos de vida
independentes. Eles não são descendentes simultâneos do core com
`resolveUpward: true`: no `AutoInjector 2.2.0`, a busca na árvore também alcança
descendentes e poderia permitir resolução lateral entre esses escopos.

O composition root cria as dependências fundamentais uma única vez e entrega
explicitamente as referências necessárias ao configurar app e onboarding.
`CoreDependencies` representa a saída completa permitida pelo core; o root a
converte em `AppDependencies` e `OnboardingDependencies`, que expõem somente as
referências usadas por cada consumidor. Os containers registram essas mesmas
instâncias; não recriam cliente HTTP, storage ou sessão e não recebem o injector
do core como service locator. O core permanece proprietário das referências
compartilhadas; os consumidores não registram callbacks para descartá-las.

Os registros do `AutoInjector` ficam em funções com nomenclatura explícita:
`registerCoreBindings`, `registerAppBindings` e
`registerOnboardingBindings`. Eles são agrupados pelo ciclo de vida do
container, não pela camada técnica do objeto registrado.

Cada função registra somente o fechamento mínimo do grafo necessário ao seu
injector. Não existem registradores gerais de services, repositories, use cases
ou view models, pois containers diferentes podem precisar de subconjuntos
diferentes dessas camadas. Um binding de outro container nunca é incluído apenas
por pertencer à mesma categoria técnica.

Bindings menores somente devem ser extraídos quando representarem uma
capacidade coesa realmente reutilizada, como o conjunto de integração necessário
ao cadastro de usuário. O container continua escolhendo explicitamente quais
capacidades compõem seu grafo. Essa extração não deve produzir módulos globais
por camada nem registrar dependências desnecessárias.

O onboarding é encapsulado por um `OnboardingScope` específico. A borda de
navegação cria um único scope por jornada, compartilha-o entre as etapas e o
encerra na saída definitiva. O scope bloqueia `get<T>()` depois de encerrado e
torna `dispose()` idempotente. Seu descarte atua somente no injector do
onboarding e nunca usa `disposeRecursive()`.

`RootContainer` inicializa primeiro o core e, quando necessário, o container da
aplicação. Isso permite compor core e onboarding sem iniciar o grafo operacional.
O router recebe `AppContainer` e a função de criação do `OnboardingScope`
explicitamente; não há injector global nem resolução de dependências dentro das
páginas. Ao encerrar, o root descarta scopes ativos, aplicação e core, nessa
ordem.

O `OnboardingUsecase` é singleton somente dentro de uma jornada e mantém o
rascunho corrente. O `OnboardingCoordinatorViewmodel` apenas inicializa esse
use case e expõe a etapa que a página raiz deve renderizar. Cada etapa possui
uma ViewModel própria, criada e descartada por sua página, e todas recebem o
mesmo use case do scope. As ViewModels não são registradas como singletons do
`OnboardingScope`.

As etapas permanecem dentro de uma única rota. O avanço ou retorno altera e
persiste a etapa no use case, e a página raiz atualiza a composição visual sem
criar outro scope. Controllers, focus nodes, códigos e senhas pertencem às
páginas. Dados persistidos pertencem ao storage do core e não são apagados pelo
encerramento do scope.

O rascunho tem validade de 24 horas desde a última alteração persistida. Nome,
e-mail, etapa, desafio e prova de verificação podem ser restaurados; código de
seis dígitos, senha, confirmação, loading e mensagens permanecem exclusivamente
na memória da UI. Consultar o rascunho pela Splash, cancelar a jornada ou
receber uma falha sem mudança de estado não renova essa validade. A conclusão da
criação da conta remove o rascunho; o simples descarte do scope não o remove.

A recuperação de senha também usa uma única rota e um scope próprio por jornada.
Esse scope recebe apenas o cliente HTTP e não expõe o storage do onboarding nem
o armazenamento de tokens de autenticação. E-mail, código, prova e nova senha
permanecem somente em memória, são descartados na saída e não aparecem em logs.
A conclusão retorna ao login sem criar uma sessão automaticamente.

Na etapa de e-mail, trocar o endereço invalida imediatamente o desafio e a
prova anteriores. A confirmação bem-sucedida permite avançar para a senha, e o
reenvio substitui integralmente o desafio. A UI deriva o contador de reenvio do
prazo persistido, sem armazenar o valor regressivo. A senha somente é enviada
para criação depois de atender aos requisitos locais e coincidir com sua
confirmação. Nenhum desses valores sensíveis é escrito em logs.

Não existe neste momento um framework genérico de scopes. Novas subdivisões
somente serão introduzidas após necessidades concretas. Da mesma forma, um use
case simples, como o `UsersUsecase` atual, não deve ser migrado ou mantido apenas
por simetria com os containers.

Os testes de composição devem demonstrar isolamento lateral, preservação de
core e app após o descarte, bloqueio de resolução no scope encerrado e criação
limpa de uma nova jornada.

## Data

`data` transforma operações do aplicativo em integração:

- API services conhecem paths, métodos, query, body e envelopes HTTP;
- DTOs representam contratos de request e response;
- adapters convertem modelos da aplicação em DTOs de request e DTOs de response
  em modelos da aplicação;
- repositories combinam APIs, armazenamento e cache;
- falhas são devolvidas como `Result`.

Dio deve permanecer dentro da implementação HTTP de `core`. API services usam
`RestClient`.

### Fronteira entre aplicação e API

DTOs pertencem exclusivamente a `data/services/apis` e não são expostos nas
interfaces públicas dos services. Repositories, use cases, view models e páginas
trabalham apenas com modelos ou valores da aplicação, mesmo quando estes possuem
os mesmos campos que um DTO.

```text
Aplicação → Repository → API service → Adapter → DTO/HTTP → RestClient
Aplicação ← Repository ← API service ← Adapter ← DTO/HTTP ← RestClient
```

O service coordena a chamada, valida o status, interpreta o envelope e converte
falhas para `AppError`. O adapter realiza somente a tradução entre os dois lados
da fronteira. A serialização e o parsing JSON permanecem nos DTOs. Adapters ficam
próximos ao respectivo service e são usados inclusive em conversões triviais,
mantendo a separação explícita e uniforme.

## Autenticação e sessão

O fluxo de autenticação separa transporte, coordenação e persistência:

```text
UI/ViewModel
    → AuthRepository
      → AuthService
        → AuthApiAdapter
          → DTO/HTTP
            → RestClient
              → Dio

AuthRepository ─┐
                ├→ AuthTokenStore → LocalSecureStorage
AuthInterceptor ┘
```

`AuthService` conhece os endpoints e coordena adapters e DTOs, mas sua interface
pública recebe e devolve somente modelos da aplicação e não armazena
credenciais.
`AuthRepository` coordena as chamadas e considera o login concluído somente
depois da gravação consistente dos dois tokens.

`restoreSession` usa o refresh token e retorna explicitamente sessão restaurada
ou ausente. Ausência da credencial e rejeição definitiva com `401` resultam em
sessão ausente. Timeout, falha de conexão, resposta `5xx`, parsing inválido ou
falha de storage permanecem como `Failure`, sem apagar uma credencial que ainda
pode ser utilizada em nova tentativa.

O `AuthInterceptor` lê o access token antes de requisições, compartilha um único
refresh entre chamadas concorrentes e repete a requisição original com o novo
Bearer. Uma marca interna impede que um segundo `401` provoque loop. O contrato
atual da API renova somente o access token; não há suporte a rotação do refresh
token.

No logout, a revogação remota é de melhor esforço e a limpeza local tem
prioridade. Se o dispositivo estiver sem conexão, o logout local ainda termina,
mas o refresh token poderá permanecer válido no servidor até sua expiração.

O mesmo singleton de `AuthTokenStore` é injetado no repository e no
interceptor. Chaves de storage ficam encapsuladas em `SecureAuthTokenStore`.

### Bootstrap da Splash

A Splash é a rota inicial e executa `BootstrapUsecase.initialize` por meio de
`SplashViewmodel.initialize`. O use case restaura a sessão e, quando ela não
existe, consulta o rascunho sem renovar sua validade. A precedência é explícita:
sessão restaurada conduz a `BootstrapDestination.home`; sem sessão, rascunho
válido conduz a `BootstrapDestination.onboarding`; ausência ou expiração do
rascunho conduz a `BootstrapDestination.login`. As rotas concretas e a
navegação permanecem na página.

A conclusão bem-sucedida substitui a Splash no histórico depois que a animação
visual inicial termina. Falhas de rede, timeout, storage e outras falhas
recuperáveis são exibidas assim que o comando termina, mantêm a página aberta e
oferecem nova tentativa sem limpar as credenciais locais.

## Domain

`domain/common` contém conceitos estáveis, por exemplo:

- usuário autenticado;
- entidade específica do domínio;
- valor ou enum de negócio;
- agregado introduzido pela aplicação.

`domain/usecases` contém coordenação de fluxos mais complexos. O domínio não
depende de widgets, navegação, Dio ou envelopes HTTP.

## UI

`ui` contém páginas, view models, componentes compartilhados e tema.

Páginas possuem:

- widgets e `BuildContext`;
- controllers, focus e form keys;
- navegação e feedback;
- observação de comandos.

View models possuem:

- repositories ou use cases injetados;
- comandos assíncronos;
- estado de apresentação independente de widgets.

## Erros e estados

`Result<T>` é o contrato para operações que podem falhar:

- `Success<T>` contém o valor;
- `Failure<T>` contém um `AppError`;
- `Unit` representa sucesso sem retorno significativo.

`Command` traduz o resultado para os estados `idle`, `running`, `success` e
`failure`. A UI deve tratar loading, vazio, sucesso e erro explicitamente.

## Configuração

Configurações de execução são fornecidas com `dart-define`:

```bash
flutter run \
  --dart-define=BASE_URL=http://10.0.2.2:8080 \
  --dart-define=APP_MODE=dev \
  --dart-define=AUTH_CLIENT_ID=foundry-stack-mobile
```

Arquivos por ambiente poderão ser adotados:

```bash
flutter run --dart-define-from-file=dev.env
```

Os testes de configuração usam fixtures próprias em `test/env/` e são
executados com `dart-define-from-file`:

```bash
make mobile-app-env-tests
```

Essas fixtures definem `AUTH_CLIENT_ID` e `EXPECTED_AUTH_CLIENT_ID`, para
validar os casos de string vazia, espaços em branco e valor com padding.

`AUTH_CLIENT_ID` identifica o cliente mobile, usa `foundry-stack-mobile` como padrão e
não é segredo. A allowlist de clientes autorizados pertence à API. Segredos
permanentes não devem ser embarcados no aplicativo.

## Decisões futuras

Ainda precisam de uma feature concreta antes da escolha:

- cache offline e sincronização;
- banco local;
- telemetria e crash reporting.

Ao adotar uma dessas decisões, atualizar este documento e os `AGENTS.md`
afetados.
