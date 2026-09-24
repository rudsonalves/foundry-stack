# AGENTS — Componentes compartilhados

Esta pasta contém widgets, formatters, feedback e tema reutilizados por mais de
uma tela.

## Um componente compartilhado deve

- ser apenas de apresentação;
- receber dados e callbacks;
- usar `Theme.of(context)`, `ColorScheme` e `TextTheme`;
- possuir API pequena e explícita;
- tratar estados relevantes, como disabled ou loading;
- ser testável isoladamente.

## Manter na página quando

- o widget é usado por uma única tela;
- o nome depende de uma feature;
- contém regras ou textos específicos do fluxo;
- a extração apenas reduziria o tamanho de um arquivo.

## Não fazer

- chamar repository, API, RestClient ou storage;
- depender de view model de página;
- esconder navegação ou efeitos colaterais;
- criar um sistema visual paralelo ao tema;
- generalizar antes de existir reutilização.
