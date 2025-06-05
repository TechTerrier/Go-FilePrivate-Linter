# Go FilePrivate Linter


This linter is designed to check if package variables are used outside of the source files they are defined in if they declared as fileprivate.

To declare variables as fileprivate, add a comment with `fileprivate` next it. Both of the following are valid examples:

~~~
// fileprivate  
var a = "Hello World!"  
~~~
  
`var b = "Hello World!" // fileprivate`

### Running the linter
To run the linter, clone this repository and build the binary with `go build` and run the `FilePrivateLinter` binary with your Go project's main directory as the first argument.

Example: `./FilePrivateLinter projects/main-project`

