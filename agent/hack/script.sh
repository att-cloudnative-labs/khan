# set appmapping cache
curl -XPOST "localhost:8443/appmapping" -H "Content-Type: application/json" -d '{"10.240.19.1":{"namespace":"unknown","name":"node"}}'
# get appmapping cache
curl "localhost:8443/appmapping"

# get connections
curl "localhost:8443/connections"

# trigger conntrack update
curl -XPOST "localhost:8443/update" -H "Content-Type: application/json" -d '{}'


# /tmp/nf_script.sh
#! /bin/bash
(printf "["; cat /tmp/nf_conntrack | grep tcp | tr -s ' ' | /usr/bin/sed -E 's/ipv4 2 tcp 6 ([0-9]+) ([A-Z_]+) src=([0-9]+.[0-9]+.[0-9]+.[0-9]+) dst=([0-9]+.[0-9]+.[0-9]+.[0-9]+) sport=([0-9]+) dport=([0-9]+) src=([0-9]+.[0-9]+.[0-9]+.[0-9]+) dst=([0-9]+.[0-9]+.[0-9]+.[0-9]+) sport=[0-9]+ dport=[0-9]+ \[([A-Z]+)\].*/\{\"req_src_ip\":\"\3\",\"req_dst_ip\":\"\4\",\"req_sport\":\"\5\",\"req_dport\":\"\6\",\"res_src_ip\":\"\7\",\"res_dst_ip\":\"\8\",\"state\":\"\2\",\"ct_flag\":\"\9\"},/g'; printf "{}]") | jq '.'


# get appmapping for single node from kubectl
kubectl get po --all-namespaces --field-selector=spec.nodeName=ip-10-223-5-75.us-west-2.compute.internal -o json | jq -r 'reduce .items[] as $i ({}; .[$i.status.podIP] = {"namespace": $i.metadata.namespace, "appName": $i.metadata.labels.app, "podName": $i.metadata.name})'

# post the appmapping
kubectl get po --all-namespaces --field-selector=spec.nodeName=ip-10-223-5-75.us-west-2.compute.internal -o json | jq -r -c 'reduce .items[] as $i ({}; .[$i.status.podIP] = {"namespace": $i.metadata.namespace, "appName": $i.metadata.labels.app, "podName": $i.metadata.name})' | curl -X POST "localhost:8443/appmapping" -H "Content-Type: application/json" -d @-

# add node mapping
kubectl get po --all-namespaces --field-selector=spec.nodeName=ip-10-223-5-75.us-west-2.compute.internal -o json | jq -r -c 'reduce .items[] as $i ({}; .[$i.status.podIP] = {"namespace": $i.metadata.namespace, "appName": $i.metadata.labels.app, "podName": $i.metadata.name}) | .["10.240.19.1"]={"namespace": "kube-node", "appName":"ip-10-223-5-75.us-west-2.compute.internal", "podName": "ip-10-223-5-75.us-west-2.compute.internal"}'
kubectl get po --all-namespaces --field-selector=spec.nodeName=ip-10-223-5-75.us-west-2.compute.internal -o json | jq -r -c 'reduce .items[] as $i ({}; .[$i.status.podIP] = {"namespace": $i.metadata.namespace, "appName": $i.metadata.labels.app, "podName": $i.metadata.name}) | .["10.240.19.1"]={"namespace": "kube-node", "appName":"ip-10-223-5-75.us-west-2.compute.internal", "podName": "ip-10-223-5-75.us-west-2.compute.internal"}' | curl -X POST "localhost:8443/appmapping" -H "Content-Type: application/json" -d @-