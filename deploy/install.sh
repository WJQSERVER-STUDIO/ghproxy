# /bin/bash

# install packages
install() {
    if [ $# -eq 0 ]; then
        echo "ARGS NOT FOUND"
        return 1
    fi

    for package in "$@"; do
        if ! command -v "$package" &>/dev/null; then
            if command -v dnf &>/dev/null; then
                dnf -y update && dnf install -y "$package"
            elif command -v yum &>/dev/null; then
                yum -y update && yum -y install "$package"
            elif command -v apt &>/dev/null; then
                apt update -y && apt install -y "$package"
            elif command -v apk &>/dev/null; then
                apk update && apk add "$package"
            else
                echo "UNKNOWN PACKAGE MANAGER"
                return 1
            fi
        fi
    done

    return 0
}

# 安装依赖包
install curl wget sed

# 查看当前架构是否为linux/amd64或linux/arm64
ARCH=$(uname -m)
if [ "$ARCH" != "x86_64" ] && [ "$ARCH" != "aarch64" ]; then
    echo " $ARCH 架构不被支持"
    exit 1
fi

# 重写架构值,改为amd64或arm64
if [ "$ARCH" == "x86_64" ]; then
    ARCH="amd64"
elif [ "$ARCH" == "aarch64" ]; then
    ARCH="arm64"
fi

read -p "请输入程序监听的端口(默认8080): " PORT
if [ -z "$PORT" ]; then
    PORT=8080
fi

# 创建目录
mkdir -p /root/data/ghproxy
mkdir -p /root/data/ghproxy/config
mkdir -p /root/data/ghproxy/log

# 获取最新版本号
VERSION=$(curl -s https://raw.githubusercontent.com/WJQSERVER-STUDIO/ghproxy/main/VERSION)
wget -O /root/data/ghproxy/VERSION https://raw.githubusercontent.com/WJQSERVER-STUDIO/ghproxy/main/VERSION

# 下载ghproxy
wget -O /root/data/ghproxy/ghproxy https://github.com/WJQSERVER-STUDIO/ghproxy/releases/download/$VERSION/ghproxy-linux-$ARCH
chmod +x /root/data/ghproxy/ghproxy

# 下载配置文件
wget -O /root/data/ghproxy/config/config.toml https://raw.githubusercontent.com/WJQSERVER-STUDIO/ghproxy/main/deploy/config.toml
# 替换 port = 8080 
sed -i "s/port = 8080/port = $PORT/g" /root/data/ghproxy/config/config.toml

# 下载systemd服务文件
wget -O /etc/systemd/system/ghproxy.service https://raw.githubusercontent.com/WJQSERVER-STUDIO/ghproxy/main/deploy/ghproxy.service

# 启动ghproxy
systemctl daemon-reload
systemctl enable ghproxy
systemctl start ghproxy

echo "ghproxy 安装成功, 监听端口为 $PORT"
