variable "registry" {
//     default = "docker.io"  # 默认值，可被环境变量覆盖
default = "registry.cn-shenzhen.aliyuncs.com"  # 默认值，可被环境变量覆盖
}

variable "image_name" {
  default = "heixiaoma/hp-lite"
}

group "default" {
  targets = ["manifest"]
}

target "manifest" {
  context = "."
  dockerfile = "Dockerfile"
  platforms = ["linux/arm64", "linux/amd64","linux/arm/v7","linux/386"]
  tags = ["${registry}/${image_name}:latest"]
  type = "manifest"
  push = true
}