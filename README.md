# kubelet-plugin

An example of a kubelet plugin that can be used to expose node devices to a 
kubelet. For more information on Kubernetes device plugins, see
https://kubernetes.io/docs/concepts/extend-kubernetes/compute-storage-net/device-plugins/.

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

## Development

To build the plugin:

```sh
make build

make test

make lint
```
