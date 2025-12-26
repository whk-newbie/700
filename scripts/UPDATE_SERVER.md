# 服务器更新代码指南

## 📋 更新步骤

### 方法1：保留本地修改（推荐）

如果服务器上有本地修改需要保留：

```bash
# 1. 查看本地修改
git status

# 2. 暂存本地修改
git stash

# 3. 拉取最新代码
git pull origin master

# 4. 恢复本地修改（如果需要）
git stash pop
```

### 方法2：丢弃本地修改（使用最新代码）

如果不需要保留本地修改：

```bash
# 1. 丢弃本地修改
git checkout -- scripts/deploy.sh

# 2. 拉取最新代码
git pull origin master
```

### 方法3：强制使用远程代码

如果确定要完全使用远程代码：

```bash
# 1. 重置到远程版本
git fetch origin
git reset --hard origin/master

# 2. 清理未跟踪文件（可选）
git clean -fd
```

## ⚠️ 注意事项

1. **备份重要修改**：在执行重置操作前，确保备份重要的本地修改
2. **检查差异**：使用 `git diff` 查看本地修改内容
3. **生产环境**：在生产环境更新前，建议先停止服务

## 🔄 完整更新流程（生产环境）

```bash
# 1. 停止服务
docker compose --profile production down

# 2. 备份当前配置（如果有自定义修改）
cp .env .env.backup

# 3. 暂存或丢弃本地修改
git stash
# 或
git checkout -- scripts/deploy.sh

# 4. 拉取最新代码
git pull origin master

# 5. 恢复配置（如果需要）
cp .env.backup .env

# 6. 重新构建并启动
docker compose --profile production up -d --build
```

## 📝 查看本地修改

```bash
# 查看修改的文件
git status

# 查看具体修改内容
git diff scripts/deploy.sh
```

