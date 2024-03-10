#!/bin/bash
#      Warning!!!!!!!!!!!This file is readonly!Don't modify this file!

cd $(dirname $0)

help() {
	echo "cmd.sh — every thing you need"
	echo "         please install git"
	echo "         please install golang(1.21+)"
	echo "         please install protoc           (github.com/protocolbuffers/protobuf)"
	echo "         please install protoc-gen-go    (github.com/protocolbuffers/protobuf-go)"
	echo "         please install codegen          (github.com/chenjie199234/Corelib)"
	echo ""
	echo "Usage:"
	echo "   ./cmd.sh <option>"
	echo ""
	echo "Options:"
	echo "   pb                        Generate the proto in this program."
	echo "   sub <sub service name>    Create a new sub service."
	echo "   kube                      Update kubernetes config."
	echo "   html                      Create html template."
	echo "   h/-h/help/-help/--help    Show this message."
}

pb() {
	rm ./api/*.pb.go
	rm ./api/*.md
	rm ./api/*.ts
	go mod tidy
	codegen -update
	corelib=$(go list -m -f "{{.Dir}}" github.com/chenjie199234/Corelib)
	protoc -I ./ -I $corelib --go_out=paths=source_relative:. ./api/*.proto
	protoc -I ./ -I $corelib --go-pbex_out=paths=source_relative:. ./api/*.proto
	protoc -I ./ -I $corelib --go-cgrpc_out=paths=source_relative:. ./api/*.proto
	protoc -I ./ -I $corelib --go-crpc_out=paths=source_relative:. ./api/*.proto
	protoc -I ./ -I $corelib --go-web_out=paths=source_relative:. ./api/*.proto
	protoc -I ./ -I $corelib --browser_out=paths=source_relative,gen_tob=true:. ./api/*.proto
	protoc -I ./ -I $corelib --markdown_out=paths=source_relative:. ./api/*.proto
	go mod tidy
}

sub() {
	go mod tidy
	codegen -update
	codegen -n im -p github.com/chenjie199234/im -sub $1
}

kube() {
	go mod tidy
	codegen -update
	codegen -n im -p github.com/chenjie199234/im -kube
}

html() {
	go mod tidy
	codegen -update
	codegen -n im -p github.com/chenjie199234/im -html
}

if !(type git >/dev/null 2>&1);then
	echo "missing dependence: git"
	exit 0
fi

if !(type go >/dev/null 2>&1);then
	echo "missing dependence: golang"
	exit 0
fi

if !(type protoc >/dev/null 2>&1);then
	echo "missing dependence: protoc"
	exit 0
fi

if !(type protoc-gen-go >/dev/null 2>&1);then
	echo "missing dependence: protoc-gen-go"
	exit 0
fi

if !(type codegen >/dev/null 2>&1);then
	echo "missing dependence: codegen"
	exit 0
fi

if [[ $# == 0 ]] || [[ "$1" == "h" ]] || [[ "$1" == "help" ]] || [[ "$1" == "-h" ]] || [[ "$1" == "-help" ]] || [[ "$1" == "--help" ]]; then
	help
	exit 0
fi

if [[ "$1" == "pb" ]];then
	pb
	exit 0
fi

if [[ "$1" == "kube" ]];then
	kube
	exit 0
fi

if [[ "$1" == "html" ]];then
	html
	exit 0
fi

if [[ $# == 2 ]] && [[ "$1" == "sub" ]];then
	sub $2
	exit 0
fi

echo "option unsupport"
help