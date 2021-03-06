Git to Lighthouse integration
=============================

This is an example Go program which can be used a Git post-receive
hook to create a new Lighthouse changeset for each commit received to
a Git repository associated with a Lighthouse project.

## Installation

``` no-highlight
go get -u github.com/nwidger/lighthouse/cmd/gittolh
```

## Usage

The program should be run as a standard post-receive hook with
arguments `<oldrev> <newrev> <refname>` passed to the program either
through command-line arguments or via stdin.

The program expects the Git config entry `lighthouse.account` to
contain your Lighthouse account name and the Git config entry
`lighthouse.project` to contain the Lighthouse project ID.

For example, if the URL to your Lighthouse project is
`http://example.lighthouseapp.com/projects/1234`, then your Lighthouse
account is `example` and your project ID is `1234`.  The following
commands should be run from the root of the Git repository:

``` no-highlight
git config lighthouse.account example
git config lighthouse.project 1234
```

In addition, the program expects a Git config entry
`lighthouse.keys.<name>` to exist for each committer on the project
where `<name>` if the name-part of their committer email.  The value
of each Git config entry should be the commit author's associated
Lighthouse API key which will be used to create a new changeset via
the Lighthouse API.  For example, if there are three committers with
emails `alice@example.com`, `bob@example2.com` and
`susan@example3.com`, the following commands should be run from the
root of the Git repository:

``` no-highlight
git config lighthouse.keys.alice 0000000000000000000000000000000000000000
git config lighthouse.keys.bob   0000000000000000000000000000000000000000
git config lighthouse.keys.susan 0000000000000000000000000000000000000000
```

The program also optionally supports appending a footer to each
Lighthouse changeset with the `lighthouse.footer` Git config entry.
If `lighthouse.footer` contains `%s`, this will be substituted with
the revision of the changeset.  For example, if
`https://git.example.com?project=example.git` is the base URL to your
Git repository in a gitweb instance, then `lighthouse.footer` could be
configured as follows to append a Markdown link to the commit in
gitweb:

``` no-highlight
git config lighthouse.footer "[gitweb](https://git.example.com/?project=example.git?a=commit;h=%s)"
```

Any errors encountered during execution are appended to the file
`/tmp/git-hooks.log`.  This file is expected to be writeable by all
users who might be running the post-receive hook.  If the file does
not exist, it is created with `777` permissions.  You may need to
modify your umask before running this program to ensure that the file
is indeed created with write permissions for all appropriate users.
