# Less Advanced Security

GitHub sells a product, [GitHub Advanced Security](https://docs.github.com/en/get-started/learning-about-github/about-github-advanced-security) which bundles CodeQL (a static analysis tool similar to [semgrep](https://semgrep.dev)), secret scanning (similar to [trufflehog](https://github.com/trufflesecurity/trufflehog)), an improved UI on dependency spec files, and [security overview](https://docs.github.com/en/code-security/security-overview/about-the-security-overview) (a way to upload static analysis findings such that they produce PR annotations, and a dashboard to view them).

Less Advanced Security is... less advanced. It enables you to bring-your-own PR annotations to any tool which outputs results in the [sarif format](https://github.com/microsoft/sarif-tutorials).

GitHub Advanced Security charges a per-active-commiter seat license of ~$600/yr. Less Advanced Security is free (bring your own compute, config, and maintenance).

## Setup

### Environment

```sh
brew install go@1.19
```

### GitHub Application

1. [Create a GitHub App](https://docs.github.com/en/developers/apps/building-github-apps/creating-a-github-app) for your user or org.
    * Set the following permissions:
        * `Repository permissions > Checks > Access: Read and write`
        * `Repository permissions > Pull requests > Access: Read and write`
    * Note the `App Id` for later.
    * Enable webhook events on installation, but submit a nonsense URL (like `https://github.com/<your_user>/dev/null`) for the destination.
1. [Install the GitHub App](https://docs.github.com/en/developers/apps/managing-github-apps/installing-github-apps), granting it access to the relevant repos.
    * Note the `Install Id` for later. To do this, return to your app configuration, look at `Advanced` settings, find the failed webhook delivery, and look at the payload.
1. [Generate a private key](https://docs.github.com/en/developers/apps/building-github-apps/authenticating-with-github-apps) and save it locally.
    * Note the path to the key for later.

## Usage

Run your sarif-producing scan, writing the sarif file to disk.

Then run
```sh
go run main.go -app_id=<app_id> -install_id=<install_id> -key_path=<path_to_key> -sha=<sha_of_target_commit> -repo=<repo_owner>/<repo_name> -pr=<pr_number> -sarif_path=<path_to_sarif_file>
```

For example:

```sh
go run main.go -app_id=12345 -install_id=87654321 -key_path=tmp/application_private_key.pem -sha=ee5dabb638b6b874c42bc3c915cf94d4b6b346b6 -repo=eliblock/less-advanced-security -pr=57 -sarif_path=/tmp/scan-results/sarif.json
```
