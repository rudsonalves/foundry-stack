# Personalizando o starter

Faça a personalização antes de publicar o primeiro build ou emitir tokens em
um ambiente persistente.

## Identidade

Atualize de forma coordenada:

- `FoundryStack`: nome exibido, documentação e e-mails;
- `foundry-stack`: binário, Compose, issuer e nomes operacionais;
- `github.com/rudsonalves/foundry-stack/api`: módulo e imports Go;
- `foundry_stack_mobile`: nome do pacote Dart;
- `br.dev.rralves.foundrystack`: Android application ID e iOS bundle ID;
- `foundry-stack-mobile` e `foundry-stack-swagger`: clientes JWT;
- `foundry_stack`: banco e usuários PostgreSQL.

Depois, mova `MainActivity.kt` para o diretório correspondente ao novo package
Kotlin e atualize os bundle IDs do target principal e dos testes iOS.

## Marca e assets

Substitua ícones, launch images, cores e textos visíveis. As fontes incluídas
podem ser removidas ou trocadas, mas seus arquivos SIL OFL e avisos devem
acompanhar os assets enquanto eles permanecerem no projeto.

## Configuração

Gere segredos exclusivos por ambiente. Revise domínios, CORS, URLs do mobile,
SMTP, proxies confiáveis, TTLs e limites antes da implantação. Não coloque
segredos em arquivos `.example`, no código ou em argumentos versionados.

## Removendo módulos

Para remover autenticação ou usuários, retire o registro de rotas e a composição
em `api/cmd`, depois remova migrations e módulos somente se nenhum ambiente
persistente depender deles. No mobile, remova primeiro rotas e bindings, depois
repositories, services, casos de uso e telas. Atualize OpenAPI e testes na mesma
mudança.

## Adicionando módulos

Crie módulos coesos e específicos do novo produto. Preserve o envelope HTTP,
o tratamento de erros e as fronteiras existentes quando forem adequados, mas
não generalize a infraestrutura antecipadamente.
