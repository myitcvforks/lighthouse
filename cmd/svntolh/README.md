SVN to Lighthouse integration
=============================

This is an example Go program which can be used an SVN commit hook to
create a new Lighthouse changeset each time a commit is made to an SVN
repository associated with a Lighthouse project.

## Installation

``` no-highlight
go get -u github.com/nwidger/lighthouse/cmd/svntolh
```

## Usage

The program should be run as `svntolh <repo-path> <revision>` where
`repo-path` is the path to the root of the SVN repository and
`revision` is the SVN revision number to create a changeset from.

The program expects two files `.lhproj` and `.lhkeys` to exist at the
root of the SVN repository.  The file `.lhproj` must contain the
Lighthouse project URL to create the changeset in on a single line:

``` no-highlight
http://<account>.lighthouseapp.com/projects/<project-id>
```

For example, if your Lighthouse account is `example` and your project
ID is `1234`, the file `.lhproj` should contain:

``` no-highlight
http://example.lighthouseapp.com/projects/1234
```

The file `.lhkeys` must contain a mapping between SVN commit authors
and their respective Lighthouse API key:

``` no-highlight
alice 0000000000000000000000000000000000000000
bob   0000000000000000000000000000000000000000
susan 0000000000000000000000000000000000000000
```

The commit author's associated Lighthouse API key will be used to
create the new changeset via the Lighthouse API.

Any errors encountered during execution are appended to the file
`/tmp/svn-hooks.log`.
