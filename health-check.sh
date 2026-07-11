#!/bin/bash

# CoAI.Dev 健康检查脚本

PORT=${PORT:-8094}
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo "=========================================="
echo "  CoAI.Dev 服务健康检查"
echo "=========================================="

# 检查容器状态
check_containers() {
    echo -e "\n${YELLOW}容器状态:${NC}"
    docker-compose -f docker-compose.prod.yml ps
}

# 检查服务端口
check_port() {
    echo -e "\n${YELLOW}端口监听检查:${NC}"
    if netstat -tuln 2>/dev/null | grep -q ":$PORT " || ss -tuln 2>/dev/null | grep -q ":$PORT "; then
        echo -e "${GREEN}✓ 端口 $PORT 正在监听${NC}"
        return 0
    else
        echo -e "${RED}✗ 端口 $PORT 未监听${NC}"
        return 1
    fi
}

# 检查 HTTP 健康端点
check_http() {
    echo -e "\n${YELLOW}HTTP 健康检查:${NC}"
    response=$(curl -s -w "\n%{http_code}" http://localhost:$PORT/health 2>/dev/null)
    http_code=$(echo "$response" | tail -n1)
    
    if [ "$http_code" = "200" ]; then
        echo -e "${GREEN}✓ 健康检查通过 (HTTP $http_code)${NC}"
        return 0
    else
        echo -e "${RED}✗ 健康检查失败 (HTTP $http_code)${NC}"
        return 1
    fi
}

# 检查数据库连接
check_database() {
    echo -e "\n${YELLOW}数据库连接检查:${NC}"
    if docker exec coai-mysql mysqladmin ping -h localhost --silent 2>/dev/null; then
        echo -e "${GREEN}✓ MySQL 连接正常${NC}"
    else
        echo -e "${RED}✗ MySQL 连接失败${NC}"
    fi
    
    if docker exec coai-redis redis-cli ping 2>/dev/null | grep -q PONG; then
        echo -e "${GREEN}✓ Redis 连接正常${NC}"
    else
        echo -e "${RED}✗ Redis 连接失败${NC}"
    fi
}

# 显示资源占用
show_resources() {
    echo -e "\n${YELLOW}资源占用:${NC}"
    docker stats --no-stream coai-backend coai-mysql coai-redis 2>/dev/null
}

# 主流程
main() {
    check_containers
    check_port
    check_http
    check_database
    show_resources
    
    echo -e "\n${YELLOW}需要查看完整日志？运行: docker-compose -f docker-compose.prod.yml logs -f${NC}"
}

main
