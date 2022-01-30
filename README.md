Using this you can access node external ip address value from your pod.

# Example
## Add init container into your pod
```
...
      initContainers:
      - args:
        - --node-name=$(NODE_NAME)
        - --pod-namespaced-name=$(NAMESPACE)/$(POD_NAME)
        env:
        - name: NAMESPACE
          valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: metadata.namespace
        - name: NODE_NAME
          valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: spec.nodeName
        - name: POD_NAME
          valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: metadata.name
        image: sputniksystemsorg/k8s-propagate-node-status:1643500620
        imagePullPolicy: IfNotPresent
        name: add-label-with-external-ip
...
```

## Setup downwardAPI mountpoint into your pod
```
...
      containers:
      - name: app
...
        volumeMounts:
        - mountPath: /etc/podinfo
          name: podinfo
...
      volumes:
      - downwardAPI:
          defaultMode: 420
          items:
          - fieldRef:
              apiVersion: v1
              fieldPath: metadata.labels
            path: labels
        name: podinfo
...
```
After this your pod labels will be accessible from container from mountpoint `/etc/podinfo`. For example, node's `.status.addresses` field will be exposed into labels `node.status.addresses/internal-ip`, `node.status.addresses/external-ip` etc.
