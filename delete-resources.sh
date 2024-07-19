#!/bin/bash

# Set the threshold time (1 hour ago)
threshold_time=$(date -d '-1 hour' --utc +%FT%TZ)

# Delete deployments less than 1 hour old
kubectl get deployments -o json | jq -r --arg threshold_time "$threshold_time" \
  '.items[] | select(.metadata.creationTimestamp > $threshold_time) | .metadata.name' | \
  xargs -I{} kubectl delete deployment {}

# Delete services less than 1 hour old
kubectl get services -o json | jq -r --arg threshold_time "$threshold_time" \
  '.items[] | select(.metadata.creationTimestamp > $threshold_time) | .metadata.name' | \
  xargs -I{} kubectl delete service {}

# Delete PVCs less than 1 hour old
kubectl get pvc -o json | jq -r --arg threshold_time "$threshold_time" \
  '.items[] | select(.metadata.creationTimestamp > $threshold_time) | .metadata.name' | \
  xargs -I{} kubectl delete pvc {}