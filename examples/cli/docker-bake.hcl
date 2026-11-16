# Customer-owned build configuration. Run from the Proven repository root.
variable "NODE_BASE" { default = "node:24-alpine3.24" }
variable "RUNTIME_BASE" { default = "alpine:3.24" }
variable "GIT_PACKAGE_VERSION" { default = "2.56.1-r0" }
variable "MUSL_PACKAGE_VERSION" { default = "1.2.6-r6" }
variable "GO_BASE" { default = "golang:1.26.10-alpine3.24" }
variable "CLAUDE_VERSION" { default = "2.1.325" }
variable "COPILOT_VERSION" { default = "1.0.112" }
group "default" { targets = ["dual"] }
target "common" {
  context = "."
  dockerfile = "examples/cli/Dockerfile"
  args = {
    NODE_BASE = NODE_BASE
    RUNTIME_BASE = RUNTIME_BASE
    GIT_PACKAGE_VERSION = GIT_PACKAGE_VERSION
    MUSL_PACKAGE_VERSION = MUSL_PACKAGE_VERSION
    GO_BASE = GO_BASE
    CLAUDE_VERSION = CLAUDE_VERSION
    COPILOT_VERSION = COPILOT_VERSION
  }
}
target "claude" {
  inherits = ["common"]
  target = "claude"
  tags = ["proven-claude:local"]
}
target "copilot" {
  inherits = ["common"]
  target = "copilot"
  tags = ["proven-copilot:local"]
}
target "dual" {
  inherits = ["common"]
  target = "dual"
  tags = ["proven-dual:local"]
}
