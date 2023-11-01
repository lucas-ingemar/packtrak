# packtrak 
Track your packages

## State - PLAN

Use sqlite. Only save whats installed: Which package manager + the "real" app-name. For dnf it would be `btop`, but for go it would be: `golang.org/x/tools/gopls`. 
Don't care about version tags, they doesn't matter for the state.

So, the packages in the state is one table that looks something like this:

```
id        manager        package
-----------------------------------------
0         dnf            btop
1         go             golang.org/x/tools/gopls
2         git            git@github.com:lucas-ingemar/packtrak.git
```

## Package file - PLAN
Should only have 1 field: `package`. This field will have the same convention as `package` in the state.
All the logic should be happening in the different managers inside the app. E.g: go package renaming from `golang.org/x/tools/gopls` to `gopls` in the UI.

### Host/environment specific
Should add rules next to the global packages. For instance for host-specific packages, or `group(?): work` packages.

## Package versions
To start with only `latest` will be installed

## Manager specific setttings.
Take dnf for example. Should be able to list coprs that should be installed. That needs to be solved
Maybe there could be something like dependencies for each manager as well
That have its own table in the state.

# KNOWN ISSUES:
* State is not updated when a package already exists when installing it. Probably better to do a proper sync, not just add/remove
