# AGENTS — Bootstrap e configuração

`bootstrap` interpreta opções de execução, carrega ambientes e valida a
configuração da API.

## Regras

- Toda configuração obrigatória deve falhar cedo com mensagem clara.
- Segredos não possuem valor padrão.
- Durações, URLs, portas e listas devem ser validadas após o parsing.
- Arquivos `.env` apoiam desenvolvimento; variáveis de processo continuam
  sendo a interface de configuração.
- Não registrar valores secretos.
- Não acoplar configuração a handlers ou repositories concretos.
- Novas variáveis devem atualizar exemplos de ambiente e testes.
- Preservar diferenças explícitas entre dev, staging, prod e test.

## Testes

Cobrir valor ausente, valor inválido, defaults seguros, precedência e seleção
de ambiente. Testes não devem depender do ambiente real da máquina.
