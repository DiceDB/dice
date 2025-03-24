# DiceDB Contributing Guide

DiceDB welcomes your contributions! Whether you're fixing bugs, adding new features, or improving the documentation, your help is valuable.

We have multiple repositories where you can contribute. So, as per your interest, you can pick one and build a deeper understanding of the project on the go.

- [dicedb/dice](https://github.com/dicedb/dice) for core database features and engine / Stack - Go
- [dicedb/dicedb-cli](https://github.com/dicedb/dicedb-cli) command line interface for DiceDB / Stack - Go
- [dicedb/dicedb-go](https://github.com/dicedb/dicedb-go) Go Client SDK for DiceDB

## Important Resources

- [Development Setup](https://github.com/dicedb/dice/blob/master/CONTRIBUTING/development-setup.md)
- [Development Setup for Docs](https://github.com/dicedb/dice/blob/master/CONTRIBUTING/docs.md)
- [Git Best Practices](https://github.com/dicedb/dice/blob/master/CONTRIBUTING/git.md)
- [Go Best Practices](https://github.com/dicedb/dice/blob/master/CONTRIBUTING/go.md)
- [Logging Best Practices](https://github.com/dicedb/dice/blob/master/CONTRIBUTING/logging.md)

## Timeline for working on Issues

### Issues

- Do not wait for anyone to assign, you can directly pick the issue up, and submit a PR.
- Assigned issues imply intent to work on them.
- Can't work on it? Unassign yourself to allow others to contribute.
- Provide updates on long-running issues to show progress.
- Inactive issues may be unassigned after a reasonable period.

### Pull Requests (PRs)

- We appreciate timely completion of PRs.
- If a PR becomes inactive, we may close it.
- Need more time? Just let us know in the comments.

## CLA

To maintain the project's quality and consistency, please follow these guidelines:

- Keep the code consistent: Use the same coding style and conventions throughout the project. 
- Keep the git repository consistent: Follow proper git practices to avoid conflicts and ensure a clean history. 

### Contributor License Agreement (CLA)

By contributing to any [DiceDB repositories](https://github.com/dicedb), you acknowledge and agree to the terms of the [DiceDB CLA](https://gist.github.com/arpitbbhayani/3bb42965630961c2b3b02f222c3338e0).

Please follow the steps outlined in the CLA to complete the agreement. If not done beforehand, this process will be triggered automatically when you open a Pull Request.

### Code documentation

Please ensure your code is adequately documented. Some things to consider for documentation:

- Always include struct, module, and package level docs. We are looking for information about what functionality is provided, what state is maintained, whether there are concurrency/thread-safety concerns and any exceptional behavior that the class might exhibit.
- Document public methods and their parameters.
