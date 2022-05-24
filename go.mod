module terraform-provider-talos

go 1.18

replace (
	// workaround checksum issue when releasing
	github.com/cilium/ebpf => github.com/cilium/ebpf v0.8.1
	// use same forked yaml that talos does.
	gopkg.in/yaml.v3 => github.com/unix4ever/yaml v0.0.0-20210315173758-8fb30b8e5a5b
	// workaround for talos siderolabs not working
	talos-systems/talos => github.com/siderolabs/talos v1.0.5
)

// terraform
require (
	github.com/hashicorp/hc-install v0.3.2
	github.com/hashicorp/terraform-exec v0.16.1
	github.com/hashicorp/terraform-json v0.13.0
	github.com/hashicorp/terraform-plugin-framework v0.8.0
	github.com/hashicorp/terraform-plugin-go v0.9.0
	github.com/hashicorp/terraform-plugin-log v0.4.0
	github.com/hashicorp/terraform-plugin-sdk/v2 v2.14.0
)

// talos
require (
	github.com/talos-systems/crypto v0.3.5
	github.com/talos-systems/talos/pkg/machinery v1.0.4
)

require (
	github.com/davecgh/go-spew v1.1.1
	github.com/dustin/go-humanize v1.0.0
	github.com/opencontainers/runtime-spec v1.0.3-0.20210326190908-1c3f411f0417
	github.com/r3labs/diff/v3 v3.0.0
	github.com/talos-systems/go-blockdevice v0.3.1
	golang.zx2c4.com/wireguard/wgctrl v0.0.0-20220504211119-3d4a969bb56b
	google.golang.org/grpc v1.46.2
	gopkg.in/yaml.v2 v2.4.0
)

require (
	github.com/AlekSi/pointer v1.2.0 // indirect
	github.com/Microsoft/go-winio v0.5.1 // indirect
	github.com/agext/levenshtein v1.2.2 // indirect
	github.com/apparentlymart/go-textseg/v13 v13.0.0 // indirect
	github.com/containerd/go-cni v1.1.3 // indirect
	github.com/containernetworking/cni v1.0.1 // indirect
	github.com/cosi-project/runtime v0.0.0-20211216175730-264f8fcd1a4f // indirect
	github.com/evanphx/json-patch v5.6.0+incompatible // indirect
	github.com/fatih/color v1.13.0 // indirect
	github.com/fsnotify/fsnotify v1.5.1 // indirect
	github.com/ghodss/yaml v1.0.0 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/go-cmp v0.5.8 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-checkpoint v0.5.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-cty v1.4.1-0.20200414143053-d3edf31b6320 // indirect
	github.com/hashicorp/go-hclog v1.2.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/go-plugin v1.4.3 // indirect
	github.com/hashicorp/go-uuid v1.0.3 // indirect
	github.com/hashicorp/go-version v1.4.0 // indirect
	github.com/hashicorp/hcl/v2 v2.11.1 // indirect
	github.com/hashicorp/logutils v1.0.0 // indirect
	github.com/hashicorp/terraform-registry-address v0.0.0-20210412075316-9b2996cce896 // indirect
	github.com/hashicorp/terraform-svchost v0.0.0-20200729002733-f050f53b9734 // indirect
	github.com/hashicorp/yamux v0.0.0-20181012175058-2f1d1f20f75d // indirect
	github.com/josharian/native v1.0.0 // indirect
	github.com/jsimonetti/rtnetlink v1.1.0 // indirect
	github.com/mattn/go-colorable v0.1.12 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	github.com/mdlayher/ethtool v0.0.0-20220213132912-856bd6cb8a38 // indirect
	github.com/mdlayher/genetlink v1.2.0 // indirect
	github.com/mdlayher/netlink v1.6.0 // indirect
	github.com/mdlayher/socket v0.2.3 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/go-testing-interface v1.14.1 // indirect
	github.com/mitchellh/go-wordwrap v1.0.0 // indirect
	github.com/mitchellh/mapstructure v1.4.3 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/oklog/run v1.0.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/ryanuber/go-glob v1.0.0 // indirect
	github.com/stretchr/objx v0.2.0 // indirect
	github.com/talos-systems/go-debug v0.2.1 // indirect
	github.com/talos-systems/net v0.3.2 // indirect
	github.com/vmihailenco/msgpack v4.0.4+incompatible // indirect
	github.com/vmihailenco/msgpack/v4 v4.3.12 // indirect
	github.com/vmihailenco/tagparser v0.1.1 // indirect
	github.com/zclconf/go-cty v1.10.0 // indirect
	golang.org/x/crypto v0.0.0-20220411220226-7b82a4e95df4 // indirect
	golang.org/x/net v0.0.0-20220418201149-a630d4f3e7a2 // indirect
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c // indirect
	golang.org/x/sys v0.0.0-20220412211240-33da011f77ad // indirect
	golang.org/x/text v0.3.7 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20220228155957-1da8797a5878 // indirect
	google.golang.org/protobuf v1.28.0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
)
