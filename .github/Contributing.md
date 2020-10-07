## <a name="commit"></a> Commit Message Guidelines

We have very precise rules over how our git commit messages can be formatted.  This leads to **more readable messages** that are easy to follow when looking through the **project history**. For full contribution guidelines visit
the [Contributors Guide](https://wiki.edgexfoundry.org/display/FA/Committing+Code+Guidelines#CommittingCodeGuidelines-Commits) on the EdgeX Wiki

### Commit Message Format
Each commit message consists of a **header**, a **body** and a **footer**.  The header has a special format that includes a **type**, a **scope** and a **subject**:

```
<type>(<scope>): <subject>
<BLANK LINE>
<body>
<BLANK LINE>
<footer>
```

The **header** with **type** is mandatory.  The **scope** of the header is optional as far as the automated PR checks are concerned, but be advised that PR reviewers **may request** you provide an applicable scope.

Any line of the commit message cannot be longer 100 characters! This allows the message to be easier to read on GitHub as well as in various git tools.

The footer should contain a [closing reference to an issue](https://help.github.com/articles/closing-issues-via-commit-messages/) if any.

Example 1:
```
feat(data): add new get event query API
```

Example 2:
```
fix(meta): correct default database port configuration 

Previously configuration used to the wrong default database port. This commit fixes the default database port for Redis in the configuration.

Closes: #123
```

### Revert
If the commit reverts a previous commit, it should begin with `revert: `, followed by the header of the reverted commit. In the body it should say: `This reverts commit <hash>.`, where the hash is the SHA of the commit being reverted.

### Type
Must be one of the following:

* **feat**: A new feature
* **fix**: A bug fix
* **docs**: Documentation only changes
* **style**: Changes that do not affect the meaning of the code (white-space, formatting, etc)
* **refactor**: A code change that neither fixes a bug nor adds a feature
* **perf**: A code change that improves performance
* **test**: Adding missing tests or correcting existing tests
* **build**: Changes that affect the CI/CD pipeline or build system or external dependencies (example scopes: travis, jenkins, makefile)
* **ci**: Changes provided by DevOps for CI purposes.
* **revert**: Reverts a previous commit.

### Scope
Should be one of the following:
Modules:
* **core-data**: (or data) A change or addition to the core data micro service 
* **core-metadata**: (or metadata or meta) A change or addition to the core metatdata micro service
* **core-command**: (or command or cmd) A change or addition to the core command micro service
* **snap**: A change or addition to snap packaging 
* **docker**: A change or addition to docker packaging
* **security**: A change or addition to security micro services
* **scheduler**: A change or addition to the supporting scheduler micro service
* **notifications**: A change or addition to the supporting notifications micro service
* **sma**: A change or addition to the system management agent (or executor) micro service
* **deps**: A change to any service due to dependencies
* **all**: A change that affects all micro services or scopes
* *no scope*:  If no scope is provided, it is assumed the PR does not apply to the above scopes

### Body
Just as in the **subject**, use the imperative, present tense: "change" not "changed" nor "changes".
The body should include the motivation for the change and contrast this with previous behavior.

### Footer
The footer should contain any information about **Breaking Changes** and is also the place to
reference GitHub issues that this commit **Closes**.

**Breaking Changes** should start with the word `BREAKING CHANGE:` with a space or two newlines. The rest of the commit message is then used for this.

