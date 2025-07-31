# Building a Minimal Kubernetes Operator with Custom Metrics Support in Go

This project documents the end-to-end process of building a Kubernetes operator using [Kubebuilder](https://book.kubebuilder.io/), with the goal of exposing both standard cluster metrics (such as memory and CPU usage, pod health) and custom operator metrics (such as reconcile events, errors, duration, etc.).

A core principle of this project is to encapsulate all the logic related to metrics in a reusable Go library, which the operator can consume with minimal coupling or code intrusion. The operator only needs to import and initialize the library — the rest (collection, exposure, and publishing) is handled internally.

In addition to building and running the operator, this guide also covers how to monitor its metrics using Prometheus and Grafana, including:

- Setting up Prometheus to scrape metrics from the operator.

- Designing and deploying a custom Grafana dashboard.

- Visualizing resource usage and reconciliation behavior in a clear, operator-focused layout.

---

## 🎯 Objectives

✅ Build a minimal, working Kubernetes operator using Kubebuilder.

✅ Collect and expose standard runtime metrics (CPU, memory, pod health) from within the operator pod.

✅ Define and expose custom operator metrics (reconcile counts, errors, latency, etc.).

✅ Package metrics logic in a standalone Go module/library that can be reused across operators.

✅ Integrate with Prometheus for metrics exposure.

✅ Create and deploy a Grafana dashboard tailored for operator observability.

✅ Maintain a clean and decoupled architecture with minimal operator code changes required.

---

## 🧱 What's included in this repository?

A working operator scaffolded with Kubebuilder.

A custom Go library (homecalling) that:

Exposes a /metrics endpoint compatible with Prometheus.

Optionally publishes metrics to Kafka or structured logs (future work).

Tracks runtime metrics like memory and CPU usage from inside the pod.

Instruments the operator's controller to emit meaningful metrics.

Kubernetes configuration files:

Deployment manifest for the operator.

ServiceMonitor for Prometheus scraping.

Sample Prometheus setup (for local Minikube use).

A ready-to-import Grafana dashboard for visualizing the operator’s internal state and activity.

## 👥 Who is this for?

This project is intended for:

- Kubernetes operator developers who want to add observability to their controllers.

- Platform engineers and SREs who need metrics to monitor and troubleshoot custom resources.

- Go developers interested in writing instrumentation-friendly code for cloud-native systems.

- Anyone learning Kubebuilder and looking to see how to connect it with real-world monitoring tools.

## ✅ Prerequisites

Before you start, ensure you have the following installed and configured:

- A local **Minikube** cluster running
- **Go** (>= 1.21)
- **Kubebuilder** (`go install sigs.k8s.io/kubebuilder/cmd@latest`)
- Docker: for building Images
- Prometheus & Grafana: either via Helm or manifests

---

## 🏗️ Project Setup

### 1.- Scaffold the Operator with Kubebuilder

```bash
kubebuilder init --domain yourdomain.com --repo github.com/your-username/simple-operator-metrics-example

# I use this:
kubebuilder init --domain demo.local --repo github.com/leobip/demo-operator

# Create the API & Controller
kubebuilder create api \
  --group=demo --version=v1 --kind=Simple \
  --resource=true --controller=true
```

- If prompted:
  - Generate Resource: ✅ Yes
  - Generate Controller: ✅ Yes

### 2.- API Definition

- Edit api/v1/simple_types.go and define a minimal spec and status (Replace or Edit the SimpleSpec & SimpleStatus Types):

```go
// SimpleSpec defines the desired state
type SimpleSpec struct {
    // +kubebuilder:validation:MinLength=1
    // Message is the string to print
    Message string `json:"message"`
}

// SimpleStatus defines the observed state
type SimpleStatus struct {
    // +optional
    // Replied indicates that we’ve seen and logged the Message
    Replied bool `json:"replied,omitempty"`
}
```

### 3.- Controller Logic

- In controllers/simple_controller.go, replace the scaffolded Reconcile logic with:

```go
func (r *SimpleReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    log := log.FromContext(ctx)

    // 1. Fetch the Simple instance
    var simple demov1.Simple
    if err := r.Get(ctx, req.NamespacedName, &simple); err != nil {
        return ctrl.Result{}, client.IgnoreNotFound(err)
    }

    // 2. Log the message
    log.Info("Hallo Welt!", "name", simple.Name, "message", simple.Spec.Message)

    // 3. Update status if not already done
    if !simple.Status.Replied {
        simple.Status.Replied = true
        if err := r.Status().Update(ctx, &simple); err != nil {
            return ctrl.Result{}, err
        }
    }

    return ctrl.Result{}, nil
}
```

- Ensure the reconcile is set up correctly in the file:

```go
func (r *SimpleReconciler) SetupWithManager(mgr ctrl.Manager) error {
    return ctrl.NewControllerManagedBy(mgr).
        For(&demoV1.Simple{}).
        Complete(r)
}
```

### 4.- 📦 Generate CRDs and Manifests

```bash
make generate
make manifests
```

- This generates:
  - CRD YAML in config/crd/bases/
  - RBAC rules in config/rbac/
  - Sample object in config/samples/demo_v1_simple.yaml

#### 4.1.- 🔓 Running without TLS (for local development) - ***Option used in this example***

By default, controllers generated by Kubebuilder expose the metrics endpoint securely over HTTPS (TLS) on port :8443.
For local development, it's often more convenient to disable TLS and use plain HTTP (port :8080) instead — for example, when Prometheus is not set up with CA certificates, or when debugging locally without cert management.

To disable TLS for the metrics endpoint, edit in cmd/main.go:

🛠 Replace:

```go
//flag.StringVar(&metricsAddr, "metrics-bind-address", "0", "The address the metrics endpoint binds to. "+
flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metrics endpoint binds to. "+
"Use :8443 for HTTPS or :8080 for HTTP, or leave as 0 to disable the metrics service.")
...
...
//flag.BoolVar(&secureMetrics, "metrics-secure", true,
flag.BoolVar(&secureMetrics, "metrics-secure", false,
"If set, the metrics endpoint is served securely via HTTPS. Use --metrics-secure=false to use HTTP instead.")
```

- ✅ This change disables HTTPS and serves metrics on plain HTTP. It’s safe for testing or local Minikube setups.
- 🚫 Not recommended for production — no encryption or client validation.

#### 4.2.- 🔐 Running with TLS (recommended for production)

In production or secured clusters, the metrics endpoint should be exposed over HTTPS, typically on port :8443. Kubebuilder supports this out of the box by enabling TLS and using a certificate/key pair.

- To enable TLS again, restore the original values in cmd/main.go
- By default, the manager expects TLS certs to be mounted from:

```bash
/tmp/k8s-webhook-server/serving-certs/tls.crt
/tmp/k8s-webhook-server/serving-certs/tls.key
```

- 🧩 Options for providing TLS certificates:
You have several options to supply valid TLS certificates:
  - Use cert-manager to automatically generate and rotate them (recommended).
    - If you're using cert-manager, create a Certificate and Issuer to automatically provision a secret. cert-manager will ensure auto-renewal.
  - Provide static certificates manually via a Kubernetes Secret.
    - 📝 Example Kubernetes secret (manual approach):📝 Example Kubernetes secret (manual approach):

        ```yaml
        apiVersion: v1
        kind: Secret
        metadata:
        name: metrics-server-cert
        namespace: your-operator-namespace
        type: kubernetes.io/tls
        data:
        tls.crt: <base64-encoded-cert>
        tls.key: <base64-encoded-key>
        ```

    - Use openssl to generate the certs, for example:

        ```bash
        openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
            -keyout tls.key -out tls.crt -subj "/CN=metrics-server/O=metrics"
        ```

        ```bash
        kubectl create secret tls metrics-server-cert \
            --cert=tls.crt --key=tls.key -n your-operator-namespace
        ```

    - 📦 Mounting the Secret in the Deployment: To make the certificates available to the controller, add a volume and volumeMount in the Deployment:

        ```yaml
        # config/manager/manager.yaml
        spec:
        containers:
            - name: manager
            volumeMounts:
                - name: cert
                mountPath: /tmp/k8s-webhook-server/serving-certs
                readOnly: true
        volumes:
            - name: cert
            secret:
                secretName: metrics-server-cert
        ```

      - Make sure this matches the expected path /tmp/k8s-webhook-server/serving-certs.

    - 🚨 TLS and Prometheus: If you are using Prometheus to scrape metrics over HTTPS:
      - Make sure it trusts the CA that signed your tls.crt.
      - Or, use insecureSkipVerify: true (only for internal/trusted environments):

        ```yaml
        spec:
          endpoints:
            - port: https
              scheme: https
              tlsConfig:
                insecureSkipVerify: true
        ```

### 5. 🚀 Install, Run & Expose the Endpoint

Once your CRDs and TLS setup are ready, it's time to install your operator and make the metrics endpoint accessible to Prometheus or any monitoring tool.

#### 5.1 🔧 Install the CRDs into the Cluster

This step installs the CustomResourceDefinitions and necessary RBAC roles:

```bash
make install
```

- 🚨 Note:
If you use the controller-runtime default manager with TLS enabled, and Prometheus is scraping your operator, ensure Prometheus is configured to trust the certificate (or skip TLS verification using insecureSkipVerify: true in your ServiceMonitor, if acceptable).

#### 5.2 ▶️ Run the Operator (Locally or In-Cluster)

- Option A: Run Locally (out-of-cluster)
  - This is useful for development and debugging:

```bash
make run
```

- Your operator will start locally, using your kubeconfig to connect to the cluster
- The metrics will be exposed at localhost:8080/metrics (or :8443 with TLS if enabled).

📈 Metrics Endpoint

This operator exposes Prometheus-compatible metrics at a configurable endpoint.

By default, metrics are served on `:8080` using HTTPS. You can override this behavior via flags.

### 🔧 Available Flags

| Flag | Description | Example |
|------|-------------|---------|
| `--metrics-bind-address` | Address to bind the metrics endpoint | `:8080`, `:8443`, or `0` (disable) |
| `--metrics-secure` | Whether to serve metrics over HTTPS (`true`) or plain HTTP (`false`) | `true` or `false` |

---

## 6.- 🚀 Running Locally with HTTP Metrics

verify the endpoint

```bash
curl http://localhost:8080/metrics
```

response:

```bash
 curl http://localhost:8080/metrics
# HELP certwatcher_read_certificate_errors_total Total number of certificate read errors
# TYPE certwatcher_read_certificate_errors_total counter
certwatcher_read_certificate_errors_total 0
# HELP certwatcher_read_certificate_total Total number of certificate reads
# TYPE certwatcher_read_certificate_total counter
certwatcher_read_certificate_total 0
# HELP controller_runtime_active_workers Number of currently used workers per controller
# TYPE controller_runtime_active_workers gauge
controller_runtime_active_workers{controller="simple"} 0
...
```

## Summary

✅ Should I Keep Using make run or Deploy to Minikube for Further Development?

- Short answer:

  - ✅ Keep using make run while developing the metrics library. Move to Minikube when you're ready to test real integrations (Prometheus scraping, TLS, Kafka, etc.)

- ✅ Advantages of make run (local development):
  - Much faster iteration cycle for testing code changes.
  - Logs appear directly in your terminal.
  - No need to build Docker images or manage Kubernetes manifests.
  - Ideal for iterating on the metrics library and validating which metrics are exposed.
- ⚠️ Limitations of make run:
  - The operator is not deployed as a real Kubernetes Deployment.
  - It's not easily scrappable by Prometheus (unless you manually expose the port via kubectl port-forward).
  - You can't fully test how it interacts with other pods or real Kubernetes resources.

### 🔄 Recommended workflow

- Continue with make run until you have a stable version of the metrics library.
- Then create a proper Deployment manifest and deploy the operator to Minikube for realistic integration tests.

### 🛠 Required Tools for Developing the Metrics Library

| Tool              | Purpose                                                                                                                              |
|-------------------|--------------------------------------------------------------------------------------------------------------------------------------|
| ***Prometheus***     | To collect the metrics from the operator endpoints. Required for Grafana integration.                                               |
| ***Grafana***        | To create and display dashboards using the metrics collected by Prometheus.                                                         |
| ***Kafka***          | *(Optional)* If the library is designed to send metrics to Kafka, deploy it for integration tests.                                  |
| ***The Metrics Library*** | Should be developed in a **separate repository** and imported into any operator. The operator should only call its public API. |
| ***Minikube***       | A lightweight Kubernetes cluster for running real deployments.                                                                      |
| ***PVCs***           | Recommended for persistence in Prometheus, Grafana, and Kafka (via StatefulSets or local PVs).                                      |
| ***Monitoring Repo***| Deploy Prometheus, Grafana, and optionally Kafka in a separate monitoring repository, using Helm or raw manifests. (<https://github.com/leobip/monitoring.git>)                 |

NOTE: To run the operator locally

```bash
make run
```

## ✅ Step-by-Step Guide: Deploying the Operator with Updated Metrics Library in Minikube

### 1. 🧱 Push changes to your metrics library repository

```bash
git add .
git commit -m "Add new metrics logic"
git push origin your-feature-branch
```

### 2. 🏷️ Create a version tag for the library

```bash
git checkout your-feature-branch
git tag v0.1.3  # Replace with the appropriate version
git push origin v0.1.3
```

### 3. 🔁 Update your operator to use the tagged version

- In your operator's go.mod file:
  - Replace the local replace line with the proper module version:

```go
require (
    github.com/your-username/metrics-libs v0.1.3
)
```

- ✅ Comment or remove the replace line like:

```go
// replace github.com/your-username/metrics-libs => ../metrics-libs
```

### 4. Set the name of the image & the Environmental Variables

- in: config/manager/manager.yaml
- Set the name of the Image

```yaml
containers:
      - command:
        - /manager
        args:
          - --leader-elect
          - --health-probe-bind-address=:8081
        image: simple-operator:v0.0.1 # <-- Here the image to use
        name: manager
```

- Set the env vars in the section containers
  - Set the KAFKA_BROKER: name_of_the_kafka_service.namespace.svc.cluster.local:9092
  - CMD to get hte services: kubectl get services -n kafka-namespace
  - value: "kafka.monitoring.svc.cluster.local:9092"

```yaml
...
containers:
      - command:
        - /manager
        args:
          - --leader-elect
          - --health-probe-bind-address=:8081
        image: simple-operator:v0.0.1
        name: manager
        env:
          - name: METRICS_NAMESPACE
            value: "metrics-ex"
          - name: METRICS_CLUSTER
            value: "local-cluster"
          - name: METRICS_RESOURCE_KIND
            value: "MyResource"
          - name: METRICS_CONTROLLER_NAME
            value: "simple-operator"
          - name: METRICS_CONTROLLER_VERSION
            value: "v0.0.1"
          - name: KAFKA_BROKER
            value: "kafka.monitoring.svc.cluster.local:9092"
          - name: KAFKA_TOPIC
            value: "metrics"
...
```

### 5. 📦 Fetch the new library version and tidy up

```bash
go get github.com/your-username/metrics-libs@v0.1.3
go mod tidy
```

### 6. 🐳 Build the operator Docker image

- ⚠️ Make sure your Docker environment is set to Minikube:

```bash
eval $(minikube docker-env)
```

- Now build the image (replace with your operator name/tag):

```bash
make docker-build IMG=demo-operator:latest
```

### 7. 📦 Load the image into Minikube

```bash
minikube image load demo-operator:latest
```

### 8. 🚀 Deploy the operator into the cluster

- Make sure your kube context points to Minikube and your operator config is updated with the right image tag:

```bash
make deploy IMG=demo-operator:latest
```

- Or, if using kustomize, edit your config/manager/kustomization.yaml:

```yaml
images:
- name: controller
  newName: demo-operator
  newTag: dev
```

- Then:

```bash
make deploy IMG=demo-operator:dev
```

### 9. ✅ Verify the deployment

```bash
kubectl get pods -n your-namespace
kubectl logs deployment/demo-operator -n your-namespace
```

### 10. To Delete or Unistall the Operator from Minikube

```bash
make undeploy
```
