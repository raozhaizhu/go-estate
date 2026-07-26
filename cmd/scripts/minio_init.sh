#!/bin/sh
set -e
# 配置客户端
mc alias set myminio http://minio:9000 "$MINIO_ACCESS_KEY_ID" "$MINIO_SECRET_ACCESS_KEY"
# 创建 avatars 桶
mc mb myminio/avatars --ignore-existing
# 设置桶下载权限为公开
mc anonymous set download myminio/avatars
# 上传默认头像
mc cp /assets/default_avatar.png myminio/avatars/default_avatar.png
echo "MinIO 初始化成功"
