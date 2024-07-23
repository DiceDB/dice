# Configure Pulumi

### Configure a new Pulumi Stack

Review and update the configuration in `Pulumi.<stack-name>.yaml`. You can also set config values using the CLI:

```bash
pulumi config set aws-native:region us-east-1
pulumi config set aws:region us-east-1
pulumi config set instance_name dicedb
pulumi config set instance_type t2.medium
```
