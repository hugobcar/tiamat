# Prometheus Adapter for Kubernetes Metrics APIs with tiamat SQS metrics.
I presumed that you installed k8s-prometheus-adapter (https://github.com/DirectXMan12/k8s-prometheus-adapter) and you install tiamat.

Your prometheus should is getting tiamat metrics.

Use this k8s-prometheus-adapter configs:

```yaml
apiVersion: v1
data:
  config.yaml: |
    externalRules:
    - metricsQuery: <<.Series>>{queue_name!=""}
      name:
        as: ""
        matches: null
      seriesQuery: '{__name__=~"tiamat.*"}'
kind: ConfigMap
metadata:
  labels:
    app: prometheus-adapter
    chart: prometheus-adapter-1.2.0
    heritage: Tiller
    release: prometheus-adapter
  name: prometheus-adapter
  namespace: kube-system
```

Note that we use `externalRules`.

```shell
$ kubectl get --raw /apis/external.metrics.k8s.io/v1beta1 | jq
{
  "kind": "APIResourceList",
  "apiVersion": "v1",
  "groupVersion": "external.metrics.k8s.io/v1beta1",
  "resources": [
    {
      "name": "tiamat_123456789012_sqs_queue1",
      "singularName": "",
      "namespaced": true,
      "kind": "ExternalMetricValueList",
      "verbs": [
        "get"
      ]
    },
    {
      "name": "tiamat_123456789012_sqs_queue2",
      "singularName": "",
      "namespaced": true,
      "kind": "ExternalMetricValueList",
      "verbs": [
        "get"
      ]
    }
  ]
}
```

```shell
$ kubectl get --raw /apis/external.metrics.k8s.io/v1beta1/namespaces/*/tiamat_123456789012_sqs_queue1 | jq .
{
  "kind": "ExternalMetricValueList",
  "apiVersion": "external.metrics.k8s.io/v1beta1",
  "metadata": {
    "selfLink": "/apis/external.metrics.k8s.io/v1beta1/namespaces/%2A/tiamat_123456789012_sqs_queue1"
  },
  "items": [
    {
      "metricName": "tiamat_123456789012_sqs_queue1",
      "metricLabels": {
        "__name__": "tiamat_123456789012__sqs_queue1",
        "endpoint": "tiamat",
        "instance": "100.XXX.XXX.XXX:5000",
        "job": "tiamat",
        "metric_type": "SQS",
        "namespace": "monitoring",
        "pod": "tiamat-67556f7c5b-bz25m",
        "queue_account": "123456789012",
        "queue_name": "queue1",
        "queue_region": "sa-east-1",
        "queue_url": "https://sqs.sa-east-1.amazonaws.com/123456789012/queue1",
        "service": "tiamat"
      },
      "timestamp": "2019-08-21T01:23:45Z",
      "value": "341547"
    }
  ]
}
```

Example scaling metrics CPU and SQS (tiamat):

```yaml
apiVersion: autoscaling/v2beta1
kind: HorizontalPodAutoscaler
metadata:
  name: app
  namespace: app
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: app
  minReplicas: 2
  maxReplicas: 5
  metrics:
  - type: Resource
    resource:
      name: cpu
      targetAverageUtilization: 75
  - type: External
    external:
      metricName: tiamat_123456789012_sqs_queue1
      targetValue: 500
```

**Note:** The `-` characters in queue names will be replaced by` _` and the queues names are converted in lowercase.
