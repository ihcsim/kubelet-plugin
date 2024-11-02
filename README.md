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

Expect the plugin `github.com.ihcsim/kvm` to register successfully with the kubelet:

```sh   
I1031 18:55:45.367492   26662 server.go:144] "Got registration request from device plugin with resource" resourceName="github.com.ihcsim/kvm"
I1031 18:55:45.367538   26662 handler.go:94] "Registered client" name="github.com.ihcsim/kvm"
# <snipped>
I1031 18:55:45.368646   26662 manager.go:229] "Device plugin connected" resourceName="github.com.ihcsim/kvm"
# <snipped>
I1031 18:55:55.375313   26662 client.go:91] "State pushed for device plugin" resource="github.com.ihcsim/kvm" resourceCapacity=1
I1031 18:55:55.382659   26662 manager.go:278] "Processed device updates for resource" resourceName="github.com.ihcsim/kvm" totalCount=1 healthyCount=1
# <snipped>
```

Deploy the provided busybox pod to the kubelet as a static pod:

```sh
make deploy
```

The kubelet logs shows that 1 `github.com.ihcsim/kvm` device is allocated to the
`kvm` pod:

```sh
I1102 20:33:21.794935    2195 config.go:398] "Receiving a new pod" pod="default/kvm-ncqxp"
# <snipped>
I1102 20:33:21.796150    2195 manager.go:580] "Need devices to allocate for pod" deviceNumber=1 resourceName="github.com.ihcsim/kvm" podUID="d2a6778a-f42d-4880-9ed8-ec5ec6ea6303" con 194569 tainerName="kvm-client"
I1102 20:33:21.796167    2195 manager.go:991] "Plugin options indicate to skip GetPreferredAllocation for resource" resourceName="github.com.ihcsim/kvm"
I1102 20:33:21.796176    2195 manager.go:859] "Making allocation request for device plugin" devices=["kvm0"] resourceName="github.com.ihcsim/kvm"
# <snipped>
I1102 20:44:05.754527   18375 manager.go:813] "Looking for needed resources" needed=1 resourceName="github.com.ihcsim/kvm"
I1102 20:44:05.754530   18375 config.go:105] "Looking for sources, have seen" sources=["api","file"] seenSources={"file":{}}
I1102 20:44:05.754536   18375 manager.go:555] "Found pre-allocated devices for resource on pod" resourceName="github.com.ihcsim/kvm" containerName="kvm-client" podUID="d2a6778a-f42d- 231430 4880-9ed8-ec5ec6ea6303" devices=["kvm0"]
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

## Debugging

Download `delve` from https://github.com/go-delve/delve.

Start by building the debug image:

```sh
make image-debug
```

Use Docker to start the debug container:

```sh
docker run -p 40000:40000 <img>
```

Now, connect to the debug container:

```sh
dlv connect 127.0.0.1:40000
```

For more information, see https://ko.build/features/debugging/

## Release

To create a new release, a new tag is required:

```sh
git tag -a <new_tag> -m "<new_tag_message>"

git push origin <new_tag>
```

Manually start the `default` GHA pipeline.
