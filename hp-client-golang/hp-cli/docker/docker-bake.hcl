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
  tags = ["${image_name}-arm64:latest"]
  push = true
}

target "amd64" {
  context = "."
  dockerfile = "Dockerfile.amd"
  platforms = ["linux/amd64"]
  args = {
    TARGETARCH = "amd64"
  }
  tags = ["${image_name}-amd64:latest"]
  push = true
}

target "manifest" {
  inherits = ["arm64", "amd64"]
  tags = ["${image_name}:latest"]
  type = "manifest"
  push = true
}