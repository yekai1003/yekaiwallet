### 卸载docker 以ubnutu为例

```sh
$ sudo apt-get remove docker docker-engine docker.io containerd runc
```

### 更新包

```sh
$ sudo apt-get update
```

### 安装依赖库

```sh
$ sudo apt-get install \
    apt-transport-https \
    ca-certificates \
    curl \
    gnupg-agent \
    software-properties-common
```

### 安装docker

```sh
$ sudo apt-get install docker-ce docker-ce-cli containerd.io
```

### 检测是否成功

```sh
$ sudo docker run hello-world
```

