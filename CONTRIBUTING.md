# Contribuindo

Obrigado por contribuir com o FoundryStack. Mudanças devem manter o starter
genérico, funcional e simples de personalizar.

## Preparação

1. Instale as versões descritas no `README.md`.
2. Execute `make env-init` para criar configurações locais ignoradas pelo Git.
3. Execute `make check` antes de iniciar uma mudança.

## Diretrizes

- mantenha API e mobile alinhados quando um contrato HTTP mudar;
- altere o banco somente por migrations versionadas;
- acompanhe mudanças de comportamento com testes e documentação;
- preserve as fronteiras arquiteturais descritas nos arquivos `AGENTS.md`;
- não inclua segredos, arquivos `.env` reais ou dados pessoais;
- evite dependências e abstrações sem necessidade concreta;
- não substitua as licenças e avisos dos assets por MIT.

## Verificação

Execute, conforme a área modificada:

```bash
make api-check
make mobile-check
make mobile-app-env-tests
git diff --check
```

Mudanças em migrations também devem ser verificadas contra um banco vazio.
Mudanças de plataforma devem incluir um build Android e, em macOS, iOS sem
assinatura.

## Pull requests

Descreva o problema, a solução, os testes executados e qualquer impacto de
configuração ou migration. Mantenha cada pull request focado em uma mudança
coerente.
