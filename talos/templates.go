package talos

func templateHaproxy() string {
	return `
{{define "haproxy"}}
global
  log         /dev/log local0
  log         /dev/log local1 notice
  daemon
defaults
  mode                    tcp
  log                     global
  option                  tcplog
  option                  tcp-check
  option                  dontlognull
  retries                 3
  timeout client          20s
  timeout server          20s
  timeout check           10s
  timeout queue           20s
  option                  redispatch
  timeout connect         5s

frontend http_stats
  bind {{ .APIProxyIP }}:8080
  mode http
  stats uri /haproxy?stats

listen k8s-apiserver
  bind {{ .APIProxyIP }}:{{ .APIProxyPort }}
  option httpchk GET /healthz
  http-check expect status 200
  option ssl-hello-chk
  balance leastconn
  server {{ .Name }}-{{ .IP }} {{ .IP }}:{{ .LocalAPIProxyPort }} check inter 5s  fall 2
{{ range .Peers }}
  server {{ $.Name }}-{{ . }} {{ . }}:{{ $.LocalAPIProxyPort }} check inter 5s  fall 2
{{ end }}

{{ if ne .WgIP "" }}
listen wireguard-http-to-ingress
  bind {{ .WgIP }}:{{ .IngressPort }}
  mode http
  option httpchk GET /healthz
  http-check disable-on-404
  balance leastconn
  server wireguard-1-http {{ .IngressIP }}:80 check inter 60s  fall 3  rise 1

listen wireguard-https-to-ingress
  bind {{ .WgIP }}:{{ .IngressSSLPort }}
  option httpchk GET /healthz
  http-check disable-on-404
  balance leastconn
  server wireguard-1-https {{ .IngressIP }}:443 check check-ssl inter 60s  fall 3  rise 1 verify none
{{ end }}
{{ end }}
`
}

func templateKeepalived() string {
	return `
{{ define "keepalived" }}
global_defs {
  router_id {{ .VRID }}
}

vrrp_instance VI_1 {
  state {{ .State }}

  interface eth0

  virtual_router_id {{ .VRID }}
  priority {{ .Priority }}

  mcast_src_ip {{ .IP }}

  authentication {
    auth_type PASS
    auth_pass {{ .VIPPass }}
  }

  unicast_peer {
{{ range .Peers }}
    {{ . }}
{{ end }}
  }

  virtual_ipaddress {
    {{ .APIProxyIP }}/24
  }
}
{{ end }}
`
}

func templateAPICheck() string {
	return `
{{ define "checkAPIServer" }}
#!/bin/sh

errorExit() {
  echo "*** $*" 1>&2
  exit 1
}

curl --silent --max-time 2 --insecure https://localhost:{{ .APIProxyPort }}/ -o /dev/null || errorExit "Error GET https://localhost:{{ .APIProxyPort }}/"
if ip addr | grep -q {{ .APIProxyIP }}; then
  curl --silent --max-time 2 --insecure https://{{ .APIProxyIP }}:{{ .APIProxyPort }}/ -o /dev/null || errorExit "Error GET https://{{ .APIProxyIP }}:{{ .APIProxyPort }}/"
fi
{{ end }}
`
}

