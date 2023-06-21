# Hasher package

## Usage

1. import package

```go
import "git.epam.com/vadym_ulitin/lets-go-chat/pkg/hasher"
```

2. Basic functionality provided by functions:

- `HashPassword(string) (string, error)`    hashes the given password data using the default hasher.

- `CheckPasswordHash(password, hash string) bool` CheckPasswordHash attempts to verifiy the password using the default hasher

default hasher defined to use hash256 algoritm.

Example:

```go
    if h, err := hasher.HashHashPassword("secret"); err == nil {
        fmt.Println(h)
    }
```

Output:

`2bb80d537b1da3e38bd30361aa855686bde0eacd7162fef6a25fe97bf527a25b`

3. There is an option to use another hash type. For this use Hasher interface that supports defining hash type used.

Example:

```go
    hi, err := hasher.New(HashSHA512)
    if err == nil {
        fmt.Println(hi.HashHashPassword("secret"))
    }
```

Output:

`bd2b1aaf7ef4f09be9f52ce2d8d599674d81aa9d6a4421696dc4d93dd0619d682ce56b4d64a9ef097761ced99e0f67265b5f76085e5b0ee7ca4696b2ad6fe2b2`

## API

### Constants

	HashSHA256 
    
	HashSHA512


### Types

    type Hasher interface {
        HashPassword(password string) (string, error)
        CheckPasswordHash(password, hash string) bool
    }

### Errors

    var (
        ErrEmptyPassword      
        ErrUnsupportedHashType
    )

### Functions

    func New(t uint8) (Hasher, error) 

    func HashPassword(password string) (string, error) 

    func CheckPasswordHash(password, hash string) bool 
