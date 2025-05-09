#/bin/bash

if [ -z "$1" ] || [ -z "$2" ] || [ -z "$3" ]; then
    echo "Usage: $0 <build_dir> <name> <ldflags>"
    exit 1
fi

# "darwin/amd64" "darwin/arm64" are not supported by uclipboard
GOOSARCHS=("linux/amd64" "linux/arm64" "windows/amd64" "windows/386" "darwin/amd64" "darwin/arm64")

build_dir=$1
name=$2
ldflags=$3
mkdir -p $build_dir
for osarch in "${GOOSARCHS[@]}"; do
    IFS=/ read -r GOOS GOARCH <<< "$osarch"
    output_name="${name}-${GOOS}-${GOARCH}"

    full_name=$build_dir/$output_name

    if [ "$GOOS" == "windows" ]; then
        full_name+=".exe"
    fi
    echo "Building $full_name with $ldflags"
    GOOS=$GOOS GOARCH=$GOARCH go build -o $full_name  -ldflags="$ldflags" .
done