func templateControl() string {
	return `
[
{{ if ne .APIProxyIP "" }}
 {
  "op": "add",
  "path": "/cluster/apiServer/extraArgs",
  "value": {
   "secure-port": {{ .LocalAPIProxyPort }}
  }
 },
 {
  "op": "add",
  "path": "/cluster/controlPlane/localAPIServerPort",
  "value": {{ .LocalAPIProxyPort }}
 },
 {
  "op": "add",
  "path": "/machine/sysctls",
  "value": {
   "net.ipv4.ip_nonlocal_bind": "1",
   "net.ipv6.ip_nonlocal_bind": "1"
  }
 },
{{ end }}
 {
  "op": "add",
  "path": "/machine/network/interfaces",
  "value": [
{{ if ne .WgAddress "" }}
   {
    "addresses": [
     "{{ .WgAddress }}"
    ],
    "interface": "{{ .WgInterface }}",
    "wireguard": {
     "peers": [
      {
       "allowedIPs": [
        "{{ .WgAllowedIPs }}"
       ],
       "endpoint": "{{ .WgEndpoint }}",
       "persistentKeepaliveInterval": "25s",
       "publicKey": "{{ .WgPublicKey }}"
      }
     ],
     "privateKey": "{{ .WgPrivateKey }}"
    }
   },
{{ end }}
   {
    "addresses": [
     "{{ .IPNetwork }}"
    ],
    "interface": "eth0",
    "routes": [
     {
      "gateway": "{{ .Gateway }}",
      "network": "0.0.0.0/0"
     }
    ]
   }
  ]
 },
{{ if ne .RegistryIP "" }}
 {
  "op": "add",
  "path": "/machine/registries",
  "value": {
   "mirrors": {
    "docker.io":  { "endpoints": [ "http://{{ .RegistryIP }}:5000" ] },
    "k8s.gcr.io": { "endpoints": [ "http://{{ .RegistryIP }}:5001" ] },
    "quay.io":    { "endpoints": [ "http://{{ .RegistryIP }}:5003" ] },
    "gcr.io":     { "endpoints": [ "http://{{ .RegistryIP }}:5003" ] },
    "ghcr.io":    { "endpoints": [ "http://{{ .RegistryIP }}:5004" ] }
   }
  }
 },
{{ end }}
 {
  "op": "add",
  "path": "/machine/network/hostname",
  "value": "{{ .Hostname }}"
 },
 {
  "op": "add",
  "path": "/machine/network/nameservers",
  "value": [

{{- range $i, $e := .Nameservers -}}
{{if $i}},
{{end}}"{{$e}}"
{{- end}}
  ]
 },
{{ if ne .APIProxyIP "" }}
 {
  "op": "add",
  "path": "/machine/kubelet/extraMounts",
  "value": [
   {
    "destination": "/var/static-confs",
    "options": [
     "rbind",
     "rshared",
     "rw"
    ],
    "source": "/var/static-confs",
    "type": "bind"
   }
  ]
  },
 {
  "op": "add",
  "path": "/machine/certSANs/0",
  "value": "{{ .APIProxyIP }}"
 },
 {
  "op": "add",
  "path": "/machine/files",
  "value": [
    {{ templateFilesValue "haproxy"    "create" "/var/static-confs/haproxy/haproxy.cfg" 438 . }},
    {{ templateFilesValue "keepalived" "create" "/var/static-confs/keepalived/keepalived.conf" 292 . }},
    {{ templateFilesValue "checkAPIServer" "create" "/var/static-confs/check_apiserver.sh" 365 . }}
  ]
 },
 {
  "op": "add",
  "path": "/machine/pods",
  "value": [
   {
    "apiVersion": "v1",
    "kind": "Pod",
    "metadata": {
     "name": "keepalived",
     "namespace": "kube-system"
    },
    "spec": {
     "containers": [
      {
       "image": "{{ .KeepalivedImage }}",
       "name": "keepalived",
       "resources": {},
       "securityContext": {
        "capabilities": {
         "add": [
          "NET_ADMIN",
          "NET_BROADCAST",
          "NET_RAW"
         ]
        }
       },
       "volumeMounts": [
        {
         "mountPath": "/usr/local/etc/keepalived/keepalived.conf",
         "name": "config"
        },
        {
         "mountPath": "/etc/keepalived/check_apiserver.sh",
         "name": "check"
        }
       ]
      }
     ],
     "hostNetwork": true,
     "volumes": [
      {
       "hostPath": {
        "path": "/var/static-confs/keepalived/keepalived.conf"
       },
       "name": "config"
      },
      {
       "hostPath": {
        "path": "/var/static-confs/check_apiserver.sh"
       },
       "name": "check"
      }
     ]
    },
    "status": {}
   },
   {
    "apiVersion": "v1",
    "kind": "Pod",
    "metadata": {
     "name": "haproxy",
     "namespace": "kube-system"
    },
    "spec": {
     "containers": [
      {
       "image": "{{ .HaproxyImage }}",
       "name": "haproxy-controlplane",
       "volumeMounts": [
        {
         "mountPath": "/usr/local/etc/haproxy/haproxy.cfg",
         "name": "haproxyconf",
         "readOnly": true
        }
       ]
      }
     ],
     "hostNetwork": true,
     "volumes": [
      {
       "hostPath": {
        "path": "/var/static-confs/haproxy/haproxy.cfg",
        "type": "FileOrCreate"
       },
       "name": "haproxyconf"
      }
     ]
    },
    "status": {}
   }
  ]
 },
{{ end }}
 {
  "op": "add",
  "path": "/cluster/proxy",
  "value": {
   "extraArgs": {
    "ipvs-strict-arp": "true"
   }
  }
 }
]
`
}

