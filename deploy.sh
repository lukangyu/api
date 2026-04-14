#!/bin/bash
set -e

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  API 网关部署脚本${NC}"
echo -e "${GREEN}========================================${NC}"

if ! command -v docker &> /dev/null; then
  echo -e "${RED}[错误] Docker 未安装${NC}"
  exit 1
fi

if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
  echo -e "${RED}[错误] Docker Compose 未安装${NC}"
  exit 1
fi

if [ ! -f .env ]; then
  echo -e "${YELLOW}[!] 未检测到 .env，自动从 .env.example 创建${NC}"
  cp .env.example .env
  if command -v openssl &> /dev/null; then
    SECRET=$(openssl rand -hex 32)
  else
    SECRET=$(date +%s | sha256sum | awk '{print $1}')
  fi
  sed -i "s/change_me_jwt_secret/${SECRET}/" .env
fi

mkdir -p data

HOST_PORT=$(grep -E '^HOST_PORT=' .env | tail -n 1 | cut -d '=' -f2)
if [ -z "$HOST_PORT" ]; then
  HOST_PORT=8081
fi

echo -e "${YELLOW}[...] 构建并启动容器${NC}"
docker compose up -d --build 2>/dev/null || docker-compose up -d --build

echo -e "${YELLOW}[...] 等待服务启动${NC}"
MAX_RETRIES=30
for i in $(seq 1 $MAX_RETRIES); do
  CODE=$(curl -s -o /dev/null -w "%{http_code}" "http://localhost:${HOST_PORT}/api/health" || true)
  if [ "$CODE" = "200" ]; then
    echo -e "${GREEN}[✓] 服务已就绪${NC}"
    break
  fi
  sleep 2
  if [ $i -eq $MAX_RETRIES ]; then
    echo -e "${RED}[!] 服务启动失败，请查看日志：docker compose logs -f${NC}"
    exit 1
  fi
done

echo -e "${GREEN}访问地址: http://服务器IP:${HOST_PORT}${NC}"
echo -e "${GREEN}默认管理员: admin / changeme123${NC}"
echo -e "${RED}请首次登录后立即修改默认密码${NC}"
