# 使用基础镜像
FROM ubuntu:latest

# 安装 iproute2 包，其中包含了 ip 命令
RUN apt-get update && apt-get install -y iproute2 && apt-get install -y net-tools && apt-get install -y sudo

# 添加一个非特权用户
RUN useradd -ms /bin/bash user

# 将非特权用户添加到 sudo 组
RUN usermod -aG sudo user

RUN echo 'user:1' | chpasswd

# 将 sudo 命令的权限设置为无密码
RUN echo '%sudo ALL=(ALL) NOPASSWD:ALL' >> /etc/sudoers

# 切换到非特权用户
USER user

# 设置容器启动时执行的默认命令
CMD ["bash"]

