#!/bin/bash

# List of services
services=("scylla" "redis" "message-broker" "grafana" "prometheus")

# Loop through the services and generate YAML files
for service in "${services[@]}"; do
  # Define file names for deployment and service
  deployment_file="${service}-deployment.yaml"
  service_file="${service}-service.yaml"

  # Create resources using kubectl create -f
  kubectl create -f $deployment_file
  kubectl create -f $service_file

  echo "$service deployment and service created with $deployment_file and $service_file."
done
