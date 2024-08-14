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
import pulumi_aws_native as aws_native
import pulumi_command as command
import boto3
from botocore.exceptions import ClientError
import time

# Input parameters
# These can be customized through Pulumi config at Pulumi.<stackname>.yaml
config = pulumi.Config()
instance_name = config.require("instance_name")
instance_type = config.get("instance_type") or "t2.micro"
ami_id = config.get("ami_id") or "ami-0b72821e2f351e396"  # Amazon Linux 2 AMI
vpc_cidr = config.get("vpc_cidr") or "10.0.0.0/16"
subnet_cidr = config.get("subnet_cidr") or "10.0.1.0/24"
region = config.get("region") or "us-east-1"
boto_profile = config.get("boto_profile") or "default"


# Create Internet gateway --> Assign to VPC --> Add internet gateway entry into route table that is assigned to ec2 subnet.
vpc = aws.ec2.Vpc(
    "dice-vpc",
    cidr_block=vpc_cidr,
    enable_dns_hostnames=True,
    enable_dns_support=True,
    tags={"Name": f"{instance_name}-vpc", "Project": "DiceDB"},
)


igw = aws.ec2.InternetGateway(
    "dice-igw",
    vpc_id=vpc.id,
    tags={"Name": f"{instance_name}-igw", "Project": "DiceDB"},
)


route_table = aws.ec2.RouteTable(
    "dice-rt",
    vpc_id=vpc.id,
    routes=[
        aws.ec2.RouteTableRouteArgs(
            cidr_block="0.0.0.0/0",
            gateway_id=igw.id,
        )
    ],
    tags={"Name": f"{instance_name}-rt", "Project": "DiceDB"},
)

subnet = aws.ec2.Subnet(
    "dice-subnet",
    vpc_id=vpc.id,
    cidr_block=subnet_cidr,
    map_public_ip_on_launch=True,
    tags={"Name": f"{instance_name}-subnet", "Project": "DiceDB"},
)

# Associate the route table with the subnet
route_table_association = aws.ec2.RouteTableAssociation(
    "dice-rta",
    subnet_id=subnet.id,
    route_table_id=route_table.id,
)

security_group = aws.ec2.SecurityGroup(
    "dice-sg",
    description="Allow inbound traffic for DiceDB",
    vpc_id=vpc.id,
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
    tags={"Name": f"{instance_name}-sg", "Project": "DiceDB"},
)

# Create a new key pair and store it in AWS Parameter Store
key_pair = aws_native.ec2.KeyPair("dice-keypair", key_name=f"{instance_name}-keypair")


# Function to retrieve the private key using boto3 with retry logic
def get_key_pair_material(key_pair_id):
    boto_session = boto3.Session(profile_name=boto_profile, region_name=region)
    ssm_client = boto_session.client("ssm")
    max_retries = 5
    retry_delay = 5  # seconds

    for attempt in range(max_retries):
        try:
            response = ssm_client.get_parameter(
                Name=f"/ec2/keypair/{key_pair_id}", WithDecryption=True
            )
            return response["Parameter"]["Value"]
        except ClientError as e:
            if e.response["Error"]["Code"] == "ParameterNotFound":
                if attempt < max_retries - 1:
                    time.sleep(retry_delay)
                    continue
                else:
                    raise Exception(
                        f"Failed to retrieve key pair after {max_retries} attempts"
                    )
            else:
                raise


# Retrieve the private key material
private_key = pulumi.Output.all(key_pair.key_pair_id).apply(
    lambda args: get_key_pair_material(args[0])
)

# Create an EC2 instance
instance = aws.ec2.Instance(
    "dicedb",
    instance_type=instance_type,
    ami=ami_id,
    subnet_id=subnet.id,
    vpc_security_group_ids=[security_group.id],
    key_name=key_pair.key_name,
    tags={"Name": instance_name, "Project": "DiceDB"},
)

# Use the remote-exec provisioner to set up DiceDB
# This installs necessary dependencies, clones the DiceDB repo,
# builds the executable, and sets up a systemd service
setup_dice = command.remote.Command(
    "setup-dice",
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
