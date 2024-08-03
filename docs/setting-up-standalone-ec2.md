# Setting Up DiceDB Standalone EC2 instance on AWS using Pulumi

## Prerequisites

- AWS CLI installed and configured
- [Pulumi CLI installed](https://www.pulumi.com/docs/install/)
- This Pulumi project uses an S3 backend for state management. Ensure that the specified S3 bucket (or any suitable bucketname mentioned in the backend of Pulumi configuration) is created before running the Pulumi script. Alternatively we can use the pulumi cloud for storing the state.

## Setup Steps

### 0. Setup ENVIRONMENT variables

```
export AWS_REGION=us-east-1
export AWS_PROFILE=<profile_name>
export STACK_NAME=standalone
export PULUMI_CONFIG_PASSPHRASE=
```

### 1. Configure AWS Credentials

Ensure the AWS credentials are properly set up:

```bash
aws configure
```

Follow the prompts to configure the AWS Access Key ID, Secret Access Key, and default region.

### 2. Ensure Key Pair

Make sure you have a keypair named `ddb-keypair` and you have uploaded the PEM file into Parameter Store
with the key `/ec2/keypair/ddb-keypair`.

### 2. Create and Activate a Virtual Environment

```bash
cd pulumi
python -m venv venv
source venv/bin/activate
```

### 3. Install Required Dependencies

```bash
pip install -r requirements.txt
```

### 4. Initialize or Select a Pulumi Stack

Create a new stack

```bash
cd $STACK_NAME
pulumi stack init $STACK_NAME
```

Or select an existing stack:

```bash
pulumi stack select $STACK_NAME
```

When asked for the password or paraphrase, just hit `ENTER`.

### 5. Preview the Infrastructure Changes

```bash
pulumi preview
```

This command shows you what changes will be made without actually applying them.

### 6. Deploy the Infrastructure

When you're ready to create the resources:

```bash
pulumi up
```

Review the proposed changes and confirm to proceed.

### 7. Access Your Resources

After deployment, Pulumi will output important information like the instance's public IP and SSH command.
Make note of these for accessing the DiceDB instance; or SSH into the instance via the console.

## Clean Up

To destroy the created resources when they're no longer needed:

```bash
pulumi destroy
```

Confirm the action when prompted.

## Troubleshooting

- If you encounter any issues with AWS permissions, ensure your AWS CLI is configured with the correct credentials and has the necessary IAM permissions.
- For Pulumi-specific problems, consult the [Pulumi documentation](https://www.pulumi.com/docs/).
- If you face issues with DiceDB setup, check the EC2 instance logs or SSH into the instance for further investigation.

## Additional Resources

- [Pulumi AWS Documentation](https://www.pulumi.com/docs/clouds/aws/)
- [AWS CLI Configuration Guide](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-configure.html)
