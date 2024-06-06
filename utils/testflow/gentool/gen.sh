#!/bin/bash

# Enable error detection, the script will terminate if any command returns a non-zero status
set -e

if [ "$#" -ne 4 ]; then
    echo "Usage: $0 <sample_dir> <patch_dir> <backup_dir> <output_dir>"
    exit 1
fi

sample_dir="$1"
patch_dir="$2"
backup_dir="$3"
output_dir="$4"

if [ -z "$sample_dir" ]; then
    echo "Input directory is required."
    exit 1
fi

if [ -z "$patch_dir" ]; then
    echo "Patch directory is required."
    exit 1
fi

if [ -z "$backup_dir" ]; then
    echo "Backup directory is required."
    exit 1
fi

if [ -z "$output_dir" ]; then
    echo "Output directory is required."
    exit 1
fi

mkdir -p "${patch_dir}"
mkdir -p "${backup_dir}"
mkdir -p "${output_dir}"

# Get the directory of the script
script_dir=$(dirname "$(realpath "$0")")

# Build the temporary binary file
temp_dir=$(mktemp -d)
trap 'rm -rf "$temp_dir"' EXIT
temp_bin=$temp_dir/testflow
(cd "$script_dir" && go build -o "$temp_bin" ./*.go)

# Function: Convert a string to snake_case
to_snake_case() {
    echo "$1" | sed -r 's/([A-Z])/_\1/g' | tr '[:upper:]' '[:lower:]' | sed 's/^_//'
}

# Iterate over all .json and .chlsj files in the specified directory
for sample_file in "${sample_dir}"/*.{json,chlsj}; do
    if [[ -f "$sample_file" ]]; then
        # Get the file name (without the path)
        filename=$(basename -- "$sample_file")
        # Get the file name without the extension
        base_name="${filename%.*}"

        snake_case_name=$(to_snake_case "$base_name")
        output_file="${output_dir}/flow_${snake_case_name}_test.go"
        patch_file="${patch_dir}/flow_${snake_case_name}_test.origin"
        if ! "$temp_bin" --sample-file="$sample_file" --output-file="$output_file" --patch-file="$patch_file" --backup-dir="$backup_dir"; then
            echo "Error processing $sample_file. Stopping execution."
            exit 1
        fi

        echo "Processed $sample_file -> $output_file with $patch_file"
    fi
done

echo "All files processed."