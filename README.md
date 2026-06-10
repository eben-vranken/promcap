# promcap
🔭 Drop-in Prometheus wrapper that caps metric cardinality at the source before unbounded labels OOM your monitoring stack.

## Drop-in scope

Capped: WithLabelValues, With, GetMetricWith, GetMetricWithLabelValues, Reset.

Not yet wrapped: CurryWith, Delete, DeleteLabelValues. Code relying on these
is not yet a drop-in replacement.
