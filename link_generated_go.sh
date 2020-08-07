#! /usr/bin/env bash
# Temporary script to create links for every file in the repo. These links are
# ignored by gazelle so having them around is no big deal but until
# https://github.com/bazelbuild/rules_go/issues/512 is fixed it's a huge PITA
# to not have code completion working and this should help.
# From
# https://gist.githubusercontent.com/kalbasit/3c9607333f5f3c794d2bd96114184301/raw/07c878cca4913a9592462866c66da32179993a1f/proto.sh

set -euo pipefail
OS="$(go env GOHOSTOS)"
ARCH="$(go env GOARCH)"
echo -e ">>> Compiling the Go proto"

for label in $(bazelisk query 'kind(go_proto_library, //...)'); do
	package="${label%%:*}"
	package="${package##//}"
	target="${label##*:}"
	# do not continue if the package does not exist
	[[ -d "${package}" ]] || continue
	# compute the path where bazel put the files
	# TODO: the _static_pure_stripped comes from a .bazelrc option `build --features=pure`, `build --features=static`.
	#out_path="bazel-bin/${package}/${OS}_${ARCH}_static_pure_stripped/${target}%/github.com/org/repo/${package}"
	out_path="bazel-bin/${package}/${target}_/github.com/bazelbuild/bazel-gazelle/${package}"
	# compute the relative_path to the
	count_paths="$(echo -n "${package}" | tr '/' '\n' | wc -l)"
	relative_path=""
	for i in $(seq 0 ${count_paths}); do
		relative_path="../${relative_path}"
	done
	bazelisk build -k "${label}" || true
	found=0
	for f in ${out_path}/*.pb.go; do
		if [[ -f "${f}" ]]; then
			found=1
      echo "Linkifying ${relative_path}${f}"
			ln -nsf "${relative_path}${f}" "${package}/"
		fi
	done
	for f in ${out_path}/*.pb.gw.go; do
		if [[ -f "${f}" ]]; then
			found=1
      echo "Linkifying ${relative_path}${f}"
			ln -nsf "${relative_path}${f}" "${package}/"
		fi
	done
	if [[ "${found}" == "0" ]]; then
		echo "ERR: no .pb.go file was found inside $out_path for the package ${package}"
		exit 1
	fi
done
