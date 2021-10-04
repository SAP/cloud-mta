# Patching Vendor dependencies 

A hacky workaround to avoid forking dependencies
and/or committing the vendor directory to SCM.

Note that patches must be applied immediately after `go mod vendor` has been run.
That is because `go mod vendor` will revert the contents of the `vendor` directory to the **original** contents.

## How to create patch files for non git tracked files

`git diff --no-index orgFile patchedFile > patches/patchName.patch`

It is recommended `patchName` will be given a meaningful name e.g:
- `moduleName-BugDescription.patch`

for example, let assume you want to modify `vendor\github.com\joho\godotenv\godotenv.go`

1. Copy the file in the same directory with a `.patch` added to its name: `godotenv.patched.go`
2. Modify `godotenv.patched.go` with the needed changes
   - Please remember to document and link to relevant references, e.g (open issues / PRs in upstream).
3. Execute: `git diff --no-index vendor\github.com\joho\godotenv\godotenv.go vendor\github.com\joho\godotenv\godotenv.patched.go > patches/godotenv-token-too-long.patch`
4. Modify the `.patch` file so:
   - all paths point to the original file name (no paths with ".patched" should remain).
      - otherwise, applying the patch will all cause a re-name.
   - Use forward slashes instead of escaped backslashes (`\\`).
     - otherwise, the circleCI build would fail. 
5. Optional but highly recommended: create a test case which would fail if the patch has not been applied.

## How to apply patches

In a shell (use git bash on Windows) execute: `./apply-patches.sh`
