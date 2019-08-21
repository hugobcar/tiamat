# Tiamat
Is used to export one ou more SQS queues metrics to Prometheus. This metrics can be used with [k8s-prometheus-adapter](https://github.com/DirectXMan12/k8s-prometheus-adapter "k8s-prometheus-adapter") to scaling pods k8s (HPA). 

Read [docs/k8s-prometheus-adapter.md](docs/k8s-prometheus-adapter.md "docs/k8s-prometheus-adapter.md") for more informations to use with k8s-prometheus-adapter.
![image](https://user-images.githubusercontent.com/10134807/63375165-a04aa000-c361-11e9-854a-be3729fbc0ea.png)

### Why use?
Currently exist some tools to getting SQS metrics, for example [CloudWatch Exporter](https://github.com/prometheus/cloudwatch_exporter "CloudWatch Exporter"). But it has a long delay of 10 minutes to getting SQS metrics.


Installation
-------------
I use namespace `monitoring`, but this is optional.

### Creating the secret. 
Is necessary getting AWS Key and Secret.

```shell
$ read awskey
<insert your key>
```

```shell
$ read awssecret
<insert your secret>
```

```shell
$ kubectl -n monitoring create secret generic aws-tiamat --from-literal=key=$awskey --from-literal=secret=$awssecret
```

AWS Policy:
```json
{
    "Version": "2012-10-17",
    "Id": "SQSTiamat",
    "Statement": [
        {
            "Sid": "SQSTiamat",
            "Effect": "Allow",
            "Action": "sqs:GetQueueAttributes",
            "Resource": "*"
        }
    ]
}
```

### Creating configMap.
In configmap do you have set URLs Queues (SQS) and region.

Ex.: configmap_tiamat.yaml
```yaml
apiVersion: v1
data:
  config.json: |
    {
        "interval": 10,
        "region": "sa-east-1",
        "format_gauge_name": true,
        "metric_type": "sqs",
        "queue_urls": [
            "https://sqs.sa-east-1.amazonaws.com/123456789012/queue1",
            "https://sqs.sa-east-1.amazonaws.com/123456789012/queue2"
        ]
    }
kind: ConfigMap
metadata:
  name: tiamat
  namespace: monitoring
```

```shell
$ kubectl apply -f configmap_tiamat.yaml
```

- `interval` - Interval to getting queues metrics (in seconds).
- `region` - Location SQS Region.
- `format_gauge_name` - If is `true` the gauge names are reports in format: `tiamat_<aws_account_id>_<queue_name>`.  Default return `tiamat` for all the gauge names.

Ex. `format_gauge_name: true`:
```
.
.
.
# HELP tiamat_123456789012_sqs_queue1 Used to export SQS metrics
# TYPE tiamat_123456789012_sqs_queue1 gauge
tiamat_123456789012_sqs_queue1{metric_type="SQS",queue_account="123456789012",queue_name="queue1",queue_region="sa-east-1",queue_url="https://sqs.sa-east-1.amazonaws.com/739171219021/queue1"} 341547
# HELP tiamat_123456789012_sqs_queue2 Used to export SQS metrics
# TYPE tiamat_123456789012_sqs_queue2 gauge
tiamat_123456789012_queue2{metric_type="SQS",queue_account="123456789012",queue_name="queue2",queue_region="sa-east-1",queue_url="https://sqs.sa-east-1.amazonaws.com/739171219021/queue2"} 110
```

### Deploying Tiamat.
deployment.yaml
```yaml
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  labels:
    run: tiamat 
    release: prometheus-operator
  name: tiamat
  namespace: monitoring
spec:
  replicas: 1
  selector:
    matchLabels:
      run: tiamat
  template:
    metadata:
      labels:
        run: tiamat
    spec:
      containers:
      - name: tiamat
        image: ifoodhub/tiamat:0.0.1
        ports:
            - containerPort: 5000
        imagePullPolicy: Always
        env:
          - name: AWSKEY
            valueFrom:
              secretKeyRef:
                name: aws-tiamat
                key: key
          - name: AWSSECRET
            valueFrom:
              secretKeyRef:
                name: aws-tiamat
                key: secret
          - name: CONFIGMAP 
            value: tiamat
        resources:
          requests:
            memory: "64Mi"
            cpu: "100m"
          limits:
            memory: "128Mi"
            cpu: "128m"
        volumeMounts:
          - name: config-volume
            mountPath: /app/config.json
            subPath: config.json
      volumes:
        - name: config-volume
          configMap:
            name: tiamat
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: tiamat
  namespace: monitoring
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: tiamat-monitoring
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cluster-admin
subjects:
- kind: ServiceAccount
  name: tiamat
  namespace: monitoring
```

```shell
$ kubectl apply -f deployment.yaml
```

### Creating the service.
service.yaml
```yaml
apiVersion: v1
kind: Service
metadata:
  labels:
    release: prometheus-operator
    run: tiamat
  name: tiamat
  namespace: monitoring
spec:
  ports:
  - name: tiamat
    port: 5000
    protocol: TCP
    targetPort: 5000
  selector:
    run: tiamat
  sessionAffinity: None
  type: ClusterIP
```

```shell
$ kubectl apply -f service.yaml
```
