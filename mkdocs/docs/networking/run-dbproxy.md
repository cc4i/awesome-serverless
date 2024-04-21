#

## Description

The Cloud Run connects to Private AlloyDB through [AlloyDB Auth Proxy](https://cloud.google.com/alloydb/docs/auth-proxy/overview), which's running along side with Cloud Run as a sidecar. All connections are pooled by levaraging the python library - [sqlalchemy](https://docs.sqlalchemy.org/en/20/).

## Guide
```bash
Skaffold build --cache-artifacts=fals
Skaffold run
```

## References
- [Code](../../../asset/run-dbproxy/)
- [AlloyDB Auth Proxy](https://cloud.google.com/alloydb/docs/auth-proxy/connect)