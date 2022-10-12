# Less Advanced Security

GitHub sells a product, [GitHub Advanced Security](https://docs.github.com/en/get-started/learning-about-github/about-github-advanced-security) which bundles CodeQL (a static analysis tool similar to [semgrep](https://semgrep.dev)), secret scanning (similar to [trufflehog](https://github.com/trufflesecurity/trufflehog)), an improved UI on dependency spec files, and [security overview](https://docs.github.com/en/code-security/security-overview/about-the-security-overview) (a way to upload static analysis findings such that they produce PR annotations, and a dashboard to view them).

Less Advanced Security is... less advanced. It enables you to bring-your-own PR annotations to any tool which outputs results in the [sarif format](https://github.com/microsoft/sarif-tutorials).

GitHub Advanced Security charges a per-active-commiter seat license of ~$600/yr. Less Advanced Security is free (bring your own compute, config, and maintenance).
