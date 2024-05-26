#!/bin/bash

# 开启错误检测，任何命令的返回非零状态时脚本将终止
set -e

# 检查参数数量
if [ "$#" -ne 4 ]; then
    echo "Usage: $0 <sample_dir> <patch_dir> <backup_dir> <output_dir>"
    exit 1
fi

sample_dir="$1"
patch_dir="$2"
backup_dir="$3"
output_dir="$4"

# 检查样本目录是否为空
if [ -z "$sample_dir" ]; then
    echo "Input directory is required."
    exit 1
fi

# 检查补丁目录是否为空
if [ -z "$patch_dir" ]; then
    echo "Patch directory is required."
    exit 1
fi

# 检查备份目录是否为空
if [ -z "$backup_dir" ]; then
    echo "Backup directory is required."
    exit 1
fi

# 检查输出目录是否为空
if [ -z "$output_dir" ]; then
    echo "Output directory is required."
    exit 1
fi

# 确保补丁目录存在
mkdir -p "${patch_dir}"

# 确保备份目录存在
mkdir -p "${backup_dir}"

# 确保输出目录存在
mkdir -p "${output_dir}"

# 获取脚本所在目录
script_dir=$(dirname "$(realpath "$0")")

# 临时二进制文件
temp_bin=$(mktemp -d)/testflow

# 构建二进制文件
(cd "$script_dir" && go build -o "$temp_bin" ./*.go)

# 函数：转换字符串为蛇形命名
to_snake_case() {
    echo "$1" | sed -r 's/([A-Z])/_\1/g' | tr '[:upper:]' '[:lower:]' | sed 's/^_//'
}

# 遍历指定目录下的所有 .json 和 .chlsj 文件
for sample_file in "${sample_dir}"/*.{json,chlsj}; do
    if [[ -f "$sample_file" ]]; then
        # 获取文件名（不包含路径）
        filename=$(basename -- "$sample_file")
        # 获取不带扩展名的文件名
        base_name="${filename%.*}"

        # 转换为蛇形命名
        snake_case_name=$(to_snake_case "$base_name")

        # 构造输出文件路径
        output_file="${output_dir}/flow_${snake_case_name}_test.go"

        # 构造补丁文件路径
        patch_file="${patch_dir}/flow_${snake_case_name}_test.origin"

        # 执行命令并检查错误
        if ! "$temp_bin" --sample-file="$sample_file" --output-file="$output_file" --patch-file="$patch_file" --backup-dir="$backup_dir"; then
            echo "Error processing $sample_file. Stopping execution."
            exit 1
        fi

        # 打印状态消息
        echo "Processed $sample_file -> $output_file with $patch_file"
    fi
done

echo "All files processed."