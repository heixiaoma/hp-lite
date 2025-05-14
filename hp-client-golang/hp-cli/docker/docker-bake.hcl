variable "registry" {
  default = "docker.io"  # 默认值，可被环境变量覆盖
//   default = "registry.cn-shenzhen.aliyuncs.com"  # 默认值，可被环境变量覆盖
}

variable "image_name" {
  default = "heixiaoma/hp-lite"
}

group "default" {
  targets = ["arm64", "amd64", "manifest"]
}

target "arm64" {
  context = "."
  dockerfile = "Dockerfile.arm"
  platforms = ["linux/arm64"]
  args = {
    TARGETARCH = "arm64"
  }
  tags = ["${registry}/${image_name}-arm64:latest"]
  push = true
}

target "amd64" {
  context = "."
  dockerfile = "Dockerfile.amd"
  platforms = ["linux/amd64"]
  args = {
    TARGETARCH = "amd64"
  }
  tags = ["${registry}/${image_name}-amd64:latest"]
  push = true
}

target "manifest" {
  inherits = ["arm64", "amd64"]
  platforms = ["linux/arm64", "linux/amd64"]
  tags = ["${registry}/${image_name}:latest"]
  type = "manifest"
  push = true
}