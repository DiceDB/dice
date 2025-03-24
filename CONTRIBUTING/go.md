# Go Best Practices

## Language Specifics

1. Functions that return something are given noun-like names

```go
func (c *Config) GetJobName(key string) (value string, ok bool)  // not okay
func (c *Config) JobName(key string) (value string, ok bool)  // okay
```

2. Functions that do something are given verb-like names

```go
func (c *Config) WriteDetail(w io.Writer) (int64, error)  // okay
```

3. Identical functions that differ only by the types involved include the name of the type at the end of the name.

```go
func ParseInt(input string) (int, error)  // okay
func ParseInt64(input string) (int64, error)  // okay
func AppendInt(buf []byte, value int) []byte  // okay
func AppendInt64(buf []byte, value int64) []byte  // okay
```

4. If there is a clear “primary” version, the type can be omitted from the name for that version:

```go
func (c *Config) Marshal() ([]byte, error)  // okay
func (c *Config) MarshalText() (string, error)  // okay
```

## Formatting and Linting

Please complete the [development setup](https://github.com/DiceDB/dice/blob/master/CONTRIBUTING/development-setup.md) and run `make lint`. This command lints the code and formats it according to best practices.
