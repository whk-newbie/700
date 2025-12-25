# 多阶段构建
# 第一阶段：构建阶段
FROM node:latest AS builder

# 设置工作目录
WORKDIR /app

# 复制package.json和package-lock.json（如果存在）
COPY package*.json ./

# 安装依赖（构建需要dev依赖）
RUN npm ci

# 复制源代码
COPY . .

# 构建应用
RUN npm run build

# 第二阶段：运行阶段
FROM nginx:latest

# 复制自定义nginx配置（可选）
# COPY nginx.conf /etc/nginx/nginx.conf

# 复制构建好的应用到nginx服务目录
COPY --from=builder /app/dist /usr/share/nginx/html

# 复制nginx配置文件
COPY nginx.conf /etc/nginx/conf.d/default.conf

# 暴露端口
EXPOSE 80

# 启动nginx
CMD ["nginx", "-g", "daemon off;"]
