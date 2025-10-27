#!/bin/bash

# GoMailer 设置脚本
# 用于快速配置和上传到 GitHub

set -e

echo "🚀 GoMailer 设置向导"
echo "===================="
echo ""

# 1. 检查 Git 是否安装
if ! command -v git &> /dev/null; then
    echo "❌ 错误: 未找到 Git，请先安装 Git"
    exit 1
fi

echo "✅ Git 已安装"
echo ""

# 2. 获取 GitHub 用户名
echo "📝 请输入你的 GitHub 用户名:"
read -p "GitHub 用户名: " github_username

if [ -z "$github_username" ]; then
    echo "❌ 错误: GitHub 用户名不能为空"
    exit 1
fi

echo ""
echo "✅ GitHub 用户名: $github_username"
echo ""

# 3. 确认替换
echo "⚠️  即将替换所有文件中的 'yourusername' 为 '$github_username'"
read -p "确认继续? (y/n): " confirm

if [ "$confirm" != "y" ] && [ "$confirm" != "Y" ]; then
    echo "❌ 操作已取消"
    exit 0
fi

echo ""
echo "🔄 正在替换模块路径..."

# 4. 替换所有文件中的 yourusername
find . -type f \( -name "*.go" -o -name "*.md" -o -name "go.mod" \) -not -path "./.git/*" -exec sed -i.bak "s/yourusername/$github_username/g" {} \;

# 清理备份文件
find . -name "*.bak" -type f -delete

echo "✅ 模块路径已替换"
echo ""

# 5. 初始化 Git（如果还没有）
if [ ! -d ".git" ]; then
    echo "📦 初始化 Git 仓库..."
    git init
    git add .
    git commit -m "feat: 初始版本 - 从 PocketBase 提取的邮件发送库"
    echo "✅ Git 仓库已初始化"
else
    echo "✅ Git 仓库已存在"
fi

echo ""

# 6. 询问是否要添加远程仓库
echo "🌐 配置远程仓库"
read -p "是否要添加远程仓库? (y/n): " add_remote

if [ "$add_remote" = "y" ] || [ "$add_remote" = "Y" ]; then
    echo ""
    echo "📝 请输入仓库名称 (默认: gomailer):"
    read -p "仓库名称: " repo_name
    repo_name=${repo_name:-gomailer}
    
    remote_url="https://github.com/$github_username/$repo_name.git"
    
    echo ""
    echo "将添加远程仓库: $remote_url"
    read -p "确认? (y/n): " confirm_remote
    
    if [ "$confirm_remote" = "y" ] || [ "$confirm_remote" = "Y" ]; then
        # 检查是否已有 origin
        if git remote get-url origin &> /dev/null; then
            echo "⚠️  远程仓库 'origin' 已存在，将移除并重新添加"
            git remote remove origin
        fi
        
        git remote add origin "$remote_url"
        git branch -M main
        echo "✅ 远程仓库已配置"
        
        echo ""
        read -p "是否现在推送到 GitHub? (y/n): " push_now
        
        if [ "$push_now" = "y" ] || [ "$push_now" = "Y" ]; then
            echo "🚀 正在推送到 GitHub..."
            git push -u origin main
            
            echo ""
            echo "🏷️  创建版本标签 v1.0.0"
            git tag v1.0.0
            git push origin v1.0.0
            
            echo "✅ 推送成功！"
        else
            echo ""
            echo "📝 稍后可以使用以下命令推送:"
            echo "   git push -u origin main"
            echo "   git tag v1.0.0"
            echo "   git push origin v1.0.0"
        fi
    fi
fi

echo ""
echo "=========================================="
echo "✨ 设置完成！"
echo "=========================================="
echo ""
echo "📦 项目信息:"
echo "   用户名: $github_username"
echo "   模块路径: github.com/$github_username/gomailer"
echo ""
echo "🔗 下一步:"
echo "   1. 在 GitHub 上创建 '$repo_name' 仓库（如果还没有）"
echo "   2. 推送代码到 GitHub（如果还没有）"
echo "   3. 创建第一个 Release"
echo "   4. 开始在其他项目中使用！"
echo ""
echo "📚 文档:"
echo "   - 快速开始: cat QUICKSTART.md"
echo "   - 完整文档: cat README.md"
echo "   - 部署指南: cat DEPLOYMENT.md"
echo ""
echo "🎉 祝你使用愉快！"

