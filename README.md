# kubelet-plugin

An example of a kubelet plugin that can be used to expose node devices to a 
kubelet. For more information on Kubernetes device plugins, see
https://kubernetes.io/docs/concepts/extend-kubernetes/compute-storage-net/device-plugins/.

It uses [CDI](https://github.com/cncf-tags/container-device-interface/) for device
discovery.

## Testing With Kubelet

For testing purposes, this repo comes with the kubelet v1.31.0 binary. The
`KubeletConfiguration` specification is defined in the `kubelet/kubelet.yaml`
file.

To uncompress and start the kubelet:

```sh
make kubelet
```

The kubelet logs will be written to `kubelet/kubelet.log`.

To run the plugin against the kubelet:

```sh
make run
```

Expect the plugin `pflex.io/block` to register successfully with the kubelet:
```sh
I0823 09:22:35.030339 1309759 server.go:144] "Got registration request from device plugin with resource" resourceName="pflex.io/block"
I0823 09:22:35.030374 1309759 handler.go:95] "Registered client" name="pflex.io/block"
I0823 09:22:35.030966 1309759 manager.go:238] "Device plugin connected" resourceName="pflex.io/block"
```

## Development

To build the plugin:

```sh
make build

make test

make lint
```
