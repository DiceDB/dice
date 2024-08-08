#!/bin/sh
set -e

DATETIME=$(date +"%Y-%m-%d-%H-%M-%S")
INSTANCE_NAME=$SERVICE-$DATETIME

echo "creating a temporary instance to prepare AMI"
INSTANCE_ID=$(aws ec2 run-instances \
    --image-id $AMI_ID \
    --key-name $SSH_KEYPAIR_NAME \
    --instance-type c5.large \
    --associate-public-ip-address \
    --block-device-mappings '[{"DeviceName":"/dev/sda1","Ebs":{"VolumeSize":20,"VolumeType":"gp2"}}]' \
    --tag-specifications "ResourceType=instance,Tags=[{Key=Name, Value=$INSTANCE_NAME}]" \
    --security-group-ids $BENCHMARK_SG \
    --output json \
    | jq -r '.Instances[0].InstanceId')

echo "instance created with ID:" $INSTANCE_ID

INSTANCE_IP=$(aws ec2 describe-instances \
    --instance-ids $INSTANCE_ID \
    --output json \
    | jq -r '.Reservations[0].Instances[0].PublicIpAddress')
echo "instance public IP:" $INSTANCE_IP

echo "waiting for instance to be in running state"
aws ec2 wait instance-running --instance-ids $INSTANCE_ID

sleep 5

echo "copying all the files to the instance"
scp -o StrictHostKeyChecking=no -i $SSH_PEM_PATH -r scripts ubuntu@$INSTANCE_IP:/home/ubuntu
