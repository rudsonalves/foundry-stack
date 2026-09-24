# AGENTS — FoundryStack

Este arquivo orienta a evolução do monorepo do FoundryStack.

## Objetivo

O FoundryStack é um starter full-stack opinativo e reutilizável para aplicações
com API em Go, PostgreSQL e aplicativo Flutter. A base oferece contas,
autenticação, verificação de e-mail e recuperação de senha sem impor regras de
um produto específico.

## Organização

- `api/`: API HTTP em Go, PostgreSQL e migrations;
- `mobile/`: aplicativo Flutter;
- `docs/`: arquitetura, configuração e personalização do starter.

## Diretrizes

- preservar baixo custo operacional e simplicidade;
- manter responsabilidades de API e mobile separadas;
- evoluir a API como monólito modular;
- alterar o banco exclusivamente por migrations versionadas;
- manter os módulos de fundação independentes do domínio da aplicação criada;
- evitar abstrações de framework sem um caso de uso concreto;
- atualizar documentação e testes junto com mudanças de comportamento.

## Fonte de verdade

1. `README.md`;
2. `docs/architecture.md`;
3. contratos e documentação dentro de `api/`;
4. código e testes executáveis.
