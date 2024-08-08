#!/bin/bash
set -e

echo "terminating the temporary instance:" $INSTANCE_ID
# X=$(aws ec2 terminate-instances \
#     --instance-ids $INSTANCE_ID \
#     --output json \
#     | jq -r '.TerminatingInstances[0].InstanceId')