func templateWorker() string {
	return `
[
{{ if .Privileged }}
 {
  "op": "add",
  "path": "/cluster/apiServer",
  "value": {
   "extraArgs": {
    "allow-privileged": "true"
   }
  }
 },
{{ end }}
{{ if eq .GPU "Cometlake" }}
 {
  "op": "add",
  "path": "/machine/install/extraKernelArgs",
  "value": [
   "i915.enable_guc=2",
   "i915.enable_dc=0"
  ]
 },
{{ end }}
 {
  "op": "add",
  "path": "/machine/network/interfaces",
  "value": [
   {
    "interface": "eth0",
    "addresses": [
     "{{ .IPNetwork }}"
    ],
    "routes": [
     {
      "network": "0.0.0.0/0",
      "gateway": "{{ .Gateway }}"
     }
    ]
   }
  ]
 },
{{ if ne .RegistryIP "" }}
 {
  "op": "add",
  "path": "/machine/registries",
  "value": {
   "mirrors": {
    "docker.io":  { "endpoints": [ "http://{{ .RegistryIP }}:5000" ] },
    "k8s.gcr.io": { "endpoints": [ "http://{{ .RegistryIP }}:5001" ] },
    "quay.io":    { "endpoints": [ "http://{{ .RegistryIP }}:5003" ] },
    "gcr.io":     { "endpoints": [ "http://{{ .RegistryIP }}:5003" ] },
    "ghcr.io":    { "endpoints": [ "http://{{ .RegistryIP }}:5004" ] }
   }
  }
 },
{{ end }}

 {
  "op": "add",
  "path": "/machine/network/hostname",
  "value": "{{ .Hostname }}"
 },
 {
  "op": "add",
  "path": "/machine/network/nameservers",
  "value": [

{{- range $i, $e := .Nameservers -}}
{{if $i}},
{{end}}"{{$e}}"
{{- end}}
  ]
 },
{{ if .Mayastor }}
 {
  "op": "add",
  "path": "/machine/kubelet/extraMounts",
  "value": [
   {
    "destination": "/var/local",
    "type": "bind",
    "source": "/var/local",
    "options": [
     "rbind",
     "rshared",
     "rw"
    ]
   }
  ]
 },
 {
  "op": "add",
  "path": "/machine/kubelet/extraArgs",
  "value": {
   "node-labels": "openebs.io/engine=mayastor"
  }
 },
 {
  "op": "add",
  "path": "/machine/sysctls",
  "value": {
   "vm.nr_hugepages": "2048"
  }
 },
{{ end }}
{{ if ne .GPU "" }}
 {
  "op": "add",
  "path": "/machine/udev",
  "value": {
   "rules": [
    "SUBSYSTEM==\"drm\", KERNEL==\"renderD*\", GROUP=\"103\", MODE=\"0666\"",
    "SUBSYSTEM==\"drm\", KERNEL==\"card*\",    GROUP=\"44\",  MODE=\"0666\""
   ]
  }
 },
{{ end }}
 {
  "op": "add",
  "path": "/cluster/proxy",
  "value": {
   "extraArgs": {
    "ipvs-strict-arp": "true"
   }
  }
 }
]
`
}
