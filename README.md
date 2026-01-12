# UCloud Upload Image

一个用于将自定义镜像上传到 UCloud 平台的 Go 工具。

## 功能特性

- 自动创建临时 UHost 实例
- 支持多种镜像格式（raw, bz2, xz, zstd）
- 通过 kexec 将镜像写入磁盘
- 自动创建自定义镜像并等待完成
- 清理临时资源（实例和密钥对）

## 环境要求

- Go 1.25.5+
- UCloud 账号和 API 密钥

## 安装

```bash
git clone https://github.com/5aaee9/ucloud-upload-image.git
cd ucloud-upload-image
go mod tidy
go build -o ucloud-upload-image
```

## 配置

设置环境变量：

```bash
export UCLOUD_PUBLIC_KEY="your_ucloud_public_key"
export UCLOUD_PRIVATE_KEY="your_ucloud_private_key"
```

## 使用方法

```bash
./ucloud-upload-image \
  --zone "cn-bj2" \
  --region "cn-bj2" \
  --name "my-custom-image" \
  --image "/path/to/image.raw.zst" \
  --format "zstd" \
  --network-type "Bgp"
```

### 参数说明

- `--zone`: UCloud 可用区（必需）
- `--region`: UCloud 区域（必需）
- `--name`: 自定义镜像名称（必需）
- `--image`: 镜像文件路径或下载 URL（必需）
- `--format`: 镜像格式，支持：raw, bz2, xz, zstd（默认：raw）
- `--network-type`: 主网卡网络类型（默认：Bgp）

## 工作流程

1. 创建临时 SSH 密钥对
2. 启动临时 UHost 实例（Debian 12）
3. 等待实例获取公网 IP
4. 通过 SSH 连接并执行 kexec
5. 将镜像写入磁盘
6. 创建自定义镜像
7. 等待镜像制作完成
8. 清理临时资源

## 项目结构

```
ucloud-upload-image/
├── main.go                 # 主程序入口
├── go.mod                  # Go 模块文件
├── pkgs/
│   ├── task/
│   │   └── ssh.go         # 任务执行逻辑
│   ├── sshutil/
│   │   ├── connect.go     # SSH 连接工具
│   │   ├── run.go         # SSH 命令执行
│   │   └── keypair.go     # 密钥对生成
│   ├── steps/
│   │   ├── kexec/         # kexec 执行步骤
│   │   ├── writedisk/     # 磁盘写入步骤
│   │   └── power/         # 电源管理步骤
│   └── utils/
│       └── env.go         # 环境变量工具
└── README.md
```

## 注意事项

- 确保镜像文件格式正确
- 临时实例会自动清理，无需手动删除
- 镜像制作过程可能需要较长时间，请耐心等待
- 建议在测试环境先验证流程

## 许可证

本项目采用 MIT 许可证。