# AGENTS — UI

`ui` contém páginas, view models, tema e componentes visuais.

## Fluxos permitidos

```text
Page → ViewModel → Repository
Page → ViewModel → UseCase
```

A UI não acessa API services, RestClient, Dio ou storage.

## Páginas

Páginas possuem widgets, controllers, form keys, focus, navegação, dialogs e
feedback. Devem observar comandos e renderizar loading, vazio, sucesso e erro.

## View models

View models recebem repositories ou use cases por construtor e expõem
`Command0`, `Command1` e getters de apresentação. Não devem possuir
`BuildContext`, controllers, focus nodes ou widgets.

## Componentes

Componentes compartilhados devem ser visuais, configuráveis por valores e
callbacks e alinhados ao tema. Widgets específicos permanecem próximos da
página até existir reutilização real.

## Testes

- testes unitários para view models;
- widget tests para páginas e componentes;
- cobrir estados importantes e efeitos de comandos;
- não fazer rede real em testes de UI.
