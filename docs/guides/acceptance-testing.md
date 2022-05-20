

---
layout: ""
page_title: "Guide: Acceptance testing"
description: |-
  Instructions for performingacceptance tests for this provider.
---

# Dependencies
Acceptance testing this provider requires a pool of virtual machines to create and delete Talos clusters on, and pull through image caches to speed up the process and avoid container registry rate limiting. As such the following programs must be installed and set up.
+ libvirt
+ virsh
+ docker
The user running the test must have appropriate permissions to access these programs. The user must have permission to connect to libvirt's `qemu:///system` uri.

# Example
```shell
TF_ACC=1 MACHINELOG_DIR=$(pwd) TALOSCONF_DIR=$(pwd) go test -v ./talos

# The default acceptance testing makefile target runs the command above.
make acc-test
```

## Environment Variables
+ `MACHINELOG_DIR` - Where virtual machine logs are kept. Must be an absolute path. Note they will be owned by root .
+ `TALOSCONF_DIR` - Where generated talos configurations are kept. Must be an absolute path.
+ `RESET_VM` - Manually reset VMs when a test is began. Workaround used when test VMs hang on reboot. 

# Tools
When a test crashes the required virtual machines and containers for running the test are not properly destroyed. In order to manually perform this step run the `tools/failedAcceptance.sh` script.