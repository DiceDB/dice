#!/bin/bash

set -e

source setup_instance1.sh

echo "executing setup_redis.sh on $INSTANCE_IP"
ssh -o StrictHostKeyChecking=no -i $SSH_PEM_PATH ubuntu@$INSTANCE_IP "/bin/bash /home/ubuntu/scripts/setup_redis.sh"
ssh -o StrictHostKeyChecking=no -i $SSH_PEM_PATH ubuntu@$INSTANCE_IP "/bin/bash /home/ubuntu/scripts/setup_memtier.sh"
ssh -o StrictHostKeyChecking=no -i $SSH_PEM_PATH ubuntu@$INSTANCE_IP "/bin/bash /home/ubuntu/scripts/memtier_1.sh"
scp -i $SSH_PEM_PATH ubuntu@$INSTANCE_IP:/home/ubuntu/results-memtier-1-redis.txt .
ssh -o StrictHostKeyChecking=no -i $SSH_PEM_PATH ubuntu@$INSTANCE_IP "/bin/bash /home/ubuntu/scripts/stop_redis.sh"
bash terminate_instance.sh
