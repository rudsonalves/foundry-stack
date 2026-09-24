SHELL := /bin/bash
.DEFAULT_GOAL := help

API_DIR ?= api
MOBILE_DIR ?= mobile
FLUTTER ?= flutter
PR ?=
BRANCH ?=

.PHONY: help check api-check mobile-check mobile-app-env-tests env-init mobile-sync-ip status push pull gitlog pullmain \
	gh-pr-create gh-pr-view gh-pr-checkout dev stag prod test colima-up db-up \
	db-setup migrate-up

help: ## Exibe os comandos de gestão do monorepo
	@awk 'BEGIN {FS = ":.*##"; printf "\nComandos:\n"} /^[a-zA-Z0-9_.-]+:.*##/ { printf "  %-18s %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

env-init: ## Cria e sincroniza os ambientes locais da API e do mobile
	@bash infra/scripts/ensure-env-files.sh

mobile-sync-ip: env-init ## Atualiza a BASE_URL mobile com o IP local desta máquina
	@bash infra/scripts/update-mobile-env-ip.sh

check: api-check mobile-check ## Valida API e mobile disponível

api-check: ## Formata, analisa e testa a API
	@$(MAKE) -C $(API_DIR) check

mobile-check: ## Analisa e testa o novo aplicativo Flutter quando inicializado
	@if [ ! -f "$(MOBILE_DIR)/pubspec.yaml" ]; then \
		echo "Mobile ainda não inicializado; validação ignorada."; \
	else \
		cd "$(MOBILE_DIR)" && $(FLUTTER) analyze && $(FLUTTER) test; \
	fi

mobile-app-env-tests: ## Testa as configurações de ambiente do aplicativo
	@cd "$(MOBILE_DIR)" && \
	for envfile in test/env/app_env_empty.env test/env/app_env_whitespace.env test/env/app_env_padded.env; do \
		$(FLUTTER) test test/core/resources/app_env_test.dart --dart-define-from-file="$$envfile"; \
	done

status: ## Exibe o estado do repositório
	@git status --short

push: ## Publica a branch atual ou BRANCH=nome
	@branch=$${BRANCH:-$$(git branch --show-current)}; \
	if [ -z "$$branch" ]; then echo "Erro: branch não encontrada"; exit 1; fi; \
	git push origin "$$branch"

pull: ## Atualiza a branch atual ou BRANCH=nome
	@branch=$${BRANCH:-$$(git branch --show-current)}; \
	if [ -z "$$branch" ]; then echo "Erro: branch não encontrada"; exit 1; fi; \
	git pull origin "$$branch"

gitlog: ## Exibe o histórico resumido
	@git log --oneline

pullmain: ## Atualiza a branch main a partir de origin/main
	@git switch main
	@git pull origin main

gh-pr-create: ## Cria um pull request da branch atual
	@gh pr create --fill --base main

gh-pr-view: ## Abre o pull request da branch atual
	@gh pr view --web

gh-pr-checkout: ## Seleciona um pull request (PR=123)
	@if [ -z "$(PR)" ]; then echo "Erro: informe PR=123"; exit 1; fi
	@gh pr checkout $(PR)

dev: env-init ## Executa a API em desenvolvimento
	@$(MAKE) -C $(API_DIR) dev

stag: ## Executa a API em staging
	@$(MAKE) -C $(API_DIR) stag

prod: ## Executa a API em produção
	@$(MAKE) -C $(API_DIR) prod

test: ## Executa os testes da API
	@$(MAKE) -C $(API_DIR) test

colima-up: ## Inicia o ambiente Docker local
	@$(MAKE) -C $(API_DIR) colima-up

db-up: env-init ## Inicia o PostgreSQL
	@$(MAKE) -C $(API_DIR) db-up

db-setup: env-init ## Inicia o PostgreSQL e aplica migrations
	@$(MAKE) -C $(API_DIR) db-setup

migrate-up: env-init ## Aplica migrations pendentes
	@$(MAKE) -C $(API_DIR) migrate-up
