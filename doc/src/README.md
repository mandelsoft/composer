# An Extensible Framework to Configure an Hierarchical Element Structure

This package provides a framework usable to compose hierarchically 
organized elements in a Go program, like a file hierarchy:

```text
{{execute}{go}{run}{../../examples/dirtree}{<extract>}{tree output}}
```

Similar to [ginkgo](https://github.com/onsi/ginkgo), every element
, like filesystem, directory, file or content, is described by a function accepting some basic element settings for the element creation and an anonymous function used to configure nested elements or refine the outer one.

The code creating the example structure above looks like this
(see [examples/dirtree](examples/dirtree/main.go))

```go
{{include}{../../examples/dirtree/main.go}{structure code}}
```

This example uses a default filesystem environment, implemented
using this framework.

```go
{{include}{../../examples/dirtree/main.go}{environment}}
```

Error handling is done by the framework. If somewhere in the
hierarchical code structure an error occurs the execution
is aborted. The `Compose` call on the method gathers the error
and provides it to its caller. As we will see later there are
other possibilities for error handling, also.

If in a code block non-framework/environment functions are called,
which provide an error, it can be propagated with `env.FailIfError(err)`
and related methods.

## Using the Framework for Creating Test Environments

This framework was initially intended to be used for the creation
of an object environment for tests, but it can be used in regular
programs, also.

When creating an environment a `FailureHandler` handler can be configured(as option for the `New` methods). THis is used to make it applicable
for test environments.

The package `testenv` supports the creation of environments prepared to be used with the `ginkgo` framework. Every 
composed environment type following the rules described in the next
section can be used, here.

Therefore, an own `testenv.New` function is provided

```go
  {{include}{../../testenv/testenv_test.go}{creation}}
```

It accepts the `New` function of an environment type, which should be
configured to be usable as test environment (here, the default filesystem
environment). Additionally, environment
options can be passed, for example the options described below, used
to provide content for the test filesystem.

The created environment leaves the failure handling to the ginkgo environment, using the `Expect` logic. A `Compose` method is not required.
The `env` methods can directly be used. Any occurring error
is handled with `ExpectWithOffset(skip, err).To(Succeed())`

If an environment incorporating the `filesystem` functionality
is used, the environment is automatically configured with a 
temporary filesystem living until `Cleanup` is called on the
environment. (Use defer or a `AfterEach` section)

This filesystem can then be enriched by further environment options:
- With the option `TestData()` a package-local folder `testdata` will be mounted (under the same name) on this temporary filesystem. Optionally, another path can be given as argument.
 This part of the filesystem is immutable, so tests cannot change your testdata provided as part of the sources.
- With `MutableTestData()`, the content will be mounted in a mutable way, by using a layered filesystem based on the provided 
  testdata folder and a mutable memory layer on-top.
- With `ProjectTestData(source)` any folder of your project can be used as test data.
- And with `ModifiableProjectTestData(source)` you again get a mutable layer on-top of the read-only project folder.

The `filesystem` functionality is based on a virtual filesystem
provided by the [`vfs` package](https://github.com/mandelsoft/vfs).
Therefore, the code using filesystem operations must use the 
appropriate functions and types from the vfs package (like `vfs.File` instead of `os.File`)

An example can be found in [example/testenv](examples/testenv)

```go
{{include}{../../examples/testenv/example_test.go}}
```

In a `BeforeEach` function we create some filesystem enriching base
content provided by the package folder `testdata` by adding a
file in a new folder and another one in the testdata folder. The `testdata` folder in the project sources provides the following
file structure:

```go
{{execute}{go}{run}{../../examples/testenv/main.go}{<extract>}{tree output}}
```

In `AfterEach` the filesystem (based on a temp filesystem in the operating system), is deleted again.

The test then checks the resulting directory.

## Implementing Environments with new Functional Elements.

The framework is designed to be extensible. Environment functions 
are organized in functional groups (like the one provided by the `filesystem` package)

### Functional Groups

Every such package provides an own `Group` type for the locally supported elements and environment functions.

A method `NewGroup(epi.EnvState)` must be offered to create an instance
for such an embeddable group. Both together, the Group type and the
constructor, will later be used to compose new environment types including
arbitrary functional groups.

The group must refer to the internal environment state 
kept in `epi.EnvState`. It is later used by all the new environment functions.

```go
{{include}{../../filesystem/group.go}{group}}
```

To be able to access the own group from any possible environment composition
every group should provide an own interface with a unique private method
returning the `Group`, which is implemented by the `Group`object.

```go
{{include}{../../filesystem/mapper.go}{mapper}}
```

A public mapping function can then be offered using this interface 
to map any environment incorporating this group to the group object.


### Environments

A final `Environment` can then be composed by a separate
environment type by embedding the desired group types, and optionally one base environment type, similar to the
preconfigured filesystem environment and group.

```go
  {{include}{../../filesystem/env.go}{environment}}
```

This example incorporates the `common.Environment` and the additional `FilesystemGroup`. To support this embedding the functional
areas should provide besides the standard type `Group`, uniquely named types, also.

Every environment type must at least incorporate the `epi.Group`, which 
provides some core functionality taken from `EnvState`, like the `Compose`
method.
This is automatically the case, if another base environment is included.

When creating a preconfigured environment, the same `epi.EnvState` must be propagated to every nested Group and the optional base environment.

```go
  {{include}{../../filesystem/env.go}{constructor}}
```

A method `New(...epi.Option)` should be offered as shown above, to create a new instance for such a preconfigured environment. The options must not be passed to the constructor of
a base environment, except the environment state. Instead, they should be applied
to the finally created environment to assure, that all groups potentially
required by the options are already available.

Any group might provide such a default environment including its own group. But new combinations can arbitrarily be created elsewhere by using the standard types and constructors provided by the functional groups.

#### Options

There are two kinds of options:
- built-in options intended for the `EnvState` setup, like the failure handler option, or an `EnvState` itself. All other options are ignored by the
 `NewEnvState` method.

- Arbitrary Options setting up a group or initial state.
  Not all options are therefore applicable to all kinds of environments.
  Those options use the group mapping described above to determine the appropriate group and provide a formal error, if the required group
  is not supported by the selected environment (`epi.ErrGroupNotSupported`).

  For example, `filesystem.Filesystem` option is only supported by the
  filesystem group.

```go
  {{include}{../../filesystem/options.go}{filesystem option}}
```

### Implementing new Group Functions

the package `epi` contains all the central types and functions of the core
framework, it stands for **e**nvironment **p**rogramming **i**nterface.

In the previous section we have already seen some basic types and functions,
like `epi.EnvState`. The state implements the core state and failure handling to support hierarchically nested elements. It is not intended to be used by an environment user, but is exclusively for implementors of new functional groups.
Therefore, all those related types and functions are in a separate package and not
exposed by the user-facing `composer` main package.

All environments always use a common mechanics. Every nesting level
is provided by a so-called `epi.Frame`, representing the state required
for this level and potentially required by nested functions and elements.

To be more polymorphic, and enable elements (or environment methods) to work
together with different outer levels, environments should offer and require
abstract state interfaces, providing access to information or methods required by 
nested elements.

For example, the `FileSystem` element supports a `FilesystemState` interface providing access to the surrounding `vfs.FileSystem`. This way nested elements requiring a filesystem can be embedded in all kinds of elements
providing such a state.

To achieve this, a frame may expose state, either by implementing directly such an interface, or by implementing the `epi.StateProvider` interface.

On the other side, every element/environment function, may restrict itself to require some outer state. The `epi` package offers functionality to request such state.

```go
  {{include}{../../filesystem/dir.go}{directory}}
```

The group method `Directory` is used to describe a directory element
and offeres the possibility to add nested elements by accepting a `epi-Block`
function.

It uses the function `epi.EvaluateWithState[DirectoryState]` 
To handle the element and nested block-.
The type parameter (`DirectoryState`) requests an outer state of this interface.
The new frame is created with the attributes passes togetjer with the `Directory` method.

As result of the evaluation the frame stack is searched for an appropriate 
state. If found the `Setup` method of the frame is called. to finally
setup the element and state for the frame. Here, it assures, that an appropriate directory exists.

```go
  {{include}{../../filesystem/dir.go}{setup}}
```

The state interface provides access to the outer filesystem and the 
current out directory.

```go
  {{include}{../../filesystem/dir.go}{state}}
```

The frame offers this state by implementing the state interface.

```go
  {{include}{../../filesystem/dir.go}{frame}}
```

The same interface is implemented by the `Filesystem` frame. Therefore, 
a directory can also be embedded into a filesystem element.

Optionally, a further embedding constraint can be given:

```go
  {{include}{../../filesystem/dir.go}{directory}{(cs :=.*)}}
```

It uses the `constraints` package to require the element
being nested either directly on-top of the found state frame or as direct child
of the state providing element.

If a more complex composed state is required, a state extractor
can be provided. It gets access to the frame stack and can compose
a complex state type based on separate independent state requests.

---

*This markdown file is proudly generated by [mdref](https://github.com/mandelsoft/mdref)*