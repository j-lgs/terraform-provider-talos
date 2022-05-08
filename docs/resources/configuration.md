---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "talos_configuration Resource - terraform-provider-talos"
subcategory: ""
description: |-
  
---

# talos_configuration (Resource)





<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `cluster_name` (String) Configures the cluster's name
- `endpoints` (List of String) A list of that the talosctl client will connect to. Can be a DNS hostname or an IP address and may include a port number. Must begin with "https://".
- `kubernetes_endpoint` (String) The kubernetes endpoint that the nodes and the kubectl client will connect to. Can be a DNS hostname or an IP address and may include a port number. Must begin with "https://".
- `kubernetes_version` (String) The version of kubernetes and all it's components (kube-apiserver, kubelet, kube-scheduler, etc) that will be deployed onto the cluster.
- `target_version` (String) The version of the Talos cluster configuration that will be generated.

### Read-Only

- `base_config` (String, Sensitive) JSON Serialised object that contains information needed to create controlplane and worker node configurations.
- `id` (String) The ID of this resource.
- `talosconfig` (String, Sensitive) Talosconfig YAML that can be used by the talosctl client to communicate with the cluster.

