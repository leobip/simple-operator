## 📈 Metrics Endpoint

This operator exposes Prometheus-compatible metrics at a configurable endpoint.

By default, metrics are served on `:8080` using HTTPS. You can override this behavior via flags.

### 🔧 Available Flags

| Flag | Description | Example |
|------|-------------|---------|
| `--metrics-bind-address` | Address to bind the metrics endpoint | `:8080`, `:8443`, or `0` (disable) |
| `--metrics-secure` | Whether to serve metrics over HTTPS (`true`) or plain HTTP (`false`) | `true` or `false` |

---

## 🚀 Running Locally with HTTP Metrics

For local development (no TLS), replace:

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

verify

```bash
curl http://localhost:8080/metrics
```
