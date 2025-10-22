SHELL := /bin/bash
.DEFAULT_GOAL := help

# Allow to pass args to targets
cmd := $(firstword $(MAKECMDGOALS))
args := $(wordlist 2,$(words $(MAKECMDGOALS)),$(MAKECMDGOALS))
%::
	@ true
# (c) https://askubuntu.com/a/1448109

##@ Targets

build: ## Build sandbox environment init image
	@ docker-compose build
.PHONY: build

up: ## Create sandbox environment, e.g. make up, make up vault
	@ mkdir -p ./vol/kind/pki ./vol/kind/kube
	@
	@ echo -e "# Create kind cluster\n"
	@ (kind get clusters | grep -q "vui-sandbox") || kind create cluster --config=./kind.yml || true
	@
	@ echo -e "\n# Export kind kubeconfig\n"
	@ kind export kubeconfig --name vui-sandbox --kubeconfig ./vol/kind/kube/config || true
	@
	@ echo -e "\n# Apply kubernetes resources\n"
	@ kubectl --context kind-vui-sandbox apply -f ./files/kubernetes.yaml || true
	@
	@ echo -e "\n# Start docker containers\n"
	@ docker-compose up -d $(args)
.PHONY: up

logs: ## Show logs for sandbox environment, e.g. make logs, make logs vault
	@ docker-compose logs -f $(args)
.PHONY: logs

ps: ## Show sandbox services
	@ docker-compose ps --format "table {{.Service}}\t{{.Ports}}\t{{.Status}}" -a --orphans
.PHONY: ps

exec: ## Execute command in sandbox environment, e.g. make exec vault
	@ docker-compose exec -it $(args) sh
	@ docker-compose exec -it $(args) sh
.PHONY: exec

env: export AWS_CONFIG_FILE=./cfg/aws
env: ## Export sandbox environment variables
	@ echo "" > .env
	@
	@ echo "# Export Vault AWS environment variables"
	@ aws configure export-credentials --profile vui-iam-role --format env >> .env
	@
	@ echo "# Export Vault LDAP environment variables"
	@ echo "export LDAP_USERNAME=testuser" >> .env
	@ echo "export LDAP_PASSWORD=testpassword" >> .env
	@
	@ echo "# Export Vault Token environment variables"
	@ echo "export VAULT_TOKEN=vui-sandbox-token" >> .env
	@
	@ echo "# Switch kind context"
	@ kubectl config use-context kind-vui-sandbox > /dev/null || true
	@
	@ echo "# Export JWT"
	@ echo "export JWT=$$(curl -s -X POST 'http://oidc:8080/realms/vui-sandbox-realm/protocol/openid-connect/token' \
			-H 'Content-Type: application/x-www-form-urlencoded' \
			-d 'client_id=vui-sandbox-oidc-client-id' \
			-d 'client_secret=vui-sandbox-oidc-client-secret' \
			-d 'username=vui' \
			-d 'password=vui' \
			-d 'grant_type=password' \
			-d 'scope=email profile' | jq -r '.access_token')" >> .env
	@ 
	@ echo "# Export UserPass credentials"
	@ echo "export USERPASS_USERNAME=vui" >> .env
	@ echo "export USERPASS_PASSWORD=vui" >> .env
	@
	@ echo "# OIDC Credentials"
	@ echo "  - Username: vui"
	@ echo "  - Password: vui"
	@
	@ sleep 1
.PHONY: env

down: ## Destroy sandbox environment
	@ echo -e "# Delete kind cluster\n"
	@ kind delete cluster --name vui-sandbox || true
	@
	@ echo -e "\n# Delete docker containers\n"
	@ docker-compose down
.PHONY: down

clean: ## Clean temporary files
	@ rm -f .env
	@ rm -rf ./vol/
.PHONY: clean


help: ## Show help message
	@awk 'BEGIN { \
		FS = ":.*##"; \
		printf "Usage:\n  make \033[36m<target>\033[0m\n" \
		} \
		/^[a-zA-Z0-9_-]+:.*?##/ { \
			desc = $$2; \
			if (match(desc, / e\.g\. /)) { \
				before = substr(desc, 1, RSTART - 1); \
				after = substr(desc, RSTART); \
				printf "  \033[36m%-20s\033[0m %s\033[32m%s\033[0m\n", $$1, before, after \
			} else { \
				printf "  \033[36m%-20s\033[0m %s\n", $$1, desc \
			} \
		} \
		/^##@/ { \
			printf "\n\033[1m%s\033[0m\n", substr($$0, 5) \
		}' $(MAKEFILE_LIST)
.PHONY: help
