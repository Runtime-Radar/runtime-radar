# Contributing to Runtime-Radar

Thank you for your interest in contributing to Runtime-Radar! We welcome contributions from the community and appreciate your help in making this project better.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Contributing code](#contributing-code)
- [Contributing without code](#contributing-without-code)
- [Development Setup](#development-setup)
- [Testing](#testing)
- [Code Style and linting](#code-style-and-linting)
- [API](#api)
- [Submitting Changes](#submitting-changes)
- [Creating Issues](#creating-issues)
- [Reporting Issues](#reporting-issues)

## Code of Conduct

By participating in this project, you agree to maintain a respectful and inclusive environment for all contributors.

## Contributing Code

1. Fork the repository on GitHub
2. Clone your fork locally
3. Set up your development environment (see [Development Setup](#development-setup))
4. Create a new branch for your changes
5. Make your changes
6. Push to your fork
7. Submit a pull request

## Contributing Without Code

You can also contribute by:
- Reporting bugs
- Suggesting new features
- Improving documentation
- Participating in discussions

## Development Setup

### Prerequisites

For local development, you will likely want to install following tools:

- A `Go` toolchain with the version [specified](https://github.com/Runtime-Radar/runtime-radar/blob/main/go.mod) in the main go.mod;
- [Task](https://taskfile.dev);
- A running `Docker` service;
- `tinygo` with with the version [specified](https://github.com/Runtime-Radar/runtime-radar/blob/main/event-processor/Taskfile.yml#L98) in the `event-processor`'s `Taskfile.yaml`;
- `NodeJS` > 20 and `Saas` > 3 for building UI (also can be built via Docker);
- `yarn` for UI's dependency management;

Also, if you want to install Runtime Radar in cluster, you additionally need to have:
- Kubernetes cluster (for example, [kind](https://kind.sigs.k8s.io))
- [Helm](https://helm.sh/docs/helm/helm_package/)

### Building and running
Build images for all backend components:
```bash
task docker-build
```

Build specific backend component's image:
```bash
task ${COMPONENT_NAME}:docker-build
```
Or:
```bash
cd ${COMPONENT_NAME} && task docker-build
```

Build specific backend component's binary:
```bash
task build
```

Run specific backend component with its dependencies via `docker compose`:
```bash
cd ${COMPONENT_NAME} && docker compose up -d
```

You can also run the whole application via `docker compose`:
```bash
docker compose up -d
```
Note that application's behavior will differ from running it it Kubernetes environment. For example, events will not contain Kubernetes' context (info about pods, nodes, namespaces, etc) which also adds limitations on setting response rules.

To run Runtime Radar in Kubernetes cluster, you need to prepare a cluster (for example, [kind](https://kind.sigs.k8s.io/docs/user/quick-start/#creating-a-cluster)) and follow the steps described in [Quickstart](https://github.com/Runtime-Radar/runtime-radar/blob/main/docs/quickstart/quickstart.md).

When working on specific backend's component, it's possible to deploy it to Kubernetes cluster without re-installing the whole application. The only requirement is running registry accessible by Kubernetes cluster:
```bash
cd ${COMPONENT_NAME} && IMAGE_REGISTRY=my.local.registry task deploy
```

Build UI using Docker:
```bash
cd radar-ui && task build
```

If you wish to run the UI locally, you can connect it to the deployed backend. To do so, rename the file `radar-ui/proxy.conf.json.sample` to `radar-ui/proxy.conf.json` and set the backend URL as the target value. The correct URL format is: `https://url.runtime-radar.com:10000/`.

To build detectors see our [guide](https://github.com/Runtime-Radar/runtime-radar/blob/main/docs/guides/detectors/guide.md) on them. 

## Testing 
We do not aim for 100% unit test coverage as a primary goal. Our strategy is to write unit tests primarily for algorithmically complex or critical functions and methods, where they provide the most value in ensuring correctness and preventing regressions.

For broader integration coverage, we focus on end-to-end (e2e) tests for API endpoints. Currently, only [policy-enforcer](https://github.com/Runtime-Radar/runtime-radar/tree/main/policy-enforcer) has all its endpoints fully covered by e2e tests. We recommend reading its [e2e-tests](https://github.com/Runtime-Radar/runtime-radar/blob/main/policy-enforcer/cmd/policy-enforcer/main_test.go) code to get familiar with our approach to writing tests. It is worth noting that a foundation for e2e testing has been laid in some other components, but only a small number of endpoints are currently covered there. We would greatly appreciate any help in extending this e2e test coverage to these and other components.

Regarding the UI layer, we mainly rely on unit tests to verify the functionality of individual components.

Here are some useful commands that will help running tests:

Test all backend components:
```bash
task test-backend
```

Test specific backend component:
```bash
cd ${COMPONENT_NAME} && task test-docker
```

Test UI:
```bash
cd radar-ui && npx nx test ${LIB_NAME}
```
where `${LIB_NAME}` refers to specific module.

## Code Style and linting

We currently don't have a specific style guide on Go code, but we are actively working on establishing one. In the meantime, we encourage contributors to follow popular Go style guides such as:

- [Google's Go Style Guide](https://google.github.io/styleguide/go/)
- [Uber's Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md)

Our main principles are:
- **Simplicity**: Avoid overcomplicated solutions
- **Readability**: Write code that is easy to understand 
- **Clarity**: Prefer explicit and straightforward implementations over clever tricks, unless the specific task requires the opposite.

When in doubt, favor code that other developers can easily read and understand.

Our current [golangci-lint configuration](https://github.com/Runtime-Radar/runtime-radar/blob/main/.golangci.yml) intentionally includes a limited set of linters enabled. This approach is designed to avoid over-constraining developers and to maintain flexibility in how code is written. However, it possibly will be changed in the nearest future. 

Lint backend components:
```bash
task lint
```

Lint specific backend component:
```bash
task ${COMPONENT_NAME}:lint
```
or
```bash
cd ${COMPONENT_NAME} && task lint
```

Regarding UI style guide we take a look at the following guides:
- [Google HTML/CSS guide](https://google.github.io/styleguide/htmlcssguide.html)

Besides, we require a specific convention for class names and folders: `[lib_name][parent_lib_name][folder_name][type]` where `[lib_name]` is a module's name, `[parent_lib_name]` is a domain or feature postfix (if it exists), `[type]` is a file's type (components/services/pipes/etc). There are the same rules in scope of folder path.

Lint UI:
```bash
cd radar-ui && npx nx lint ${LIB_NAME} --fix
```

## API

Our project uses gRPC for inter-component communication. Please note that any changes to the API contracts require the regeneration of client and server stubs. After regeneration, it is necessary to update the dependencies in all components that import clients of the modified API. To generate stubs from `.proto`-files run:
```bash
task ${COMPONENT_NAME}:proto
```
or
```bash
cd ${COMPONENT_NAME} && task proto
```

To re-generate stubs for all components run:
```bash
task proto
```
in the project's root.

## Submitting Changes

### Commit messages

Currently, our team doesn't enforce a strict, mandatory convention for commit messages. However, for those looking for guidance on how to structure clear and informative commits, we recommend the [Conventional Commits specification](https://www.conventionalcommits.org/en/v1.0.0/) as a great source of inspiration.

Following its general principles is optional but encouraged, as it helps create a more readable and maintainable project history.

### Pull Request Checklist

- [ ] Tests pass locally
- [ ] Code follows the project's style guidelines
- [ ] Linter passes without errors
- [ ] Documentation has been updated (if applicable)
- [ ] Commit messages follow the convention
- [ ] Branch is up to date with the `main` branch

### Review Process

Currently, for a Pull Request to be merged, it requires an approval from at least one of the project's maintainers who is also a member of our organization.

## Creating Issues

Before creating a new issue, please:

1. **Search existing issues** to avoid duplicates
2. **Check the documentation** to ensure it's not a usage question
3. **Verify you're using the latest version** of the project

### How to Create an Issue

1. Go to the GitHub Issues page
2. Click "New Issue"
3. Choose the appropriate issue template (if available)
4. Fill in all required information
5. Add relevant labels (if you have permission)
6. Submit the issue

### Issue Guidelines

- **Use a clear and descriptive title**
- **Provide detailed information** following the template
- **Stay on topic** - one issue per problem or feature
- **Be respectful** and follow the Code of Conduct
- **Respond to questions** from maintainers in a timely manner

## Reporting Issues

### Security Issues

Please do not report security vulnerabilities in public GitHub issues. Instead, email us at runtimeradar@outlook.com — see our [Security Policy](https://github.com/Runtime-Radar/runtime-radar/blob/main/SECURITY.md) for details.

### Bug Reports

When reporting bugs, please include:

- A clear and descriptive title
- Steps to reproduce the issue
- Expected behavior
- Actual behavior
- Environment details (OS, version, etc.)
- Screenshots (if applicable)

### Feature Requests

When suggesting features, please include:

- A clear and descriptive title
- Detailed description of the proposed feature
- Use cases and benefits
- Any potential drawbacks or considerations

To simplify the process of creating issue, we've added templates for reporting bugs and suggesting features.

### Your First Code Contribution

Not sure where to begin? You can start by looking at issues labeled `good first issue` or `help wanted`.

*   [Good First Issues](https://github.com/runtime-radar/runtime-radar/labels/good%20first%20issue)
*   [Help Wanted Issues](https://github.com/runtime-radar/runtime-radar/labels/help%20wanted)

## Questions?

- [Telegram](https://t.me/runtimeradar)
- [Slack](https://runtimeradar.slack.com)

## License

By contributing to Runtime-Radar, you agree that your contributions will be licensed under the project's license.

---

Thank you for contributing to Runtime Radar! 🎉
