#!/bin/bash

# 跨平台打包脚本
# 为不同操作系统和架构编译 sim-sms-forward

set -e

# 项目信息
PROJECT_NAME="sim-sms-forward"
VERSION="v1.0.0"
BUILD_TIME=$(date +"%Y-%m-%d %H:%M:%S")
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# 输出目录
OUTPUT_DIR="dist"
BINARY_NAME="sim-sms-forward"

# 清理之前的构建
echo "清理之前的构建文件..."
rm -rf ${OUTPUT_DIR}
mkdir -p ${OUTPUT_DIR}

# 构建信息
LDFLAGS="-s -w"
LDFLAGS="${LDFLAGS} -X 'main.Version=${VERSION}'"
LDFLAGS="${LDFLAGS} -X 'main.BuildTime=${BUILD_TIME}'"
LDFLAGS="${LDFLAGS} -X 'main.GitCommit=${GIT_COMMIT}'"

# 支持的平台和架构
declare -a PLATFORMS=(
    "linux/amd64"
    "linux/arm64" 
    "linux/arm"
    "windows/amd64"
    "windows/arm64"
    "darwin/amd64"
    "darwin/arm64"
    "freebsd/amd64"
    "openbsd/amd64"
)

echo "开始构建 ${PROJECT_NAME} ${VERSION}"
echo "构建时间: ${BUILD_TIME}"
echo "Git 提交: ${GIT_COMMIT}"
echo "======================================"

# 编译各平台版本
for platform in "${PLATFORMS[@]}"; do
    IFS='/' read -r GOOS GOARCH <<< "$platform"
    
    # 设置输出文件名
    output_name="${BINARY_NAME}-${GOOS}-${GOARCH}"
    if [ "$GOOS" = "windows" ]; then
        output_name="${output_name}.exe"
    fi
    
    output_path="${OUTPUT_DIR}/${output_name}"
    
    echo "构建 ${GOOS}/${GOARCH}..."
    
    # 编译
    env GOOS="$GOOS" GOARCH="$GOARCH" go build \
        -ldflags="${LDFLAGS}" \
        -o "$output_path" \
        main.go
    
    if [ $? -eq 0 ]; then
        file_size=$(ls -lh "$output_path" | awk '{print $5}')
        echo "  ✓ 构建成功: $output_path ($file_size)"
    else
        echo "  ✗ 构建失败: ${GOOS}/${GOARCH}"
        exit 1
    fi
done

echo "======================================"
echo "所有平台构建完成！"
echo ""
echo "构建文件列表:"
ls -la ${OUTPUT_DIR}/

# 计算文件哈希
echo ""
echo "文件哈希值:"
cd ${OUTPUT_DIR}
for file in *; do
    if [ -f "$file" ]; then
        hash=$(shasum -a 256 "$file" | cut -d' ' -f1)
        echo "$file: $hash"
    fi
done
cd ..

# 创建发布包
echo ""
echo "创建发布包..."
cd ${OUTPUT_DIR}

# 为每个平台创建压缩包
for platform in "${PLATFORMS[@]}"; do
    IFS='/' read -r GOOS GOARCH <<< "$platform"
    
    binary_name="${BINARY_NAME}-${GOOS}-${GOARCH}"
    if [ "$GOOS" = "windows" ]; then
        binary_name="${binary_name}.exe"
    fi
    
    if [ -f "$binary_name" ]; then
        archive_name="${PROJECT_NAME}-${VERSION}-${GOOS}-${GOARCH}"
        
        # 创建临时目录
        mkdir -p "${archive_name}"
        
        # 复制文件
        cp "$binary_name" "${archive_name}/${BINARY_NAME}"
        if [ "$GOOS" = "windows" ]; then
            mv "${archive_name}/${BINARY_NAME}" "${archive_name}/${BINARY_NAME}.exe"
        else
            cp ../run.sh "${archive_name}/"
            cp ../watchdog.sh "${archive_name}/"
        fi
        # 复制配置文件示例
        cp ../conf/config.example.json "${archive_name}/config.json"
        
        # 创建README
        cat > "${archive_name}/README.txt" << EOF
${PROJECT_NAME} ${VERSION}
构建时间: ${BUILD_TIME}
平台: ${GOOS}/${GOARCH}

使用方法:
1. 编辑 config.json 配置文件
2. 运行: ./${BINARY_NAME} config.json

更多信息请查看项目文档。
EOF
        
        # 创建压缩包
        if [ "$GOOS" = "windows" ]; then
            zip -r "${archive_name}.zip" "${archive_name}" > /dev/null
            echo "  ✓ 创建 ${archive_name}.zip"
        else
            tar -czf "${archive_name}.tar.gz" "${archive_name}" > /dev/null
            echo "  ✓ 创建 ${archive_name}.tar.gz"
        fi
        
        # 清理临时目录
        rm -rf "${archive_name}"
    fi
done

cd ..

echo ""
echo "======================================"
echo "打包完成！输出目录: ${OUTPUT_DIR}"
echo "发布包已创建，可以直接分发使用。"