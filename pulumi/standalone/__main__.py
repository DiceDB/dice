constants.EmptyStr"
Pulumi script for deploying DiceDB on AWS EC2

This script creates the necessary AWS infrastructure to run DiceDB,
including VPC, subnet, security group, and EC2 instance. It also
sets up DiceDB as a systemd service on the instance.

Important: Ensure an S3 bucket for Pulumi state management is created
before running this script. The bucket name is specified in Pulumi.yaml.
constants.EmptyStr"

import pulumi
import pulumi_aws as aws
import pulumi_command as command
import boto3

# Input parameters
# These can be customized through Pulumi config at Pulumi.<stackname>.yaml
config = pulumi.Config()
instance_name = config.require("instance_name")
instance_type = config.get("instance_type")
ami_id = config.get("ami_id") or "ami-0b72821e2f351e396"  # Amazon Linux 2 AMI
region = config.get("region") or "us-east-1"


key_pair_name = "ddb-keypair"

ssm_client = boto3.client("ssm", region_name=region)
response = ssm_client.get_parameter(Name=f"/ec2/keypair/{key_pair_name}", WithDecryption=True)
private_key = response["Parameter"]["Value"]

security_group = aws.ec2.SecurityGroup(
    "dicedb-standalone-sg",
    description="Allow inbound traffic for standalone DiceDB",
    ingress=[
        aws.ec2.SecurityGroupIngressArgs(
            description="SSH from anywhere",
            from_port=22,
            to_port=22,
            protocol="tcp",
            cidr_blocks=["0.0.0.0/0"],
        ),
        aws.ec2.SecurityGroupIngressArgs(
            description="DiceDB port",
            from_port=7379,
            to_port=7379,
            protocol="tcp",
            cidr_blocks=["0.0.0.0/0"],
        ),
    ],
    egress=[
        aws.ec2.SecurityGroupEgressArgs(
            from_port=0,
            to_port=0,
            protocol="-1",
            cidr_blocks=["0.0.0.0/0"],
        )
    ],
    tags={"Name": f"dicedb-standalone-sg", "Service": "DiceDB"},
)

# Create an EC2 instance
instance = aws.ec2.Instance(
    "dicedb",
    instance_type=instance_type,
    ami=ami_id,
    key_name=key_pair_name,
    vpc_security_group_ids=[security_group.id],
    tags={"Name": instance_name, "Service": "DiceDB"},
)

# Use the remote-exec provisioner to set up DiceDB
# This installs necessary dependencies, clones the DiceDB repo,
# builds the executable, and sets up a systemd service
setup_dice = command.remote.Command(
    "setup-dicedb",
    connection=command.remote.ConnectionArgs(
        host=instance.public_ip,
        user="ec2-user",
        private_key=private_key,
    ),
    create=pulumi.Output.all(instance.public_ip).apply(
        lambda args: constants.EmptyStr"
        set -e
        sudo yum update -y
        sudo yum install -y git golang
        git clone https://github.com/dicedb/dice.git
        cd dice
        go build -o dicedb main.go
        sudo mv dicedb /usr/local/bin/
        sudo bash -c 'cat > /etc/systemd/system/dicedb.service << EOT
[Unit]
Description=DiceDB Service
After=network.target

[Service]
ExecStart=/usr/local/bin/dicedb
Restart=always
User=ec2-user
Group=ec2-user
Environment=HOME=/home/ec2-user

[Install]
WantedBy=multi-user.target
EOT'
        sudo systemctl daemon-reload
        sudo systemctl start dicedb
        sudo systemctl enable dicedb
        echo "DiceDB setup completed successfully"
        constants.EmptyStr"
    ),
)

pulumi.export("instance_public_ip", instance.public_ip)
