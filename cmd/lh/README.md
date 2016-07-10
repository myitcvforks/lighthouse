Lighthouse CLI client
=====================

A Golang CLI client for interacting with the
[Lighthouse API](http://lighthouseapp.com/api).

## Installation

```
go get -u github.com/nwidger/lighthouse/cmd/lh
```

## Usage

Run `lh` with no arguments for usage help:

``` shell
$ lh
lh provides CLI access to the Lighthouse API http://help.lighthouseapp.com/kb/api

Please specify your Lighthouse account name via -a, --account, the
LH_ACCOUNT environment variable or the config file.  If your
Lighthouse URL is 'https://your-account-name.lighthouseapp.com' then
your account name is 'your-account-name'.

Lighthouse requires a valid API token or email/password to
authenticate API requests.  Please specify a Lighthouse API token via
-t, --token, the LH_TOKEN environment variable or the config file.  If
you'd prefer to authenticate with an email/password, please specify it
via -e, --email, the LH_EMAIL environment variable, -p, --password,
the LH_PASSWORD environment variable or the config file.  If the
specified password has the form '@FILE', the password is instead read
from FILE.

Many subcommands work on resources that are Lighthouse
project-specific.  These commands require the project ID to be
specified via -p, --project, the LH_PROJECT environment variable or
the config file.

The default config file is $HOME/.lh.yaml but can be overridden with
--config.

Usage:
  lh [command]

Available Commands:
  get         Get Lighthouse resources by ID
  list        List Lighthouse resources

Flags:
  -a, --account string    Lighthouse account name
      --config string     config file (default is $HOME/.lh.yaml)
      --email string      Lighthouse email (cannot be used with --token)
      --password string   Lighthouse password (cannot be used with --token)
  -p, --project int       Lighthouse project ID
  -t, --token string      Lighthouse API token

Use "lh [command] --help" for more information about a command.
```

Use `lh get` to retrieve a specific Lighthouse resource (usually by
ID):

``` shell
$ lh get -h
Get Lighthouse resources by ID

Usage:
  lh get [command]

Available Commands:
  bin         Get a ticket bin (requires -p)
  changeset   Get a changeset (requires -p)
  message     Get a message (requires -p)
  milestone   Get a milestone (requires -p)
  plan        Get your Lighthouse plan
  profile     Get your Lighthouse profile
  ticket      Get a ticket (requires -p)
  token       Get information about an API token
  user        Get information about a Lighthouse user

Global Flags:
  -a, --account string    Lighthouse account name
      --config string     config file (default is $HOME/.lh.yaml)
      --email string      Lighthouse email (cannot be used with --token)
      --password string   Lighthouse password (cannot be used with --token)
  -p, --project int       Lighthouse project ID
  -t, --token string      Lighthouse API token

Use "lh get [command] --help" for more information about a command.
```

Use `lh list` to list Lighthouse resources:

``` shell
List Lighthouse resources

Usage:
  lh list [command]

Available Commands:
  bins        List ticket bins (requires -p)
  changesets  List changesets (requires -p)
  memberships List a project's memberships
  messages    List messages (requires -p)
  milestones  List milestones (requires -p)
  projects    List projects
  tickets     List tickets (requires -p)

Global Flags:
  -a, --account string    Lighthouse account name
      --config string     config file (default is $HOME/.lh.yaml)
      --email string      Lighthouse email (cannot be used with --token)
      --password string   Lighthouse password (cannot be used with --token)
  -p, --project int       Lighthouse project ID
  -t, --token string      Lighthouse API token

Use "lh list [command] --help" for more information about a command.
```

## Config File

Modify the following example config file with your own account name,
API token and project ID and save it to `$HOME/.lh.yaml`.

``` yaml
account: your-account-name
token: deadbeefdeadbeefdeadbeefdeadbeefdeadbeef
project: 1234
```

## Output

All commands return resources as JSON.  Piping `lh`'s output to a JSON
processor such as [jq](https://stedolan.github.io/jq/) may be helpful
to pretty-print the output or retrieve specific fields.

## Examples

The following examples assume you have configured your account name,
API token and project ID in the config file.

Get ticket `2428`:

``` shell
$ lh get ticket 2428
```

List all tickets matching query `milestone:"XYZ v9"`

``` shell
$ lh list tickets --all --query 'milestone:"XYZ v9"`
```

List account memberships for user `999999`:

``` shell
$ lh get user 999999 --memberships
```
