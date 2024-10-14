# /bin/bash

# 停止 ghproxy 服务
systemctl stop ghproxy

# 删除 ghproxy 服务
systemctl disable ghproxy
rm /etc/systemd/system/ghproxy.service

# 删除 ghproxy 文件夹
rm -r /root/data/ghproxy

echo "ghproxy 已成功卸载"