# /bin/bash
# https://github.com/WJQSERVER-STUDIO/ghproxy

ghproxy_dir="/usr/local/ghproxy"

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



# 检查是否为root用户
if [ "$EUID" -ne 0 ]; then
    echo "请以root用户运行此脚本"
    exit 1
fi

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

# 获取监听端口
read -p "请输入程序监听的端口(默认8080): " PORT
if [ -z "$PORT" ]; then
    PORT=8080
fi

# 本机监听/泛监听(127.0.0.1/0.0.0.0)
read -p "请键入程序监听的IP(默认127.0.0.1)(0.0.0.0为泛监听): " IP
if [ -z "$IP" ]; then
    IP="127.0.0.1"
fi

# 安装目录
read -p "请输入安装目录(默认/usr/local/ghproxy): " ghproxy_dir
if [ -z "$ghproxy_dir" ]; then
    ghproxy_dir="/usr/local/ghproxy"
fi

# 创建目录
mkdir -p ${ghproxy_dir}
mkdir -p ${ghproxy_dir}/config
mkdir -p ${ghproxy_dir}/log
mkdir -p ${ghproxy_dir}/pages

# 获取最新版本号
VERSION=$(curl -s https://raw.githubusercontent.com/WJQSERVER-STUDIO/ghproxy/main/VERSION)
wget -q -O ${ghproxy_dir}/VERSION https://raw.githubusercontent.com/WJQSERVER-STUDIO/ghproxy/main/VERSION

# 下载ghproxy
wget -q -O ${ghproxy_dir}/ghproxy-linux-$ARCH.tar.gz https://github.com/WJQSERVER-STUDIO/ghproxy/releases/download/${VERSION}/ghproxy-linux-${ARCH}.tar.gz
install tar
tar -zxvf ${ghproxy_dir}/ghproxy-linux-$ARCH.tar.gz -C ${ghproxy_dir}
chmod +x ${ghproxy_dir}/ghproxy

# 下载pages
wget -q -O ${ghproxy_dir}/pages/index.html https://raw.githubusercontent.com/WJQSERVER-STUDIO/ghproxy/main/pages/bootstrap/index.html
wget -q -O ${ghproxy_dir}/pages/favicon.ico https://raw.githubusercontent.com/WJQSERVER-STUDIO/ghproxy/main/pages/bootstrap/favicon.ico


# 下载配置文件
if [ -f ${ghproxy_dir}/config/config.toml ]; then
    echo "配置文件已存在, 跳过下载"
    echo "[WARNING] 请检查配置文件是否正确，DEV版本升级时请注意配置文件兼容性"
    sleep 2
else
    wget -q -O ${ghproxy_dir}/config/config.toml https://raw.githubusercontent.com/WJQSERVER-STUDIO/ghproxy/main/deploy/config.toml
fi

# 替换 port = 8080 
sed -i "s/port = 8080/port = $PORT/g" ${ghproxy_dir}/config/config.toml
sed -i 's/host = "127.0.0.1"/host = "'"$IP"'"/g' ${ghproxy_dir}/config/config.toml
sed -i "s|staticDir = \"/usr/local/ghproxy/pages\"|staticDir = \"${ghproxy_dir}/pages\"|g" ${ghproxy_dir}/config/config.toml
sed -i "s|logFilePath = \"/usr/local/ghproxy/log/ghproxy.log\"|logFilePath = \"${ghproxy_dir}/log/ghproxy.log\"|g" ${ghproxy_dir}/config/config.toml
sed -i "s|blacklistFile = \"/usr/local/ghproxy/config/blacklist.json\"|blacklistFile = \"${ghproxy_dir}/config/blacklist.json\"|g" ${ghproxy_dir}/config/config.toml
sed -i "s|whitelistFile = \"/usr/local/ghproxy/config/whitelist.json\"|whitelistFile = \"${ghproxy_dir}/config/whitelist.json\"|g" ${ghproxy_dir}/config/config.toml

# 下载systemd服务文件
if [ "$ghproxy_dir" = "/usr/local/ghproxy" ]; then
    wget -q -O /etc/systemd/system/ghproxy.service https://raw.githubusercontent.com/WJQSERVER-STUDIO/ghproxy/main/deploy/ghproxy.service
else

    cat <<EOF > /etc/systemd/system/ghproxy.service

[Unit]
Description=Github Proxy Service
After=network.target

[Service]
ExecStart=/bin/bash -c '$ghproxy_dir/ghproxy -c $ghproxy_dir/config/config.toml > $ghproxy_dir/log/run.log 2>&1'
WorkingDirectory=$ghproxy_dir
Restart=always
User=root
Group=root

[Install]
WantedBy=multi-user.target

EOF

fi

# 启动ghproxy
systemctl daemon-reload
systemctl enable ghproxy
systemctl start ghproxy

echo "ghproxy 安装成功, 监听端口为 $PORT"
