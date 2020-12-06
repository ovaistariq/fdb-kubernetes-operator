
# Image URL to use all building/pushing image targets
IMG ?= fdb-kubernetes-operator:latest

# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true,maxDescLen=0"

CONTROLLER_GEN_VERSION ?= 0.2.4

ifneq "$(FDB_WEBSITE)" ""
	docker_build_args := $(docker_build_args) --build-arg FDB_WEBSITE=$(FDB_WEBSITE)
endif

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

GO_SRC=$(shell find . -name "*.go" -not -name "zz_generated.*.go")
GENERATED_GO=api/v1beta1/zz_generated.deepcopy.go
GO_ALL=${GO_SRC} ${GENERATED_GO}
MANIFESTS=config/crd/bases/apps.foundationdb.org_foundationdbbackups.yaml config/crd/bases/apps.foundationdb.org_foundationdbclusters.yaml config/crd/bases/apps.foundationdb.org_foundationdbrestores.yaml
CONTROLLER_GEN=$(GOBIN)/controller-gen

all: generate fmt vet manager manifests samples documentation test_if_changed

.PHONY: clean all manager samples documentation run install uninstall deploy manifests fmt vet generate docker-build docker-push rebuild-operator bounce lint

clean:
	find config/crd/bases -type f -name "*.yaml" -delete
	find api -type f -name "zz_generated.*.go" -delete
	mkdir -p bin
	rm -r bin
	find config/samples -type f -name deployment.yaml -delete
	find . -name "cover.out" -delete

# Run tests
test:
ifneq "$(SKIP_TEST)" "1"
	go test ./... -coverprofile cover.out
endif

test_if_changed: cover.out

cover.out: ${GO_ALL} ${MANIFESTS}
ifneq "$(SKIP_TEST)" "1"
	go test ./... -coverprofile cover.out -tags test
endif

# Build manager binary
manager: bin/manager

bin/manager: ${GO_SRC}
	go build -o bin/manager main.go

# Build kubectl-fdb binary
plugin: bin/kubectl-fdb

bin/kubectl-fdb: ${GO_SRC}
	go build -o bin/kubectl-fdb ./kubectl-fdb

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate manifests
	go run ./main.go

# Install CRDs into a cluster
install: manifests
	kustomize build config/crd | kubectl apply -f -

# Uninstall CRDs from a cluster
uninstall: manifests
	kustomize build config/crd | kubectl delete -f -

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: install manifests
	cd config/manager && kustomize edit set image controller=${IMG}
	kustomize build config/default | kubectl apply -f -

# Generate manifests e.g. CRD, RBAC etc.
manifests: ${MANIFESTS}

${MANIFESTS}: ${CONTROLLER_GEN} ${GO_SRC}
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases

# Run go fmt against code
fmt: bin/fmt_check

bin/fmt_check: ${GO_ALL}
	gofmt -w -s .
	mkdir -p bin
	@touch bin/fmt_check

# Run go vet against code
vet: bin/vet_check

bin/vet_check: ${GO_ALL}
	go vet ./...
	mkdir -p bin
	@touch bin/vet_check

# Generate code
generate: ${GENERATED_GO}

${GENERATED_GO}: ${GO_SRC} hack/boilerplate.go.txt ${CONTROLLER_GEN}
	$(CONTROLLER_GEN) object:headerFile=./hack/boilerplate.go.txt paths="./..."

# Build the docker image
docker-build: test_if_changed
	docker build ${docker_build_args} . -t ${IMG}

# Push the docker image
docker-push:
	docker push ${IMG}

# Rebuilds, deploys, and bounces the operator
rebuild-operator: docker-build deploy bounce

bounce:
	kubectl delete pod -l app=fdb-kubernetes-operator-controller-manager

samples: config/samples/deployment.yaml

config/samples/deployment.yaml: config/samples/deployment/*.yaml
	kustomize build config/samples/deployment > config/samples/deployment.yaml

bin/po-docgen: cmd/po-docgen/*.go
	go build -o bin/po-docgen cmd/po-docgen/main.go  cmd/po-docgen/api.go

docs/cluster_spec.md: bin/po-docgen api/v1beta1/foundationdbcluster_types.go
	bin/po-docgen api api/v1beta1/foundationdbcluster_types.go > docs/cluster_spec.md

docs/backup_spec.md: bin/po-docgen api/v1beta1/foundationdbbackup_types.go
	bin/po-docgen api api/v1beta1/foundationdbbackup_types.go > docs/backup_spec.md

docs/restore_spec.md: bin/po-docgen api/v1beta1/foundationdbrestore_types.go
	bin/po-docgen api api/v1beta1/foundationdbrestore_types.go > docs/restore_spec.md

documentation: docs/cluster_spec.md docs/backup_spec.md docs/restore_spec.md

lint:
	golangci-lint run ./...

# find or download controller-gen
# download controller-gen if necessary
controller-gen: ${CONTROLLER_GEN}

${CONTROLLER_GEN}:
	@{ \
	set -e ;\
	CONTROLLER_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$CONTROLLER_GEN_TMP_DIR ;\
	go mod init tmp ;\
	go get sigs.k8s.io/controller-tools/cmd/controller-gen@v${CONTROLLER_GEN_VERSION} ;\
	rm -rf $$CONTROLLER_GEN_TMP_DIR ;\
	}
