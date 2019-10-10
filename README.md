# Spotinst CLI

A unified command-line interface to manage your [Spotinst](https://www.spotinst.com/) resources.

* [Installation](https://github.com/spotinst/spotinst-cli#installation)
* [Getting Started](https://github.com/spotinst/spotinst-cli#getting-started)
* [Documentation](https://github.com/spotinst/spotinst-cli#documentation)
* [Examples](https://github.com/spotinst/spotinst-cli#examples)
* [Frequently Asked Questions](https://github.com/spotinst/spotinst-cli#frequently-asked-questions)
* [Getting Help](https://github.com/spotinst/spotinst-cli#getting-help)
* [Community](https://github.com/spotinst/spotinst-cli#community)
* [License](https://github.com/spotinst/spotinst-cli#license)

## Installation

For macOS users, the easiest way to install `spotinst-cli` is to use [Homebrew](https://brew.sh/):

```
$ brew install spotinst/tap/spotinst-cli
```

Otherwise, please download the latest release from the [Releases](https://github.com/spotinst/spotinst-cli/releases/) page.

## Getting Started

Before using `spotinst-cli`, you need to configure your Spotinst credentials. You can do this in several ways:

* Environment variables
* Credentials file

The quickest way to get started is to run the `spotinst configure` command:

```
$ spotinst configure
```
    
[![asciicast](https://asciinema.org/a/266181.png)](https://asciinema.org/a/266181)
    
To use environment variables, do the following:

```
$ export SPOTINST_TOKEN=<spotinst_token>
$ export SPOTINST_ACCOUNT=<spotinst_account>
```

To use the credentials file, run the `spotinst configure` command or manually create an INI formatted file like this:

```ini
[default]
token   = <spotinst_token>
account = <spotinst_account>
```

and place it in:
* Unix/Linux/macOS: `~/.spotinst/credentials`
* Windows: `%UserProfile%\.spotinst/credentials` 

If you wish to place the credentials file in a different location than the one specified above, you need to tell `spotinst-cli` where to find it.  Do this by setting the following environment variable:

```
$ export SPOTINST_CREDENTIALS_FILE=/path/to/credentials_file
```

## Documentation

If you're new to Spotinst and want to get started, please checkout our [Getting Started](https://api.spotinst.com/getting-started-with-spotinst/) guide, available on the [Spotinst Documentation](https://api.spotinst.com/) website.

## Examples

Create a new quickstart Kubernetes cluster on AWS using kops with Ocean node instance groups:

```
$ spotinst ocean quickstart cluster kubernetes aws
```
    
[![asciicast](https://asciinema.org/a/264624.png)](https://asciinema.org/a/264624)

## Frequently Asked Questions

* **How do I set up credentials for the Spotinst CLI?**<br/>
See [Getting Started](https://github.com/spotinst/spotinst-cli#getting-started/).
 
## Getting Help

We use GitHub issues for tracking bugs and feature requests. Please use these community resources for getting help:

* Ask a question on [Stack Overflow](https://stackoverflow.com/) and tag it with [spotinst-cli](https://stackoverflow.com/questions/tagged/spotinst-cli/).
* Join our Spotinst community on [Slack](http://slack.spotinst.com/).
* Open an [issue](https://github.com/spotinst/spotinst-cli/issues/new/choose/).

## Community

* [Slack](http://slack.spotinst.com/)
* [Twitter](https://twitter.com/spotinst/)

## License
Code is licensed under the [Apache License 2.0](https://github.com/spotinst/spotinst-cli/blob/master/LICENSE/).