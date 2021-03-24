# Tiamat
Is used to export one or more SQS queues metrics to Prometheus. This metrics can be used with [k8s-prometheus-adapter](https://github.com/DirectXMan12/k8s-prometheus-adapter "k8s-prometheus-adapter") to scaling pods k8s (HPA). 

Read [docs/k8s-prometheus-adapter.md](docs/k8s-prometheus-adapter.md "docs/k8s-prometheus-adapter.md") for more informations to use with k8s-prometheus-adapter.
![image](https://user-images.githubusercontent.com/10134807/64730080-e1435980-d4b4-11e9-8156-93d3312e1bfb.png)

### Why use?
Currently exist some tools to getting SQS metrics, for example [CloudWatch Exporter](https://github.com/prometheus/cloudwatch_exporter "CloudWatch Exporter"). But it has a long delay of 5~10 minutes to getting SQS metrics.


Installation
-------------
I use namespace `monitoring`, but this is optional.

### AWS Permissions

The following AWS policy is needed to use this program:

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

If the program is running outside AWS and you don't have any configured credentials in the instance environment, you can set an AWS key and secret this way:

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

If the program is running inside an instance that already has an IAM role attached and you don't want to use explicit credentials, set the key and secret as empty by running:

```shell
$ kubectl -n monitoring create secret generic aws-tiamat --from-literal=key="" --from-literal=secret=""
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
        "logs_enabled": true,
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
- `format_gauge_name` - If is `true` the gauge names are reports in format: `tiamat_<aws_account_id>_sqs_<queue_name>`.  Default return `tiamat` for all the gauge names. **Note**: The `-` characters in queue names will be replaced by` _` and the queues names are converted in lowercase.

Ex. `format_gauge_name: true`:
```
.
.
.
# HELP tiamat_123456789012_sqs_queue1 Legacy metric, use total metrics instead
# TYPE tiamat_123456789012_sqs_queue1 gauge
tiamat_123456789012_sqs_queue1{metric_type="sqs",queue_account="123456789012",queue_name="queue1",queue_region="sa-east-1",queue_url="https://sqs.sa-east-1.amazonaws.com/739171219021/queue1"} 341547
# HELP tiamat_123456789012_sqs_queue1_total SQS Total Messages metrics
# TYPE tiamat_123456789012_sqs_queue1_total gauge
tiamat_123456789012_sqs_queue1_total{metric_type="sqs",queue_account="123456789012",queue_name="queue1",queue_region="sa-east-1",queue_url="https://sqs.sa-east-1.amazonaws.com/739171219021/queue1"} 341547
# HELP tiamat_123456789012_sqs_queue1_visible SQS Visible Messages metrics
# TYPE tiamat_123456789012_sqs_queue1_visible gauge
tiamat_123456789012_sqs_queue1_visible{metric_type="sqs",queue_account="123456789012",queue_name="queue1",queue_region="sa-east-1",queue_url="https://sqs.sa-east-1.amazonaws.com/739171219021/queue1"} 302473
# HELP tiamat_123456789012_sqs_queue1_in_flight SQS In Fight Messages metrics
# TYPE tiamat_123456789012_sqs_queue1_in_flight gauge
tiamat_123456789012_sqs_queue1_in_flight{metric_type="sqs",queue_account="123456789012",queue_name="queue1",queue_region="sa-east-1",queue_url="https://sqs.sa-east-1.amazonaws.com/739171219021/queue1"} 1253
```
- `metric_type` - In this moment is possible getting 'sqs' metrics, but in the future it will be possible to set others metrics (redis, dynamodb...).
- `logs_enabled` - If is `true` logs are disabled. Default is `false`.

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
        image: hugobcar/tiamat:0.0.2
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
