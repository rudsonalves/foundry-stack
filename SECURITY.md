# Política de segurança

## Versões suportadas

Enquanto o projeto não publicar uma matriz de releases estáveis, apenas a
versão mais recente da branch `main` recebe correções de segurança.

## Reportando uma vulnerabilidade

Não abra uma issue pública com detalhes exploráveis. Use o recurso privado de
reporte de vulnerabilidades do repositório no GitHub. Inclua a área afetada,
passos de reprodução, impacto esperado e, quando possível, uma mitigação.

Não inclua tokens, senhas, dados pessoais ou credenciais reais no relatório.

## Escopo operacional

O FoundryStack é uma fundação e precisa ser revisado para o ambiente em que for
implantado. Em particular, configure segredos próprios, TLS, CORS, SMTP,
observabilidade, backups e políticas de retenção. Os rate limiters em memória
são adequados a uma única instância e precisam de armazenamento compartilhado
antes de escalar horizontalmente.
