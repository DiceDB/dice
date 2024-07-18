
# Setting Up DiceDB Infrastructure on AWS using Pulumi

## Prerequisites

- AWS CLI installed and configured
- Pulumi CLI installed
- This Pulumi project uses an S3 backend for state management. Ensure that the specified S3 bucket (dice-pulumi or any suitable bucketname) is created before running the Pulumi script. Alternatively we can use the pulumi cloud for storing the state

## Setup Steps

### 1. Configure AWS Credentials

Ensure the AWS credentials are properly set up:

```bash
aws configure
```

Follow the prompts to configure the AWS Access Key ID, Secret Access Key, and default region.

### 2. Create and Activate a Virtual Environment

```bash
python -m venv venv
source venv/bin/activate # On Windows, use `venv\Scripts\activate`
```

### 3. Install Required Dependencies

```bash
cd pulumi
pip install -r requirements.txt
```

### 4. Initialize or Select a Pulumi Stack

Create a new stack:
```bash
pulumi stack init <stack-name>
```
Or select an existing stack:
```bash
pulumi stack select <stack-name>
```

### 5. Configure the Pulumi Stack

Review and update the configuration in `Pulumi.<stack-name>.yaml`. You can also set config values using the CLI:

```bash
pulumi config set aws:region us-east-1
pulumi config set instance_name dice-db
pulumi config set instance_type t2.medium
```

### 6. Preview the Infrastructure Changes

```bash
pulumi preview
```

This command shows you what changes will be made without actually applying them.

### 7. Deploy the Infrastructure

When you're ready to create the resources:

```bash
pulumi up
```

Review the proposed changes and confirm to proceed.

### 8. Access Your Resources

After deployment, Pulumi will output important information like the instance's public IP and SSH command. Make note of these for accessing the DiceDB instance; or SSH into the instance via the console.

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
