# /bin/bash

# 停止 ghproxy 服务
systemctl stop ghproxy

# 删除 ghproxy 服务
systemctl disable ghproxy
rm /etc/systemd/system/ghproxy.service

# 获取安装文件夹
read -p "请输入 ghproxy 安装文件夹路径(默认 /usr/local/ghproxy): " install_path
if [ -z "$install_path" ]; then
    install_path="/usr/local/ghproxy"
fi

# 删除 ghproxy 文件夹
# 检查目录是否存在ghproxy文件
if [ -f "$install_path" ]; then
    echo "ghproxy 未安装或安装路径错误"
    exit 1
else
    echo "ghproxy 安装目录已确认，正在卸载..."
    rm -r $install_path
fi


echo "ghproxy 已成功卸载"