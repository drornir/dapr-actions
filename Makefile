default: help

SHELL=/bin/zsh

include .bingo/Variables.mk
DAPR := $(shell brew --prefix)/bin/dapr # TODO


KIND_CONFIG_FILE := './kind/default.yaml'

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk commands is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

##@ General

.PHONY: help
help: ## shows this message
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<command>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.PHONY: help-bingo
help-bingo: ## shows help about dev-deps
	@cat ./.bingo/README.md

##@ Development

.PHONY: build
build: ## builds
	echo todo

.PHONY: fmt
fmt: ## format
	$(GOFUMPT) -l -extra .

.PHONY: dev-deps
dev-deps: $(BINGO) ## installs dev-deps and links them for local use (see below)
	$(BINGO) get -l

.PHONY: kind-create
kind-create: $(KIND) ## create the kind cluster
	$(KIND) create cluster --config ${KIND_CONFIG_FILE}

.PHONY: kind-delete
kind-delete: $(KIND) ## delete the kind cluster
	kind delete cluster -n dapr-actions

.PHONY: kctx
kctx: ## set kubectl context to kind
	kubectl config use-context kind-dapr-actions

.PHONY: dapr-init
dapr-init: ## init dev cluster
	$(DAPR) init --kubernetes --dev

.PHONY: dapr-dash
dapr-dash: ## opens dapr dashboard on port 9999
	$(DAPR) dashboard -k -p 9999

.PHONY: dapr-status
dapr-status: ## get status of dapr in the cli
	$(DAPR) status -k
