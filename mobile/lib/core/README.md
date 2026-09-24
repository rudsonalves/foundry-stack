# Core

Infraestrutura compartilhada do aplicativo Flutter.

Esta base contém apenas elementos compartilhados e independentes de domínio:

- `result/`: resultados, erros e comandos assíncronos;
- `resources/`: configuração de ambiente e cabeçalhos HTTP;
- `services/client_http/`: contrato HTTP e implementação com Dio;
- `services/logging/`: logging para desenvolvimento.

O `core` não deve depender das camadas de dados, domínio ou interface. Rotas,
autenticação, injeção de dependência e regras específicas da aplicação devem ser
adicionadas quando os respectivos fluxos forem definidos.
