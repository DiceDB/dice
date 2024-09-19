# Contribution Guidelines

Before you begin to contribute, make sure you have reviewed [Dev Environment Setup](https://github.com/dicedb/dice/blob/master/README.md) sections and that you have created your own fork of the source code.

## Create a design document

If your change is relatively minor, you can skip this step. If you are adding new major feature, we suggest that you add a design document and solicit comments from the community before submitting any code.

## Create an issue for the change

Create an issue [here](https://github.com/dicedb/dice/issues) for the change you would like to make. Provide information on why the change is needed and how you plan to address it. Use the conversations on the issue as a way to validate assumptions and the right way to proceed.

If you have a design document, please refer to the design documents in your Issue. You may even want to create multiple issues depending on the extent of your change.

Once you are clear about what you want to do, proceed with the next steps listed below.

## Create a branch for your change

```text
$ cd dice
#
# ensure you are starting from the latest code base
# the following steps, ensure your fork's (origin's) master is up-to-date
#
$ git fetch upstream
$ git checkout master
$ git merge upstream/master
# create a branch for your issue
$ git checkout -b <your issue branch>
```

Make the necessary changes. If the changes you plan to make are too big, make sure you break it down into smaller tasks.

## Making the changes

Follow the best-practices when you are making changes.

### Code documentation

Please ensure your code is adequately documented. Some things to consider for documentation:

- Always include struct, module, and package level docs. We are looking for information about what functionality is provided, what state is maintained, whether there are concurrency/thread-safety concerns and any exceptional behavior that the class might exhibit.
- Document public methods and their parameters.

### Logging

- Ensure there is adequate logging for positive paths as well as exceptional paths. As a corollary to this, ensure logs are not noisy.
- Do not use fmt.Println to log messages. Use the `logrus` logger.
- Use logging levels correctly: set level to `debug` for verbose logs useful for only for debugging.

### Code Formatting

To ensure code quality and consistency in our Go project, we use [golangci-lint](https://golangci-lint.run/), a popular linter aggregator for Go. This tool combines multiple linters to catch common issues and enforce best practices in Go code.

To use golangci-lint, you need to have it installed. You can install it using the following methods:

- Via Homebrew (macOS):

```sh
brew install golangci/tap/golangci-lint
```

- Via Go:

```sh
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

- Via Binary:

  You can download the latest binary release from the golangci-lint [releases page](https://github.com/golangci/golangci-lint/releases)

We already have preconfigured [.golangci.yml](https://github.com/DiceDB/dice/blob/master/.golangci.yaml) file in the repository. Please follow the rules defined in the file.

To run golangci-lint and check your code for issues, use the following command:

```sh
golangci-lint run
```

### Pre Commit Hook

Pre-commit Hook for golangci-lint ensures that linting checks are performed automatically before changes are committed.

To use pre-commit, you need to have it installed on your system. You can install it using following methods:

- Via Pip:

```sh
pip install pre-commit
```

- Via Homebrew (macOS):

```sh
brew install pre-commit
```

- Via apt-get (Ubuntu/Debian):

```sh
sudo apt-get install pre-commit
```

Verify the installation:

```sh
$ pre-commit --version
pre-commit 3.8.0
```

Install the pre commit hook:

```sh
$ pre-commit install
pre-commit installed at .git/hooks/pre-commit
```

With the pre-commit hook configured, `golangci-lint` will automatically run every time you execute `git commit`. This ensures that your code adheres to the linting rules specified in the [.golangci.yml](https://github.com/DiceDB/dice/blob/master/.golangci.yaml) file before any changes are committed.

### Backward and Forward compatibility changes

Make sure you consider both backward and forward compatibility issues while making your changes.

- For backward compatibility, consider cases where one component is using the new version and another is still on the old version. Will it break?
- For forward compatibility, consider rollback cases.

## Creating a Pull Request (PR)

- **Verify code-style**
- **Push changes and create a PR for review**

  Commit your changes with a meaningful commit message.

```text
$ git add <files required for the change>
$ git commit -m "Meaningful oneliner for the change"
$ git push origin <your issue branch>

After this, create a PullRequest in `github <https://github.com/dicedb/docs/pulls>`_. Make sure you have linked the relevant Issue in the description with "Closes #number" or "Fixes #number".
```

- Once you receive comments on github on your changes, be sure to respond to them on github and address the concerns. If any discussions happen offline for the changes in question, make sure to capture the outcome of the discussion, so others can follow along as well.

  It is possible that while your change is being reviewed, other changes were made to the master branch. Be sure to pull rebase your change on the new changes thus:

```text
# commit your changes
$ git add <updated files>
$ git commit -m "Meaningful message for the udpate"
# pull new changes
$ git checkout master
$ git merge upstream/master
$ git checkout <your issue branch>
$ git rebase master

At this time, if rebase flags any conflicts, resolve the conflicts and follow the instructions provided by the rebase command.

Run additional tests/validations for the new changes and update the PR by pushing your changes:
```

```text
git push origin <your issue branch>
```

- When you have addressed all comments and have an approved PR, one of the committers can merge your PR.
- After your change is merged, check to see if any documentation needs to be updated. If so, create a PR for documentation.

## Update Documentation

Usually for new features, functionalities, API changes, documentation update is required to keep users up to date and keep track of our development.

## Timeline for working on Issues

### Issues

- Assigned issues imply intent to work on them.
- Can't work on it? Unassign yourself to allow others to contribute.
- Provide updates on long-running issues to show progress.
- Inactive issues may be unassigned after a reasonable period.

### Pull Requests (PRs)

- We appreciate timely completion of PRs.
- If a PR becomes inactive, we may close it.
- Need more time? Just let us know in the comments.
