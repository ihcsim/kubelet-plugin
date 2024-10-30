# kubelet-plugin

This repo contains examples of kubelet plugins that can be used to expose the
following node devices to the Kubelet:

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
make run.kvm
```

Expect the plugin `github.com.ihcsim/crand` to register successfully with the kubelet:

```sh
I0823 09:22:35.030339 1309759 server.go:144] "Got registration request from device plugin with resource" resourceName="github.com.ihcsim/crand"
I0823 09:22:35.030374 1309759 handler.go:95] "Registered client" name="github.com.ihcsim/crand"
I0823 09:22:35.030966 1309759 manager.go:238] "Device plugin connected" resourceName="github.com.ihcsim/crand"
# ...
I0823 09:04:14.473749 1352537 setters.go:329] "Updated capacity for device plugin" plugin="github.com.ihcsim/crand" capacity=3
```

Similarly, the plugin `github.com.ihcsim/kvm` should also register successfully:
```sh   
I0908 13:46:44.317313  211403 server.go:144] "Got registration request from device plugin with resource" resourceName="github.com.ihcsim/kvm"
I0908 13:46:44.317346  211403 handler.go:95] "Registered client" name="github.com.ihcsim/kvm"
I0908 13:46:44.318228  211403 manager.go:238] "Device plugin connected" resourceName="github.com.ihcsim/kvm"
# ...
```


Deploy the provided busybox pod to the kubelet as a static pod:

```sh
make deploy
```

The kubelet logs shows that 2 `github.com.ihcsim/crand` device is allocated to 
the `busybox-crand` pod and 1 `github.com.ihcsim/kvm` device is allocated to the
`busybox-kvm` pod:

```sh
I0908 13:52:08.084307  211403 config.go:397] "Receiving a new pod" pod="default/busybox-crand-localhost"
I0908 13:52:08.084353  211403 kubelet.go:2407] "SyncLoop ADD" source="file" pods=["default/busybox-crand-localhost"]
I0908 13:52:08.084371  211403 manager.go:836] "Looking for needed resources" needed=2 resourceName="github.com.ihcsim/crand"
I0908 13:52:08.084384  211403 manager.go:560] "Pods to be removed" podUIDs=["a9dc80a0d8f74cefb3be144bbfc1b898"]
I0908 13:52:08.084393  211403 manager.go:601] "Need devices to allocate for pod" deviceNumber=2 resourceName="github.com.ihcsim/crand" podUID="8b5e7c6badf1ce0c12118bdb12ce9a8c" containerName="busybox"
I0908 13:52:08.084404  211403 manager.go:1014] "Plugin options indicate to skip GetPreferredAllocation for resource" resourceName="github.com.ihcsim/crand"
I0908 13:52:08.084409  211403 file.go:201] "Reading config file" path="/home/isim/workspace/kubelet-plugin/kubelet/run/pods/busybox-kvm.yaml"
I0908 13:52:08.084415  211403 manager.go:882] "Making allocation request for device plugin" devices=["crand1","crand0"] resourceName="github.com.ihcsim/crand"
# ...
I0908 13:54:14.259136  211403 config.go:397] "Receiving a new pod" pod="default/busybox-kvm-localhost"
I0908 13:54:14.259192  211403 kubelet.go:2407] "SyncLoop ADD" source="file" pods=["default/busybox-kvm-localhost"]
I0908 13:54:14.259236  211403 manager.go:836] "Looking for needed resources" needed=1 resourceName="github.com.ihcsim/kvm"
I0908 13:54:14.259350  211403 manager.go:601] "Need devices to allocate for pod" deviceNumber=1 resourceName="github.com.ihcsim/kvm" podUID="79afb85449be9e045489922c8d983fe8" containerName="busybox"
I0908 13:54:14.259386  211403 manager.go:1014] "Plugin options indicate to skip GetPreferredAllocation for resource" resourceName="github.com.ihcsim/kvm"
I0908 13:54:14.259425  211403 manager.go:882] "Making allocation request for device plugin" devices=["github.com.ihcsim/kvm"] resourceName="github.com.ihcsim/kvm"
# ...
```

With the `ResourceHealthStatus` feature gate enabled, the kubelet also reports 
the `allocatedResourcesStatus` field in the pod status container status, 
showing that the healthy device `github.com.ihcsim/crand=crand1` is allocated to the pod:

```json
{
  "allocatedResourcesStatus": [
    {
      "name": "github.com.ihcsim/crand",
      "resources": [
        {
          "resourceID": "crand1",
          "health": "Healthy"
        },
        {
          "resourceID": "crand0",
          "health": "Healthy"
        }
      ]
    }
  ]
}

{
  "allocatedResourcesStatus": [
    {
      "name": "github.com.ihcsim/kvm",
      "resources": [
        {
          "resourceID": "github.com.ihcsim/kvm",
          "health": "Healthy"
        }
      ]
    }
  ]
}
```

## Development

To build the plugins:

```sh
make build

make test

make lint
```

To update the Docker image and DaemonSet YAML, download [ko](https://ko.build/).

To build and push a new Docker image:

```sh
make image [KO_DOCKER_REPO=<docker_repo>]
```

To update the DaemonSet YAML with the latest image:

```sh
make yaml [KO_DOCKER_REPO=<docker_repo>]
```

To create a new release, a new tag is required:

```sh
git tag -a <new_tag> -m "<new_tag_message>"

git push origin <new_tag>
```

Manually start the `default` GHA pipeline.
