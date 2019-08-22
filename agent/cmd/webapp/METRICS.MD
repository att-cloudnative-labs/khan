# Service metrics
| Metric | Description | Type | Labels |
|--------|-------------|------|--------|
| client_error | DME client errors | counter | call_alias, error_type, x_att_clientid, x_aeg_client_id, x_aeg_route_offer |
| current_concurrency | Current in-progress REST requests | gauge |  |
| custom_metric_example | This is the number of times the metric was called. | counter | label1 |
| http_downstream_requests_ms | Response time ms | histogram | call_alias, method |
| http_downstream_requests_summary_ms | Response time ms | summary | call_alias, method, status, x_att_clientid, x_aeg_client_id, x_aeg_route_offer |
| http_requests_ms | Service response time | histogram | path, method |
| http_requests_summary_ms | Service response time with more labels | summary | path, method, status_code, x_att_clientid, x_aeg_client_id, x_aeg_route_offer |
| yawl | log statistics | counter | loglevel, status, code, decision |


Last update: 2019-07-11 18:07:29.904128 -0700 PDT m=+0.013939641
		