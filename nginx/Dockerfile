FROM nginx:latest

# 安装OpenSSL和gettext-base（用于envsubst）
RUN apt-get update && \
    apt-get install -y openssl gettext-base && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

# 复制配置模板和启动脚本
COPY nginx.conf.template /etc/nginx/nginx.conf.template
COPY entrypoint.sh /docker-entrypoint.sh
RUN chmod +x /docker-entrypoint.sh

# 使用自定义entrypoint
ENTRYPOINT ["/docker-entrypoint.sh"]
CMD ["nginx", "-g", "daemon off;"]

