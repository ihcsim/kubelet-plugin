# kubelet-plugin

An example of a kubelet plugin that can be used to expose node devices to a 
kubelet. For more information on Kubernetes device plugins, see
https://kubernetes.io/docs/concepts/extend-kubernetes/compute-storage-net/device-plugins/.

## Development

To build the plugin, run:

```sh
make build

make test

make lint
```
