BUILD_IMAGE = pesto
PREFIX = packages.dns.ad.zopa.com:5002/${BUILD_IMAGE}
VERSION := $(shell versioner VERSION 2>/dev/null || echo `cat VERSION`dev)

build: get-zopa-certs
	docker build \
	-t $(PREFIX):$(VERSION) \
	--build-arg app_version=$(VERSION) \
	--build-arg github_access_key=$(GO_GHE_ACCESS_KEY) \
	-f Dockerfile \
	.

get-zopa-certs:
	mkdir -p .zopa_certs
	curl -skS -L https://packages.dns.ad.zopa.com/artifactory/tools-dev-local/zopa_certs.tar | tar xvf - -C .zopa_certs

publish: build tag push

push:
	@echo "***Pushing git tags***"
	git push --tags
	@echo "***Pushing docker image***"
	docker push $(PREFIX):$(VERSION)

tag:
	@echo "***Tagging git $(VERSION)***"
	git tag v$(VERSION)
	@echo "***Tagging docker image***"
	docker tag $(PREFIX):$(VERSION) $(PREFIX):latest

version:
	@echo $(VERSION)