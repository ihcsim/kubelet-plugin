# kubelet-plugin

This repo contains examples of kubelet plugins that can be used to expose the
following node devices to the Kubelet:

* [Character device file](https://man7.org/linux/man-pages/man2/mknod.2.html).
The `crand` plugin manages character special files pointing to the local
`/dev/random`.
* KVM device file. The `kvm` plugin points to the local `/dev/kvm`.

For more information on Kubernetes device plugins, see
https://kubernetes.io/docs/concepts/extend-kubernetes/compute-storage-net/device-plugins/.

## CDI Configuration - containerd

This plugin uses [CDI](https://github.com/cncf-tags/container-device-interface/)
for device discovery.

To enable CDI support with containerd v1.6+, update the 
`/etc/containerd/config.toml` file to include:

```sh
[plugins."io.containerd.grpc.v1.cri"]
  enable_cdi = true
  cdi_spec_dirs = ["/etc/cdi", "/var/run/cdi"]
```

Restart containerd for the changes to take effect.

Set up the sample devices and CDI configuration files:

```sh
# sudo required
make cdi
```

The CDI configuration files can be found in the local `/etc/cdi` directory.

## Flatcar

Download and start a Flatcar Linux VM with `virsh`:

```sh
make flatcar-start
```

If the host uses AppArmor, allow qemu to access the config files:

```sh
echo "  `pwd`/flatcar/provision.ign r," >> /etc/apparmor.d/abstractions/libvirt-qemu
```

## Testing With Kubelet

For ease of testing purposes, this repo comes with the kubelet v1.31.0 binary. 
The `KubeletConfiguration` specification is defined in the `kubelet/kubelet.yaml`
file.

To uncompress and start the kubelet:

```sh
# sudo required
make kubelet
```

The kubelet logs will be written to `kubelet/kubelet.log`.

Run the plugin against the kubelet:

```sh
# sudo required
make run
```

Expect the plugin `github.com.ihcsim/crand` to register successfully with the kubelet:
```sh
I0823 09:22:35.030339 1309759 server.go:144] "Got registration request from device plugin with resource" resourceName="github.com.ihcsim/crand"
I0823 09:22:35.030374 1309759 handler.go:95] "Registered client" name="github.com.ihcsim/crand"
I0823 09:22:35.030966 1309759 manager.go:238] "Device plugin connected" resourceName="github.com.ihcsim/crand"
# ...
I0823 09:04:14.473749 1352537 setters.go:329] "Updated capacity for device plugin" plugin="github.com.ihcsim/crand" capacity=3
```

Deploy the provided busybox pod to the kubelet as a static pod:

```sh
make deploy
```

The kubelet logs shows that a `github.com.ihcsim/crand` device is allocated to the pod:

```sh
I0823 20:09:35.750554 1358062 kubelet.go:2407] "SyncLoop ADD" source="file" pods=["default/busybox-crand-localhost"]
I0823 20:09:35.750607 1358062 manager.go:836] "Looking for needed resources" needed=1 resourceName="github.com.ihcsim/crand"
I0823 20:09:35.750642 1358062 manager.go:576] "Found pre-allocated devices for resource on pod" resourceName="github.com.ihcsim/crand" containerName="busybox" podUID="a9dc80a0d8f74cefb3be144bbfc1b898" devices=["pfl   2117 ex1"]
# ...
I0823 20:09:58.293380 1358062 kubelet.go:1758] "SyncPod enter" pod="default/busybox-crand-localhost" podUID="a9dc80a0d8f74cefb3be144bbfc1b898"
I0823 20:09:58.293433 1358062 kubelet_pods.go:1774] "Generating pod status" podIsTerminal=false pod="default/busybox-crand-localhost"
I0823 20:09:58.293490 1358062 kubelet_pods.go:1787] "Got phase for pod" pod="default/busybox-crand-localhost" oldPhase="Running" phase="Running"
I0823 20:09:58.293629 1358062 status_manager.go:691] "Ignoring same status for pod" pod="default/busybox-crand-localhost" status={"phase":"Running","conditions":[{"type":"PodReadyToStartContainers","status":"True","lastProbeTime":null,"lastTransitionTime":"2024-08-24T03:09:35Z"},{"type":"Initialized","status":"True","lastProbeTime":null,"lastTransitionTime":"2024-08-24T03:09:35Z"},{"type":"Ready","status":"True","lastProbeTime":null,"lastTransitionTime":"2024-08-24T03:09:35Z"},{"type":"ContainersReady","status":"True","lastProbeTime":null,"lastTransitionTime":"2024-08-24T03:09:35Z"},{"type":"PodScheduled","status":"True","lastProbeTime":null,"lastTransitionTime":"2024-08-24T03:09:35Z"}],"podIP":"172.16.16.4","podIPs":[{"ip":"172.16.16.4"}],"startTime":"2024-08-24T03:09:35Z","containerStatuses":[{"name":"busybox","state":{"running":{"startedAt":"2024-08-24T02:11:58Z"}},"lastState":{},"ready":true,"restartCount":0,"image":"docker.io/library/busybox:latest","imageID":"docker.io/library/busybox@sha256:9ae97d36d26566ff84e8893c64a6dc4fe8ca6d1144bf5b87b2b85a32def253c7","containerID":"containerd://72ebbaf688f4454f47eec5991d36ec02fa82299e92ff6f849751c828f3c69ac0","started":true,"allocatedResourcesStatus":[{"name":"github.com.ihcsim/crand","resources":[{"resourceID":"crand1","health":"Healthy"}]}]}],"qosClass":"BestEffort"}
```

With the `ResourceHealthStatus` feature gate enabled, the kubelet also reports 
the `allocatedResourcesStatus` field in the pod status container status, 
showing that the healthy device `github.com.ihcsim/crand=crand1` is allocated to the pod:

```json
"allocatedResourcesStatus": [
  {
    "name": "github.com.ihcsim/crand",
    "resources": [
      {
        "resourceID": "crand1",
        "health": "Healthy"
      }
    ]
  }
]
```

## Development

To build the plugins:

```sh
make build

make test

make lint
```
