#!/bin/sh
# kill parent processes
cleanup() {
   pkill -P $$
}

# Setup signals to kill child processes on exit.
for sig in INT QUIT HUP TERM; do
  trap "
    cleanup
    trap - $sig EXIT
    kill -s $sig "'"$$"' "$sig"
done
trap cleanup EXIT

mkdir -p test/bin

slirp_version="1.2.0"
slirp_arch="x86_64"
slirp_sha256="11080fdfb2c47b99f2b0c2b72d92cc64400d0eaba11c1ec34f779e17e8844360"

# Get dependencies
if [ ! -f test/bin/slirp4netns-v${slirp_version} ]; then
    echo "acctest -> downloading slirp4netns binary"
    curl -o test/bin/slirp4netns-v${slirp_version} --fail -L \
	 "https://github.com/rootless-containers/slirp4netns/releases/download/v${slirp_version}/slirp4netns-$(uname -m)"

    echo "$slirp_sha256 test/bin/slirp4netns-v${slirp_version}" | sha256sum -c -;

    chmod +x test/bin/slirp4netns-v${slirp_version}
fi

talos_version="1.0.5"
talos_arch="amd64"
talos_sha256="73103abf33ecbc5c2df1d0e157fe13ecea0510867b13060b625b260e4346addc"

mkdir -p test/opt

if [ ! -f test/opt/talos-amd64-v${talos_version}.iso ]; then
    echo "acctest -> downloading talos iso"

    rm -f test/opt/talos*

    curl -o test/opt/talos-${talos_arch}-v${talos_version}.iso --fail -L \
	 https://github.com/siderolabs/talos/releases/download/v${talos_version}/talos-${talos_arch}.iso
    echo "$talos_sha256 test/opt/talos-${talos_arch}-v${talos_version}" | sha256sum -c -;
fi

talosctl_version="1.0.5"
talosctl_arch="amd64"
talosctl_sha256="300eb0663fdf1826dd34cb0a72331cc56242830cf5100b69100327aad79a9c44"

if [ ! -f test/bin/talosctl-v${talosctl_version} ]; then
    echo "acctest -> downloading talosctl"

    rm -f test/bin/talosctl*

    curl -o test/bin/talosctl-v${talosctl_version} --fail -L \
	 https://github.com/siderolabs/talos/releases/download/v${talos_version}/talosctl-linux-${talos_arch}
    echo "$talosctl_sha256 test/bin/talosctl-v${talosctl_version}" | sha256sum -c -;

    chmod +x test/bin/talosctl-v${talosctl_version}
fi

registry_version="2.8.1"
registry_arch="amd64"
registry_sha256="f1a376964912a5fd7d588107ebe5185da77803244e15476d483c945959347ee2"

if [ ! -f test/bin/registry-v${registry_version} ]; then
    echo "acctest -> downloading registry binary"

    rm -f test/bin/registry*

    curl -o registry.tar.gz --fail -L \
	 "https://github.com/distribution/distribution/releases/download/v${registry_version}/registry_${registry_version}_linux_${registry_arch}.tar.gz";
    echo "$registry_sha256 test/bin/registry.tar.gz" | sha256sum -c -;
    tar --extract --verbose --file registry.tar.gz --directory test/bin/ registry;

    mv test/bin/registry test/bin/registry-v${registry_version}
    rm registry.tar.gz;
fi

mkdir -p test/etc
cat << EOF > test/etc/registry.yml
version: 0.1
log:
  fields:
    service: registry
storage:
  cache:
    blobdescriptor: inmemory
EOF

cat << EOF > test/etc/resolv.conf
nameserver 10.0.2.3
EOF

# Create namespace and run tests inside of it
unshare --user --map-root-user --net --mount sh -c 'sleep 1800' &
pid=$!
echo $pid > test/run/namespace.pid
sleep 0.1
test/bin/slirp4netns-v${slirp_version} --configure --mtu=65520 --disable-host-loopback $pid tap9 > /dev/null 2>&1 &
nsenter -U --wd="$(pwd)" -t $pid -m -n --preserve tools/runtest.sh
