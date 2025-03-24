# Git Best Practices

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

Make the necessary changes. If the changes you plan to make are too big, make sure you break them down into smaller tasks.

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

- Once you receive comments on GitHub on your changes, be sure to respond to them on GitHub and address the concerns. If any discussions happen offline for the changes in question, make sure to capture the outcome of the discussion, so others can follow along as well.

  It is possible that while your change is being reviewed, other changes were made to the master branch. Be sure to pull rebase your change on the new changes thus:

```text
# commit your changes
$ git add <updated files>
$ git commit -m "Meaningful message for the update"
# pull new changes
$ git checkout master
$ git fetch upstream
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
