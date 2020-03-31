# Spotctl

A unified command-line interface to manage your [Spot](https://spot.io/) resources.

## Contents

- [Installation](#installation)
- [Getting Started](#getting-started)
- [Documentation](#documentation)
- [Examples](#examples)
- [Frequently Asked Questions](#frequently-asked-questions)
- [Getting Help](#getting-help)
- [Community](#community)
- [Contributing](#contributing)
- [License](#license)

## Installation

For macOS users, the easiest way to install `spotctl` is to use [Homebrew](https://brew.sh/):

```
$ brew install spotinst/tap/spotctl
```

Otherwise, please download the latest release from the [Releases](https://github.com/spotinst/spotctl/releases/) page.

## Getting Started

Before using `spotctl`, you need to configure your Spot credentials. You can do this in several ways:

- Environment variables
- Credentials file

The quickest way to get started is to run the `spotctl configure` command:

```
$ spotctl configure
```

[![asciicast](https://asciinema.org/a/266181.png)](https://asciinema.org/a/266181)

To use environment variables, do the following:

```
$ export SPOTINST_TOKEN=<spotinst_token>
$ export SPOTINST_ACCOUNT=<spotinst_account>
```

To use the credentials file, run the `spotctl configure` command or manually create an INI formatted file like this:

```ini
[default]
token   = <spotinst_token>
account = <spotinst_account>
```

and place it in:

- Unix/Linux/macOS: `~/.spotinst/credentials`
- Windows: `%UserProfile%\.spotinst/credentials`

If you wish to place the credentials file in a different location than the one specified above, you need to tell `spotctl` where to find it. Do this by setting the following environment variable:

```
$ export SPOTINST_CREDENTIALS_FILE=/path/to/credentials_file
```

## Documentation

If you're new to Spot and want to get started, please checkout our [Getting Started](https://help.spot.io/getting-started-with-spotinst/) guide, available on the [Spot Documentation](https://help.spot.io/) website.

## Examples

Create a new quickstart Kubernetes cluster on AWS using kops with Ocean node instance groups:

```
$ spotctl ocean quickstart cluster kubernetes aws
```

[![asciicast](https://asciinema.org/a/264624.png)](https://asciinema.org/a/264624)

## Frequently Asked Questions

- **How do I set up credentials for `spotctl`**<br/>
  See [Getting Started](#getting-started/).

## Getting Help

We use GitHub issues for tracking bugs and feature requests. Please use these community resources for getting help:

- Ask a question on [Stack Overflow](https://stackoverflow.com/) and tag it with [spotctl](https://stackoverflow.com/questions/tagged/spotctl/).
- Join our Spot community on [Slack](http://slack.spot.io/).
- Open an [issue](https://github.com/spotinst/spotctl/issues/new/choose/).

## Community

- [Slack](http://slack.spot.io/)
- [Twitter](https://twitter.com/spot_hq/)

## Contributing

Please see the [contribution guidelines](.github/CONTRIBUTING.md).

## License

Code is licensed under the [Apache License 2.0](LICENSE). See [NOTICE.md](NOTICE.md) for complete details, including software and third-party licenses and permissions.
