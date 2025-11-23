# Contributing to Push From K8s Back to Docker Registry

First off, thank you for considering contributing! 🎉

Every contribution helps, whether it's:
- Reporting a bug
- Discussing the current state of the code
- Submitting a fix
- Proposing new features
- Becoming a maintainer

## We Develop with GitHub

We use GitHub to host code, track issues and feature requests, as well as accept pull requests.

## We Use [GitHub Flow](https://guides.github.com/introduction/flow/index.html)

Pull requests are the best way to propose changes to the codebase:

1. Fork the repo and create your branch from `main`
2. If you've added code that should be tested, add tests
3. If you've changed APIs, update the documentation
4. Ensure the test suite passes
5. Make sure your code lints
6. Issue that pull request!

## Any Contributions You Make Will Be Under the MIT License

In short, when you submit code changes, your submissions are understood to be under the same [MIT License](LICENSE) that covers the project.

## Report Bugs Using GitHub's [Issue Tracker](https://github.com/tazhate/push-from-k8s-back-to-docker-registry/issues)

We use GitHub issues to track public bugs. Report a bug by [opening a new issue](https://github.com/tazhate/push-from-k8s-back-to-docker-registry/issues/new).

### Write Bug Reports with Detail, Background, and Sample Code

**Great Bug Reports** tend to have:

- A quick summary and/or background
- Steps to reproduce
  - Be specific!
  - Give sample code if you can
- What you expected would happen
- What actually happens
- Notes (possibly including why you think this might be happening, or stuff you tried that didn't work)

Example:

```markdown
## Bug: Images from Docker Hub fail to sync

### Environment
- Kubernetes: v1.28.0 (k3s)
- Tool version: v2.3.0
- Registry: Harbor 2.10

### Steps to Reproduce
1. Deploy the helm chart with `registry.url=harbor.example.com`
2. Create a deployment using `nginx:latest` from Docker Hub
3. Wait for sync cycle

### Expected Behavior
Image should be copied to Harbor

### Actual Behavior
Error in logs:
```
failed to copy image: 401 Unauthorized
```

### Additional Context
- Harbor has anonymous pulls disabled
- Same error with redis:alpine
```

## Development Process

### Setting Up Development Environment

```bash
# Clone your fork
git clone https://github.com/YOUR_USERNAME/push-from-k8s-back-to-docker-registry.git
cd push-from-k8s-back-to-docker-registry

# Install dependencies
go mod download

# Run tests
go test ./...

# Build
go build -o bin/syncer ./cmd/syncer
```

### Running Locally

```bash
# Set environment variables
export TARGET_REGISTRY_URL=your-registry.com
export TARGET_REGISTRY_USERNAME=admin
export TARGET_REGISTRY_PASSWORD=password
export NAMESPACES=default
export SYNC_PERIOD=1m
export LOG_LEVEL=debug

# Run
./bin/syncer
```

### Code Style

We use standard Go formatting:

```bash
# Format code
go fmt ./...

# Run linters
golangci-lint run
```

### Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific test
go test -v -run TestSyncOnce ./internal/syncer
```

### Building Docker Image

```bash
docker build -f Dockerfile.new -t your-registry/image-sync:dev .
```

## Pull Request Process

1. **Update Documentation**: Update the README.md with details of changes if needed
2. **Update Chart**: If you change configs, update `chart/values.yaml` and templates
3. **Add Tests**: New features should include tests
4. **Follow Commit Message Convention**:
   ```
   type(scope): subject

   body (optional)

   footer (optional)
   ```

   Types:
   - `feat`: New feature
   - `fix`: Bug fix
   - `docs`: Documentation changes
   - `style`: Code style changes (formatting, etc.)
   - `refactor`: Code refactoring
   - `test`: Adding or updating tests
   - `chore`: Maintenance tasks

   Examples:
   ```
   feat(registry): add support for multiple target registries
   fix(k8s): handle pods with no images gracefully
   docs(readme): update installation instructions
   ```

5. **Update CHANGELOG**: Add your changes to `CHANGELOG.md` under `[Unreleased]`

6. **Request Review**: Tag maintainers or wait for automatic assignment

## Code Review Process

The core team looks at Pull Requests on a regular basis. After feedback has been given, we expect responses within two weeks. After two weeks, we may close the PR if it isn't showing any activity.

## Community

- **GitHub Discussions**: For questions and ideas
- **GitHub Issues**: For bugs and feature requests
- **Twitter**: [@tazhate](https://twitter.com/tazhate)

## Recognition

Contributors will be added to:
- README.md acknowledgments section
- Release notes for their contributions
- GitHub's contributor graph (automatic)

## First Time Contributors

Looking for an easy first issue? Check out issues labeled [`good first issue`](https://github.com/tazhate/push-from-k8s-back-to-docker-registry/labels/good%20first%20issue).

## Questions?

Feel free to open an issue with the `question` label or start a discussion!

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

---

Thank you for contributing! 🙏
