## Khan - Pod Connection Tracking Metrics Exporter
[![Go Report Card](https://goreportcard.com/badge/github.com/att-cloudnative-labs/khan)](https://goreportcard.com/report/github.com/att-cloudnative-labs/khan)
[![Build Status](https://travis-ci.org/att-cloudnative-labs/khan.svg?branch=master)](https://travis-ci.org/att-cloudnative-labs/khan)
#

Khan captures connection tracking snapshots on Pods, and Nodes and exposes them as  prometheus metrics. Note that the metrics don't constitute realtime connection info, only snapshots that are polled with a default period of 30s.

The use case for this application is for tracking down pods/services that are leaking connections or finding an unknown client that is overloading a server.

This application is composed of a 'controller' that runs as a deployment. The controller is mainly an API for the node agents to retrieve mappings of IP-to-pod for IPs found in the conntrack table. The 'agent' runs as a daemonset on each node and captures the conntrack table and converts it to a set of prometheus metrics.
