#!/bin/bash

set -e

echo "=========================================="
echo "  CoAI.Dev Linux 生产环境部署脚本"
echo "=========================================="

# 颜色定义
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

# 检查 Docker 和 Docker Compose
check_dependencies() {
    echo -e "${YELLOW}检查依赖...${NC}"
    
    if ! command -v docker &> /dev/null; then
        echo -e "${RED}错误: 未安装 Docker${NC}"
        echo "请先安装 Docker: https://docs.docker.com/engine/install/"
        exit 1
    fi
    
    if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
        echo -e "${RED}错误: 未安装 Docker Compose${NC}"
        echo "请先安装 Docker Compose"
        exit 1
    fi
    
    echo -e "${GREEN}✓ 依赖检查通过${NC}"
}

# 检查端口占用
check_ports() {
    echo -e "${YELLOW}检查端口占用...${NC}"
    PORT=${PORT:-8094}
    
    if netstat -tuln 2>/dev/null | grep -q ":$PORT " || ss -tuln 2>/dev/null | grep -q ":$PORT "; then
        echo -e "${RED}警告: 端口 $PORT 已被占用${NC}"
        read -p "是否继续部署？(y/N) " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            exit 1
        fi
    else
        echo -e "${GREEN}✓ 端口 $PORT 可用${NC}"
    fi
}

# 创建环境变量文件
create_env() {
    if [ ! -f .env ]; then
        echo -e "${YELLOW}创建 .env 配置文件...${NC}"
        cat > .env <<EOF
# 数据库配置
MYSQL_DB=chatnio
MYSQL_PASSWORD=$(openssl rand -base64 32 2>/dev/null || date +%s | sha256sum | base64 | head -c 32)

# Redis 配置
REDIS_PASSWORD=

# JWT 密钥（请务必修改）
SECRET=$(openssl rand -base64 64 2>/dev/null || date +%s | sha256sum | base64 | head -c 64)

# 服务端口
PORT=8094
EOF
        echo -e "${GREEN}✓ 已创建 .env 文件，请根据需要修改配置${NC}"
        echo -e "${YELLOW}重要: 请修改 SECRET 为随机字符串${NC}"
    else
        echo -e "${GREEN}✓ .env 文件已存在${NC}"
    fi
}

# 构建并启动服务
deploy() {
    echo -e "${YELLOW}开始构建镜像...${NC}"
    docker-compose -f docker-compose.prod.yml build --no-cache
    
    echo -e "${YELLOW}启动服务...${NC}"
    docker-compose -f docker-compose.prod.yml up -d
    
    echo -e "${GREEN}✓ 服务已启动${NC}"
}

# 等待服务就绪
wait_for_service() {
    echo -e "${YELLOW}等待服务启动...${NC}"
    sleep 10
    
    for i in {1..30}; do
        if curl -f http://localhost:${PORT:-8094}/health &>/dev/null; then
            echo -e "${GREEN}✓ 服务已就绪${NC}"
            return 0
        fi
        echo -n "."
        sleep 2
    done
    
    echo -e "${RED}服务启动超时，请检查日志${NC}"
    docker-compose -f docker-compose.prod.yml logs --tail=50
    exit 1
}

# 显示部署信息
show_info() {
    PORT=${PORT:-8094}
    echo ""
    echo "=========================================="
    echo -e "${GREEN}部署成功！${NC}"
    echo "=========================================="
    echo "访问地址: http://localhost:$PORT"
    echo "管理员账号: root"
    echo "默认密码: chatnio123456"
    echo ""
    echo "常用命令:"
    echo "  查看日志: docker-compose -f docker-compose.prod.yml logs -f"
    echo "  重启服务: docker-compose -f docker-compose.prod.yml restart"
    echo "  停止服务: docker-compose -f docker-compose.prod.yml down"
    echo "  更新服务: ./deploy.sh"
    echo ""
    echo -e "${YELLOW}注意: 首次登录后请立即修改管理员密码！${NC}"
    echo "=========================================="
}

# 主流程
main() {
    check_dependencies
    check_ports
    create_env
    deploy
    wait_for_service
    show_info
}

main
