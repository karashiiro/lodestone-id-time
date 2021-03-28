SHELL :=/bin/bash
.DEFAULT_GOAL := help 

help: ## Show this help
	@echo Dependencies: go python
	@egrep -h '\s##\s' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

install: ## Install Python dependencies
	python -m pip install -r ./requirements.txt

build_scraper: ## Build the Lodestone scraper application
	go build -ldflags="-s -w"

estimate: ## Run the estimation script (requires converting the date column to a number column)
	python ./estimate.py
