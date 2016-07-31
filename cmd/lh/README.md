Lighthouse CLI client
=====================

A Golang CLI client for interacting with the
[Lighthouse API](http://lighthouseapp.com/api).

## Installation

``` no-highlight
go get -u github.com/nwidger/lighthouse/cmd/lh
```

## Usage

Run `lh` with no arguments for usage help:

``` no-highlight
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
project-specific.  These commands require the project ID or name to be
specified via -p, --project, the LH_PROJECT environment variable or
the config file.

On Unix systems, the default config file is $HOME/.lh.yaml.  On
Windows systems, the default config file is
%HOMEDRIVE%\%HOMEPATH%\.lh.yaml, falling back to
%USERPROFILE%\.lh.yaml if necessary.  On all systems, the default can
be overridden with --config.

Usage:
  lh [command]

Available Commands:
  create      Create Lighthouse resources
  delete      Delete Lighthouse resources
  get         Get Lighthouse resources
  list        List Lighthouse resources
  update      Update Lighthouse resources

Flags:
  -a, --account string    Lighthouse account name
      --config string     config file (default is $HOME/.lh.yaml)
      --email string      Lighthouse email (cannot be used with --token)
  -M, --monochrome        Monochrome (don't colorize JSON)
  -h, --help              help for lh
      --password string   Lighthouse password (cannot be used with --token)
  -p, --project string    Lighthouse project ID or name
  -t, --token string      Lighthouse API token

Use "lh [command] --help" for more information about a command.
```

Use `lh create` to create Lighthouse resources:

``` no-highlight
$ lh create
Create Lighthouse resources

Usage:
  lh create [command]

Available Commands:
  bin         Create a bin (requires -p)
  changeset   Create a changeset (requires -p)
  message     Create a message (requires -p)
  milestone   Create a milestone (requires -p)
  project     Create a project
  ticket      Create a ticket (requires -p)

Flags:
  -h, --help   help for create

Global Flags:
  -a, --account string    Lighthouse account name
      --config string     config file (default is $HOME/.lh.yaml)
      --email string      Lighthouse email (cannot be used with --token)
  -M, --monochrome        Monochrome (don't colorize JSON)
      --password string   Lighthouse password (cannot be used with --token)
  -p, --project string    Lighthouse project ID or name
  -t, --token string      Lighthouse API token

Use "lh create [command] --help" for more information about a command.
```

Use `lh delete` to delete Lighthouse resources:

``` no-highlight
Delete Lighthouse resources

Usage:
  lh delete [command]

Available Commands:
  bin         Delete a bin (requires -p)
  changeset   Delete a changeset (requires -p)
  message     Delete a message (requires -p)
  milestone   Delete a milestone (requires -p)
  project     Delete a project (requires -p)
  ticket      Delete a ticket (requires -p)

Flags:
  -h, --help   help for delete

Global Flags:
  -a, --account string    Lighthouse account name
      --config string     config file (default is $HOME/.lh.yaml)
      --email string      Lighthouse email (cannot be used with --token)
  -M, --monochrome        Monochrome (don't colorize JSON)
      --password string   Lighthouse password (cannot be used with --token)
  -p, --project string    Lighthouse project ID or name
  -t, --token string      Lighthouse API token

Use "lh delete [command] --help" for more information about a command.
```

Use `lh get` to retrieve a specific Lighthouse resource:

``` no-highlight
$ lh get
Get Lighthouse resources

Usage:
  lh get [command]

Available Commands:
  bin         Get a ticket bin (requires -p)
  changeset   Get a changeset (requires -p)
  message     Get a message (requires -p)
  milestone   Get a milestone (requires -p)
  plan        Get your Lighthouse plan
  profile     Get your Lighthouse profile
  project     Get your Lighthouse project
  ticket      Get a ticket (requires -p)
  token       Get information about an API token
  user        Get information about a Lighthouse user

Flags:
  -h, --help   help for get

Global Flags:
  -a, --account string    Lighthouse account name
      --config string     config file (default is $HOME/.lh.yaml)
      --email string      Lighthouse email (cannot be used with --token)
  -M, --monochrome        Monochrome (don't colorize JSON)
      --password string   Lighthouse password (cannot be used with --token)
  -p, --project string    Lighthouse project ID or name
  -t, --token string      Lighthouse API token

Use "lh get [command] --help" for more information about a command.
```

Use `lh list` to list Lighthouse resources:

``` no-highlight
$ lh list
List Lighthouse resources

Usage:
  lh list [command]

Available Commands:
  bins        List ticket bins (requires -p)
  changesets  List changesets (requires -p)
  messages    List messages (requires -p)
  milestones  List milestones (requires -p)
  projects    List projects
  tickets     List tickets (requires -p)

Flags:
  -h, --help   help for list

Global Flags:
  -a, --account string    Lighthouse account name
      --config string     config file (default is $HOME/.lh.yaml)
      --email string      Lighthouse email (cannot be used with --token)
  -M, --monochrome        Monochrome (don't colorize JSON)
      --password string   Lighthouse password (cannot be used with --token)
  -p, --project string    Lighthouse project ID or name
  -t, --token string      Lighthouse API token

Use "lh list [command] --help" for more information about a command.
```

Use `lh update` to update a specific Lighthouse resource:

``` no-highlight
$ lh update
Update Lighthouse resources

Usage:
  lh update [command]

Available Commands:
  bin         Update a bin (requires -p)
  message     Update a message (requires -p)
  milestone   Update a milestone (requires -p)
  project     Update a project
  ticket      Update a ticket (requires -p)
  tickets     Bulk update tickets (requires -p)
  user        Update information about a Lighthouse user

Flags:
  -h, --help   help for update

Global Flags:
  -a, --account string    Lighthouse account name
      --config string     config file (default is $HOME/.lh.yaml)
      --email string      Lighthouse email (cannot be used with --token)
  -M, --monochrome        Monochrome (don't colorize JSON)
      --password string   Lighthouse password (cannot be used with --token)
  -p, --project string    Lighthouse project ID or name
  -t, --token string      Lighthouse API token

Use "lh update [command] --help" for more information about a command.
```

## Config File

Modify the following example config file with your own account name,
API token and project and save it to `$HOME/.lh.yaml`.

``` yaml
account: your-account-name
token: deadbeefdeadbeefdeadbeefdeadbeefdeadbeef
project: your-project-name
```

## Output

All commands return resources as JSON.  By default, the output is
colorized using the [jsoncolor](https://github.com/nwidger/jsoncolor)
package.  This can be disabled using `-M` or `--monochrome`.  Piping
`lh`'s output to a JSON processor such as
[jq](https://stedolan.github.io/jq/) may be helpful to retrieve
specific fields.

## Examples

The following examples assume you have configured your account name,
API token and project in the config file.

Create bin `Fred's Open Tickets`:

``` no-highlight
$ lh create bin --name "Fred's Open Tickets" --query "assigned:fred state:open"
```

Delete milestone `v9`:

``` no-highlight
$ lh delete milestone v9
```

Get ticket `2428`:

``` no-highlight
$ lh get ticket 2428
```

Download attachment `bad.conf` from ticket `2428`:

``` no-highlight
$ lh get ticket 2428 --attachment bad.conf > bad.conf
```

List all tickets matching query `milestone:"XYZ v9"`

``` no-highlight
$ lh list tickets --all --query 'milestone:"XYZ v9"'
```

Update ticket `2428`:

``` no-highlight
$ lh update ticket 2428 --comment "Looks good to me" --state resolved --assigned fred
```
