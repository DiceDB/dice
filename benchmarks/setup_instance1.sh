#!/bin/sh
set -e

export INSTANCE_ID=i-07a3f5fc71bc63046
export INSTANCE_IP=107.20.129.231
echo "export INSTANCE_ID=$INSTANCE_ID"
echo "export INSTANCE_IP=$INSTANCE_IP"


echo "copying all the files to the instance"
scp -o StrictHostKeyChecking=no -i $SSH_PEM_PATH -r scripts ubuntu@$INSTANCE_IP:/home/ubuntu
